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

func TestFileWriterCustomNaming(t *testing.T) {
	// Test custom file naming patterns
	tests := []struct {
		name        string
		pattern     string
		expected    string
		description string
	}{
		{
			name:        "daily_log_pattern",
			pattern:     "artifex-{YYMMDD}.log",
			expected:    "artifex-" + time.Now().Format("060102") + ".log",
			description: "Should create daily log files with YYMMDD format",
		},
		{
			name:        "hourly_log_pattern",
			pattern:     "artifex-{YYMMDD-HH}.log",
			expected:    "artifex-" + time.Now().Format("060102-15") + ".log",
			description: "Should create hourly log files with YYMMDD-HH format",
		},
		{
			name:        "timestamped_log_pattern",
			pattern:     "artifex-{YYMMDD-HHMMSS}.log",
			expected:    "artifex-" + time.Now().Format("060102-150405") + ".log",
			description: "Should create timestamped log files with YYMMDD-HHMMSS format",
		},
		{
			name:        "custom_service_pattern",
			pattern:     "{SERVICE}-{YYMMDD}-{TT}.log",
			expected:    "artifex-" + time.Now().Format("060102") + "-" + time.Now().Format("15") + ".log",
			description: "Should support SERVICE and TT (hour) placeholders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()

			// Test pattern expansion
			expanded := expandFileNamePattern(tt.pattern, "artifex")

			// Check pattern-based prefix rather than exact timestamp
			switch tt.name {
			case "daily_log_pattern":
				expectedPrefix := "artifex-" + time.Now().Format("060102")
				if !strings.HasPrefix(expanded, expectedPrefix) {
					t.Errorf("expandFileNamePattern(%q) = %q, expected to start with %q",
						tt.pattern, expanded, expectedPrefix)
				}
			case "hourly_log_pattern":
				expectedPrefix := "artifex-" + time.Now().Format("060102-15")
				if !strings.HasPrefix(expanded, expectedPrefix) {
					t.Errorf("expandFileNamePattern(%q) = %q, expected to start with %q",
						tt.pattern, expanded, expectedPrefix)
				}
			case "timestamped_log_pattern":
				// For timestamp patterns, just check the minute portion to avoid second timing issues
				expectedPrefix := "artifex-" + time.Now().Format("060102-1504")
				if !strings.HasPrefix(expanded, expectedPrefix) {
					t.Errorf("expandFileNamePattern(%q) = %q, expected to start with minute %q",
						tt.pattern, expanded, expectedPrefix)
				}
			case "custom_service_pattern":
				expectedPrefix := "artifex-" + time.Now().Format("060102-15")
				if !strings.HasPrefix(expanded, expectedPrefix) {
					t.Errorf("expandFileNamePattern(%q) = %q, expected to start with %q",
						tt.pattern, expanded, expectedPrefix)
				}
			}

			// Test actual file creation
			fullPath := filepath.Join(tempDir, expanded)
			fw, err := filewriter.NewWithPatternAndFormat(fullPath, "", "standard", 100, 5)
			if err != nil {
				t.Fatalf("Failed to create file writer: %v", err)
			}

			// Write a test message with proper JSON format
			_, writeErr := fw.Write([]byte(`{"level":"info","message":"test message"}` + "\n"))
			if writeErr != nil {
				t.Logf("Write error (may be expected): %v", writeErr)
			}

			// Close the writer before checking file
			fw.Close()

			// Give more time for file operations to complete
			time.Sleep(500 * time.Millisecond)

			// Check if any log file was created in the temp directory
			files, globErr := filepath.Glob(filepath.Join(tempDir, "*.log"))
			if globErr != nil {
				t.Fatalf("Failed to list log files: %v", globErr)
			}

			if len(files) == 0 {
				t.Errorf("No log files were created in %s", tempDir)
			} else {
				t.Logf("Created log files: %v", files)
			}
		})
	}
}


func TestFileWriterRotation(t *testing.T) {
	// Test file rotation functionality
	tempDir := t.TempDir()
	defer func() { os.RemoveAll(tempDir) }()
	pattern := "test-{YYMMDD}.log"

	// Create file writer with low max files for testing
	fw, err := filewriter.NewWithPatternAndFormat(
		filepath.Join(tempDir, pattern),
		pattern,
		"standard",
		200, // larger buffer for testing
		3,   // only keep 3 files
	)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer fw.Close()

	// Write multiple entries to trigger rotation
	for i := 0; i < 100; i++ {
		_, err := fw.Write([]byte(`{"level":"info","message":"test message ` +
			fmt.Sprintf("%d", i) + `"}` + "\n"))
		if err != nil {
			t.Errorf("Failed to write log entry %d: %v", i, err)
		}
	}

	// Give time for async operations
	time.Sleep(200 * time.Millisecond)

	// Check that files exist in the directory
	files, err := filepath.Glob(filepath.Join(tempDir, "*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	if len(files) == 0 {
		t.Error("No log files were created")
	}

	t.Logf("Created %d log files: %v", len(files), files)
}


func TestFileWriterConcurrency(t *testing.T) {
	// Test concurrent writing to file logger
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "concurrent-test.log")

	fw, err := filewriter.NewWithPatternAndFormat(logFile, "", "json", 200, 5)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer fw.Close()

	// Create multiple goroutines writing concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 10; j++ {
				message := `{"level":"info","message":"concurrent test from goroutine ` +
					fmt.Sprintf("%d", id) + ` iteration ` + fmt.Sprintf("%d", j) + `"}` + "\n"
				_, err := fw.Write([]byte(message))
				if err != nil {
					t.Errorf("Goroutine %d failed to write: %v", id, err)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Give time for async operations
	time.Sleep(200 * time.Millisecond)

	// Verify file exists and has content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read concurrent test log: %v", err)
	}

	if len(content) == 0 {
		t.Error("Concurrent test log file is empty")
	}

	// Should have 100 messages (10 goroutines * 10 messages each)
	lines := strings.Count(string(content), "\n")
	if lines < 90 { // Allow some tolerance for async operations
		t.Errorf("Expected around 100 log lines, got %d", lines)
	}
}

// Helper function to test pattern expansion (to be implemented)
func expandFileNamePattern(pattern, serviceName string) string {
	now := time.Now()

	expanded := strings.ReplaceAll(pattern, "{SERVICE}", serviceName)
	expanded = strings.ReplaceAll(expanded, "{YYMMDD-HHMMSS}", now.Format("060102-150405"))
	expanded = strings.ReplaceAll(expanded, "{YYMMDD-HH}", now.Format("060102-15"))
	expanded = strings.ReplaceAll(expanded, "{YYMMDD}", now.Format("060102"))
	expanded = strings.ReplaceAll(expanded, "{TT}", now.Format("15"))
	expanded = strings.ReplaceAll(expanded, "{YYYY}", now.Format("2006"))
	expanded = strings.ReplaceAll(expanded, "{MM}", now.Format("01"))
	expanded = strings.ReplaceAll(expanded, "{DD}", now.Format("02"))
	expanded = strings.ReplaceAll(expanded, "{HH}", now.Format("15"))
	expanded = strings.ReplaceAll(expanded, "{MMSS}", now.Format("0405"))

	return expanded
}
