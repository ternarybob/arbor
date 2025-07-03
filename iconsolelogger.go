package arbor

import (
	"io"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type IConsoleLogger interface {
	GetLogger() *zerolog.Logger

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

	GinWriter() io.Writer

	GetMemoryLogs(correlationid string, minLevel Level) (map[string]string, error)
}
