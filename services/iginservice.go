package services

import (
	"github.com/ternarybob/arbor/models"
)

// IGinService defines the interface for Gin log formatting service
type IGinService interface {
	IsGinLog(logContent string) bool
	FormatGinLog(logContent string) (string, error)
	ParseGinLog(p []byte) *models.GinLogEvent
	// ShouldLogLevel(level string) bool
	// FormatConsoleOutput(logEntry *models.GinLogEvent) string
	// ToJSON(logEntry *models.GinLogEvent) ([]byte, error)
}
