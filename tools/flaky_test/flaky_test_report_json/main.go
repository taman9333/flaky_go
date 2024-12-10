package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type TestResult struct {
	Test    string `json:"Test"`
	Action  string `json:"Action"`
	Package string `json:"Package"`
}

type FlakyTest struct {
	TestName    string `json:"test_name"`
	Occurrences int    `json:"occurrences"`
}

func main() {
	// Read from standard input
	scanner := bufio.NewScanner(os.Stdin)

	// Process NDJSON line by line
	flakyCounts := make(map[string]int)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse each line as a JSON object
		var testResult TestResult
		if err := json.Unmarshal([]byte(line), &testResult); err != nil {
			log.Fatalf("Failed to decode JSON: %v", err)
		}

		// Count flaky tests
		if isValidTest(testResult) {
			fullTestName := fmt.Sprintf("%v.%v", testResult.Package, testResult.Test)
			flakyCounts[fullTestName]++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	// Convert flakyCounts to JSON array
	if len(flakyCounts) == 0 {
		fmt.Println("[]") // Output empty JSON array
		return
	}

	var flakyTests []FlakyTest
	for test, count := range flakyCounts {
		if count <= 10 { // Exclude permanently failed tests
			flakyTests = append(flakyTests, FlakyTest{
				TestName:    test,
				Occurrences: count,
			})
		}
	}

	// Output flaky tests as JSON
	jsonOutput, err := json.MarshalIndent(flakyTests, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON output: %v", err)
	}

	fmt.Println(string(jsonOutput))
}

// isValidTest checks if the test result is a valid flaky test
func isValidTest(testResult TestResult) bool {
	return testResult.Action == "fail" && containsSpecificTest(testResult.Test)
}

// containsSpecificTest checks if the test name contains a '/' to identify specific test cases
func containsSpecificTest(testName string) bool {
	return len(testName) > 0 && strings.Contains(testName, "/")
}
