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

package config

import (
	"github.com/palantir/pkg/matcher"

	"github.com/palantir/godel-test-plugin/testplugin"
	v0 "github.com/palantir/godel-test-plugin/testplugin/config/internal/v0"
)

type Test v0.Config

func (cfg *Test) ToParam() testplugin.TestParam {
	m := make(map[string]matcher.Matcher, len(cfg.Tags))
	for k, v := range cfg.Tags {
		m[k] = v.Matcher()
	}
	return testplugin.TestParam{
		Tags:    m,
		Exclude: cfg.Exclude.Matcher(),
	}
}
