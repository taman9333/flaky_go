package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
)

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

	if len(flakyCounts) == 0 {
		return
	}

	for test, count := range flakyCounts {
		if count <= 10 {
			fmt.Printf("- `%s`: %d occurrences\n", test, count)
		}
	}
}
