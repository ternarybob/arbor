package ginwriter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	consolewriter "github.com/ternarybob/arbor/consolewriter"

	"github.com/gookit/color"
	"github.com/rs/zerolog"
)

type GinWriter struct {
	Out   io.Writer
	Level zerolog.Level
}

var (
	loglevel    zerolog.Level  = zerolog.WarnLevel
	prefix      string         = "GinWriter"
	internallog zerolog.Logger = zerolog.New(&consolewriter.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger().Level(loglevel)
)

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
		log      = internallog.With().Str("prefix", "writeline").Logger()
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
		log.Warn().Err(err).Msg("")
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
	Level     zerolog.Level
	ColorCode func(a ...interface{}) string
}

var Levels = map[Level]*LevelMetadata{
	DisableLevel: {
		Name:      "disable",
		ShortName: "DIS",
		Level:     zerolog.Disabled,
		ColorCode: color.Debug.Render,
	},
	FatalLevel: {
		Name:      "fatal",
		ShortName: "FTL",
		Level:     zerolog.FatalLevel,
		ColorCode: color.Danger.Render,
	},
	ErrorLevel: {
		Name:      "error",
		ShortName: "ERR",
		Level:     zerolog.ErrorLevel,
		ColorCode: color.Error.Render,
	},
	WarnLevel: {
		Name:      "warn",
		ShortName: "WRN",
		Level:     zerolog.WarnLevel,
		ColorCode: color.Warn.Render,
	},
	InfoLevel: {
		Name:      "info",
		ShortName: "INF",
		Level:     zerolog.InfoLevel,
		ColorCode: color.Info.Render,
	},
	DebugLevel: {
		Name:      "debug",
		ShortName: "DBG",
		Level:     zerolog.DebugLevel,
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
		log := internallog.With().Str("prefix", "toJson").Logger()
		log.Err(err).Msg("Object marshaling error")
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
