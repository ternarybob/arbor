package main

import (
	"github.com/ternarybob/arbor"
)

func main() {
	logger := arbor.ConsoleLogger().WithPrefix("TEST")
	log := logger.GetLogger()

	log.Info().Msg("Testing console output format")
	log.Debug().Msg("Debug message with colors")
	log.Warn().Msg("Warning message with colors")
	log.Error().Msg("Error message with colors")
}
