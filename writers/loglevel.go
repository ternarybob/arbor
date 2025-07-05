package writers

import (
	"strings"

	"github.com/gookit/color"
)

// levelprint formats the log level with color - consolidated utility function
func levelprint(level string, colour bool) string {
	switch strings.ToLower(level) {
	case "fatal":
		if colour {
			return color.Red.Render("FTL")
		}
		return "FTL"
	case "error":
		if colour {
			return color.Red.Render("ERR")
		}
		return "ERR"
	case "warn", "warning":
		if colour {
			return color.Yellow.Render("WRN")
		}
		return "WRN"
	case "info":
		if colour {
			return color.Green.Render("INF")
		}
		return "INF"
	case "debug":
		if colour {
			return color.Cyan.Render("DBG")
		}
		return "DBG"
	case "trace":
		if colour {
			return color.Gray.Render("TRC")
		}
		return "TRC"
	default:
		return strings.ToUpper(level)
	}
}
