package arbor

import (
	"testing"

	"github.com/ternarybob/arbor/models"
	"github.com/ternarybob/arbor/writers"
)

// init ensures the default logger is created and registered during package initialization
func init() {
	// This will create and register the default logger
	Logger()
}

func TestGlobalWriterRegistry(t *testing.T) {
	// Test registering a file writer
	fileConfig := models.WriterConfiguration{
		Type:       models.LogWriterTypeFile,
		TimeFormat: "01-02 15:04:05.000",
		FileName:   "test.log",
	}

	fileWriter := writers.FileWriter(fileConfig)
	RegisterWriter("test-file", fileWriter)

	// Test retrieving the writer
	retrievedWriter := GetRegisteredWriter("test-file")
	if retrievedWriter == nil {
		t.Error("Test file writer should be retrievable")
	}

	// Test writer count
	count := GetWriterCount()
	if count < 1 {
		t.Errorf("Expected at least 1 writer, got %d", count)
	}

	// Test writer names
	names := GetRegisteredWriterNames()
	if len(names) < 1 {
		t.Errorf("Expected at least 1 writer name, got %d", len(names))
	}

	// Test unregistering
	UnregisterWriter("test-file")
	retrievedWriter = GetRegisteredWriter("test-file")
	if retrievedWriter != nil {
		t.Error("Test file writer should be unregistered")
	}
}

func TestMemoryWriterRegistry(t *testing.T) {
	// Test registering a memory writer
	memoryConfig := models.WriterConfiguration{
		Type:       models.LogWriterTypeMemory,
		TimeFormat: "01-02 15:04:05.000",
	}

	memoryWriter := writers.MemoryWriter(memoryConfig)
	RegisterWriter("test-memory", memoryWriter)

	// Test retrieving the memory writer
	retrievedMemoryWriter := GetRegisteredMemoryWriter("test-memory")
	if retrievedMemoryWriter == nil {
		t.Error("Test memory writer should be retrievable")
	}

	// Test that regular GetRegisteredWriter also works
	retrievedWriter := GetRegisteredWriter("test-memory")
	if retrievedWriter == nil {
		t.Error("Test memory writer should be retrievable as IWriter")
	}

	// Clean up
	UnregisterWriter("test-memory")
}

func TestLoggerWithRegisteredWriters(t *testing.T) {
	// Create a logger and register a memory writer
	logger := Logger()
	
	memoryConfig := models.WriterConfiguration{
		Type:       models.LogWriterTypeMemory,
		TimeFormat: "01-02 15:04:05.000",
	}

	// This should register the memory writer
	logger.WithMemoryWriter(memoryConfig).WithCorrelationId("test-correlation")

	// Test logging
	logger.Info().Msg("Test message")

	// Verify the message was written to the registered memory writer
	logs, err := logger.GetMemoryLogs("test-correlation", DebugLevel)
	if err != nil {
		t.Errorf("Error retrieving logs: %v", err)
	}

	if len(logs) == 0 {
		t.Error("Expected test message to be logged to memory writer")
	}

	// Clean up
	UnregisterWriter(WRITER_MEMORY)
}
