package transformers

import (
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/models"
	"github.com/ternarybob/arbor/writers"
)

// ginTransformer implements io.Writer for Gin framework integration
// It captures Gin logs, transforms them to arbor format, and outputs to registered writers
type ginTransformer struct {
	config         models.WriterConfiguration
	correlationID  string
	correlationMux sync.RWMutex
	getWriters     func() map[string]writers.IWriter // Function to get registered writers
}

// NewGinTransformer creates a new Gin transformer that integrates with arbor's logging infrastructure
func NewGinTransformer(config models.WriterConfiguration, getWriters func() map[string]writers.IWriter) io.Writer {
	return &ginTransformer{
		config:     config,
		getWriters: getWriters,
	}
}

// Write implements io.Writer interface for Gin integration
func (gt *ginTransformer) Write(p []byte) (n int, err error) {
	internalLog := common.NewLogger().WithContext("function", "GinTransformer.Write").GetLogger()

	if len(p) == 0 {
		return 0, nil
	}

	logContent := strings.TrimSpace(string(p))
	if logContent == "" {
		return len(p), nil
	}

	// Parse the Gin log and convert to arbor LogEvent
	logEvent := gt.parseGinLog(logContent)
	if logEvent == nil {
		internalLog.Debug().Msgf("Failed to parse Gin log: %s", logContent)
		return len(p), nil
	}

	// Check if we should log this level
	if !gt.shouldLogLevel(logEvent.Level) {
		return len(p), nil
	}

	// Transform to arbor format and send to all registered writers
	gt.outputToRegisteredWriters(logEvent)

	internalLog.Trace().Msgf("Processed Gin log: %s", logContent)
	return len(p), nil
}

// parseGinLog parses a Gin log entry and converts it to arbor LogEvent format
func (gt *ginTransformer) parseGinLog(logContent string) *models.LogEvent {
	internalLog := common.NewLogger().WithContext("function", "GinTransformer.parseGinLog").GetLogger()

	// Extract correlation ID from current context if available
	correlationID := gt.extractCorrelationID()

	// Determine log level from content
	level := gt.determineLogLevel(logContent)

	// Create log event in arbor format
	logEvent := &models.LogEvent{
		Level:         level,
		Timestamp:     time.Now(),
		Prefix:        "gin",
		Message:       logContent,
		CorrelationID: correlationID,
		Function:      gt.extractFunction(logContent),
	}

	internalLog.Trace().Msgf("Parsed Gin log - Level: %s, CorrelationID: %s",
		level.String(), correlationID)

	return logEvent
}

// extractCorrelationID attempts to extract correlation ID
func (gt *ginTransformer) extractCorrelationID() string {
	gt.correlationMux.RLock()
	defer gt.correlationMux.RUnlock()

	// Return the stored correlation ID (will be set by SetCorrelationID method)
	return gt.correlationID
}

// determineLogLevel determines the log level based on Gin log content
func (gt *ginTransformer) determineLogLevel(logContent string) log.Level {
	lowerContent := strings.ToLower(logContent)

	// Check for specific patterns to determine level
	if strings.Contains(lowerContent, "fatal") || strings.Contains(lowerContent, "panic") {
		return log.FatalLevel
	}
	if strings.Contains(lowerContent, "error") {
		return log.ErrorLevel
	}
	if strings.Contains(lowerContent, "warning") || strings.Contains(lowerContent, "warn") {
		return log.WarnLevel
	}
	if strings.Contains(lowerContent, "[gin]") {
		return log.DebugLevel
	}
	if strings.Contains(lowerContent, "[gin-debug]") {
		return log.DebugLevel
	}
	if strings.Contains(lowerContent, "debug") {
		return log.DebugLevel
	}

	// Default to info level for standard Gin logs
	return log.InfoLevel
}

// extractFunction attempts to extract function information from Gin log content
func (gt *ginTransformer) extractFunction(logContent string) string {
	// Pattern to match Gin route handler functions
	// Example: "GET /health --> artifex-receiver/handlers.(*HealthHandler).HandleHealthCheck-fm"
	functionPattern := regexp.MustCompile(`-->\s*([^(]+(?:\([^)]*\))?[^-]*(?:-fm)?)`)
	matches := functionPattern.FindStringSubmatch(logContent)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Pattern for other Gin log formats
	if strings.Contains(logContent, "[GIN]") || strings.Contains(logContent, "[GIN-debug]") {
		return "gin.Engine"
	}

	return ""
}

// shouldLogLevel checks if the log level should be processed based on configuration
func (gt *ginTransformer) shouldLogLevel(level log.Level) bool {
	configLevel := gt.config.Level.ToLogLevel()
	return level >= configLevel
}

// outputToRegisteredWriters sends the log event to all registered writers
func (gt *ginTransformer) outputToRegisteredWriters(logEvent *models.LogEvent) {
	internalLog := common.NewLogger().WithContext("function", "GinTransformer.outputToRegisteredWriters").GetLogger()

	// Get all registered writers using the injected function
	if gt.getWriters == nil {
		internalLog.Debug().Msg("No getWriters function provided")
		return
	}

	registeredWriters := gt.getWriters()
	if len(registeredWriters) == 0 {
		internalLog.Debug().Msg("No registered writers found")
		return
	}

	// Marshal the log event to JSON for writers
	jsonData, err := json.Marshal(logEvent)
	if err != nil {
		internalLog.Error().Err(err).Msg("Failed to marshal log event")
		return
	}

	// Send to all registered writers
	for writerName, writer := range registeredWriters {
		if writer != nil {
			writer.Write(jsonData)
			internalLog.Trace().Msgf("Sent Gin log to writer: %s", writerName)
		}
	}
}

// SetCorrelationID sets the correlation ID for the Gin transformer
func (gt *ginTransformer) SetCorrelationID(correlationID string) {
	gt.correlationMux.Lock()
	defer gt.correlationMux.Unlock()
	gt.correlationID = correlationID
}

// GetCorrelationID gets the current correlation ID
func (gt *ginTransformer) GetCorrelationID() string {
	gt.correlationMux.RLock()
	defer gt.correlationMux.RUnlock()
	return gt.correlationID
}

// ClearCorrelationID clears the correlation ID for the Gin transformer
func (gt *ginTransformer) ClearCorrelationID() {
	gt.correlationMux.Lock()
	defer gt.correlationMux.Unlock()
	gt.correlationID = ""
}
