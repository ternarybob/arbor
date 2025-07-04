package arbor

import (
	"bytes"
	"strings"
	"testing"
	"github.com/ternarybob/arbor/ginwriter"
)

func TestGinWriterTimestampFormat(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	
	// Create a GIN writer that writes to our buffer
	ginWriter := &ginwriter.GinWriter{
		Out:   &buf,
		Level: 5, // Debug level to catch all messages
	}
	
	// Test various GIN log levels
	testMessages := []string{
		"[GIN-debug] This is a debug message\n",
		"[GIN-information] This is an info message\n", 
		"[GIN-warning] This is a warning message\n",
		"[GIN-error] This is an error message\n",
		"[GIN-fatal] This is a fatal message\n",
	}
	
	for _, msg := range testMessages {
		buf.Reset() // Clear buffer before each test
		
		_, err := ginWriter.Write([]byte(msg))
		if err != nil {
			t.Errorf("Error writing GIN message: %v", err)
			continue
		}
		
		output := buf.String()
		if output == "" {
			continue // Skip if no output (level filtering)
		}
		
		t.Logf("GIN Writer Output: %s", strings.TrimSpace(output))
		
		// Check that the timestamp format is HH:MM:SS.sss (no date, no timezone)
		// The output should contain a timestamp in format like "12:08:11.123"
		if !strings.Contains(output, ":") {
			t.Errorf("Expected timestamp with colons in output: %s", output)
		}
		
		// Check that it doesn't contain the old verbose timestamp format
		if strings.Contains(output, "2025") || strings.Contains(output, "+") {
			t.Errorf("Output contains verbose timestamp format, expected short format: %s", output)
		}
		
		// The format should be roughly: LEVEL|HH:MM:SS.sss|PREFIX|MESSAGE
		parts := strings.Split(output, "|")
		if len(parts) < 3 {
			t.Errorf("Expected at least 3 parts separated by |, got %d: %s", len(parts), output)
		}
	}
}
