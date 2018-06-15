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
	"bufio"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func longestPkgNameLen(pkgPaths []string, projectDir string) (int, error) {
	goListCmd := exec.Command("go", append([]string{"list"}, pkgPaths...)...)
	goListCmd.Dir = projectDir

	listedPkgsBytes, err := goListCmd.CombinedOutput()
	if err != nil {
		return 0, errors.Wrapf(err, "%v failed: %s", goListCmd.Args, string(listedPkgsBytes))
	}

	longestPkgLen := 0
	listedPkgs := strings.Split(string(listedPkgsBytes), "\n")
	for _, pkgName := range listedPkgs {
		if len(pkgName) > longestPkgLen {
			longestPkgLen = len(pkgName)
		}
	}
	return longestPkgLen, nil
}

// executeTestCommand executes the provided command. The output produced by the command's Stdout and Stderr calls are
// processed as they are written and an aligned version of the output is written to the Stdout of the current process.
// The "longestPkgNameLen" parameter specifies the longest package name (used to align the console output). This
// function returns a slice that contains the packages that had test failures (output line started with "FAIL"). The
// error value will contain any error that was encountered while executing the command, including if the command
// executed successfully but any tests failed. In either case, the packages that encountered errors will also be
// returned.
func executeTestCmd(execCmd *exec.Cmd, stdout, rawOutputWriter io.Writer, longestPkgNameLen int) (rFailedPkgs []string, rErr error) {
	bw := bufio.NewWriter(rawOutputWriter)

	// stream output to Stdout
	multiWriter := multiWriter{
		consoleWriter:     stdout,
		rawOutputWriter:   bw,
		failedPkgs:        []string{},
		longestPkgNameLen: longestPkgNameLen,
	}

	// flush buffered writer at the end of the function
	defer func() {
		if err := bw.Flush(); err != nil && rErr == nil {
			rErr = errors.Wrapf(err, "failed to flush buffered writer in defer")
		}
	}()

	// set Stdout and Stderr of command to multiwriter
	execCmd.Stdout = &multiWriter
	execCmd.Stderr = &multiWriter

	// run command (which will print its Stdout and Stderr to the Stdout of current process) and return output
	err := execCmd.Run()
	return multiWriter.failedPkgs, err
}

type multiWriter struct {
	consoleWriter     io.Writer
	rawOutputWriter   io.Writer
	failedPkgs        []string
	longestPkgNameLen int
}

var setupFailedRegexp = regexp.MustCompile(`(^FAIL\t.+) (\[setup failed\]$)`)

func (w *multiWriter) Write(p []byte) (int, error) {
	// write unaltered output to file
	n, err := w.rawOutputWriter.Write(p)
	if err != nil {
		return n, err
	}

	lines := strings.Split(string(p), "\n")
	for i, currLine := range lines {
		// test output for valid case always starts with "Ok" or "FAIL"
		if strings.HasPrefix(currLine, "ok") || strings.HasPrefix(currLine, "FAIL") || strings.HasPrefix(currLine, "?") {
			if setupFailedRegexp.MatchString(currLine) {
				// if line matches "setup failed" output, modify output to conform to expected style
				// (namely, replace space between package name and "[setup failed]" with a tab)
				currLine = setupFailedRegexp.ReplaceAllString(currLine, "$1\t$2")
			}

			// split into at most 4 parts
			fields := strings.SplitN(currLine, "\t", 4)

			// valid test lines have at least 3 parts: "[ok|FAIL|?]\t[pkgName]\t[time|no test files]"
			if len(fields) >= 3 {
				currPkgName := strings.TrimSpace(fields[1])
				lines[i] = alignLine(fields, w.longestPkgNameLen)
				// append package name to failures list if this was a failure
				if strings.HasPrefix(currLine, "FAIL") {
					w.failedPkgs = append(w.failedPkgs, currPkgName)
				}
			}
		}
	}

	// write formatted version to console writer
	if n, err := w.consoleWriter.Write([]byte(strings.Join(lines, "\n"))); err != nil {
		return n, err
	}

	// n and err are from the unaltered write to the rawOutputWriter
	return n, err
}

// alignLine returns a string where the length of the second field (fields[1]) is padded with spaces to make its length
// equal to the value of maxPkgLen and the fields are joined with tab characters. Assuming that the first field is
// always the same length, this method ensures that the third field will always be aligned together for any fixed value
// of maxPkgLen.
func alignLine(fields []string, maxPkgLen int) string {
	currPkgName := fields[1]
	repeat := maxPkgLen - len(currPkgName)
	if repeat < 0 {
		// this should not occur under normal circumstances. However, it appears that it is possible if tests
		// create test packages in the directory structure while tests are already running. If such a case is
		// encountered, having output that isn't aligned optimally is better than crashing, so set repeat to 0.
		repeat = 0
	}
	fields[1] = currPkgName + strings.Repeat(" ", repeat)
	return strings.Join(fields, "\t")
}
