package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ternarybob/arbor/models"
)

// GinService provides Gin log formatting functionality
type GinService struct {
}

// NewGinService creates a new Gin log formatting service with the specified minimum log level
func NewGinService() IGinService {
	return &GinService{}
}

// IsGinLog detects if the log message is from Gin framework
func (g *GinService) IsGinLog(logContent string) bool {
	// Check for Gin-specific patterns
	ginPatterns := []string{
		`\[GIN\]`,             // Standard Gin logs
		`\[GIN-debug\]`,       // Gin debug logs
		`\|\s*\d+\s*\|\s*\d+`, // HTTP status code pattern: | 200 | 123ms |
	}

	for _, pattern := range ginPatterns {
		matched, _ := regexp.MatchString(pattern, logContent)
		if matched {
			return true
		}
	}

	return false
}

// FormatGinLog formats a Gin log entry to JSON
func (g *GinService) FormatGinLog(logContent string) (string, error) {
	logEvent := g.ParseGinLog([]byte(logContent))
	jsonData, err := g.ToJSON(logEvent)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// ParseGinLog parses a Gin log entry and extracts relevant information
func (g *GinService) ParseGinLog(p []byte) *models.GinLogEvent {
	logContent := strings.TrimSpace(string(p))

	// Default values
	logEntry := &models.GinLogEvent{
		Level:     "info",
		Timestamp: time.Now(),
		Prefix:    "GIN",
		Message:   logContent,
	}

	// Check for different Gin log levels based on content
	lowerContent := strings.ToLower(logContent)
	if strings.Contains(lowerContent, "fatal") {
		logEntry.Level = "fatal"
	} else if strings.Contains(lowerContent, "error") {
		logEntry.Level = "error"
	} else if strings.Contains(lowerContent, "warning") || strings.Contains(lowerContent, "warn") {
		logEntry.Level = "warn"
	} else if strings.Contains(lowerContent, "debug") {
		logEntry.Level = "debug"
	}

	return logEntry
}

// FormatConsoleOutput formats a Gin log entry for console display
func (g *GinService) FormatConsoleOutput(logEntry *models.GinLogEvent) string {
	// Get level print format
	levelStr := g.getLevelPrint(logEntry.Level)

	// Format timestamp
	timeStr := logEntry.Timestamp.Format("15:04:05.000")

	// Build formatted output
	return fmt.Sprintf("%s|%s|%s|%s",
		levelStr,
		timeStr,
		logEntry.Prefix,
		logEntry.Message)
}

// ToJSON converts a Gin log entry to JSON format
func (g *GinService) ToJSON(logEntry *models.GinLogEvent) ([]byte, error) {
	return json.Marshal(logEntry)
}

// getLevelPrint returns the formatted level string for console output
func (g *GinService) getLevelPrint(level string) string {
	switch strings.ToLower(level) {
	case "fatal":
		return "FTL"
	case "error":
		return "ERR"
	case "warn", "warning":
		return "WRN"
	case "info":
		return "INF"
	case "debug":
		return "DBG"
	case "trace":
		return "TRC"
	default:
		return "INF"
	}
}
