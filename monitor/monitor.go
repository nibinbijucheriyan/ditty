package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// RiskAlert represents the structure of a risk alert payload
type RiskAlert struct {
	File       string `json:"file"`
	Function   string `json:"function"`
	IssueType  string `json:"issue_type"`
	Severity   string `json:"severity"`
	LineNumber int    `json:"line_number"`
	Timestamp  string `json:"timestamp"`
	Message    string `json:"message"`
}

// Mock eBPF probe scenarios
var mockScenarios = []RiskAlert{
	{
		File:       "src/api/handlers.go",
		Function:   "ProcessRequest",
		IssueType:  "high_cpu_cycles",
		Severity:   "high",
		LineNumber: 145,
		Message:    "Detected inefficient loop consuming excessive CPU cycles",
	},
	{
		File:       "src/database/query.go",
		Function:   "FetchUserData",
		IssueType:  "memory_leak",
		Severity:   "critical",
		LineNumber: 89,
		Message:    "Potential memory leak detected in database connection pool",
	},
	{
		File:       "src/utils/parser.go",
		Function:   "ParseJSON",
		IssueType:  "inefficient_algorithm",
		Severity:   "medium",
		LineNumber: 234,
		Message:    "O(n²) algorithm detected, consider optimization",
	},
	{
		File:       "src/services/carbon.go",
		Function:   "CalculateEmissions",
		IssueType:  "high_carbon_footprint",
		Severity:   "high",
		LineNumber: 67,
		Message:    "High carbon emissions detected from inefficient compute operations",
	},
	{
		File:       "src/network/client.go",
		Function:   "SendRequest",
		IssueType:  "blocking_io",
		Severity:   "medium",
		LineNumber: 112,
		Message:    "Blocking I/O operation detected, consider async implementation",
	},
}

func main() {
	// Get webhook URL from environment variable or use default
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		webhookURL = "http://localhost:8080/api/risk-alert"
		log.Printf("WEBHOOK_URL not set, using default: %s", webhookURL)
	}

	// Get monitoring interval from environment variable or use default (30 seconds)
	interval := 30 * time.Second
	if intervalStr := os.Getenv("MONITOR_INTERVAL"); intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			interval = d
		}
	}

	log.Printf("Starting Risk Monitor...")
	log.Printf("Webhook URL: %s", webhookURL)
	log.Printf("Monitoring interval: %s", interval)
	log.Println("Simulating eBPF performance and carbon monitoring probe...")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Start monitoring loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send initial alert immediately
	sendRiskAlert(webhookURL)

	// Continue monitoring
	for range ticker.C {
		sendRiskAlert(webhookURL)
	}
}

func sendRiskAlert(webhookURL string) {
	// Select a random scenario
	scenario := mockScenarios[rand.Intn(len(mockScenarios))]
	
	// Add timestamp
	scenario.Timestamp = time.Now().UTC().Format(time.RFC3339)

	// Marshal to JSON
	payload, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	log.Printf("\n=== Sending Risk Alert ===")
	log.Printf("File: %s", scenario.File)
	log.Printf("Function: %s", scenario.Function)
	log.Printf("Issue: %s (Severity: %s)", scenario.IssueType, scenario.Severity)
	log.Printf("Line: %d", scenario.LineNumber)
	log.Printf("Message: %s", scenario.Message)

	// Send HTTP POST request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error sending alert: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("✓ Alert sent successfully (Status: %d)", resp.StatusCode)
	} else {
		log.Printf("✗ Alert failed (Status: %d)", resp.StatusCode)
	}
}

// Made with Bob
