// Copyright 2026 Palantir Technologies, Inc.
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

package testplugin

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPkgsToTest(t *testing.T) {
	// Create a temp directory with some Go packages for testing
	tmpDir := t.TempDir()
	goModContent := []byte("module testmod\n\ngo 1.21\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), goModContent, 0644))
	pkgContent := []byte("package foo\n")
	for _, pkg := range []string{"a", "b", "c", "d"} {
		pkgDir := filepath.Join(tmpDir, pkg)
		require.NoError(t, os.MkdirAll(pkgDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "foo.go"), pkgContent, 0644))
	}

	for _, tc := range []struct {
		name      string
		partition string
		wantPkgs  []string
		wantErr   string
	}{
		{
			name:      "no partition returns all packages",
			partition: "",
			wantPkgs:  []string{"./a", "./b", "./c", "./d"},
		},
		{
			name:      "partition 0,2 returns first half",
			partition: "0,2",
			wantPkgs:  []string{"./a", "./b"},
		},
		{
			name:      "partition 1,2 returns second half",
			partition: "1,2",
			wantPkgs:  []string{"./c", "./d"},
		},
		{
			name:      "partition 0,4 returns first package",
			partition: "0,4",
			wantPkgs:  []string{"./a"},
		},
		{
			name:      "partition 3,4 returns last package",
			partition: "3,4",
			wantPkgs:  []string{"./d"},
		},
		{
			name:      "invalid partition format returns error",
			partition: "invalid",
			wantErr:   "failed to parse partition flag",
		},
		{
			name:      "partition out of bounds returns error",
			partition: "5,4",
			wantErr:   "failed to parse partition flag",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			pkgs, err := PkgsToTest(tmpDir, nil, tc.partition, TestParam{}, &stdout)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantPkgs, pkgs)
		})
	}
}

func TestPkgsToTestEmptyPartition(t *testing.T) {
	// Create a temp directory with fewer packages than partitions
	tmpDir := t.TempDir()
	goModContent := []byte("module testmod\n\ngo 1.21\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), goModContent, 0644))
	pkgContent := []byte("package foo\n")
	for _, pkg := range []string{"a", "b"} {
		pkgDir := filepath.Join(tmpDir, pkg)
		require.NoError(t, os.MkdirAll(pkgDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "foo.go"), pkgContent, 0644))
	}

	var stdout bytes.Buffer
	// Partition 3 of 4 with only 2 packages should return empty slice (not error)
	pkgs, err := PkgsToTest(tmpDir, nil, "3,4", TestParam{}, &stdout)
	require.NoError(t, err)
	assert.Empty(t, pkgs)
}

func TestPkgsToTestNoPackages(t *testing.T) {
	// Create a temp directory with no Go packages
	tmpDir := t.TempDir()
	goModContent := []byte("module testmod\n\ngo 1.21\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), goModContent, 0644))

	var stdout bytes.Buffer
	// No partition, no packages should return empty slice
	pkgs, err := PkgsToTest(tmpDir, nil, "", TestParam{}, &stdout)
	require.NoError(t, err)
	assert.Empty(t, pkgs)
}
