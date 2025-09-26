// Copyright 2025 Palantir Technologies, Inc.
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
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/jstemmer/go-junit-report/v2/junit"
	"github.com/jstemmer/go-junit-report/v2/parser/gotest"
	"github.com/pkg/errors"
)

// provided an output file, returns a pipe writer to which 'go test' output should be written
// and a function that should be deferred until after the 'go test' command has completed which
// flushes and closes the unit output file.
func startJUnitReporter(junitOutput string) (writer io.Writer, deferFunc func(), err error) {
	junitOutputFile, err := os.Create(junitOutput)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create JUnit output file")
	}

	junitInputPipeReader, junitInputPipeWriter := io.Pipe()
	done := make(chan error)
	go func() {
		defer close(done)

		report, err := gotest.NewParser().Parse(junitInputPipeReader)
		if err != nil {
			done <- fmt.Errorf("error parsing input: %w", err)
			return
		}
		if _, err := io.WriteString(junitOutputFile, xml.Header); err != nil {
			done <- fmt.Errorf("error writing xml header: %w", err)
			return
		}
		testsuites := junit.CreateFromReport(report, "")
		if err := testsuites.WriteXML(junitOutputFile); err != nil {
			done <- fmt.Errorf("error writing xml testsuites: %w", err)
			return
		}
	}()

	finish := func() {
		// Close the reporter input to signal it should finish the report
		if err := junitInputPipeWriter.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to close JUnit reporter input reader: %v\n", err)
		}
		// Blocks until the reporter has finished writing its output
		if err := <-done; err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "JUnit reporter failed: %v\n", err)
		}
		if err := junitOutputFile.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to close JUnit reporter output file: %v\n", err)
		}
	}

	return junitInputPipeWriter, finish, nil
}
