package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	spreadsheetID = "1abPAJi43FxnYgcjx6wnex_GbiTha3UEvqWKNCCAGpFI" // Replace with your Google Sheets ID
	sheetName     = "Flaky tests"                                  // Replace with your sheet name
)

type FlakyTest struct {
	TestName    string `json:"test_name"`
	Occurrences int    `json:"occurrences"`
}

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

	// Read flaky report JSON
	reportData, err := os.ReadFile("flaky_report.json")
	if err != nil {
		log.Fatalf("Unable to read flaky_report.json: %v", err)
	}

	var flakyTests []FlakyTest
	if err := json.Unmarshal(reportData, &flakyTests); err != nil {
		log.Fatalf("Unable to parse flaky report JSON: %v", err)
	}

	if len(flakyTests) == 0 {
		fmt.Println("No flaky tests detected. Skipping update.")
		return
	}

	// Read existing data from the spreadsheet
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

	// Prepare data for update or append
	var rowsToAppend [][]interface{}
	var dataToUpdate []*sheets.ValueRange
	for _, test := range flakyTests {
		lastFlakied := time.Now().Format(time.RFC3339)
		if rowIndex, exists := existingData[test.TestName]; exists {
			// Update existing row
			totalFlakes := existingFlakes[test.TestName] + test.Occurrences
			updateRange := fmt.Sprintf("B%d:C%d", rowIndex, rowIndex)
			vr := &sheets.ValueRange{
				Range:  updateRange,
				Values: [][]interface{}{{totalFlakes, lastFlakied}},
			}
			dataToUpdate = append(dataToUpdate, vr)
		} else {
			// Append new row
			rowsToAppend = append(rowsToAppend, []interface{}{
				test.TestName,
				test.Occurrences,
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
