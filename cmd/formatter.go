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
	"fmt"
	"io"
	"os"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

// PrintTextResult writes the default test result link.
func PrintTextResult(summary *connectors.TestResultSummary, serverAddr, testResultID string) {
	printTextResult(os.Stdout, serverAddr, testResultID)
}

// PrintGitHubActionsResult writes per-operation failure annotations and a step summary table.
func PrintGitHubActionsResult(summary *connectors.TestResultSummary, serverAddr, testResultID string) {
	printGitHubActionsResult(os.Stdout, summary, serverAddr, testResultID, os.Getenv("GITHUB_STEP_SUMMARY"))
}

func printTextResult(w io.Writer, serverAddr, testResultID string) {
	fmt.Fprintf(w, "Full TestResult details are available here: %s/#/tests/%s \n", serverAddr, testResultID)
}

func printGitHubActionsResult(w io.Writer, summary *connectors.TestResultSummary, serverAddr, testResultID, summaryFile string) {
	if summary == nil {
		return
	}

	for _, tc := range summary.TestCaseResults {
		if !tc.Success {
			fmt.Fprintf(w, "::error title=Test Failed::%s - operation %s failed after %dms\n",
				summary.ServiceID, tc.OperationName, tc.ElapsedTime)
		}
	}

	if summaryFile != "" {
		f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			fmt.Fprintf(f, "## Microcks test result - %s\n\n", summary.ServiceID)
			fmt.Fprintf(f, "| operation | result | elapsed (ms) |\n")
			fmt.Fprintf(f, "|---|---|---|\n")
			for _, tc := range summary.TestCaseResults {
				status := "pass"
				if !tc.Success {
					status = "fail"
				}
				fmt.Fprintf(f, "| %s | %s | %d |\n", tc.OperationName, status, tc.ElapsedTime)
			}
			fmt.Fprintf(f, "\nFull result: %s/#/tests/%s\n", serverAddr, testResultID)
		}
	}

	printTextResult(w, serverAddr, testResultID)
}
