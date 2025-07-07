package writers

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
)

// ginWriter implements io.Writer for Gin framework integration
// It captures Gin logs and routes them through arbor's existing memory writer infrastructure
type ginWriter struct {
	config         models.WriterConfiguration
	memoryWriter   IMemoryWriter
	correlationID  string
	correlationMux sync.RWMutex
}

// GinWriter creates a new Gin writer that integrates with arbor's memory writer
func GinWriter(config models.WriterConfiguration, memoryWriter IMemoryWriter) io.Writer {
	return &ginWriter{
		config:       config,
		memoryWriter: memoryWriter,
	}
}

// Write implements io.Writer interface for Gin integration
func (gw *ginWriter) Write(p []byte) (n int, err error) {
	internalLog := common.NewLogger().WithContext("function", "GinWriter.Write").GetLogger()

	if len(p) == 0 {
		return 0, nil
	}

	logContent := strings.TrimSpace(string(p))
	if logContent == "" {
		return len(p), nil
	}

	// Parse the Gin log and convert to arbor LogEvent
	logEvent := gw.parseGinLog(logContent)
	if logEvent == nil {
		internalLog.Debug().Msgf("Failed to parse Gin log: %s", logContent)
		return len(p), nil
	}

	// Always output to console in arbor format (unified with other logs)
	// This ensures we see all Gin messages including route registration
	gw.outputToConsole(logEvent)

	// Check if we should log this level for memory storage
	if !gw.shouldLogLevel(logEvent.Level) {
		return len(p), nil
	}

	// Convert to JSON for memory writer
	jsonData, err := json.Marshal(logEvent)
	if err != nil {
		internalLog.Warn().Err(err).Msg("Failed to marshal Gin log event")
		return len(p), err
	}

	// Only write to memory store if there's a correlation ID (i.e., it's from an HTTP request)
	if logEvent.CorrelationID != "" {
		// Write to memory writer if available
		if gw.memoryWriter != nil {
			_, err = gw.memoryWriter.Write(jsonData)
			if err != nil {
				internalLog.Warn().Err(err).Msg("Failed to write to memory writer")
			}
		}
	} else {
		internalLog.Trace().Msgf("Skipping memory store for Gin log without correlation ID: %s", logContent)
	}

	internalLog.Trace().Msgf("Processed Gin log: %s", logContent)
	return len(p), nil
}

// parseGinLog parses a Gin log entry and converts it to arbor LogEvent format
func (gw *ginWriter) parseGinLog(logContent string) *models.LogEvent {
	internalLog := common.NewLogger().WithContext("function", "GinWriter.parseGinLog").GetLogger()

	// Extract correlation ID from current Gin context if available
	correlationID := gw.extractCorrelationID()

	// Determine log level from content
	level := gw.determineLogLevel(logContent)

	// Create log event in arbor format
	logEvent := &models.LogEvent{
		Level:         level,
		Timestamp:     time.Now(),
		Prefix:        "gin",
		Message:       logContent,
		CorrelationID: correlationID,
		Function:      gw.extractFunction(logContent),
	}

	internalLog.Trace().Msgf("Parsed Gin log - Level: %s, CorrelationID: %s",
		level.String(), correlationID)

	return logEvent
}

// extractCorrelationID attempts to extract correlation ID
func (gw *ginWriter) extractCorrelationID() string {
	gw.correlationMux.RLock()
	defer gw.correlationMux.RUnlock()

	// Return the stored correlation ID (will be set by SetCorrelationID method)
	return gw.correlationID
}

// determineLogLevel determines the log level based on Gin log content
func (gw *ginWriter) determineLogLevel(logContent string) log.Level {
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
	if strings.Contains(lowerContent, "[gin-debug]") || strings.Contains(lowerContent, "debug") {
		return log.DebugLevel
	}

	// Default to info level for standard Gin logs
	return log.InfoLevel
}

// extractFunction attempts to extract function information from Gin log content
func (gw *ginWriter) extractFunction(logContent string) string {
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
func (gw *ginWriter) shouldLogLevel(level log.Level) bool {
	configLevel := gw.config.Level.ToLogLevel()
	return level >= configLevel
}

// SetCorrelationID sets the correlation ID for the Gin writer
func (gw *ginWriter) SetCorrelationID(correlationID string) {
	gw.correlationMux.Lock()
	defer gw.correlationMux.Unlock()
	gw.correlationID = correlationID
}

// GetCorrelationID gets the current correlation ID
func (gw *ginWriter) GetCorrelationID() string {
	gw.correlationMux.RLock()
	defer gw.correlationMux.RUnlock()
	return gw.correlationID
}

// outputToConsole formats and outputs the log event to console using phuslu with color support
func (gw *ginWriter) outputToConsole(logEvent *models.LogEvent) {
	// Create a phuslu console logger with color support
	consoleLogger := log.Logger{
		Level:      log.TraceLevel, // Allow all levels through
		TimeFormat: "15:04:05.000",
		Writer: &log.ConsoleWriter{
			ColorOutput:    true,
			EndWithMessage: true,
		},
	}

	// Create the appropriate log entry based on level
	var entry *log.Entry
	switch logEvent.Level {
	case log.TraceLevel:
		entry = consoleLogger.Trace()
	case log.DebugLevel:
		entry = consoleLogger.Debug()
	case log.InfoLevel:
		entry = consoleLogger.Info()
	case log.WarnLevel:
		entry = consoleLogger.Warn()
	case log.ErrorLevel:
		entry = consoleLogger.Error()
	case log.FatalLevel:
		entry = consoleLogger.Fatal()
	case log.PanicLevel:
		entry = consoleLogger.Panic()
	default:
		entry = consoleLogger.Info()
	}

	// Add arbor-specific fields
	entry = entry.Str("prefix", logEvent.Prefix)

	if logEvent.Function != "" {
		entry = entry.Str("function", logEvent.Function)
	}

	if logEvent.CorrelationID != "" {
		entry = entry.Str("correlationid", logEvent.CorrelationID)
	}

	// Output the message with phuslu formatting and colors
	entry.Msg(logEvent.Message)
}
