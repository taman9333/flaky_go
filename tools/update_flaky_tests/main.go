package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	spreadsheetID = "1abPAJi43FxnYgcjx6wnex_GbiTha3UEvqWKNCCAGpFI" // Replace with your Google Sheets ID
	sheetName     = "Flaky tests"                                  // Replace with your sheet name
)

func main() {
	// Load Google credentials from environment variable
	credentials := os.Getenv("GOOGLE_CREDENTIALS")
	if credentials == "" {
		log.Fatal("GOOGLE_CREDENTIALS environment variable not set")
	}

	// Parse credentials
	config := []byte(credentials)
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(config))
	if err != nil {
		log.Fatalf("Unable to create Sheets service: %v", err)
	}

	// Read flaky report
	reportData, err := os.ReadFile("flaky_report.txt")
	if err != nil {
		log.Fatalf("Unable to read flaky_report.txt: %v", err)
	}

	report := string(reportData)
	if report == "" {
		fmt.Println("No flaky tests detected. Skipping update.")
		return
	}

	// Parse flaky report
	var rows [][]interface{}
	lines := splitLines(report)
	for _, line := range lines {
		parts := splitFlakyLine(line)
		if len(parts) != 2 {
			continue
		}
		testName, flakes := parts[0], parts[1]
		rows = append(rows, []interface{}{
			testName,
			flakes,
			time.Now().Format(time.RFC3339),
		})
	}

	// Prepare the range and values
	writeRange := fmt.Sprintf("%s!A:C", sheetName)
	valueRange := &sheets.ValueRange{
		Values: rows,
	}

	// Append data to the sheet
	_, err = srv.Spreadsheets.Values.Append(spreadsheetID, writeRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		log.Fatalf("Unable to append data to sheet: %v", err)
	}

	fmt.Println("Flaky tests updated successfully.")
}

// splitLines splits the report into lines.
func splitLines(report string) []string {
	lines := []string{}
	for _, line := range strings.Split(report, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

// splitFlakyLine splits a line into test name and occurrence count.
func splitFlakyLine(line string) []string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		parts[1] = strings.TrimSpace(strings.ReplaceAll(parts[1], "occurrences", ""))
	}
	return parts
}
