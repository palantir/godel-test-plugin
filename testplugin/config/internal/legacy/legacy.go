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

package legacy

import (
	"github.com/palantir/godel/pkg/versionedconfig"
	"github.com/palantir/pkg/matcher"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	v0 "github.com/palantir/godel-test-plugin/testplugin/config/internal/v0"
)

type Config struct {
	versionedconfig.ConfigWithLegacy `yaml:",inline"`

	// Tags group tests into different sets. The key is the name of the tag and the value is a
	// matcher.NamesPathsWithExcludeCfg that specifies the rules for matching the tests that are part of the tag.
	// Any test that matches the provided matcher is considered part of the tag.
	Tags map[string]matcher.NamesPathsWithExcludeCfg `yaml:"tags" json:"tags"`

	// Exclude specifies the files that should be excluded from tests.
	Exclude matcher.NamesPathsCfg `yaml:"exclude" json:"exclude"`
}

func UpgradeConfig(cfgBytes []byte) ([]byte, error) {
	var legacyCfg Config
	if err := yaml.UnmarshalStrict(cfgBytes, &legacyCfg); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal test-plugin legacy configuration")
	}
	cfg := v0.Config{
		Tags:    legacyCfg.Tags,
		Exclude: legacyCfg.Exclude,
	}
	upgradedBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal test-plugin v0 configuration")
	}
	return upgradedBytes, nil
}
