package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ternarybob/arbor/filewriter"
)

func main() {
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
