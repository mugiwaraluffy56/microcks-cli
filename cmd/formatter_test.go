/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/stretchr/testify/assert"
)

func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

func testSummary() *connectors.TestResultSummary {
	return &connectors.TestResultSummary{
		ID:             "test-123",
		ServiceID:      "Weather API:1.0",
		TestedEndpoint: "http://localhost:3000",
		Success:        false,
		InProgress:     false,
		TestCaseResults: []connectors.TestCaseResult{
			{OperationName: "GET /forecast", Success: true, ElapsedTime: 120},
			{OperationName: "POST /forecast", Success: false, ElapsedTime: 250},
		},
	}
}

func TestPrintGitHubActionsResult_ErrorAnnotationPerFailedOperation(t *testing.T) {
	out := captureStdout(func() {
		PrintGitHubActionsResult(testSummary(), "http://microcks.example.com", "test-123")
	})

	assert.Contains(t, out, "::error title=Test Failed::")
	assert.Contains(t, out, "POST /forecast")
	assert.NotContains(t, out, "GET /forecast")
}

func TestPrintGitHubActionsResult_NoAnnotationsWhenAllPass(t *testing.T) {
	summary := testSummary()
	summary.Success = true
	for i := range summary.TestCaseResults {
		summary.TestCaseResults[i].Success = true
	}

	out := captureStdout(func() {
		PrintGitHubActionsResult(summary, "http://microcks.example.com", "test-123")
	})

	assert.NotContains(t, out, "::error")
}

func TestPrintGitHubActionsResult_WritesStepSummaryFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "step-summary-*.md")
	assert.NoError(t, err)
	tmp.Close()
	defer os.Remove(tmp.Name())

	t.Setenv("GITHUB_STEP_SUMMARY", tmp.Name())

	captureStdout(func() {
		PrintGitHubActionsResult(testSummary(), "http://microcks.example.com", "test-123")
	})

	content, err := os.ReadFile(tmp.Name())
	assert.NoError(t, err)

	body := string(content)
	assert.True(t, strings.Contains(body, "GET /forecast"))
	assert.True(t, strings.Contains(body, "POST /forecast"))
	assert.True(t, strings.Contains(body, "pass"))
	assert.True(t, strings.Contains(body, "fail"))
	assert.True(t, strings.Contains(body, "test-123"))
}

func TestPrintTextResult_WritesResultLink(t *testing.T) {
	out := captureStdout(func() {
		PrintTextResult(testSummary(), "http://microcks.example.com", "test-123")
	})

	assert.Equal(t, "Full TestResult details are available here: http://microcks.example.com/#/tests/test-123 \n", out)
}

func TestPrintGitHubActionsResult_NilSummaryDoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		captureStdout(func() {
			PrintGitHubActionsResult(nil, "http://microcks.example.com", "test-123")
		})
	})
}
