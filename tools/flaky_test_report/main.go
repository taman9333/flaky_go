package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
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

	// Check exit code
	cmd := exec.Command("bash", "-c", "echo $?")
	exitCodeOutput, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting exit code: %v\n", err)
		os.Exit(1)
	}

	exitCode := strings.TrimSpace(string(exitCodeOutput))
	if exitCode != "0" {
		// Non-zero exit code, indicate a retry is needed
		fmt.Println("Non-zero exit code detected. Tests need to be retried.")
		return
	}

	// Process output for flaky tests
	flakyTestPattern := regexp.MustCompile(`FAIL ([\w./]+) \((re-run \d+)\)`)
	flakyCounts := make(map[string]int)

	scanner = bufio.NewScanner(strings.NewReader(output.String()))
	for scanner.Scan() {
		line := scanner.Text()
		matches := flakyTestPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			testName := matches[1]
			flakyCounts[testName]++
		}
	}

	// If there are no flaky tests, exit
	if len(flakyCounts) == 0 {
		fmt.Println("No flaky tests detected.")
		return
	}

	// Log flaky tests and their counts
	for testName, count := range flakyCounts {
		fmt.Printf("- `%s`: %d flakiness occurrence(s)\n", testName, count)
	}
}
