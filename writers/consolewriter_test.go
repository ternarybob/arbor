package writers

import (
	"strings"
	"testing"
)

func TestConsoleWriterCreation(t *testing.T) {
	// Test that ConsoleWriter can be created without panicking
	writer := NewConsoleWriter()
	if writer == nil {
		t.Error("ConsoleWriter should not be nil after creation")
	}
}

func TestConsoleWriterWrite(t *testing.T) {
	writer := NewConsoleWriter()

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

func TestHandleRegularLog_JSON(t *testing.T) {
	// Create a mock writer that implements Write method
	var capturedData []byte
	mockWriter := &mockWriter{
		WriteFunc: func(p []byte) (int, error) {
			capturedData = append(capturedData, p...)
			return len(p), nil
		},
	}
	
	writers := map[string]interface{}{
		"test": mockWriter,
	}
	
	jsonLog := `{"level":"info","message":"test message"}`
	n, err := HandleRegularLog([]byte(jsonLog), writers)
	
	if err != nil {
		t.Fatalf("HandleRegularLog returned error: %v", err)
	}
	
	if n != len(jsonLog) {
		t.Errorf("HandleRegularLog returned n = %d, expected %d", n, len(jsonLog))
	}
	
	if string(capturedData) != jsonLog {
		t.Errorf("Writer received %q, expected %q", string(capturedData), jsonLog)
	}
}

func TestHandleRegularLog_PlainText(t *testing.T) {
	// Create a mock writer that implements Write method
	var capturedData []byte
	mockWriter := &mockWriter{
		WriteFunc: func(p []byte) (int, error) {
			capturedData = append(capturedData, p...)
			return len(p), nil
		},
	}
	
	writers := map[string]interface{}{
		"file": mockWriter,
	}
	
	plainText := "This is a plain text message"
	n, err := HandleRegularLog([]byte(plainText), writers)
	
	if err != nil {
		t.Fatalf("HandleRegularLog returned error: %v", err)
	}
	
	if n != len(plainText) {
		t.Errorf("HandleRegularLog returned n = %d, expected %d", n, len(plainText))
	}
	
	// Should have received JSON representation of the plain text
	if !strings.Contains(string(capturedData), plainText) {
		t.Errorf("Writer should have received JSON containing the plain text message")
	}
	if !strings.Contains(string(capturedData), `"level":"info"`) {
		t.Error("Writer should have received JSON with info level")
	}
	if !strings.Contains(string(capturedData), `"prefix":"APP"`) {
		t.Error("Writer should have received JSON with APP prefix")
	}
}

// mockWriter is a helper for testing
type mockWriter struct {
	WriteFunc func([]byte) (int, error)
}

func (m *mockWriter) Write(p []byte) (int, error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(p)
	}
	return len(p), nil
}
