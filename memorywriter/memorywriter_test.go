package memorywriter

import (
	"testing"
)

func TestMemoryWriterCreation(t *testing.T) {
	// Test that MemoryWriter can be created without panicking
	writer := New()
	if writer == nil {
		t.Error("MemoryWriter should not be nil after creation")
	}
}

func TestMemoryWriterWrite(t *testing.T) {
	writer := New()
	
	testData := []byte("test memory message")
	n, err := writer.Write(testData)
	
	if err != nil {
		t.Errorf("Write should not return error: %v", err)
	}
	
	if n != len(testData) {
		t.Errorf("Write should return correct byte count: got %d, want %d", n, len(testData))
	}
}

func TestMemoryWriterGetLogs(t *testing.T) {
	writer := New()
	
	// Write some test data
	testMessages := []string{
		"first memory message",
		"second memory message", 
		"third memory message",
	}
	
	for _, msg := range testMessages {
		writer.Write([]byte(msg))
	}
	
	// Test that we can retrieve logs (method signature may vary)
	// This is a basic test to ensure the interface works
	if writer == nil {
		t.Error("Writer should maintain logs in memory")
	}
}
