package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
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

	// Read existing data from the spreadsheet
	// readRange := fmt.Sprintf("%s!A:C", sheetName)
	readRange := "A:C"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	// Build a map of existing test names to their row indices and flake counts
	existingData := map[string]int{}   // Map of testName to rowIndex
	existingFlakes := map[string]int{} // Map of testName to existing flakes
	if len(resp.Values) > 0 {
		for i, row := range resp.Values {
			if len(row) < 2 {
				continue
			}
			testName := fmt.Sprintf("%v", row[0])
			flakesStr := fmt.Sprintf("%v", row[1])
			flakes, err := strconv.Atoi(flakesStr)
			if err != nil {
				flakes = 0
			}
			existingData[testName] = i + 1 // Google Sheets rows are 1-indexed
			existingFlakes[testName] = flakes
		}
	}

	// Parse flaky report and prepare data for update or append
	lines := splitLines(report)
	var rowsToAppend [][]interface{}
	var dataToUpdate []*sheets.ValueRange
	for _, line := range lines {
		parts := splitFlakyLine(line)
		if len(parts) != 2 {
			continue
		}
		testName, flakesStr := parts[0], parts[1]
		newFlakes, err := strconv.Atoi(flakesStr)
		if err != nil {
			log.Printf("Invalid flakes number for test %s: %v", testName, err)
			continue
		}
		lastFlakied := time.Now().Format(time.RFC3339)
		if rowIndex, exists := existingData[testName]; exists {
			// Update existing row
			totalFlakes := existingFlakes[testName] + newFlakes
			// updateRange := fmt.Sprintf("%s!B%d:C%d", sheetName, rowIndex, rowIndex)
			updateRange := fmt.Sprintf("B%d:C%d", rowIndex, rowIndex)
			vr := &sheets.ValueRange{
				Range:  updateRange,
				Values: [][]interface{}{{totalFlakes, lastFlakied}},
			}
			dataToUpdate = append(dataToUpdate, vr)
		} else {
			// Append new row
			rowsToAppend = append(rowsToAppend, []interface{}{
				testName,
				newFlakes,
				lastFlakied,
			})
		}
	}

	// Update existing rows
	if len(dataToUpdate) > 0 {
		rb := &sheets.BatchUpdateValuesRequest{
			ValueInputOption: "USER_ENTERED",
			Data:             dataToUpdate,
		}
		_, err = srv.Spreadsheets.Values.BatchUpdate(spreadsheetID, rb).Do()
		if err != nil {
			log.Fatalf("Unable to batch update data: %v", err)
		}
	}

	// Append new rows
	if len(rowsToAppend) > 0 {
		// writeRange := fmt.Sprintf("%s!A:C", sheetName)
		writeRange := "A:C"
		valueRange := &sheets.ValueRange{
			Values: rowsToAppend,
		}
		_, err = srv.Spreadsheets.Values.Append(spreadsheetID, writeRange, valueRange).
			ValueInputOption("USER_ENTERED").
			Context(ctx).
			Do()
		if err != nil {
			log.Fatalf("Unable to append data to sheet: %v", err)
		}
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
