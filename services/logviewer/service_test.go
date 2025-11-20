package logviewer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"testing"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

func TestGetLogContent(t *testing.T) {
	// Create temp directory for logs
	tempDir := t.TempDir()

	// 1. Create a JSON log file
	jsonLogPath := filepath.Join(tempDir, "test.json.log")
	jsonFile, err := os.Create(jsonLogPath)
	if err != nil {
		t.Fatalf("Failed to create json log file: %v", err)
	}

	jsonEntry := models.LogEvent{
		Timestamp: time.Now(),
		Level:     log.InfoLevel,
		Message:   "JSON Message",
		Fields:    map[string]interface{}{"foo": "bar"},
	}
	jsonBytes, _ := json.Marshal(jsonEntry)
	jsonFile.Write(jsonBytes)
	jsonFile.WriteString("\n")
	jsonFile.Close()

	// 2. Create a Text log file
	textLogPath := filepath.Join(tempDir, "test.text.log")
	textFile, err := os.Create(textLogPath)
	if err != nil {
		t.Fatalf("Failed to create text log file: %v", err)
	}

	// Format: TIMESTAMP LEVEL > MESSAGE
	timestamp := time.Now().Format(time.RFC3339)
	textLine := timestamp + " INF > Text Message\n"
	textFile.WriteString(textLine)
	textFile.Close()

	// Test Service
	service := NewService(models.WriterConfiguration{
		FileName:   jsonLogPath, // We need to point to the dir, but NewService takes config with FileName
		TextOutput: models.TextOutputFormatJSON,
	})
	// But wait, NewService derives LogDirectory from FileName.
	// If we pass jsonLogPath, LogDirectory will be tempDir.
	// And GetLogContent takes a filename relative to LogDirectory.
	// So if we pass jsonLogPath (full path), LogDirectory is tempDir.
	// Then GetLogContent("test.json.log") will join tempDir + test.json.log -> correct.

	// However, we want to test both JSON and Text files in the same directory.
	// NewService takes one config.
	// If we initialize with jsonLogPath, LogDirectory is set.
	// Then we can read both files if they are in the same dir.

	// Let's use a dummy filename in the same dir to init service
	service = NewService(models.WriterConfiguration{
		FileName:   filepath.Join(tempDir, "dummy.log"),
		TextOutput: models.TextOutputFormatJSON, // Force JSON for this test setup, though GetLogContent auto-detects
	})

	// Test JSON reading
	entries, err := service.GetLogContent("test.json.log", 0, nil)
	if err != nil {
		t.Errorf("Failed to read JSON log: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Message != "JSON Message" {
		t.Errorf("Expected message 'JSON Message', got '%s'", entries[0].Message)
	}

	// Test Text reading
	entries, err = service.GetLogContent("test.text.log", 0, nil)
	if err != nil {
		t.Errorf("Failed to read Text log: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Message != "Text Message" {
		t.Errorf("Expected message 'Text Message', got '%s'", entries[0].Message)
	}
	if entries[0].Level != log.InfoLevel {
		t.Errorf("Expected level Info, got %v", entries[0].Level)
	}
}

func TestParseTextLog(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantMsg string
		wantLvl log.Level
		wantErr bool
	}{
		{
			name:    "Valid Info",
			line:    "2025-11-19T22:43:44+11:00 INF > Message",
			wantMsg: "Message",
			wantLvl: log.InfoLevel,
			wantErr: false,
		},
		{
			name:    "Valid Error with fields",
			line:    "2025-11-19T22:43:44+11:00 ERR > key=val Error occurred",
			wantMsg: "key=val Error occurred",
			wantLvl: log.ErrorLevel,
			wantErr: false,
		},
		{
			name:    "Pipe Delimited Info",
			line:    "2025-11-19T22:43:44+11:00 | INF | Message",
			wantMsg: "Message",
			wantLvl: log.InfoLevel,
			wantErr: false,
		},
		{
			name:    "Pipe Delimited Error with fields",
			line:    "2025-11-19T22:43:44+11:00 | ERR | Message | key=val",
			wantMsg: "Message",
			wantLvl: log.ErrorLevel,
			wantErr: false,
			// Note: Fields are not checked in this test helper, but we check message.
			// In pipe format, fields are parsed into Fields map, not part of Message.
			// So wantMsg should be "Message".
		},
		{
			name:    "Pipe Delimited with multiple fields",
			line:    "2025-11-19T22:43:44+11:00 | INF | Message | key1=val1 key2=val2",
			wantMsg: "Message",
			wantLvl: log.InfoLevel,
			wantErr: false,
		},
		{
			name:    "Logfmt Info",
			line:    "time=2025-11-19T22:43:44+11:00 level=INF message=\"Message\" foo=bar",
			wantMsg: "Message",
			wantLvl: log.InfoLevel,
			wantErr: false,
		},
		{
			name:    "Logfmt Error",
			line:    "time=2025-11-19T22:43:44+11:00 level=ERR message=\"Error occurred\"",
			wantMsg: "Error occurred",
			wantLvl: log.ErrorLevel,
			wantErr: false,
		},
		{
			name:    "Invalid Separator",
			line:    "2025-11-19T22:43:44+11:00 INF Message",
			wantErr: true,
		},
		{
			name:    "Short Line",
			line:    "Short",
			wantErr: true, // parseTextLog checks length implicitly via separator search or explicit check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTextLog(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTextLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Message != tt.wantMsg {
					t.Errorf("parseTextLog() message = %v, want %v", got.Message, tt.wantMsg)
				}
				if got.Level != tt.wantLvl {
					t.Errorf("parseTextLog() level = %v, want %v", got.Level, tt.wantLvl)
				}
				// Check Time string
				if got.Time == "" {
					t.Errorf("parseTextLog() time is empty")
				}

				if strings.HasPrefix(tt.line, "time=") {
					// For logfmt, expect time field value after "time="
					after := strings.TrimPrefix(tt.line, "time=")
					if idx := strings.IndexByte(after, ' '); idx >= 0 {
						after = after[:idx]
					}
					if got.Time != after {
						t.Errorf("parseTextLog() time = %v, want %v", got.Time, after)
					}
				} else {
					// For pipe and legacy formats, time should match the start of the line
					if !strings.HasPrefix(tt.line, got.Time) {
						t.Errorf("parseTextLog() time %v does not match start of line %v", got.Time, tt.line)
					}
				}
			}
		})
	}
}
