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
	"path"

	"github.com/palantir/godel/framework/godellauncher"
	"github.com/palantir/godel/framework/pluginapi"
	"github.com/spf13/cobra"
)

var (
	debugFlagVal           bool
	projectDirFlagVal      string
	godelConfigFileFlagVal string
	testConfigFileFlagVal  string
	junitOutputFlagVal     string
	tagsFlagVal            []string
)

var rootCmd = &cobra.Command{
	Use:   "test-plugin",
	Short: "Run test on project packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		param, err := testParamFromFlags(testConfigFileFlagVal, godelConfigFileFlagVal)
		if err != nil {
			return err
		}
		return runTestCmd(projectDirFlagVal, args, tagsFlagVal, junitOutputFlagVal, param, cmd.OutOrStdout())
	},
}

func testParamFromFlags(testConfigFile, godelConfigFile string) (TestParam, error) {
	var testCfg TestConfig
	if testConfigFile != "" {
		cfg, err := readTestConfigFromFile(testConfigFile)
		if err != nil {
			return TestParam{}, err
		}
		testCfg = cfg
	}
	if godelConfigFile != "" {
		cfg, err := godellauncher.ReadGodelConfig(path.Dir(godelConfigFile))
		if err != nil {
			return TestParam{}, err
		}
		testCfg.Exclude.Add(cfg.Exclude)
	}
	return testCfg.ToParam(), nil
}

func init() {
	pluginapi.AddDebugPFlagPtr(rootCmd.PersistentFlags(), &debugFlagVal)
	pluginapi.AddGodelConfigPFlagPtr(rootCmd.PersistentFlags(), &godelConfigFileFlagVal)
	pluginapi.AddConfigPFlagPtr(rootCmd.PersistentFlags(), &testConfigFileFlagVal)
	pluginapi.AddProjectDirPFlagPtr(rootCmd.PersistentFlags(), &projectDirFlagVal)
	if err := rootCmd.MarkPersistentFlagRequired(pluginapi.ProjectDirFlagName); err != nil {
		panic(err)
	}
	rootCmd.Flags().StringVar(&junitOutputFlagVal, "junit-output", "", "file to which JUnit output is written")
	rootCmd.Flags().StringSliceVar(&tagsFlagVal, "tags", nil, "run tests that are part of the provided tags")
}
