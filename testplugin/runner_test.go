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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiWriterFailedPkgs verifies that the packages that had failures are determined correctly
// even if the content written to the writer is not line-aligned (the writes performed by the "go
// test" command are not guaranteed to end on a line boundary).
func TestMultiWriterFailedPkgs(t *testing.T) {
	const output = `--- FAIL: TestFoo (0.00s)
    foo_test.go:13: assertion failed
FAIL
FAIL	github.com/palantir/project/pkgone	0.021s
ok  	github.com/palantir/project/pkgtwo	0.004s
FAIL
`
	for _, tc := range []struct {
		name       string
		splitAtIdx int
	}{
		{
			name:       "single write",
			splitAtIdx: len(output),
		},
		{
			name:       "write boundary inside the FAIL summary line",
			splitAtIdx: bytes.Index([]byte(output), []byte("FAIL\tgithub.com/palantir/project/pkgone")) + 20,
		},
		{
			name:       "write boundary at the start of the FAIL summary line",
			splitAtIdx: bytes.Index([]byte(output), []byte("FAIL\tgithub.com/palantir/project/pkgone")),
		},
		{
			name:       "write boundary after every byte",
			splitAtIdx: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var console bytes.Buffer
			w := &multiWriter{
				consoleWriter:     &console,
				rawOutputWriter:   io.Discard,
				failedPkgs:        []string{},
				longestPkgNameLen: len("github.com/palantir/project/pkgone"),
			}

			for remaining := []byte(output); len(remaining) > 0; {
				currWriteLen := min(tc.splitAtIdx, len(remaining))
				n, err := w.Write(remaining[:currWriteLen])
				require.NoError(t, err)
				require.Equal(t, currWriteLen, n)
				remaining = remaining[currWriteLen:]
			}
			require.NoError(t, w.Flush())

			assert.Equal(t, []string{"github.com/palantir/project/pkgone"}, w.failedPkgs)
			// all of the content written to the writer must be written to the console exactly once
			assert.Equal(t, output, console.String())
		})
	}
}
