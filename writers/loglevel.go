package writers

import (
	"github.com/gookit/color"
	"github.com/ternarybob/arbor/common"
)

// levelprint formats the log level with color - consolidated utility function
func levelprint(level string, colour bool) string {
	// Get standardized 3-letter level
	lvl := common.LevelStringTo3Letter(level)

	switch lvl {
	case "FTL":
		if colour {
			return color.Red.Render(lvl)
		}
		return lvl
	case "ERR":
		if colour {
			return color.Red.Render(lvl)
		}
		return lvl
	case "WRN":
		if colour {
			return color.Yellow.Render(lvl)
		}
		return lvl
	case "INF":
		if colour {
			return color.Green.Render(lvl)
		}
		return lvl
	case "DBG":
		if colour {
			return color.Cyan.Render(lvl)
		}
		return lvl
	case "TRC":
		if colour {
			return color.Gray.Render(lvl)
		}
		return lvl
	default:
		return lvl
	}
}
