package writers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/phuslu/log"
)

// GinLogEvent represents a structured log event from Gin framework
type GinLogEvent struct {
	Level         string    `json:"level"`
	Timestamp     time.Time `json:"time"`
	Prefix        string    `json:"prefix"`
	CorrelationID string    `json:"correlationid"`
	Message       string    `json:"message"`
	Error         string    `json:"error"`
}

// GinLogDetector handles detection and processing of Gin framework logs
type GinLogDetector struct {
	level log.Level
}

// NewGinDetector creates a new GinLogDetector with the specified log level
func NewGinDetector(level log.Level) *GinLogDetector {
	return &GinLogDetector{level: level}
}

// IsGinLog detects if the log content is from Gin framework
func (g *GinLogDetector) IsGinLog(content string) bool {
	lower := strings.ToLower(content)

	// Primary detection: Gin prefixes
	if strings.Contains(lower, "[gin") {
		return true
	}

	// Secondary detection: HTTP request pattern
	// Look for pattern like: | 200 | 1.234ms | 127.0.0.1 | GET
	if strings.Contains(content, " | ") &&
		(strings.Contains(content, "GET") ||
			strings.Contains(content, "POST") ||
			strings.Contains(content, "PUT") ||
			strings.Contains(content, "DELETE") ||
			strings.Contains(content, "PATCH") ||
			strings.Contains(content, "HEAD") ||
			strings.Contains(content, "OPTIONS")) {
		return true
	}

	return false
}

// ParseGinLog parses a gin log entry and returns a structured log event
func (g *GinLogDetector) ParseGinLog(p []byte) *GinLogEvent {
	logentry := &GinLogEvent{
		Prefix:    "GIN",
		Timestamp: time.Now(),
		Message:   strings.TrimSuffix(string(p), "\n"),
	}

	logstring := string(p)

	// Parse Gin log level
	switch {
	case strings.Contains(strings.ToLower(logstring), "gin-fatal"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-fatal] ", "")
		logentry.Level = "fatal"
	case strings.Contains(strings.ToLower(logstring), "gin-error"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-error] ", "")
		logentry.Level = "error"
	case strings.Contains(strings.ToLower(logstring), "gin-warning"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-warning] ", "")
		logentry.Level = "warn"
	case strings.Contains(strings.ToLower(logstring), "gin-information"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-information] ", "")
		logentry.Level = "info"
	case strings.Contains(strings.ToLower(logstring), "gin-debug"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-debug] ", "")
		logentry.Level = "debug"
	default:
		// Default to info level for standard Gin logs
		logentry.Level = "info"
	}

	return logentry
}

// ShouldLogLevel checks if the log level should be written based on configured level
func (g *GinLogDetector) ShouldLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "fatal":
		return g.level <= log.FatalLevel
	case "error":
		return g.level <= log.ErrorLevel
	case "warn":
		return g.level <= log.WarnLevel
	case "info":
		return g.level <= log.InfoLevel
	case "debug":
		return g.level <= log.DebugLevel
	default:
		return true
	}
}

// FormatConsoleOutput formats Gin log entries to match ConsoleWriter format
func (g *GinLogDetector) FormatConsoleOutput(l *GinLogEvent) string {
	timestamp := l.Timestamp.Format("15:04:05.000")

	output := fmt.Sprintf("%s|%s", levelprint(l.Level, true), timestamp)

	if l.Prefix != "" {
		output += fmt.Sprintf("|%s", l.Prefix)
	}

	if l.Message != "" {
		output += fmt.Sprintf("|%s", l.Message)
	}

	if l.Error != "" {
		output += fmt.Sprintf("|%s", l.Error)
	}

	return output
}

// ToJSON converts the GinLogEvent to JSON bytes for file writers
func (g *GinLogDetector) ToJSON(event *GinLogEvent) ([]byte, error) {
	return json.Marshal(event)
}
