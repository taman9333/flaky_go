package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

type FlakyTest struct {
	TestName    string `json:"test_name"`
	Occurrences int    `json:"occurrences"`
}

func main() {
	// Read filtered output from stdin
	var output bytes.Buffer
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		output.WriteString(scanner.Text() + "\n")
	}

	// Regex to detect test failures with package and test name
	flakyTestPattern := regexp.MustCompile(`===\s+FAIL:\s+(\w+)\s+(\w+)`)

	// Map to track test failures
	flakyCounts := make(map[string]int)

	// Parse the filtered output
	scanner = bufio.NewScanner(&output)
	for scanner.Scan() {
		line := scanner.Text()
		matches := flakyTestPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			packageName := matches[1]
			testName := matches[2]
			fullTestName := fmt.Sprintf("%s.%s", packageName, testName)
			flakyCounts[fullTestName]++
		}
	}

	// Convert the results to JSON
	if len(flakyCounts) == 0 {
		fmt.Println("[]") // Output empty JSON array
		return
	}

	var flakyTests []FlakyTest
	for test, count := range flakyCounts {
		if count <= 10 {
			flakyTests = append(flakyTests, FlakyTest{
				TestName:    test,
				Occurrences: count,
			})
		}
	}

	jsonOutput, err := json.MarshalIndent(flakyTests, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
