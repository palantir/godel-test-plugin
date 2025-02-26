// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testplugin

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/palantir/pkg/matcher"
	"github.com/palantir/pkg/pkgpath"
	"github.com/pkg/errors"
)

const GoJUnitReport = "gojunitreport"

func RunTestCmd(projectDir string, testArgs, tags []string, junitOutput string, param TestParam, stdout io.Writer) (rErr error) {
	if err := param.Validate(); err != nil {
		return err
	}
	pkgs, err := PkgsForTags(projectDir, tags, param)
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return errors.Errorf("no packages to test")
	}

	args := []string{
		"test",
	}
	args = append(args, testArgs...)

	if junitOutput != "" {
		args = append(args, "-v")
	}
	args = append(args, pkgs...)
	cmd := exec.Command("go", args...)
	cmd.Dir = projectDir

	maxPkgLen, err := longestPkgNameLen(pkgs, projectDir)
	if err != nil {
		return err
	}

	var junitOutputCmd *exec.Cmd
	rawOutputWriter := ioutil.Discard
	done := make(chan error)

	if junitOutput != "" {
		pathToSelf, err := os.Executable()
		if err != nil {
			return errors.Wrapf(err, "failed to determine path for current executable")
		}
		junitOutputCmd = exec.Command(pathToSelf, "__"+GoJUnitReport)

		junitOutputFile, err := os.Create(junitOutput)
		if err != nil {
			return errors.Wrapf(err, "failed to create JUnit output file")
		}
		defer func() {
			if err := junitOutputFile.Close(); err != nil && rErr == nil {
				rErr = errors.Wrapf(err, "failed to close output file")
			}
		}()
		junitOutputCmd.Stdout = junitOutputFile
		junitOutputCmd.Stderr = junitOutputFile

		wc, err := junitOutputCmd.StdinPipe()
		if err != nil {
			return errors.Wrapf(err, "failed to create stdin pipe")
		}
		rawOutputWriter = wc

		go func() {
			done <- junitOutputCmd.Run()
		}()
	} else {
		close(done)
	}

	failedPkgs, err := executeTestCmd(cmd, stdout, rawOutputWriter, maxPkgLen)
	if closer, ok := rawOutputWriter.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return errors.Wrapf(err, "failed to close raw output writer")
		}
	}

	if err != nil && err.Error() != "exit status 1" {
		// only re-throw if error is not "exit status 1", since those errors are generally recoverable
		return err
	}

	defer func() {
		// Blocks until the reporter has finished writing its output
		if err := <-done; err != nil && rErr == nil {
			rErr = errors.Wrapf(err, "JUnit reporter failed: %v", junitOutputCmd.Args)
		}
	}()

	if len(failedPkgs) > 0 {
		numFailedPkgs := len(failedPkgs)
		outputParts := append([]string{fmt.Sprintf("%d package(s) had failing tests:", numFailedPkgs)}, failedPkgs...)
		return errors.Errorf("%s", strings.Join(outputParts, "\n\t"))
	}

	return nil
}

func PkgsForTags(projectDir string, tags []string, param TestParam) ([]string, error) {
	// tagsMatcher is a matcher that matches the specified tags (or nil if no tags were specified)
	tagsMatcher, err := matcherForTags(tags, param)
	if err != nil {
		return nil, err
	}
	excludeMatchers := []matcher.Matcher{param.Exclude}
	if tagsMatcher != nil {
		// if tagsMatcher is non-nil, should exclude all files that do not match the tags
		excludeMatchers = append(excludeMatchers, matcher.Not(tagsMatcher))
	}
	return pkgPaths(projectDir, matcher.Any(excludeMatchers...))
}

// matcherForTags returns a Matcher that matches all packages that are matched by the provided tags. If no tags are
// provided, returns nil. If the tags consist of a single tag named "all", the returned matcher matches the union of all
// known tags. If the tags consist of a single tag named "none", the returned matcher matches everything except the
// union of all known tags (untagged tests).
func matcherForTags(tags []string, cfg TestParam) (matcher.Matcher, error) {
	if len(tags) == 0 {
		// if no tags were provided, does not match anything
		return nil, nil
	}

	if len(tags) == 1 {
		var allMatchers []matcher.Matcher
		for _, matcher := range cfg.Tags {
			allMatchers = append(allMatchers, matcher)
		}
		anyTagMatcher := matcher.Any(allMatchers...)
		switch tags[0] {
		case AllTagName:
			// if tags contains only a single tag that is the "all" tag, return matcher that matches union of all tags
			return anyTagMatcher, nil
		case NoneTagName:
			// if tags contains only a single tag that is the "none" tag, return matcher that matches not of union of all tags
			return matcher.Not(anyTagMatcher), nil
		}
	}

	// due to previous check, if "all" or "none" tag exists at this point it means that it was one of multiple tags
	for _, tag := range tags {
		switch tag {
		case AllTagName, NoneTagName:
			return nil, errors.Errorf("if %q tag is specified, it must be the only tag specified", tag)
		}
	}

	var tagMatchers []matcher.Matcher
	var missingTags []string
	for _, tag := range tags {
		if include, ok := cfg.Tags[tag]; ok {
			tagMatchers = append(tagMatchers, include)
		} else {
			missingTags = append(missingTags, fmt.Sprintf("%q", tag))
		}
	}

	if len(missingTags) > 0 {
		var allTags []string
		for tag := range cfg.Tags {
			allTags = append(allTags, fmt.Sprintf("%q", tag))
		}
		sort.Strings(allTags)
		validTagsOutput := fmt.Sprintf("valid tags: %v", strings.Join(allTags, ", "))
		if len(allTags) == 0 {
			validTagsOutput = "no tags are defined"
		}
		return nil, fmt.Errorf("tag(s) %v not defined in configuration: %s", strings.Join(missingTags, ", "), validTagsOutput)
	}

	// not possible: if initial tags were empty then should have already returned, if specified tags did not match then
	// missing block should have executed and returned, so at this point matchers must exist
	if len(tagMatchers) == 0 {
		panic("no matching tags found")
	}

	// OR of tags
	return matcher.Any(tagMatchers...), nil
}

// pkgPaths returns a slice that contains the relative package paths for all of the packages in the provided project
// directory relative to the project directory excluding any of the paths that match the provided "exclude" Matcher.
func pkgPaths(projectDir string, exclude matcher.Matcher) ([]string, error) {
	pkgs, err := pkgpath.PackagesInDirMatchingRootModule(projectDir, exclude)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list packages in %s", projectDir)
	}
	resultPkgPaths, err := pkgs.Paths(pkgpath.Relative)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get relative paths for packages %v", pkgs)
	}
	return resultPkgPaths, nil
}
