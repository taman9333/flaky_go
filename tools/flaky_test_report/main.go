package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Read gotestsum output from stdin
	var output bytes.Buffer
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		output.WriteString(scanner.Text() + "\n")
	}

	// Regex to detect flaky test failures
	flakyTestPattern := regexp.MustCompile(`===\s+FAIL:\s+([^\s(]+)`)

	// Map to count flaky tests
	flakyCounts := make(map[string]int)

	// Log start of parsing
	fmt.Println("Parsing gotestsum output:")
	scanner = bufio.NewScanner(strings.NewReader(output.String()))
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("Processing line: %s\n", line) // Debug each line

		// Match flaky tests
		matches := flakyTestPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			testName := matches[1]
			fmt.Printf("Matched flaky test: %s\n", testName) // Log matched test
			flakyCounts[testName]++
		} else {
			fmt.Printf("No match for line: %s\n", line) // Log unmatched lines
		}
	}

	// If no flaky tests are detected
	if len(flakyCounts) == 0 {
		fmt.Println("No flaky tests detected.")
		return
	}

	// Log flaky tests
	fmt.Println("Flaky counts:", flakyCounts)
	fmt.Println("Detected flaky tests:")
	for testName, count := range flakyCounts {
		fmt.Printf("- `%s`: %d occurrences\n", testName, count)
	}
}
