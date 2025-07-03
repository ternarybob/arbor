package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ternarybob/arbor/filewriter"
)

func main() {
	fmt.Println("=== Combined FileWriter Test Suite ===")
	
	// Run the improved filewriter tests
	runImprovedFilewriterTests()
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	
	// Run the JSON fix tests
	runJSONFixTests()
}

func runImprovedFilewriterTests() {
	fmt.Println("\n=== IMPROVED FILEWRITER TESTS ===")
	
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

func runJSONFixTests() {
	fmt.Println("\n=== JSON FIX TESTS ===")
	
	// Create a temporary log file
	tempDir := os.TempDir()
	logFile := filepath.Join(tempDir, "json_fix_test.log")

	// Create filewriter with standard format
	fw, err := filewriter.NewWithPatternAndFormat(logFile, "", "standard", 200, 5)
	if err != nil {
		fmt.Printf("Failed to create file writer: %v\n", err)
		return
	}
	defer fw.Close()

	// Test cases that match your original problem
	testCases := []string{
		// Valid JSON that should be converted to pipe format
		`{"level":"info","prefix":"main","time":"2025-07-03T23:06:19+10:00","message":"Starting Artifex File Receiver..."}`,
		`{"level":"debug","prefix":"startup","time":"2025-07-03T23:06:19+10:00","message":"Starting WebServer Service(s)"}`,
		`{"level":"info","prefix":"RavenDBService","database":"PrimaryGoDatabase","connection":"PrimaryRavenDB","time":"2025-07-03T23:06:19+10:00","message":"Database assumed to exist or will be created"}`,

		// Already formatted entries (should pass through unchanged)
		`INF|Jul  3 23:06:19|startup|Gin set to Debug mode (Development)`,
		`DBG|Jul  3 23:06:19|startup|Starting WebServer Service(s)`,

		// Malformed JSON (should be handled gracefully)
		`{"level":"info","prefix":"startup","time":"2025-07-03T23:06:19+10:00","message":"Starting WebServer Service(s)"}`,
		`{"level":"debug","message":"incomplete`,
		`corrupted json {"level":"info"} extra data`,
	}

	fmt.Println("Writing test cases to log file...")
	for i, testCase := range testCases {
		fmt.Printf("Test case %d: %s\n", i+1, testCase)
		_, err := fw.Write([]byte(testCase + "\n"))
		if err != nil {
			fmt.Printf("Error writing test case %d: %v\n", i+1, err)
		}
	}

	// Give time for async processing
	time.Sleep(500 * time.Millisecond)

	// Read and display the results
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("RESULTS - Log file content:")
	fmt.Println(strings.Repeat("=", 80))

	content, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Printf("Failed to read log file: %v\n", err)
		return
	}

	fmt.Print(string(content))

	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Log file location: %s\n", logFile)

	// Verify no JSON remains in output
	contentStr := string(content)
	if containsJSON(contentStr) {
		fmt.Println("❌ FAILED: Output still contains JSON format")
	} else {
		fmt.Println("✅ SUCCESS: All output is in proper pipe-separated format")
	}
}

// containsJSON checks if the content contains JSON-like structures
func containsJSON(content string) bool {
	return (len(content) > 0 &&
		(content[0] == '{' ||
			(len(content) > 10 && content[10:20] == `{"level":"`)))
}
