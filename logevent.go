package arbor

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/models"
)

// logEvent implements the ILogEvent interface
type logEvent struct {
	logger *logger
	level  log.Level
	fields map[string]interface{}
	err    error
}

// newLogEvent creates a new log event
func newLogEvent(logger *logger, level log.Level) *logEvent {
	return &logEvent{
		logger: logger,
		level:  level,
		fields: make(map[string]interface{}),
	}
}

// Strs adds a string slice field to the log event
func (le *logEvent) Strs(key string, values []string) ILogEvent {
	le.fields[key] = values
	return le
}

// Str adds a string field to the log event
func (le *logEvent) Str(key, value string) ILogEvent {
	le.fields[key] = value
	return le
}

// Err adds an error field to the log event
func (le *logEvent) Err(err error) ILogEvent {
	le.err = err
	return le
}

// Int adds an integer field to the log event
func (le *logEvent) Int(key string, value int) ILogEvent {
	le.fields[key] = value
	return le
}

// Int32 adds an int32 field to the log event
func (le *logEvent) Int32(key string, value int32) ILogEvent {
	le.fields[key] = value
	return le
}

// Int64 adds an int64 field to the log event
func (le *logEvent) Int64(key string, value int64) ILogEvent {
	le.fields[key] = value
	return le
}

// Float32 adds a float32 field to the log event
func (le *logEvent) Float32(key string, value float32) ILogEvent {
	le.fields[key] = value
	return le
}

// Dur adds a duration field to the log event
func (le *logEvent) Dur(key string, value time.Duration) ILogEvent {
	le.fields[key] = value.String()
	return le
}

// Float64 adds a float64 field to the log event
func (le *logEvent) Float64(key string, value float64) ILogEvent {
	le.fields[key] = value
	return le
}

// Msg logs the message with the accumulated fields
func (le *logEvent) Msg(message string) {
	le.writeLog(message)
}

// Msgf logs the formatted message with the accumulated fields
func (le *logEvent) Msgf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	le.writeLog(message)
}

// writeLog writes the log event to all configured writers
func (le *logEvent) writeLog(message string) {
	// Create a log event model
	logEvent := &models.LogEvent{
		Level:     le.level,
		Timestamp: time.Now(),
		Message:   message,
		Fields:    le.fields,
	}

	// Add error if present
	if le.err != nil {
		logEvent.Error = le.err.Error()
	}

	// Add correlation ID if present
	if correlationID, exists := le.logger.contextData[CORRELATION_ID_KEY]; exists {
		logEvent.CorrelationID = correlationID
	}

	// Add prefix if present
	if prefix, exists := le.logger.contextData[PREFIX_KEY]; exists {
		logEvent.Prefix = prefix
	}

	// Add function name
	logEvent.Function = le.logger.getFunctionName()

	// Write to all registered writers
	registeredWriters := GetAllRegisteredWriters()
	for writerKey, writer := range registeredWriters {
		if writerKey == WRITER_CONSOLE {
			// Console writer expects JSON data (phuslu will handle formatting)
			if jsonData, err := json.Marshal(logEvent); err == nil {
				writer.Write(jsonData)
			}
		} else if writerKey == WRITER_MEMORY {
			// Memory writers need JSON format
			if jsonData, err := json.Marshal(logEvent); err == nil {
				writer.Write(jsonData)
			}
		} else {
			// File writers get formatted string
			logEntry := le.formatLogEntry(logEvent)
			writer.Write([]byte(logEntry))
		}
	}
}

// formatLogEntry formats the log event into a string
func (le *logEvent) formatLogEntry(event *models.LogEvent) string {
	timestamp := event.Timestamp.Format("15:04:05.000")
	levelStr := le.levelToString(event.Level)

	output := fmt.Sprintf("%s|%s", levelStr, timestamp)

	if event.Prefix != "" {
		output += fmt.Sprintf("|%s", event.Prefix)
	}

	if event.Function != "" {
		output += fmt.Sprintf("|%s", event.Function)
	}

	if event.CorrelationID != "" {
		output += fmt.Sprintf("|%s", event.CorrelationID)
	}

	// Add custom fields
	for key, value := range event.Fields {
		output += fmt.Sprintf("|%s=%v", key, value)
	}

	if event.Error != "" {
		output += fmt.Sprintf("|error=%s", event.Error)
	}

	if event.Message != "" {
		output += fmt.Sprintf("|%s", event.Message)
	}

	return output + "\n"
}

// levelToString converts log level to string representation
func (le *logEvent) levelToString(level log.Level) string {
	switch level {
	case log.TraceLevel:
		return "TRACE"
	case log.DebugLevel:
		return "DEBUG"
	case log.InfoLevel:
		return "INFO"
	case log.WarnLevel:
		return "WARN"
	case log.ErrorLevel:
		return "ERROR"
	case log.FatalLevel:
		return "FATAL"
	case log.PanicLevel:
		return "PANIC"
	default:
		return "INFO"
	}
}

// LevelToString converts log level to string representation (exported for writers)
func LevelToString(level log.Level) string {
	switch level {
	case log.TraceLevel:
		return "trace"
	case log.DebugLevel:
		return "debug"
	case log.InfoLevel:
		return "info"
	case log.WarnLevel:
		return "warn"
	case log.ErrorLevel:
		return "error"
	case log.FatalLevel:
		return "fatal"
	case log.PanicLevel:
		return "panic"
	default:
		return "info"
	}
}
