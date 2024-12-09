package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	jiraBaseURL = "https://payrails.atlassian.net/rest/api/2"
	email       = "abdelrahman.taman@payrails.com"
	jiraProject = "PR"
	component   = "Tests"
	label       = "flakytests"
)

var jiraClient = &http.Client{
	Timeout: 5 * time.Second,
}

type FlakyTest struct {
	TestName    string  `json:"test_name"`
	Occurrences float64 `json:"occurrences"`
}

type JiraIssue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary        string  `json:"summary"`
		FlakinessCount float64 `json:"customfield_10016"`
	} `json:"fields"`
}

func main() {
	apiToken := os.Getenv("JIRA_API_TOKEN")

	if apiToken == "" {
		log.Fatal("JIRA_API_TOKEN environment variables are not set")
	}

	data, err := os.ReadFile("flaky_report.json")
	if err != nil {
		log.Fatalf("Failed to read flaky_report.json: %v", err)
	}

	var flakyTests []FlakyTest
	if err := json.Unmarshal(data, &flakyTests); err != nil {
		log.Fatalf("Failed to parse flaky_report.json: %v", err)
	}

	if len(flakyTests) == 0 {
		fmt.Println("No flaky tests detected. Skipping Jira update.")
		return
	}

	existingTickets := fetchExistingTickets(apiToken, flakyTests)

	// Process flaky tests
	for _, test := range flakyTests {
		if ticket, exists := existingTickets[test.TestName]; exists {
			updateJiraTicket(apiToken, ticket, test)
		} else {
			createJiraTicket(apiToken, test)
		}
	}

	fmt.Println("Jira update completed successfully.")
}

func fetchExistingTickets(apiToken string, flakyTests []FlakyTest) map[string]JiraIssue {
	existingTickets := make(map[string]JiraIssue)

	var testNames []string
	for _, test := range flakyTests {
		testNames = append(testNames, fmt.Sprintf(`summary ~ "%s"`, test.TestName))
	}

	testNamesCondition := strings.Join(testNames, " OR ")

	jqlQuery := fmt.Sprintf(
		"project = %s AND issuetype = Spike AND component = %s AND labels = %s AND status != Done AND (%s)",
		jiraProject, component, label, testNamesCondition,
	)

	url := fmt.Sprintf("%s/search?jql=%s", jiraBaseURL, url.QueryEscape(jqlQuery))
	resp, err := makeJiraRequest("GET", url, apiToken, nil)
	if err != nil {
		log.Fatalf("Failed to fetch Jira tickets: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to fetch Jira tickets, status: %s, response: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var result struct {
		Issues []JiraIssue `json:"issues"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Failed to parse Jira response: %v", err)
	}

	for _, issue := range result.Issues {
		existingTickets[issue.Fields.Summary] = issue
	}

	return existingTickets
}

func createJiraTicket(apiToken string, test FlakyTest) {
	url := fmt.Sprintf("%s/issue", jiraBaseURL)
	data := map[string]interface{}{
		"fields": map[string]interface{}{
			"project":           map[string]string{"key": jiraProject},
			"summary":           test.TestName,
			"issuetype":         map[string]string{"name": "Spike"},
			"components":        []map[string]string{{"name": component}},
			"customfield_10016": test.Occurrences,
			"labels":            []string{label},
			"description": fmt.Sprintf("This test has flaked %v times. Last detected: %s",
				test.Occurrences, time.Now().Format(time.RFC3339)),
		},
	}

	payload, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Failed to create Jira payload: %v", err)
	}

	resp, err := makeJiraRequest("POST", url, apiToken, payload)
	if err != nil {
		log.Fatalf("Failed to create Jira ticket: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to create Jira ticket, status: %s, response: %s", resp.Status, string(body))
	}

	fmt.Printf("Created Jira ticket for test: %s\n", test.TestName)
}

func updateJiraTicket(apiToken string, ticket JiraIssue, test FlakyTest) {
	currentFlakinessCount := ticket.Fields.FlakinessCount

	// Might have concurrency issue
	newFlakinessCount := currentFlakinessCount + test.Occurrences

	url := fmt.Sprintf("%s/issue/%s", jiraBaseURL, ticket.ID)
	data := map[string]interface{}{
		"fields": map[string]interface{}{
			"customfield_10016": newFlakinessCount,
			"description": fmt.Sprintf(
				"This test has flaked %v times. Last detected: %s",
				newFlakinessCount, time.Now().Format(time.RFC3339),
			),
		},
	}

	payload, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Failed to create Jira update payload: %v", err)
	}

	resp, err := makeJiraRequest("PUT", url, apiToken, payload)
	if err != nil {
		log.Fatalf("Failed to update Jira ticket: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to update Jira ticket, status: %s, response: %s", resp.Status, string(body))
	}

	fmt.Printf("Updated Jira ticket for test: %s with new flakiness count: %v\n", test.TestName, newFlakinessCount)
}

func makeJiraRequest(method, url string, apiToken string, payload []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(email, apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := jiraClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}
