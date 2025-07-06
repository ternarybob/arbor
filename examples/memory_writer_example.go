package main

import (
	"fmt"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor"
	"github.com/ternarybob/arbor/models"
)

func memoryWriterExample() {
	fmt.Println("=== BoltDB Memory Writer Example ===\n")

	// Create logger with memory writer
	config := models.WriterConfiguration{}
	logger := arbor.Logger().WithMemoryWriter(config)

	// Set correlation ID for tracking related log entries
	correlationID := "user-session-123"
	logger = logger.WithCorrelationId(correlationID).WithPrefix("WEBAPP")

	fmt.Printf("Logging messages with correlation ID: %s\n", correlationID)

	// Log various messages with different levels
	logger.Info().Msg("User logged in successfully")
	logger.Debug().Str("username", "john.doe").Msg("Debug: User authentication details")
	logger.Warn().Msg("Password will expire in 3 days")
	logger.Error().Str("error", "database connection timeout").Msg("Failed to load user preferences")
	logger.Info().Msg("User navigated to dashboard")

	// Small delay to ensure all writes are processed
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n=== Retrieving All Logs (Info level and above) ===")
	allLogs, err := logger.GetMemoryLogs(correlationID, arbor.LogLevel(log.InfoLevel))
	if err != nil {
		fmt.Printf("Error retrieving logs: %v\n", err)
		return
	}

	for index, logEntry := range allLogs {
		fmt.Printf("[%s] %s\n", index, logEntry)
	}

	fmt.Println("\n=== Retrieving Error Logs Only ===")
	errorLogs, err := logger.GetMemoryLogs(correlationID, arbor.LogLevel(log.ErrorLevel))
	if err != nil {
		fmt.Printf("Error retrieving error logs: %v\n", err)
		return
	}

	for index, logEntry := range errorLogs {
		fmt.Printf("[%s] %s\n", index, logEntry)
	}

	// Demonstrate with different correlation ID
	fmt.Println("\n=== Different User Session ===")
	correlationID2 := "user-session-456"
	logger2 := logger.WithCorrelationId(correlationID2)

	logger2.Info().Msg("Different user logged in")
	logger2.Warn().Msg("Suspicious login attempt detected")

	time.Sleep(50 * time.Millisecond)

	// Show logs for second correlation ID
	fmt.Printf("\nLogs for correlation ID: %s\n", correlationID2)
	user2Logs, err := logger.GetMemoryLogs(correlationID2, arbor.LogLevel(log.InfoLevel))
	if err != nil {
		fmt.Printf("Error retrieving logs: %v\n", err)
		return
	}

	for index, logEntry := range user2Logs {
		fmt.Printf("[%s] %s\n", index, logEntry)
	}

	// Demonstrate non-existent correlation ID
	fmt.Println("\n=== Non-existent Correlation ID ===")
	emptyLogs, err := logger.GetMemoryLogs("non-existent-id", arbor.LogLevel(log.InfoLevel))
	if err != nil {
		fmt.Printf("Error retrieving logs: %v\n", err)
		return
	}

	fmt.Printf("Logs found for non-existent ID: %d\n", len(emptyLogs))

	fmt.Println("\n=== Key Features Demonstrated ===")
	fmt.Println("✓ Self-expiring storage (24 hour TTL)")
	fmt.Println("✓ Persistent storage using BoltDB")
	fmt.Println("✓ Level-based filtering")
	fmt.Println("✓ Correlation ID-based grouping")
	fmt.Println("✓ Automatic cleanup (no manual intervention)")
	fmt.Println("✓ Thread-safe operations")
	fmt.Println("✓ Multiple correlation IDs supported")

	fmt.Println("\n=== Database Information ===")
	fmt.Println("• Database file: {TempDir}/arbor_logs.db")
	fmt.Println("• Automatic cleanup: Every 1 hour")
	fmt.Println("• Entry expiration: 24 hours")
	fmt.Println("• Storage format: JSON with expiration timestamps")
}
