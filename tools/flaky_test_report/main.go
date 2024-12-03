package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type FlakyReport struct {
	Tests []FlakyTest `json:"tests"`
}

type FlakyTest struct {
	Name        string `json:"name"`
	Occurrences int    `json:"occurrences"`
}

func main() {
	var output bytes.Buffer
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		output.WriteString(scanner.Text() + "\n")
	}

	// Debug: Print the raw test output
	fmt.Println("Raw gotestsum output:")
	fmt.Println(output.String())

	// Regex to detect flaky tests
	flakyTestPattern := regexp.MustCompile(`FAIL\s+([\w./]+)\s+\(re-run\s+\d+\)`)

	// Count flaky test occurrences
	flakyCounts := make(map[string]int)
	fmt.Println("Parsing gotestsum output:")
	scanner = bufio.NewScanner(strings.NewReader(output.String()))
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line) // Debug: Log each line being processed
		matches := flakyTestPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			testName := matches[1]
			flakyCounts[testName]++
		}
	}

	// If no flaky tests are detected
	if len(flakyCounts) == 0 {
		fmt.Println("No flaky tests detected.")
		_ = os.WriteFile("flaky_report.json", []byte("No flaky tests detected."), 0644)
		return
	}

	// Build the flaky test report
	var report FlakyReport
	for testName, count := range flakyCounts {
		report.Tests = append(report.Tests, FlakyTest{Name: testName, Occurrences: count})
	}

	// Write the report to a JSON file
	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile("flaky_report.json", reportBytes, 0644)
	if err != nil {
		fmt.Printf("Error writing flaky report: %v\n", err)
		os.Exit(1)
	}

	// Print the detected flaky tests
	fmt.Println("Flaky tests detected:")
	for _, test := range report.Tests {
		fmt.Printf("- `%s`: %d flakiness occurrence(s)\n", test.Name, test.Occurrences)
	}
}
