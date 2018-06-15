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

package integration_test

import (
	"testing"

	"github.com/palantir/godel/framework/pluginapitester"
	"github.com/palantir/godel/pkg/products/v2/products"
	"github.com/stretchr/testify/require"
)

func TestUpgradeConfig(t *testing.T) {
	pluginPath, err := products.Bin("test-plugin")
	require.NoError(t, err)
	pluginProvider := pluginapitester.NewPluginProvider(pluginPath)

	pluginapitester.RunUpgradeConfigTest(t,
		pluginProvider,
		nil,
		[]pluginapitester.UpgradeConfigTestCase{
			{
				Name: "legacy test config is upgraded",
				ConfigFiles: map[string]string{
					"godel/config/test.yml": `
tags:
  integration:
    names:
      - "^integration$"
    paths:
      # These tests are tightly coupled to unexported types in the package, so
      # it's more effort than it's worth to separate the tests out into an
      # integration package.
      - "notification"
exclude:
  names:
    - testdata
`,
				},
				Legacy:     true,
				WantOutput: "Upgraded configuration for test-plugin.yml\n",
				WantFiles: map[string]string{
					"godel/config/test-plugin.yml": `tags:
  integration:
    names:
    - ^integration$
    paths:
    - notification
exclude:
  names:
  - testdata
`,
				},
			},
			{
				Name: "current config is unmodified",
				ConfigFiles: map[string]string{
					"godel/config/test-plugin.yml": `
tags:
  integration:
    names:
      - "^integration$"
    paths:
      # These tests are tightly coupled to unexported types in the package, so
      # it's more effort than it's worth to separate the tests out into an
      # integration package.
      - "notification"
exclude:
  names:
    - testdata
`,
				},
				WantOutput: "",
				WantFiles: map[string]string{
					"godel/config/test-plugin.yml": `
tags:
  integration:
    names:
      - "^integration$"
    paths:
      # These tests are tightly coupled to unexported types in the package, so
      # it's more effort than it's worth to separate the tests out into an
      # integration package.
      - "notification"
exclude:
  names:
    - testdata
`,
				},
			},
		},
	)
}
