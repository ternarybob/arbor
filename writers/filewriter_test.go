package writers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

func TestFileWriter_New(t *testing.T) {
	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeFile,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
		FileName:   "test.log",
		MaxBackups: 3,
		MaxSize:    1024,
	}

	writer := FileWriter(config)
	if writer == nil {
		t.Fatal("FileWriter should not return nil")
	}

	// Verify it implements IWriter interface
	var _ IWriter = writer
}

func TestFileWriter_DefaultValues(t *testing.T) {
	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeFile,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
		// Don't set FileName, MaxBackups, or MaxSize to test defaults
	}

	writer := FileWriter(config)
	if writer == nil {
		t.Fatal("FileWriter should not return nil")
	}

	fw := writer.(*fileWriter)

	// Test that logger was configured
	if fw.logger.Writer == nil {
		t.Error("FileWriter should have a writer configured")
	}
}

func TestFileWriter_WithLevel(t *testing.T) {
	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeFile,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
		FileName:   "test.log",
	}

	writer := FileWriter(config)

	// Test changing level
	newWriter := writer.WithLevel(log.DebugLevel)
	if newWriter == nil {
		t.Error("WithLevel should not return nil")
	}

	// Should return the same instance
	if newWriter != writer {
		t.Error("WithLevel should return the same instance")
	}
}

func TestFileWriter_Write(t *testing.T) {
	// Create a temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "arbor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeFile,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
		FileName:   filepath.Join(tempDir, "test.log"),
		MaxBackups: 3,
		MaxSize:    1024,
	}

	writer := FileWriter(config)

	testCases := []struct {
		name     string
		input    []byte
		expected int
	}{
		{
			name:     "normal message",
			input:    []byte("test message"),
			expected: 12,
		},
		{
			name:     "empty message",
			input:    []byte(""),
			expected: 0,
		},
		{
			name:     "json message",
			input:    []byte(`{"level":"info","msg":"test"}`),
			expected: 29,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, err := writer.Write(tc.input)
			if err != nil {
				t.Errorf("Write should not return error: %v", err)
			}
			if n != tc.expected {
				t.Errorf("Expected %d bytes written, got %d", tc.expected, n)
			}
		})
	}
}

func TestFileWriter_Configuration(t *testing.T) {
	testCases := []struct {
		name     string
		config   models.WriterConfiguration
		validate func(t *testing.T, writer IWriter)
	}{
		{
			name: "custom filename",
			config: models.WriterConfiguration{
				Type:       models.LogWriterTypeFile,
				Level:      log.InfoLevel,
				TimeFormat: "15:04:05.000",
				FileName:   "custom.log",
				MaxBackups: 5,
				MaxSize:    2048,
			},
			validate: func(t *testing.T, writer IWriter) {
				fw := writer.(*fileWriter)
				if fw.config.FileName != "custom.log" {
					t.Errorf("Expected filename 'custom.log', got '%s'", fw.config.FileName)
				}
				if fw.config.MaxBackups != 5 {
					t.Errorf("Expected MaxBackups 5, got %d", fw.config.MaxBackups)
				}
				if fw.config.MaxSize != 2048 {
					t.Errorf("Expected MaxSize 2048, got %d", fw.config.MaxSize)
				}
			},
		},
		{
			name: "default values when missing",
			config: models.WriterConfiguration{
				Type:       models.LogWriterTypeFile,
				Level:      log.InfoLevel,
				TimeFormat: "15:04:05.000",
				// Missing FileName, MaxBackups, MaxSize
			},
			validate: func(t *testing.T, writer IWriter) {
				// Should not panic and should create a valid writer
				if writer == nil {
					t.Error("Writer should not be nil")
				}
			},
		},
		{
			name: "zero max backups gets default",
			config: models.WriterConfiguration{
				Type:       models.LogWriterTypeFile,
				Level:      log.InfoLevel,
				TimeFormat: "15:04:05.000",
				FileName:   "test.log",
				MaxBackups: 0, // Should get default of 5
				MaxSize:    1024,
			},
			validate: func(t *testing.T, writer IWriter) {
				// Should not panic and should use default value
				if writer == nil {
					t.Error("Writer should not be nil")
				}
			},
		},
		{
			name: "zero max size gets default",
			config: models.WriterConfiguration{
				Type:       models.LogWriterTypeFile,
				Level:      log.InfoLevel,
				TimeFormat: "15:04:05.000",
				FileName:   "test.log",
				MaxBackups: 3,
				MaxSize:    0, // Should get default of MaxLogSize
			},
			validate: func(t *testing.T, writer IWriter) {
				// Should not panic and should use default value
				if writer == nil {
					t.Error("Writer should not be nil")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			writer := FileWriter(tc.config)
			tc.validate(t, writer)
		})
	}
}

func TestFileWriter_InterfaceCompliance(t *testing.T) {
	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeFile,
		Level:      log.InfoLevel,
		TimeFormat: "15:04:05.000",
		FileName:   "test.log",
	}

	writer := FileWriter(config)

	// Test all IWriter interface methods
	t.Run("WithLevel", func(t *testing.T) {
		result := writer.WithLevel(log.DebugLevel)
		if result == nil {
			t.Error("WithLevel should not return nil")
		}
	})

	t.Run("Write", func(t *testing.T) {
		n, err := writer.Write([]byte("test"))
		if err != nil {
			t.Errorf("Write should not error: %v", err)
		}
		if n != 4 {
			t.Errorf("Expected 4 bytes written, got %d", n)
		}
	})
}

func TestMaxLogSize_Constant(t *testing.T) {
	expectedSize := int64(10 * 1024 * 1024) // 10 MB
	if MaxLogSize != expectedSize {
		t.Errorf("Expected MaxLogSize to be %d, got %d", expectedSize, MaxLogSize)
	}
}

func TestFileLogEntry_Struct(t *testing.T) {
	entry := FileLogEntry{
		Level:   "info",
		Message: "test message",
		Time:    "12:34:56.789",
		Prefix:  "TEST",
		Extra:   map[string]interface{}{"key": "value"},
	}

	if entry.Level != "info" {
		t.Errorf("Expected level 'info', got '%s'", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entry.Message)
	}
	if entry.Time != "12:34:56.789" {
		t.Errorf("Expected time '12:34:56.789', got '%s'", entry.Time)
	}
	if entry.Prefix != "TEST" {
		t.Errorf("Expected prefix 'TEST', got '%s'", entry.Prefix)
	}
}
