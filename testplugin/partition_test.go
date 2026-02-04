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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePartition(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		want    *Partition
		wantErr string
	}{
		{
			name:  "empty string returns nil",
			input: "",
			want:  nil,
		},
		{
			name:  "valid partition 0,4",
			input: "0,4",
			want:  &Partition{Index: 0, Total: 4},
		},
		{
			name:  "valid partition 3,4",
			input: "3,4",
			want:  &Partition{Index: 3, Total: 4},
		},
		{
			name:  "valid partition with spaces",
			input: " 1 , 3 ",
			want:  &Partition{Index: 1, Total: 3},
		},
		{
			name:    "invalid format - missing comma",
			input:   "14",
			wantErr: `invalid partition format "14": expected format X,N (e.g., 0,4)`,
		},
		{
			name:    "invalid format - too many parts",
			input:   "1,2,3",
			wantErr: `invalid partition format "1,2,3": expected format X,N (e.g., 0,4)`,
		},
		{
			name:    "invalid index - not a number",
			input:   "a,4",
			wantErr: `invalid partition index "a"`,
		},
		{
			name:    "invalid total - not a number",
			input:   "1,b",
			wantErr: `invalid partition total "b"`,
		},
		{
			name:    "invalid total - zero",
			input:   "1,0",
			wantErr: "partition total must be at least 1, got 0",
		},
		{
			name:    "invalid index - negative",
			input:   "-1,4",
			wantErr: "partition index must be between 0 and 3, got -1",
		},
		{
			name:    "invalid index - equal to total",
			input:   "4,4",
			wantErr: "partition index must be between 0 and 3, got 4",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePartition(tc.input)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPartitionApply(t *testing.T) {
	for _, tc := range []struct {
		name      string
		partition *Partition
		pkgs      []string
		want      []string
	}{
		{
			name:      "nil partition returns original",
			partition: nil,
			pkgs:      []string{"c", "a", "b"},
			want:      []string{"c", "a", "b"},
		},
		{
			name:      "empty packages returns empty",
			partition: &Partition{Index: 0, Total: 2},
			pkgs:      []string{},
			want:      []string{},
		},
		{
			name:      "4 packages, partition 0 of 4",
			partition: &Partition{Index: 0, Total: 4},
			pkgs:      []string{"d", "c", "b", "a"},
			want:      []string{"a"},
		},
		{
			name:      "4 packages, partition 3 of 4",
			partition: &Partition{Index: 3, Total: 4},
			pkgs:      []string{"d", "c", "b", "a"},
			want:      []string{"d"},
		},
		{
			name:      "5 packages, partition 0 of 2 (gets 3)",
			partition: &Partition{Index: 0, Total: 2},
			pkgs:      []string{"e", "d", "c", "b", "a"},
			want:      []string{"a", "b", "c"},
		},
		{
			name:      "5 packages, partition 1 of 2 (gets 2)",
			partition: &Partition{Index: 1, Total: 2},
			pkgs:      []string{"e", "d", "c", "b", "a"},
			want:      []string{"d", "e"},
		},
		{
			name:      "3 packages, partition 3 of 4 (empty)",
			partition: &Partition{Index: 3, Total: 4},
			pkgs:      []string{"c", "b", "a"},
			want:      nil,
		},
		{
			name:      "10 packages, partition 0 of 3",
			partition: &Partition{Index: 0, Total: 3},
			pkgs:      []string{"j", "i", "h", "g", "f", "e", "d", "c", "b", "a"},
			want:      []string{"a", "b", "c", "d"},
		},
		{
			name:      "10 packages, partition 1 of 3",
			partition: &Partition{Index: 1, Total: 3},
			pkgs:      []string{"j", "i", "h", "g", "f", "e", "d", "c", "b", "a"},
			want:      []string{"e", "f", "g"},
		},
		{
			name:      "10 packages, partition 2 of 3",
			partition: &Partition{Index: 2, Total: 3},
			pkgs:      []string{"j", "i", "h", "g", "f", "e", "d", "c", "b", "a"},
			want:      []string{"h", "i", "j"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.partition.Apply(tc.pkgs)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPartitionString(t *testing.T) {
	assert.Equal(t, "no partition", (*Partition)(nil).String())
	assert.Equal(t, "partition 0 of 4", (&Partition{Index: 0, Total: 4}).String())
	assert.Equal(t, "partition 2 of 5", (&Partition{Index: 2, Total: 5}).String())
}
