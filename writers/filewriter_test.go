package writers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test File Writeer
func TestFileWriterCreation(t *testing.T) {

	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	writer, err := NewFileWriterWithPattern(logFile, "", "standard", 100, 5)
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}

	// Immediately close to avoid file lock issues
	if writer != nil {
		writer.Close()
	}

	if writer == nil {
		t.Error("FileWriter should not be nil after creation")
	}
}

func TestFileWriterWrite(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	writer, err := NewFileWriterWithPattern(logFile, "", "standard", 100, 5)
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}
	defer writer.Close()

	testData := []byte("test file message\n")
	n, err := writer.Write(testData)

	if err != nil {
		t.Errorf("Write should not return error: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Write should return correct byte count: got %d, want %d", n, len(testData))
	}

	// Give time for async writing
	time.Sleep(100 * time.Millisecond)

	// Verify file was created and has content
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestFileWriterDifferentLogLevels(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "loglevels-test.log")

	writer, err := NewFileWriterWithPattern(logFile, "", "standard", 100, 5)
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}
	defer writer.Close()

	// Test different log levels with JSON format input
	testEntries := []struct {
		level   string
		message string
		counter int
	}{
		{"trace", "Test message for trace level", 1},
		{"debug", "Test message for debug level", 2},
		{"info", "Test message for info level", 3},
		{"warn", "Test message for warn level", 4},
		{"error", "Test message for error level", 5},
		{"fatal", "Test message for fatal level", 6},
	}

	for _, entry := range testEntries {
		// Create JSON log entry
		jsonEntry := fmt.Sprintf(`{"level":"%s","message":"%s","counter":%d}`, entry.level, entry.message, entry.counter)

		_, err := writer.Write([]byte(jsonEntry))
		if err != nil {
			t.Errorf("Failed to write %s level entry: %v", entry.level, err)
		}

		// Small delay between writes
		time.Sleep(10 * time.Millisecond)
	}

	// Give time for async writing
	time.Sleep(200 * time.Millisecond)

	// Read and verify the log file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)
	t.Logf("Log output:\n%s", output)

	// Verify that all log levels are present in the output
	// phuslu/log uses TRC, DBG, INF, WRN, ERR format (FTL becomes ERR with level=fatal)
	expectedLevels := []string{"TRC", "DBG", "INF", "WRN", "ERR", "ERR"} // Last two are ERR because fatal/panic become error
	for i, level := range expectedLevels {
		if !strings.Contains(output, level) {
			t.Errorf("Log output missing level: %s", level)
		}

		// Also check that the message content is present
		expectedMessage := testEntries[i].message
		if !strings.Contains(output, expectedMessage) {
			t.Errorf("Log output missing message: %s", expectedMessage)
		}
	}

	// Verify console writer format (timestamp LEVEL > message)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != len(testEntries) {
		t.Errorf("Expected %d log lines, got %d", len(testEntries), len(lines))
	}

	for i, line := range lines {
		if line == "" {
			continue
		}

		// Check that line contains timestamp, level code, and message
		// Format should be: YYYY-MM-DDTHH:MM:SS.sss+TZ LVL > message
		if !strings.Contains(line, ">") {
			t.Errorf("Log line should contain '>' separator: %s", line)
		}

		// Check that the expected level is present
		if i < len(expectedLevels) {
			if !strings.Contains(line, expectedLevels[i]) {
				t.Errorf("Log line should contain level %s: %s", expectedLevels[i], line)
			}
		}

		// Check that the expected message is present
		if i < len(testEntries) {
			if !strings.Contains(line, testEntries[i].message) {
				t.Errorf("Log line should contain message '%s': %s", testEntries[i].message, line)
			}
		}
	}
}
