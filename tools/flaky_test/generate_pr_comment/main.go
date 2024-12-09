package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type FlakyTest struct {
	TestName    string `json:"test_name"`
	Occurrences int    `json:"occurrences"`
}

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	var flakyTests []FlakyTest
	if err := json.Unmarshal(data, &flakyTests); err != nil {
		log.Fatalf("Failed to parse flaky_report.json: %v", err)
	}

	if len(flakyTests) == 0 {
		fmt.Println("No flaky tests detected. Skipping PR comment generation.")
		return
	}

	commentBody := "### Flaky Tests Detected:\n\n"
	for _, test := range flakyTests {
		commentBody += fmt.Sprintf("- `%s`: %d occurrences\n", test.TestName, test.Occurrences)
	}

	if err := os.WriteFile("pr_comment_body.txt", []byte(commentBody), 0644); err != nil {
		log.Fatalf("Failed to write pr_comment_body.txt: %v", err)
	}

	fmt.Println("PR comment body generated successfully.")
}
