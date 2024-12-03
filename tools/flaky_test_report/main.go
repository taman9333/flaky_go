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

	// Regex for flaky test detection
	flakyTestPattern := regexp.MustCompile(`FAIL\s+([\w.\/]+)\s+\(re-run\s+(\d+)\)\s+\(\d+\.\d+s\)`)

	// Count flaky tests
	flakyCounts := make(map[string]int)
	fmt.Println("Parsing gotestsum output:")
	scanner = bufio.NewScanner(strings.NewReader(output.String()))
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	fmt.Println("Processing line:", line) // Debug each line
	// 	matches := flakyTestPattern.FindStringSubmatch(line)
	// 	if len(matches) > 0 {
	// 		testName := matches[1]
	// 		fmt.Printf("Matched flaky test: %s\n", testName) // Log matched test
	// 		flakyCounts[testName]++
	// 	}
	// }

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("Processing line:", line) // Log every line
		matches := flakyTestPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			fmt.Printf("Matched flaky test: %s\n", matches[1]) // Log matched lines
			flakyCounts[matches[1]]++
		} else {
			fmt.Println("No match for line:", line) // Log lines that don't match
		}
	}

	// If no flaky tests are detected
	fmt.Println("Flaky counts:", flakyCounts)
	if len(flakyCounts) == 0 {
		fmt.Println("No flaky tests detected.")
		return
	}

	// Log flaky tests
	fmt.Println("Detected flaky tests:")
	for testName, count := range flakyCounts {
		fmt.Printf("- `%s`: %d occurrences\n", testName, count)
	}
}
