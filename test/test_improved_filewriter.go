package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ternarybob/arbor/filewriter"
)

func main() {
	// Create a temporary file for testing
	file, err := os.CreateTemp("", "test_improved_*.log")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer os.Remove(file.Name())
	defer file.Close()

	fmt.Printf("Testing improved filewriter with file: %s\n", file.Name())

	// Create filewriter with standard format (pipe-delimited)
	fw := filewriter.New(file, 100)

	// Test 1: Write some JSON log entries
	fmt.Println("\n=== Test 1: JSON to Pipe-Delimited Conversion ===")

	logEntries := []map[string]interface{}{
		{
			"level":   "info",
			"time":    time.Now().Format(time.RFC3339),
			"prefix":  "main",
			"message": "Starting Artifex File Receiver...",
		},
		{
			"level":   "info",
			"time":    time.Now().Format(time.RFC3339),
			"prefix":  "startup",
			"message": "Initializing database manager...",
		},
		{
			"level":      "info",
			"time":       time.Now().Format(time.RFC3339),
			"prefix":     "RavenDBService",
			"message":    "Successfully connected to database",
			"connection": "PrimaryRavenDB",
			"url":        "http://localhost:8080",
		},
		{
			"level":   "debug",
			"time":    time.Now().Format(time.RFC3339),
			"prefix":  "startup",
			"message": "Starting WebServer Service(s)",
		},
	}

	for _, entry := range logEntries {
		jsonData, _ := json.Marshal(entry)
		fw.Write(jsonData)
	}

	// Test 2: Write duplicate entries to test deduplication
	fmt.Println("\n=== Test 2: Duplicate Detection ===")

	duplicateEntry := map[string]interface{}{
		"level":   "debug",
		"time":    time.Now().Format(time.RFC3339),
		"prefix":  "startup",
		"message": "Starting WebServer Service(s)",
	}

	// Write the same entry multiple times
	for i := 0; i < 5; i++ {
		jsonData, _ := json.Marshal(duplicateEntry)
		fw.Write(jsonData)
		time.Sleep(10 * time.Millisecond) // Small delay
	}

	// Test 3: Write malformed JSON
	fmt.Println("\n=== Test 3: Malformed JSON Handling ===")

	malformedEntries := []string{
		`{"level":"info","prefix":"test","message":"incomplete`,
		`not json at all`,
		`{"level":"warn","message":"box chars: ╚══════════════════════════════════════════════════════╝"}`,
	}

	for _, entry := range malformedEntries {
		fw.Write([]byte(entry))
	}

	// Wait for all writes to complete
	time.Sleep(100 * time.Millisecond)
	fw.Close()

	// Read and display the results
	fmt.Println("\n=== Results ===")
	content, err := os.ReadFile(file.Name())
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	fmt.Printf("File content:\n%s\n", string(content))

	// Count lines to show deduplication worked
	lines := 0
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}
	fmt.Printf("Total lines written: %d (should be less than total entries due to deduplication)\n", lines)
}
