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
	"github.com/palantir/godel/framework/godellauncher"
	"github.com/palantir/godel/framework/pluginapi/v2/pluginapi"
	"github.com/palantir/godel/framework/verifyorder"
	"github.com/palantir/pkg/cobracli"
)

var PluginInfo = pluginapi.MustNewPluginInfo(
	"com.palantir.godel-test-plugin",
	"test-plugin",
	cobracli.Version,
	pluginapi.PluginInfoUsesConfigFile(),
	pluginapi.PluginInfoGlobalFlagOptions(
		pluginapi.GlobalFlagOptionsParamDebugFlag("--"+pluginapi.DebugFlagName),
		pluginapi.GlobalFlagOptionsParamProjectDirFlag("--"+pluginapi.ProjectDirFlagName),
		pluginapi.GlobalFlagOptionsParamGodelConfigFlag("--"+pluginapi.GodelConfigFlagName),
		pluginapi.GlobalFlagOptionsParamConfigFlag("--"+pluginapi.ConfigFlagName),
	),
	pluginapi.PluginInfoTaskInfo(
		"test",
		"Test packages",
		pluginapi.TaskInfoCommand("run"),
		pluginapi.TaskInfoVerifyOptions(
			pluginapi.VerifyOptionsTaskFlags(
				pluginapi.NewVerifyFlag(
					"junit-output",
					"path to JUnit XML output (only used if 'test' task is run)",
					godellauncher.StringFlag,
				),
				pluginapi.NewVerifyFlag(
					"tags",
					"specify tags that should be used for tests (only used if 'test' task is run)",
					godellauncher.StringFlag,
				),
			),
			pluginapi.VerifyOptionsOrdering(intPtr(verifyorder.Test)),
		),
	),
	pluginapi.PluginInfoTaskInfo(
		"test-tags",
		"Print the test packages that match the provided test tags",
		pluginapi.TaskInfoCommand("tags"),
	),
	pluginapi.PluginInfoUpgradeConfigTaskInfo(
		pluginapi.UpgradeConfigTaskInfoCommand("upgrade-config"),
		pluginapi.LegacyConfigFile("test.yml"),
	),
)

func intPtr(val int) *int {
	return &val
}
