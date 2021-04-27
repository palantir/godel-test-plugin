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

package config_test

import (
	"fmt"
	"testing"

	"github.com/palantir/godel-test-plugin/testplugin/config"
	"github.com/palantir/pkg/matcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestLoadConfig(t *testing.T) {
	for i, tc := range []struct {
		yml           string
		want          config.Test
		wantParamKeys map[string]struct{}
	}{
		{
			yml: `
tags:
  integration:
    names:
      - "integration_tests"
    paths:
      - "test"
    exclude:
      names:
        - "ignore"
      paths:
        - "test/foo"
exclude:
  names:
    - ".*test"
    - "m?cks"
    - "gunit"
  paths:
    - "vendor"
    - "generated_src"
`,
			want: config.Test{
				Tags: map[string]matcher.NamesPathsWithExcludeCfg{
					"integration": {
						NamesPathsCfg: matcher.NamesPathsCfg{
							Names: []string{`integration_tests`},
							Paths: []string{`test`},
						},
						Exclude: matcher.NamesPathsCfg{
							Names: []string{`ignore`},
							Paths: []string{`test/foo`},
						},
					},
				},
				Exclude: matcher.NamesPathsCfg{
					Names: []string{`.*test`, `m?cks`, `gunit`},
					Paths: []string{`vendor`, `generated_src`},
				},
			},
			wantParamKeys: map[string]struct{}{
				"integration": {},
			},
		},
		{
			yml: `
tags:
  integration:
    names:
      - "integration_tests"
  mixedCasing:
    paths:
      - "test"
`,
			want: config.Test{
				Tags: map[string]matcher.NamesPathsWithExcludeCfg{
					"integration": {
						NamesPathsCfg: matcher.NamesPathsCfg{
							Names: []string{`integration_tests`},
						},
					},
					"mixedCasing": {
						NamesPathsCfg: matcher.NamesPathsCfg{
							Paths: []string{`test`},
						},
					},
				},
			},
			wantParamKeys: map[string]struct{}{
				"integration": {},
				"mixedcasing": {},
			},
		},
	} {
		var got config.Test
		err := yaml.Unmarshal([]byte(tc.yml), &got)
		require.NoError(t, err, "Case %d", i)
		p := got.ToParam()
		err = p.Validate()
		require.NoError(t, err, "Case %d", i)
		assert.Equal(t, tc.want, got, "Case %d", i)

		gotParamKeys := make(map[string]struct{})
		for k := range p.Tags {
			gotParamKeys[k] = struct{}{}
		}
		assert.Equal(t, tc.wantParamKeys, gotParamKeys, "Case %d", i)
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	for i, tc := range []struct {
		name      string
		yml       string
		wantError string
	}{
		{
			name: "tags cannot contain illegal characters",
			yml: `
tags:
  integration:
    names:
      - "integration_tests"
  foo-bar:
    paths:
      - "foo-bar"
  foo_bar:
    names:
      - "foo_bar"
  "invalid,entry":
    names:
      - "invalid"
  "another bad":
    names:
      - "another bad"
`,
			wantError: "invalid tag names: [another bad invalid,entry]",
		},
		{
			name: "tags must be unique in a case-insensitive manner",
			yml: `
tags:
  integration:
    names:
      - "integration_tests"
  INTEGRATION:
    paths:
      - "foo-bar"
`,
			wantError: "tag names were defined multiple times (names must be unique in case-insensitive manner): [integration]",
		},
		{
			name: `"all" is a reserved tag name`,
			yml: `
tags:
  all:
    names:
      - "integration_tests"
`,
			wantError: `"all" is a reserved name that cannot be used as a tag name`,
		},
	} {
		var got config.Test
		err := yaml.Unmarshal([]byte(tc.yml), &got)
		require.NoError(t, err, fmt.Sprintf("Case %d: %s", i, tc.name))
		p := got.ToParam()
		err = p.Validate()
		require.Error(t, err, fmt.Sprintf("Case %d: %s", i, tc.name))
		assert.Equal(t, tc.wantError, err.Error(), "Case %d: %s", i, tc.name)
	}
}
