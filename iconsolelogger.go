package arbor

import (
	"io"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type IConsoleLogger interface {
	GetLogger() zerolog.Logger

	GetLevel() zerolog.Level

	WithRequestContext(ctx echo.Context) IConsoleLogger

	WithWriter(name string, writer io.Writer) IConsoleLogger

	WithPrefix(value string) IConsoleLogger

	WithCorrelationId(value string) IConsoleLogger

	WithLevel(lvl zerolog.Level) IConsoleLogger

	WithContext(key string, value string) IConsoleLogger

	WithFunction() IConsoleLogger

	WithFileWriterPath(name string, filePath string, bufferSize int) (IConsoleLogger, error)

	GinWriter() io.Writer

	GetMemoryLogs(correlationid string, minLevel zerolog.Level) (map[string]string, error)
}
