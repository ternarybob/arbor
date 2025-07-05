package main

import (
	"errors"
	"fmt"

	"github.com/ternarybob/arbor"
)

func main() {
	fmt.Println("=== Arbor Fluent Logging Usage Examples ===")
	fmt.Println()

	// Example 1: Direct global logging functions (like log.Info().Msg)
	fmt.Println("1. Direct global logging:")
	arbor.Info().Msg("Application started")
	arbor.Error().Err(errors.New("sample error")).Msg("Something went wrong")
	arbor.Debug().Str("component", "auth").Msg("Debug information")

	fmt.Println("\n2. Using GetLogger() instance:")
	// Example 2: Using logger instance (like logger.GetLogger().Warn().Str(...))
	logger := arbor.GetLogger()
	logger.Warn().Str("connection", "database").Err(errors.New("timeout")).Msg("Connection failed")

	fmt.Println("\n3. Logger with context:")
	// Example 3: Logger with context (correlation ID and prefix)
	contextLogger := logger.WithCorrelationId("req-123").WithPrefix("API")
	contextLogger.Info().Str("endpoint", "/users").Str("method", "GET").Msg("Request processed")

	fmt.Println("\n4. Formatted messages:")
	// Example 4: Formatted messages (like log.Info().Msgf(...))
	port := 8080
	host := "localhost"
	arbor.Info().Msgf("Server starting on %s:%d", host, port)

	fmt.Println("\n5. Chaining multiple fields:")
	// Example 5: Chaining multiple fields
	arbor.Warn().
		Str("user", "john_doe").
		Str("action", "login").
		Err(errors.New("invalid credentials")).
		Msg("Authentication failed")

	fmt.Println("\n6. Different log levels:")
	// Example 6: Different log levels
	arbor.Trace().Msg("Trace level message")
	arbor.Debug().Msg("Debug level message")
	arbor.Info().Msg("Info level message")
	arbor.Warn().Msg("Warning level message")
	arbor.Error().Msg("Error level message")

	fmt.Println("\n=== Examples completed ===")
}
