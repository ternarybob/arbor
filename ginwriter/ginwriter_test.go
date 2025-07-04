package ginwriter

import (
	"testing"
)

func TestGinWriterCreation(t *testing.T) {
	// Test that GinWriter can be created without panicking
	writer := New()
	if writer == nil {
		t.Error("GinWriter should not be nil after creation")
	}
}

func TestGinWriterWrite(t *testing.T) {
	writer := New()

	// Test with typical Gin log message format
	testData := []byte("[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.\n")
	_, err := writer.Write(testData)

	if err != nil {
		t.Errorf("Write returned unexpected error: %v", err)
	}
}
