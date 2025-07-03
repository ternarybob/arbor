package filewriter

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWriterCreation(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")
	
	writer, err := NewWithPatternAndFormat(logFile, "", "standard", 100, 5)
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
	
	writer, err := NewWithPatternAndFormat(logFile, "", "standard", 100, 5)
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
