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
	"os"

	"github.com/palantir/godel/framework/pluginapi"
	"github.com/palantir/pkg/cobracli"

	"github.com/palantir/godel-test-plugin/gojunit/generated_src"
)

const goJUnitReport = "gojunitreport"

func main() {
	if ok := pluginapi.InfoCmd(os.Args, os.Stdout, pluginInfo); ok {
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "__"+goJUnitReport {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		amalgomated.Instance().Run(goJUnitReport)
		return
	}
	os.Exit(cobracli.ExecuteWithDefaultParamsWithVersion(rootCmd, &debugFlagVal, ""))
}
