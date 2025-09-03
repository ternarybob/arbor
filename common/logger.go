package common

import "github.com/phuslu/log"

var (
	InternalLevel      log.Level = log.ErrorLevel
	InternalTimeFormat string    = "15:04:05.000" // Console logger, should not need the date!
)

type logger struct {
	internalLog log.Logger
}

func NewLogger() *logger {

	return &logger{
		internalLog: log.Logger{
			Level:      InternalLevel,
			TimeFormat: InternalTimeFormat,
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				EndWithMessage: true,
			},
		},
	}
}

func (l *logger) WithLevel(level log.Level) *logger {
	l.internalLog.Level = log.Level(level)
	return l
}

func (l *logger) WithTimeFormat(format string) *logger {
	l.internalLog.TimeFormat = format
	return l
}

func (l *logger) WithContext(name string, value string) *logger {
	l.internalLog.Context = log.NewContext(nil).Str(name, value).Value()
	return l
}

func (l *logger) GetLogger() log.Logger {
	return l.internalLog
}
