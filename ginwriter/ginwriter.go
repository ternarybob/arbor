package ginwriter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/phuslu/log"
)

type GinWriter struct {
	Out   io.Writer
	Level log.Level
}

var (
	loglevel    log.Level  = log.WarnLevel
	prefix      string     = "GinWriter"
	internallog log.Logger = log.Logger{
		Level:  loglevel,
		Writer: &log.ConsoleWriter{},
	}
)

// New creates a new GinWriter instance
func New() *GinWriter {
	return &GinWriter{
		Out:   os.Stdout,
		Level: log.InfoLevel,
	}
}

type LogEvent struct {
	Level         *LevelMetadata `json:"level"`
	Time          time.Time      `json:"-"`
	Timestamp     int64          `json:"timestamp,omitempty"`
	Prefix        string         `json:"prefix"`
	CorrelationID string         `json:"correlationid"`
	Message       string         `json:"message"`
	Error         string         `json:"error"`
}

type Frame struct {
	Function string `json:"function"`
	Source   string `json:"source"`
}

func (w *GinWriter) Write(e []byte) (n int, err error) {

	n = len(e)

	if n == 0 {
		return n, nil
	}

	// fmt.Printf(string(e))

	err = w.writeline(e)
	if err != nil {
		return 0, err
	}

	return 0, nil
}

func (w *GinWriter) writeline(event []byte) error {

	var (
		// Use direct logging instead of chained logger
		logentry LogEvent
	)

	if len(event) <= 0 {
		return fmt.Errorf("[%s] entry is Empty", prefix)
	}

	logentry.Prefix = "GIN"
	logentry.Time = time.Now()
	logentry.Message = strings.TrimSuffix(string(event), "\n")

	logstring := string(event)

	switch {
	case stringContains(logstring, "GIN-fatal"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-fatal] ", "")
		logentry.Level = Levels[1]
	case stringContains(logstring, "GIN-error"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-error] ", "")
		logentry.Level = Levels[2]
	case stringContains(logstring, "GIN-warning"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-warning] ", "")
		logentry.Level = Levels[3]
	case stringContains(logstring, "GIN-information"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-information] ", "")
		logentry.Level = Levels[4]
	case stringContains(logstring, "GIN-debug"):
		logentry.Message = strings.ReplaceAll(logentry.Message, "[GIN-debug] ", "")
		logentry.Level = Levels[5]
	default:
		// Disabled or Trace
		return nil
	}

	if w.Level > logentry.Level.Level {
		return nil
	}

	_, err := fmt.Printf("%s\n", w.format(&logentry, true))
	if err != nil {
		internallog.Warn().Str("prefix", "writeline").Err(err).Msg("")
		return err
	}

	return nil
}

func (w *GinWriter) format(l *LogEvent, colour bool) string {

	timestamp := l.Time.Format(time.Stamp)

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

func levelprint(level *LevelMetadata, colour bool) string {

	if level == nil {
		return ""
	}

	if colour {
		return level.ColorCode(level.ShortName)
	} else {
		return level.ShortName
	}

}

type Level int8

const (
	DisableLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

// LevelMetadata describes the information
// behind a log Level, each level has its own unique metadata.
type LevelMetadata struct {
	Name      string
	ShortName string
	Level     log.Level
	ColorCode func(a ...interface{}) string
}

var Levels = map[Level]*LevelMetadata{
	DisableLevel: {
		Name:      "disable",
		ShortName: "DIS",
		Level:     log.PanicLevel + 1, // Effectively disabled
		ColorCode: color.Debug.Render,
	},
	FatalLevel: {
		Name:      "fatal",
		ShortName: "FTL",
		Level:     log.FatalLevel,
		ColorCode: color.Danger.Render,
	},
	ErrorLevel: {
		Name:      "error",
		ShortName: "ERR",
		Level:     log.ErrorLevel,
		ColorCode: color.Error.Render,
	},
	WarnLevel: {
		Name:      "warn",
		ShortName: "WRN",
		Level:     log.WarnLevel,
		ColorCode: color.Warn.Render,
	},
	InfoLevel: {
		Name:      "info",
		ShortName: "INF",
		Level:     log.InfoLevel,
		ColorCode: color.Info.Render,
	},
	DebugLevel: {
		Name:      "debug",
		ShortName: "DBG",
		Level:     log.DebugLevel,
		ColorCode: color.Debug.Render,
	},
}

func parseLevel(levelName string) Level {

	if isEmpty(levelName) {
		return DisableLevel
	}

	levelName = strings.ToLower(levelName)

	for level, meta := range Levels {
		if strings.ToLower(meta.Name) == levelName {
			return level
		}

	}

	return DisableLevel
}

func toJson(input interface{}) string {

	output, err := json.MarshalIndent(input, "", "\t")

	if err != nil {
		internallog.Error().Str("prefix", "toJson").Msgf("Object marshaling error: %v", err)
		return ""
	}

	return string(output)
}

func isEmpty(input string) bool {
	return (len(strings.TrimSpace(input)) <= 0)
}

func stringContains(this string, contains string) bool {

	if strings.Contains(strings.ToLower(this), strings.ToLower(contains)) {
		return true
	}

	return false
}
