package arraywriter

import (
	"strings"
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
}

var Levels = map[Level]*LevelMetadata{
	DisableLevel: {
		Name:      "disable",
		ShortName: "",
	},
	FatalLevel: {
		Name:      "fatal",
		ShortName: "FTL",
	},
	ErrorLevel: {
		Name:      "error",
		ShortName: "ERR",
	},
	WarnLevel: {
		Name:      "warn",
		ShortName: "WRN",
	},
	InfoLevel: {
		Name:      "info",
		ShortName: "INF",
	},
	DebugLevel: {
		Name:      "debug",
		ShortName: "DBG",
	},
	TraceLevel: {
		Name:      "trace",
		ShortName: "TRC",
	},
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
