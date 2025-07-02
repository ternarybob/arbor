package arbor

import "github.com/rs/zerolog"

// Level type for arbor logging levels
type Level = zerolog.Level

// Level constants that mirror zerolog levels
const (
	// TraceLevel defines trace log level.
	TraceLevel Level = zerolog.TraceLevel
	// DebugLevel defines debug log level.
	DebugLevel Level = zerolog.DebugLevel
	// InfoLevel defines info log level.
	InfoLevel Level = zerolog.InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel Level = zerolog.WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel Level = zerolog.ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel Level = zerolog.FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel Level = zerolog.PanicLevel
	// NoLevel defines an absent log level.
	NoLevel Level = zerolog.NoLevel
	// Disabled disables the logger.
	Disabled Level = zerolog.Disabled
)

// ParseLevel converts a level string to a Level value.
func ParseLevel(levelStr string) (Level, error) {
	return zerolog.ParseLevel(levelStr)
}
