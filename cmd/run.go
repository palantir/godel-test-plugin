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

package cmd

import (
	"os"

	"github.com/palantir/godel-test-plugin/testplugin"
	"github.com/palantir/godel-test-plugin/testplugin/config"
	godelconfig "github.com/palantir/godel/v2/framework/godel/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests using the provided arguments",
	RunE: func(cmd *cobra.Command, args []string) error {
		param, err := testParamFromFlags(testConfigFileFlagVal, godelConfigFileFlagVal)
		if err != nil {
			return err
		}
		partition, err := testplugin.ParsePartition(partitionFlagVal)
		if err != nil {
			return err
		}
		return testplugin.RunTestCmd(projectDirFlagVal, args, tagsFlagVal, junitOutputFlagVal, partition, param, cmd.OutOrStdout())
	},
}

func init() {
	runCmd.Flags().StringVar(&junitOutputFlagVal, "junit-output", "", "file to which JUnit output is written")
	runCmd.Flags().StringSliceVar(&tagsFlagVal, "tags", nil, "run tests that are part of the provided tags")
	runCmd.Flags().StringVar(&partitionFlagVal, "partition", "", "partition packages for parallel testing (format: X,N where X is 0-indexed partition and N is total partitions)")
	RootCmd.AddCommand(runCmd)
}

func testParamFromFlags(testConfigFile, godelConfigFile string) (testplugin.TestParam, error) {
	var testCfg config.Test
	if testConfigFile != "" {
		cfg, err := readTestConfigFromFile(testConfigFile)
		if err != nil {
			return testplugin.TestParam{}, err
		}
		testCfg = cfg
	}
	if godelConfigFile != "" {
		excludes, err := godelconfig.ReadGodelConfigExcludesFromFile(godelConfigFile)
		if err != nil {
			return testplugin.TestParam{}, err
		}
		testCfg.Exclude.Add(excludes)
	}
	return testCfg.ToParam(), nil
}

func readTestConfigFromFile(cfg string) (config.Test, error) {
	bytes, err := os.ReadFile(cfg)
	if os.IsNotExist(err) {
		return config.Test{}, nil
	}
	if err != nil {
		return config.Test{}, errors.Wrapf(err, "failed to read config file")
	}
	var testCfg config.Test
	if err := yaml.Unmarshal(bytes, &testCfg); err != nil {
		return config.Test{}, errors.Wrapf(err, "failed to unmarshal YAML")
	}
	return testCfg, nil
}
