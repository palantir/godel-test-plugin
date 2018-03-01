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

package main

import (
	"io/ioutil"
	"regexp"
	"sort"
	"strings"

	"github.com/palantir/pkg/matcher"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func readTestConfigFromFile(cfg string) (TestConfig, error) {
	bytes, err := ioutil.ReadFile(cfg)
	if err != nil {
		return TestConfig{}, errors.Wrapf(err, "failed to read config file")
	}
	return readTestConfig(bytes)
}

func readTestConfig(cfg []byte) (TestConfig, error) {
	var testCfg TestConfig
	if err := yaml.Unmarshal(cfg, &testCfg); err != nil {
		return TestConfig{}, errors.Wrapf(err, "failed to unmarshal YAML")
	}
	return testCfg, nil
}

type TestConfig struct {
	// Tags group tests into different sets. The key is the name of the tag and the value is a
	// matcher.NamesPathsWithExcludeCfg that specifies the rules for matching the tests that are part of the tag.
	// Any test that matches the provided matcher is considered part of the tag.
	Tags map[string]matcher.NamesPathsWithExcludeCfg `yaml:"tags" json:"tags"`

	// Exclude specifies the files that should be excluded from tests.
	Exclude matcher.NamesPathsCfg `yaml:"exclude" json:"exclude"`
}

func (r *TestConfig) ToParam() TestParam {
	m := make(map[string]matcher.Matcher, len(r.Tags))
	for k, v := range r.Tags {
		m[k] = v.Matcher()
	}
	return TestParam{
		Tags:    m,
		Exclude: r.Exclude.Matcher(),
	}
}

const (
	AllTagName  = "all"
	NoneTagName = "none"
)

type TestParam struct {
	// Tags group tests into different sets. The key is the name of the tag and the value is a matcher.NamesPathsCfg
	// that specifies the rules for matching the tests that are part of the tag. Any test that matches the provided
	// matcher is considered part of the tag.
	Tags map[string]matcher.Matcher

	// Exclude specifies the files that should be excluded from tests.
	Exclude matcher.Matcher
}

func (p *TestParam) Validate() error {
	var invalidTagNames []string
	seenTagNames := make(map[string]struct{})
	duplicateTagNames := make(map[string]struct{})

	for k := range p.Tags {
		if !validTagName(k) {
			invalidTagNames = append(invalidTagNames, k)
		}
		normalized := strings.ToLower(k)
		if _, ok := seenTagNames[normalized]; ok {
			duplicateTagNames[normalized] = struct{}{}
		}
		seenTagNames[normalized] = struct{}{}

		switch normalized {
		case AllTagName, NoneTagName:
			return errors.Errorf("%q is a reserved name that cannot be used as a tag name", normalized)
		}
	}

	if len(invalidTagNames) > 0 {
		sort.Strings(invalidTagNames)
		return errors.Errorf("invalid tag names: %v", invalidTagNames)
	}

	if len(duplicateTagNames) > 0 {
		var sorted []string
		for k := range duplicateTagNames {
			sorted = append(sorted, k)
		}
		sort.Strings(sorted)
		return errors.Errorf("tag names were defined multiple times (names must be unique in case-insensitive manner): %v", sorted)
	}

	// normalize tags to all lowercase
	for k, v := range p.Tags {
		delete(p.Tags, k)
		p.Tags[strings.ToLower(k)] = v
	}

	return nil
}

var tagRegExp = regexp.MustCompile(`[A-Za-z0-9_-]+`)

func validTagName(tag string) bool {
	return len(tagRegExp.ReplaceAllString(tag, "")) == 0
}
