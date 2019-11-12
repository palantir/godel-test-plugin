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
	"github.com/palantir/godel/v2/framework/pluginapi"
	"github.com/spf13/cobra"
)

var (
	DebugFlagVal bool

	projectDirFlagVal      string
	godelConfigFileFlagVal string
	testConfigFileFlagVal  string
	junitOutputFlagVal     string
	tagsFlagVal            []string
)

var RootCmd = &cobra.Command{
	Use:   "test-plugin",
	Short: "Run test on project packages",
}

func init() {
	pluginapi.AddDebugPFlagPtr(RootCmd.PersistentFlags(), &DebugFlagVal)
	pluginapi.AddGodelConfigPFlagPtr(RootCmd.PersistentFlags(), &godelConfigFileFlagVal)
	pluginapi.AddConfigPFlagPtr(RootCmd.PersistentFlags(), &testConfigFileFlagVal)
	pluginapi.AddProjectDirPFlagPtr(RootCmd.PersistentFlags(), &projectDirFlagVal)
	if err := RootCmd.MarkPersistentFlagRequired(pluginapi.ProjectDirFlagName); err != nil {
		panic(err)
	}
}
