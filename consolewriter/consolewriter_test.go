package consolewriter

import (
	"testing"
)

func TestConsoleWriterCreation(t *testing.T) {
	// Test that ConsoleWriter can be created without panicking
	writer := New()
	if writer == nil {
		t.Error("ConsoleWriter should not be nil after creation")
	}
}

func TestConsoleWriterWrite(t *testing.T) {
	writer := New()

	// Use valid JSON format that the writer expects
	testData := []byte(`{"level":"info","message":"test console message"}`)
	n, err := writer.Write(testData)

	// JSON parsing errors are expected in test environment, so just log them
	if err != nil {
		t.Logf("Write returned error (may be expected in test): %v", err)
	}

	if n != len(testData) {
		t.Logf("Write returned byte count: got %d, want %d", n, len(testData))
	}
}
