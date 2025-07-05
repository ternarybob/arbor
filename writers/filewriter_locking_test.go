package writers

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFileLockingPrevention(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.log")

	// Create a FileWriter
	fw, err := NewFileWriter(testFile, 100, 5)
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}
	defer fw.Close()

	// Write some initial data
	testData := []byte(`{"level":"info","time":"2024-01-01T12:00:00Z","message":"Test message"}`)
	_, err = fw.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// Wait for write to complete
	time.Sleep(100 * time.Millisecond)

	// Verify the file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatalf("Test file was not created: %s", testFile)
	}

	t.Logf("Successfully created and wrote to: %s", testFile)
}

func TestNumberedFileCreation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "locked.log")

	// Create a FileWriter first to establish the primary file
	fw1, err := NewFileWriter(testFile, 100, 5)
	if err != nil {
		t.Fatalf("Failed to create first FileWriter: %v", err)
	}

	// Write some data to the first writer to establish the file
	testData1 := []byte(`{"level":"info","time":"2024-01-01T12:00:00Z","message":"First writer message"}`)
	_, err = fw1.Write(testData1)
	if err != nil {
		t.Fatalf("Failed to write test data to first writer: %v", err)
	}

	// Wait for write to complete
	time.Sleep(100 * time.Millisecond)

	// Create a second FileWriter that should try to use the same file
	// This should trigger the numbered file creation mechanism
	fw2, err := NewFileWriter(testFile, 100, 5)
	if err != nil {
		t.Fatalf("Failed to create second FileWriter: %v", err)
	}
	defer fw2.Close()

	// Write some data with the second writer
	testData2 := []byte(`{"level":"info","time":"2024-01-01T12:00:00Z","message":"Second writer message"}`)
	_, err = fw2.Write(testData2)
	if err != nil {
		t.Fatalf("Failed to write test data to second writer: %v", err)
	}

	// Wait for write to complete
	time.Sleep(200 * time.Millisecond)

	// Close the first writer
	fw1.Close()

	// List all files in the directory to see what was created
	files, err := filepath.Glob(filepath.Join(tempDir, "*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	t.Logf("Created files: %v", files)

	// We should have at least one file
	if len(files) == 0 {
		t.Fatalf("No log files were created")
	}

	// Check if we have the primary file
	primaryExists := false
	numberedExists := false

	for _, file := range files {
		baseName := filepath.Base(file)
		if baseName == "locked.log" {
			primaryExists = true
		} else if strings.Contains(baseName, "locked.") && strings.HasSuffix(baseName, ".log") {
			numberedExists = true
		}
	}

	if !primaryExists {
		t.Fatalf("Primary file 'locked.log' was not created")
	}

	t.Logf("Primary file exists: %v, Numbered file exists: %v", primaryExists, numberedExists)

	// Verify content was written to at least one file
	totalContent := ""
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read file %s: %v", file, err)
		}
		totalContent += string(content)
	}

	if !strings.Contains(totalContent, "First writer message") {
		t.Fatalf("First writer message not found in any file. Total content: %s", totalContent)
	}

	if !strings.Contains(totalContent, "Second writer message") {
		t.Fatalf("Second writer message not found in any file. Total content: %s", totalContent)
	}

	t.Logf("Both messages were successfully written to log files")
}

func TestConcurrentWritesToSameFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "concurrent.log")

	// Create multiple FileWriters trying to write to the same file
	const numWriters = 5
	const messagesPerWriter = 10

	var wg sync.WaitGroup
	var writers []*FileWriter

	// Create writers
	for i := 0; i < numWriters; i++ {
			fw, err := NewFileWriter(testFile, 100, 5)
		if err != nil {
			t.Fatalf("Failed to create FileWriter %d: %v", i, err)
		}
		writers = append(writers, fw)
	}

	// Start concurrent writes
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int, fw *FileWriter) {
			defer wg.Done()
			for j := 0; j < messagesPerWriter; j++ {
				testData := []byte(`{"level":"info","time":"2024-01-01T12:00:00Z","message":"Writer ` +
					string(rune('0'+writerID)) + ` message ` + string(rune('0'+j)) + `"}`)
				fw.Write(testData)
				time.Sleep(10 * time.Millisecond) // Small delay between writes
			}
		}(i, writers[i])
	}

	// Wait for all writes to complete
	wg.Wait()

	// Close all writers
	for _, fw := range writers {
		fw.Close()
	}

	// Wait a bit more for any pending writes
	time.Sleep(200 * time.Millisecond)

	// Check what files were created
	files, err := filepath.Glob(filepath.Join(tempDir, "*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	t.Logf("Created files: %v", files)

	// Verify at least one file was created
	if len(files) == 0 {
		t.Fatalf("No log files were created")
	}

	// Count total messages written
	totalMessages := 0
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read file %s: %v", file, err)
		}
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			totalMessages += len(lines)
		}
	}

	expectedMessages := numWriters * messagesPerWriter
	t.Logf("Expected %d messages, found %d messages", expectedMessages, totalMessages)

	// We should have all messages written (some might be in numbered files)
	if totalMessages < expectedMessages {
		t.Fatalf("Expected at least %d messages, but found %d", expectedMessages, totalMessages)
	}
}
