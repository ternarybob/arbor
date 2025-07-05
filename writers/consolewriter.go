package writers

import (
	"fmt"

	"github.com/ternarybob/arbor/models"
	"github.com/ternarybob/arbor/services"

	"github.com/gookit/color"
	"github.com/phuslu/log"
)

var (
	loglevel    log.Level  = log.WarnLevel
	internallog log.Logger = log.Logger{
		Level:  loglevel,
		Writer: &log.ConsoleWriter{},
	}
)

func init() {
	// Enable color output for Windows terminals
	color.ForceOpenColor()
}

type consoleWriter struct {
	logger     log.Logger
	config     models.WriterConfiguration
	ginService services.IGinService // Optional Gin formatting service
}

// NewConsoleWriter creates a new ConsoleWriter with optional configuration
func ConsoleWriter(config models.WriterConfiguration) IWriter {

	cw := &consoleWriter{
		logger: log.Logger{
			Level:      config.Level,
			TimeFormat: config.TimeFormat,
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				EndWithMessage: true,
			},
		},
		config: config,
	}

	return cw
}

func (cw *consoleWriter) WithLevel(level log.Level) IWriter {

	cw.logger.SetLevel(level)

	return cw
}

func (cw *consoleWriter) Write(e []byte) (n int, err error) {
	n = len(e)
	if n <= 0 {
		return n, err
	}

	return n, nil
}

func (cw *consoleWriter) format(l *models.LogEvent, colour bool) string {

	timestamp := l.Timestamp.Format("15:04:05.000")

	output := fmt.Sprintf("%s|%s", levelprint(l.Level, colour), timestamp)

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
