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
	"strings"

	"github.com/spf13/cobra"

	"github.com/palantir/godel-test-plugin/testplugin"
)

var tagPkgsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Print the packages that match the provided tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		param, err := testParamFromFlags(testConfigFileFlagVal, godelConfigFileFlagVal)
		if err != nil {
			return err
		}
		pkgs, err := testplugin.PkgsForTags(projectDirFlagVal, args, param)
		if err != nil {
			return err
		}
		cmd.Println(strings.Join(pkgs, "\n"))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(tagPkgsCmd)
}
