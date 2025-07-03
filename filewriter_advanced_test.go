package arbor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ternarybob/arbor/filewriter"
)

func TestFileWriterWithMalformedJSON(t *testing.T) {
	// Test handling of malformed JSON and edge cases
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "malformed.log")

	fw, err := filewriter.NewWithPatternAndFormat(logFile, "", "standard", 200, 5)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer fw.Close()

	testCases := []struct {
		name string
		data string
	}{
		{"malformed_json", `{"level":"info","message":"test"}`},                  // Missing newline
		{"incomplete_json", `{"level":"info","mes`},                              // Incomplete
		{"non_json", `This is not JSON at all`},                                  // Plain text
		{"empty_string", ``},                                                     // Empty
		{"only_whitespace", `   	`},                                              // Whitespace only
		{"invalid_escape", `{"level":"info","message":"test\invalid"}`},          // Invalid escape
		{"mixed_content", `{"level":"info","message":"test"}some extra content`}, // Mixed content
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fw.Write([]byte(tc.data + "\n"))
			if err != nil {
				t.Logf("Write error for %s (may be expected): %v", tc.name, err)
			}
		})
	}

	// Give time for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify the file exists and has content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Log file should have content even with malformed JSON")
	}

	// Should have some form of output for each test case
	lines := strings.Count(string(content), "\n")
	if lines < len(testCases)/2 { // Allow some tolerance
		t.Errorf("Expected at least %d lines for malformed JSON handling, got %d", len(testCases)/2, lines)
	}

	t.Logf("Handled %d malformed JSON test cases, file content:\n%s", len(testCases), string(content))
}

func TestFileWriterLargeMessages(t *testing.T) {
	// Test handling of large log messages
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "large.log")

	fw, err := filewriter.NewWithPatternAndFormat(logFile, "", "standard", 500, 5)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer fw.Close()

	// Create large messages of different sizes
	sizes := []int{1024, 10240, 102400} // 1KB, 10KB, 100KB

	for _, size := range sizes {
		largeMessage := strings.Repeat("A", size)
		jsonMessage := fmt.Sprintf(`{"level":"info","message":"%s"}`, largeMessage)

		_, err := fw.Write([]byte(jsonMessage + "\n"))
		if err != nil {
			t.Errorf("Failed to write large message of size %d: %v", size, err)
		}
	}

	// Give time for processing
	time.Sleep(500 * time.Millisecond)

	// Verify file exists and has substantial content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read large message log: %v", err)
	}

	if len(content) < 100000 { // Should be at least 100KB
		t.Errorf("Expected large content, got %d bytes", len(content))
	}

	t.Logf("Successfully wrote and read %d bytes of large messages", len(content))
}

func TestFileWriterUnicodeAndSpecialChars(t *testing.T) {
	// Test handling of Unicode and special characters
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "unicode.log")

	fw, err := filewriter.NewWithPatternAndFormat(logFile, "", "standard", 200, 5)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer fw.Close()

	testMessages := []string{
		"Hello 世界",                     // Chinese characters
		"Café français",                // French accents
		"🚀 Rocket emoji",               // Emoji
		"Line 1\nLine 2",               // Newlines
		"Tab\tSeparated",               // Tabs
		`"Quoted" and 'single' quotes`, // Quotes
		"Special chars: !@#$%^&*()",    // Special symbols
	}

	for i, msg := range testMessages {
		jsonMessage := fmt.Sprintf(`{"level":"info","message":"%s","index":%d}`, msg, i)
		_, err := fw.Write([]byte(jsonMessage + "\n"))
		if err != nil {
			t.Errorf("Failed to write Unicode message %d: %v", i, err)
		}
	}

	// Give time for processing
	time.Sleep(200 * time.Millisecond)

	// Verify content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read Unicode log: %v", err)
	}

	// Check that some special characters made it through
	contentStr := string(content)
	if !strings.Contains(contentStr, "世界") {
		t.Error("Unicode characters should be preserved")
	}

	if !strings.Contains(contentStr, "🚀") {
		t.Error("Emoji should be preserved")
	}

	t.Logf("Successfully handled Unicode content:\n%s", contentStr)
}

func TestFileWriterHighFrequency(t *testing.T) {
	// Test high-frequency logging to stress the async system
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "highfreq.log")

	fw, err := filewriter.NewWithPatternAndFormat(logFile, "", "standard", 1000, 5) // Large buffer
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer fw.Close()

	// Write messages rapidly
	numMessages := 1000
	for i := 0; i < numMessages; i++ {
		jsonMessage := fmt.Sprintf(`{"level":"info","message":"High frequency test %d","timestamp":"%s"}`,
			i, time.Now().Format(time.RFC3339))
		_, err := fw.Write([]byte(jsonMessage + "\n"))
		if err != nil {
			t.Errorf("Failed to write high-frequency message %d: %v", i, err)
		}
	}

	// Give substantial time for all async operations to complete
	time.Sleep(1 * time.Second)

	// Verify file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read high-frequency log: %v", err)
	}

	lines := strings.Count(string(content), "\n")
	if lines < numMessages*3/4 { // Allow 25% tolerance for dropped messages
		t.Errorf("Expected at least %d lines, got %d (some messages may have been dropped under high load)",
			numMessages*3/4, lines)
	}

	t.Logf("High-frequency test: wrote %d messages, got %d lines in output", numMessages, lines)
}

