package consolewriter

import (
	"strings"

	"github.com/gookit/color"
)

type Level uint32

const (
	DisableLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

// LevelMetadata describes the information
// behind a log Level, each level has its own unique metadata.
type LevelMetadata struct {
	Name      string
	ShortName string
	ColorCode func(a ...interface{}) string
}

var Levels = map[Level]*LevelMetadata{
	DisableLevel: {
		Name:      "disable",
		ShortName: "DIS",
		ColorCode: color.Gray.Render,
	},
	FatalLevel: {
		Name:      "fatal",
		ShortName: "FTL",
		ColorCode: color.Danger.Render,
	},
	ErrorLevel: {
		Name:      "error",
		ShortName: "ERR",
		ColorCode: color.Error.Render,
	},
	WarnLevel: {
		Name:      "warn",
		ShortName: "WRN",
		ColorCode: color.Warn.Render,
	},
	InfoLevel: {
		Name:      "info",
		ShortName: "INF",
		ColorCode: color.Info.Render,
	},
	DebugLevel: {
		Name:      "debug",
		ShortName: "DBG",
		ColorCode: color.Debug.Render,
	},
	TraceLevel: {
		Name:      "trace",
		ShortName: "TRC",
		ColorCode: color.Light.Render,
	},
}

func levelprint(level string, colour bool) string {

	parsedLevel := parselevel(level)
	_level := Levels[parsedLevel]

	// Handle nil level metadata or missing ColorCode
	if _level == nil {
		return level // Fallback to original level string
	}

	// Debug output to see what's happening
	// fmt.Printf("[DEBUG] level=%s, parsedLevel=%d, colour=%v, hasColorCode=%v\n", level, parsedLevel, colour, _level.ColorCode != nil)

	if colour && _level.ColorCode != nil {
		return _level.ColorCode(_level.ShortName)
	} else {
		return _level.ShortName
	}

}

func parselevel(levelName string) Level {

	levelName = strings.ToLower(levelName)

	for level, meta := range Levels {
		if strings.ToLower(meta.Name) == levelName {
			return level
		}

	}

	return DisableLevel
}
