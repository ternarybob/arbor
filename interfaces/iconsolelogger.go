package interfaces

import (
	"io"

	"github.com/labstack/echo/v4"
	"github.com/phuslu/log"
)

// Level type for arbor logging levels
type Level = log.Level

// Level constants that mirror phuslu/log levels
const (
	// TraceLevel defines trace log level.
	TraceLevel Level = log.TraceLevel
	// DebugLevel defines debug log level.
	DebugLevel Level = log.DebugLevel
	// InfoLevel defines info log level.
	InfoLevel Level = log.InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel Level = log.WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel Level = log.ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel Level = log.FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel Level = log.PanicLevel
	// Disabled disables the logger.
	Disabled Level = log.PanicLevel + 1
)


type IConsoleLogger interface {
	// io.Writer interface for direct usage with frameworks like Gin
	Write(p []byte) (n int, err error)

	GetLogger() *log.Logger

	GetLevel() Level

	WithRequestContext(ctx echo.Context) IConsoleLogger

	WithWriter(name string, writer io.Writer) IConsoleLogger

	WithPrefix(value string) IConsoleLogger

	WithPrefixExtend(value string) IConsoleLogger

	WithCorrelationId(value string) IConsoleLogger

	WithLevel(lvl Level) IConsoleLogger

	WithContext(key string, value string) IConsoleLogger

	WithFunction() IConsoleLogger

	WithFileWriterPath(name string, filePath string, bufferSize, maxFiles int) (IConsoleLogger, error)

	WithFileWriterCustom(name string, fileWriter io.Writer) (IConsoleLogger, error)

	WithFileWriterPattern(name string, pattern string, format string, bufferSize, maxFiles int) (IConsoleLogger, error)

	GetMemoryLogs(correlationid string, minLevel Level) (map[string]string, error)
}
