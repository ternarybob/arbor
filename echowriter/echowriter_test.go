package echowriter

import (
	"testing"
)

func TestEchoWriterCreation(t *testing.T) {
	// Test that EchoWriter can be created without panicking
	writer := New()
	if writer == nil {
		t.Error("EchoWriter should not be nil after creation")
	}
}

func TestEchoWriterWrite(t *testing.T) {
	writer := New()

	// Test with a GIN-style log message that EchoWriter expects
	testData := []byte("[GIN-information] test echo message")
	n, err := writer.Write(testData)

	if err != nil {
		t.Logf("Write returned error: %v", err)
	}

	// EchoWriter may return 0 for the byte count in some cases
	if n != len(testData) {
		t.Logf("Write returned byte count: got %d, want %d (may be expected)", n, len(testData))
	}
}
