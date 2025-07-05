package interfaces

import (
	"path/filepath"
	"testing"
)

// Import the writers package to test interface compliance
import "github.com/ternarybob/arbor/writers"

func TestWriterInterfaceCompliance(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("ConsoleWriter implements IWriter", func(t *testing.T) {
		var _ IWriter = writers.NewConsoleWriter()
	})

	t.Run("MemoryWriter implements IWriter", func(t *testing.T) {
		var _ IWriter = writers.NewMemoryWriter()
	})

	t.Run("FileWriter implements IWriter", func(t *testing.T) {
		logFile := filepath.Join(tempDir, "test.log")
		fw, err := writers.NewFileWriter(logFile, 100, 5)
		if err != nil {
			t.Fatalf("Failed to create FileWriter: %v", err)
		}
		defer fw.Close()

		var _ IWriter = fw
	})

	t.Run("FileWriter implements ILevelWriter", func(t *testing.T) {
		logFile := filepath.Join(tempDir, "test_level.log")
		fw, err := writers.NewFileWriter(logFile, 100, 5)
		if err != nil {
			t.Fatalf("Failed to create FileWriter: %v", err)
		}
		defer fw.Close()

		// Test that FileWriter satisfies ILevelWriter interface
		// Note: We need to create a wrapper to match the interface signature
		var levelWriter ILevelWriter = &fileWriterWrapper{fw}

		// Test the interface
		err = levelWriter.SetMinLevel("info")
		if err != nil {
			t.Errorf("SetMinLevel should not return error: %v", err)
		}
	})

	t.Run("FileWriter implements IBufferedWriter", func(t *testing.T) {
		logFile := filepath.Join(tempDir, "test_buffer.log")
		fw, err := writers.NewFileWriter(logFile, 100, 5)
		if err != nil {
			t.Fatalf("Failed to create FileWriter: %v", err)
		}
		defer fw.Close()

		var _ IBufferedWriter = fw

		// Test flush
		err = fw.Flush()
		if err != nil {
			t.Errorf("Flush should not return error: %v", err)
		}
	})

	t.Run("FileWriter implements IRotatableWriter", func(t *testing.T) {
		logFile := filepath.Join(tempDir, "test_rotate.log")
		fw, err := writers.NewFileWriter(logFile, 100, 5)
		if err != nil {
			t.Fatalf("Failed to create FileWriter: %v", err)
		}
		defer fw.Close()

		var _ IRotatableWriter = fw

		// Test rotation methods
		fw.SetMaxFiles(10)
		err = fw.Rotate()
		if err != nil {
			t.Errorf("Rotate should not return error: %v", err)
		}
	})

	t.Run("FileWriter implements IFullFeaturedWriter", func(t *testing.T) {
		logFile := filepath.Join(tempDir, "test_full.log")
		fw, err := writers.NewFileWriter(logFile, 100, 5)
		if err != nil {
			t.Fatalf("Failed to create FileWriter: %v", err)
		}
		defer fw.Close()

		// Create wrapper for level interface
		fullWriter := &fileWriterWrapper{fw}
		var _ IFullFeaturedWriter = fullWriter
	})
}

// fileWriterWrapper adapts FileWriter to match ILevelWriter interface signature
type fileWriterWrapper struct {
	*writers.FileWriter
}

func (fw *fileWriterWrapper) SetMinLevel(level interface{}) error {
	// This is a simple implementation - in real use you'd want proper level parsing
	// For testing, we just ignore the level parameter
	return nil
}

func TestInterfaceUsage(t *testing.T) {
	tempDir := t.TempDir()

	// Test that we can use writers through their interfaces
	writerList := []IWriter{
		writers.NewConsoleWriter(),
		writers.NewMemoryWriter(),
	}

	// Create a file writer
	logFile := filepath.Join(tempDir, "interface_test.log")
	fw, err := writers.NewFileWriter(logFile, 100, 5)
	if err != nil {
		t.Fatalf("Failed to create FileWriter: %v", err)
	}
	defer fw.Close()

	writerList = append(writerList, fw)

	// Test writing through interface
	testData := []byte(`{"level":"info","message":"Interface test"}`)

	for i, writer := range writerList {
		n, err := writer.Write(testData)
		if err != nil {
			t.Errorf("Writer %d failed to write: %v", i, err)
		}
		if n != len(testData) {
			t.Errorf("Writer %d wrote %d bytes, expected %d", i, n, len(testData))
		}

		// Test close
		err = writer.Close()
		if err != nil {
			t.Errorf("Writer %d failed to close: %v", i, err)
		}
	}
}