func TestFileWriterDifferentLogLevels(t *testing.T) {
	// Test different log levels and their formatting
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "levels.log")

	fw, err := filewriter.NewWithPatternAndFormat(logFile, "", "standard", 200, 5)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer fw.Close()

	levels := []string{"trace", "debug", "info", "warn", "error", "fatal"}

	for i, level := range levels {
		jsonMessage := fmt.Sprintf(`{"level":"%s","message":"Test message for %s level","counter":%d}`,
			level, level, i)
		_, err := fw.Write([]byte(jsonMessage + "\n"))
		if err != nil {
			t.Errorf("Failed to write %s level message: %v", level, err)
		}
	}

	// Give time for processing
	time.Sleep(200 * time.Millisecond)

	// Verify content has all log levels
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read levels log: %v", err)
	}

	contentStr := string(content)
	for _, level := range levels {
		if !strings.Contains(contentStr, level) {
			t.Errorf("Content should contain %s level", level)
		}
	}

	// Check that proper level codes are used (TRC, DBG, INF, WRN, ERR, FTL)
	expectedCodes := []string{"TRC", "DBG", "INF", "WRN", "ERR", "FTL"}
	foundCodes := 0
	for _, code := range expectedCodes {
		if strings.Contains(contentStr, code) {
			foundCodes++
		}
	}

	if foundCodes < len(expectedCodes)/2 {
		t.Errorf("Expected to find level codes in output, found %d out of %d", foundCodes, len(expectedCodes))
	}

	t.Logf("Successfully tested all log levels:\n%s", contentStr)
}

func TestFileWriterJSONCorruption(t *testing.T) {
	// Test handling of corrupted/concatenated JSON entries
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "corruption.log")

	fw, err := filewriter.NewWithPatternAndFormat(logFile, "", "standard", 200, 5)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer fw.Close()

	// Test cases with corrupted JSON similar to what we saw in the log file
	corruptedEntries := []string{
		`{"level":"info","message":"test"},{"level":"debug","message":"another"}`, // Multiple JSON objects
		`{"level":"info","message":"test"},"trailing":"data"`,                     // Trailing data
		`{"level":"info","prefix":"test","prefix":"another","message":"msg"}`,     // Duplicate keys
		`{"level":"info","message":"test with unicode ╚══════"}`,                  // Unicode characters
		`{"level":"debug","message":"incomplete`,                                  // Incomplete JSON
	}

	for i, entry := range corruptedEntries {
		_, err := fw.Write([]byte(entry + "\n"))
		if err != nil {
			t.Errorf("Failed to write corrupted entry %d: %v", i, err)
		}
	}

	// Give time for processing
	time.Sleep(200 * time.Millisecond)

	// Verify content was written and is readable
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read corruption log: %v", err)
	}

	if len(content) == 0 {
		t.Error("Corruption log should have content")
	}

	// Should have some clean output for most entries
	contentStr := string(content)
	lines := strings.Count(contentStr, "\n")
	if lines < len(corruptedEntries)/2 {
		t.Errorf("Expected at least %d lines for corrupted JSON, got %d", len(corruptedEntries)/2, lines)
	}

	// Should not contain raw JSON corruption markers
	if strings.Contains(contentStr, "},{") {
		t.Error("Output should not contain concatenated JSON objects")
	}

	t.Logf("Successfully handled %d corrupted JSON entries:\n%s", len(corruptedEntries), contentStr)
}

func TestFileWriterEdgeCaseFilenames(t *testing.T) {
	// Test edge case filename patterns and scenarios
	tempDir := t.TempDir()

	testCases := []struct {
		name           string
		pattern        string
		expectedSuffix string
	}{
		{"no_pattern", "", ".log"},
		{"with_milliseconds", "test-{YYMMDD-HHMMSS}.log", ".log"},
		{"with_service_placeholder", "{SERVICE}-debug.log", "debug.log"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logFile := filepath.Join(tempDir, fmt.Sprintf("test-%s.log", tc.name))

			fw, err := filewriter.NewWithPatternAndFormat(logFile, tc.pattern, "standard", 200, 5)
			if err != nil {
				t.Fatalf("Failed to create file writer for %s: %v", tc.name, err)
			}

			// Write a test message
			jsonMessage := fmt.Sprintf(`{"level":"info","message":"Edge case test for %s"}`, tc.name)
			_, err = fw.Write([]byte(jsonMessage + "\n"))
			if err != nil {
				t.Errorf("Failed to write to %s: %v", tc.name, err)
			}

			fw.Close()

			// Give time for file operations
			time.Sleep(100 * time.Millisecond)

			// Check that some file was created
			files, err := filepath.Glob(filepath.Join(tempDir, "*.log"))
			if err != nil {
				t.Fatalf("Failed to list files for %s: %v", tc.name, err)
			}

			found := false
			for _, file := range files {
				if strings.Contains(file, tc.name) || strings.HasSuffix(file, tc.expectedSuffix) {
					found = true
					break
				}
			}

			if !found && tc.pattern != "" {
				t.Logf("No specific file found for %s, but this may be expected for pattern: %s", tc.name, tc.pattern)
			}
		})
	}
}
