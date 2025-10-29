package arbor

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/phuslu/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ternarybob/arbor/models"
)

func TestLogger_New(t *testing.T) {
	logger := Logger()
	if logger == nil {
		t.Fatal("Logger() should not return nil")
	}

	// Verify it implements ILogger interface
	var _ ILogger = logger
}

func TestLogger_WithCorrelationId(t *testing.T) {
	logger := Logger()

	// Test with provided correlation ID
	correlationID := "test-correlation-123"
	newLogger := logger.WithCorrelationId(correlationID)

	if newLogger == nil {
		t.Error("WithCorrelationId should not return nil")
	}

	// Test with empty correlation ID (should generate UUID)
	newLogger2 := logger.WithCorrelationId("")
	if newLogger2 == nil {
		t.Error("WithCorrelationId with empty string should not return nil")
	}
}

func TestLogger_WithPrefix(t *testing.T) {
	logger := Logger()

	// Test with valid prefix
	prefix := "API"
	newLogger := logger.WithPrefix(prefix)

	if newLogger == nil {
		t.Error("WithPrefix should not return nil")
	}

	// Test with empty prefix
	newLogger2 := logger.WithPrefix("")
	if newLogger2 == nil {
		t.Error("WithPrefix with empty string should not return nil")
	}
}

func TestLogger_WithLevel(t *testing.T) {
	logger := Logger()

	testLevels := []LogLevel{
		TraceLevel,
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		FatalLevel,
		PanicLevel,
	}

	for _, level := range testLevels {
		t.Run(string(rune(level)), func(t *testing.T) {
			newLogger := logger.WithLevel(level)
			if newLogger == nil {
				t.Error("WithLevel should not return nil")
			}
		})
	}
}

func TestLogger_WithContext(t *testing.T) {
	logger := Logger()

	// Test with valid key-value pair
	newLogger := logger.WithContext("key", "value")
	if newLogger == nil {
		t.Error("WithContext should not return nil")
	}

	// Test with empty key
	newLogger2 := logger.WithContext("", "value")
	if newLogger2 == nil {
		t.Error("WithContext with empty key should not return nil")
	}

	// Test with empty value
	newLogger3 := logger.WithContext("key", "")
	if newLogger3 == nil {
		t.Error("WithContext with empty value should not return nil")
	}
}

func TestLogger_WithFileWriter(t *testing.T) {
	logger := Logger()

	config := models.WriterConfiguration{
		Type:       models.LogWriterTypeFile,
		Level:      InfoLevel,
		TimeFormat: "15:04:05.000",
		FileName:   "temp/test.log",
	}

	newLogger := logger.WithFileWriter(config)
	if newLogger == nil {
		t.Error("WithFileWriter should not return nil")
	}
}

func TestLogger_FluentMethods(t *testing.T) {
	logger := Logger()

	// Test all fluent logging methods
	testCases := []struct {
		name   string
		method func() ILogEvent
	}{
		{"Trace", logger.Trace},
		{"Debug", logger.Debug},
		{"Info", logger.Info},
		{"Warn", logger.Warn},
		{"Error", logger.Error},
		{"Fatal", logger.Fatal},
		{"Panic", logger.Panic},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := tc.method()
			if event == nil {
				t.Errorf("%s() should not return nil", tc.name)
			}

			// Verify it implements ILogEvent interface
			var _ ILogEvent = event
		})
	}
}

func TestGlobalLogger_Functions(t *testing.T) {
	// Test that global functions work
	testCases := []struct {
		name   string
		method func() ILogEvent
	}{
		{"Trace", Trace},
		{"Debug", Debug},
		{"Info", Info},
		{"Warn", Warn},
		{"Error", Error},
		{"Fatal", Fatal},
		{"Panic", Panic},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := tc.method()
			if event == nil {
				t.Errorf("%s() should not return nil", tc.name)
			}

			// Verify it implements ILogEvent interface
			var _ ILogEvent = event
		})
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger() should not return nil")
	}

	// Should return the same instance on multiple calls
	logger2 := GetLogger()
	if logger != logger2 {
		t.Error("GetLogger() should return the same instance")
	}
}

func TestLevelToString_Function(t *testing.T) {
	testCases := []struct {
		level    log.Level
		expected string
	}{
		{log.TraceLevel, "trace"},
		{log.DebugLevel, "debug"},
		{log.InfoLevel, "info"},
		{log.WarnLevel, "warn"},
		{log.ErrorLevel, "error"},
		{log.FatalLevel, "fatal"},
		{log.PanicLevel, "panic"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := LevelToString(tc.level)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestParseLevelString(t *testing.T) {
	testCases := []struct {
		input    string
		expected log.Level
		hasError bool
	}{
		{"trace", log.TraceLevel, false},
		{"debug", log.DebugLevel, false},
		{"info", log.InfoLevel, false},
		{"warn", log.WarnLevel, false},
		{"warning", log.WarnLevel, false},
		{"error", log.ErrorLevel, false},
		{"fatal", log.FatalLevel, false},
		{"panic", log.PanicLevel, false},
		{"disabled", log.PanicLevel + 1, false},
		{"off", log.PanicLevel + 1, false},
		{"invalid", log.InfoLevel, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := ParseLevelString(tc.input)

			if tc.hasError {
				if err == nil {
					t.Error("Expected error for invalid level")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	testCases := []struct {
		input    int
		expected log.Level
	}{
		{int(TraceLevel), log.TraceLevel},
		{int(DebugLevel), log.DebugLevel},
		{int(InfoLevel), log.InfoLevel},
		{int(WarnLevel), log.WarnLevel},
		{int(ErrorLevel), log.ErrorLevel},
		{int(FatalLevel), log.FatalLevel},
		{int(PanicLevel), log.PanicLevel},
		{int(Disabled), 0},
		{999, log.InfoLevel}, // Default case
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.input)), func(t *testing.T) {
			result := ParseLogLevel(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestLogger_GetFunctionName(t *testing.T) {
	logger := Logger().(*logger)

	// This will test the function name detection
	funcName := logger.getFunctionName()

	// Should contain test function name or be empty (acceptable for edge cases)
	if funcName != "" && !strings.Contains(funcName, "Test") {
		// This is informational - function name detection can vary
		t.Logf("Function name detected: %s", funcName)
	}
}

func TestLogger_ChainedUsage(t *testing.T) {
	// Test complex chained usage
	logger := Logger().WithCorrelationId("test-123").WithPrefix("TEST")

	// This should not panic and should work end-to-end
	event := logger.Info().Str("key1", "value1").Str("key2", "value2")
	if event == nil {
		t.Error("Chained usage should not return nil")
	}

	// Verify we can add an error and still chain
	err := errors.New("test error")
	event2 := event.Err(err)
	if event2 == nil {
		t.Error("Chained usage with error should not return nil")
	}

	// Should be the same instance
	if event != event2 {
		t.Error("Chained methods should return the same instance")
	}
}

func TestLogger_Copy(t *testing.T) {
	// Create a logger with some context
	originalLogger := Logger().WithCorrelationId("original-123").WithPrefix("ORIGINAL")

	// Create a copy
	copiedLogger := originalLogger.Copy()

	// Verify copy is not nil
	if copiedLogger == nil {
		t.Error("Copy should not return nil")
	}

	// Verify copy is a different instance
	if originalLogger == copiedLogger {
		t.Error("Copy should return a different instance")
	}

	// Verify both loggers implement ILogger interface
	var _ ILogger = originalLogger
	var _ ILogger = copiedLogger

	// Verify that the copied logger has NO context data (fresh/clean)
	// This is the key behavior - Copy should give you a clean logger
	originalLoggerTyped := originalLogger.(*logger)
	copiedLoggerTyped := copiedLogger.(*logger)

	// Original should have context data
	if originalLoggerTyped.contextData == nil || len(originalLoggerTyped.contextData) == 0 {
		t.Error("Original logger should have context data")
	}

	// Copied logger should have empty context data (fresh/clean)
	if copiedLoggerTyped.contextData == nil {
		t.Error("Copied logger should have initialized (but empty) context data")
	}
	if len(copiedLoggerTyped.contextData) != 0 {
		t.Error("Copied logger should have empty context data (fresh/clean)")
	}

	// Test that modifying the copy doesn't affect the original
	copiedLogger.WithCorrelationId("copied-456").WithPrefix("COPIED")

	// Both should still be usable for logging
	originalEvent := originalLogger.Info().Str("source", "original")
	copiedEvent := copiedLogger.Info().Str("source", "copied")

	if originalEvent == nil || copiedEvent == nil {
		t.Error("Both original and copied loggers should be usable for logging")
	}
}

func TestLogger_ClearContext(t *testing.T) {
	// Create a logger and add multiple context items
	testLogger := NewLogger()
	testLogger.WithCorrelationId("test-correlation-123")
	testLogger.WithContext("key1", "value1")
	testLogger.WithContext("key2", "value2")
	testLogger.WithPrefix("test-prefix")

	// Cast to internal logger type to access contextData
	testLoggerTyped := testLogger.(*logger)

	// Verify context is set
	expectedCount := 4 // correlationid, key1, key2, prefix
	if len(testLoggerTyped.contextData) != expectedCount {
		t.Errorf("Logger should have %d context items, got %d. Context: %v", expectedCount, len(testLoggerTyped.contextData), testLoggerTyped.contextData)
	}

	// Clear all context
	testLogger.ClearContext()

	// Verify context is cleared
	if len(testLoggerTyped.contextData) != 0 {
		t.Errorf("Logger should have empty context after ClearContext(), got %d items", len(testLoggerTyped.contextData))
	}

	// Verify we can add new context after clearing
	testLogger.WithCorrelationId("new-correlation-456")
	if testLoggerTyped.contextData["correlationid"] != "new-correlation-456" {
		t.Errorf("Logger should accept new context after clearing")
	}
	if len(testLoggerTyped.contextData) != 1 {
		t.Errorf("Logger should have 1 context item after adding new correlation ID, got %d", len(testLoggerTyped.contextData))
	}
}

// TestLogger_SetChannel_Basic tests basic SetChannel functionality with default batching (5 events, 1 second)
func TestLogger_SetChannel_Basic(t *testing.T) {
	logger := NewLogger()
	logChan := make(chan []models.LogEvent, 10)
	channelName := "test-channel-basic"

	// Register channel with default batching
	logger.SetChannel(channelName, logChan)
	defer logger.UnregisterChannel(channelName)

	// Verify writer is registered in global registry
	writer := GetRegisteredWriter(channelName)
	require.NotNil(t, writer, "Channel writer should be registered")

	// Log 5 messages (should trigger batch flush due to batch size)
	for i := 0; i < 5; i++ {
		logger.Info().Msgf("Test message %d", i)
	}

	// Wait for batch with timeout
	select {
	case batch := <-logChan:
		require.NotEmpty(t, batch, "Batch should contain events")
		assert.LessOrEqual(t, len(batch), 5, "Batch size should not exceed 5")

		// Verify first event has correct message
		if len(batch) > 0 {
			assert.Contains(t, batch[0].Message, "Test message", "Message should match")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for batch")
	}
}

// TestLogger_SetChannelWithBuffer_CustomBatching tests custom batching parameters (3 events, 50ms)
func TestLogger_SetChannelWithBuffer_CustomBatching(t *testing.T) {
	logger := NewLogger()
	logChan := make(chan []models.LogEvent, 10)
	channelName := "test-channel-custom"

	// Register channel with custom batching (3 events, 50ms)
	logger.SetChannelWithBuffer(channelName, logChan, 3, 50*time.Millisecond)
	defer logger.UnregisterChannel(channelName)

	// Verify writer is registered
	writer := GetRegisteredWriter(channelName)
	require.NotNil(t, writer, "Channel writer should be registered")

	// Log 3 messages (should trigger batch flush due to batch size)
	for i := 0; i < 3; i++ {
		logger.Info().Msgf("Custom batch message %d", i)
	}

	// Wait for batch with timeout
	select {
	case batch := <-logChan:
		assert.LessOrEqual(t, len(batch), 3, "Batch size should not exceed 3")
		assert.NotEmpty(t, batch, "Batch should contain events")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for custom batch")
	}
}

// TestLogger_SetChannel_MultipleChannels tests multiple independent channels receiving logs
func TestLogger_SetChannel_MultipleChannels(t *testing.T) {
	logger := NewLogger()

	logChan1 := make(chan []models.LogEvent, 10)
	logChan2 := make(chan []models.LogEvent, 10)
	channelName1 := "test-channel-multi-1"
	channelName2 := "test-channel-multi-2"

	// Register two independent channels
	logger.SetChannel(channelName1, logChan1)
	logger.SetChannel(channelName2, logChan2)
	defer logger.UnregisterChannel(channelName1)
	defer logger.UnregisterChannel(channelName2)

	// Verify both writers are registered
	writer1 := GetRegisteredWriter(channelName1)
	writer2 := GetRegisteredWriter(channelName2)
	require.NotNil(t, writer1, "Channel 1 writer should be registered")
	require.NotNil(t, writer2, "Channel 2 writer should be registered")

	// Log messages
	for i := 0; i < 5; i++ {
		logger.Info().Msgf("Multi-channel message %d", i)
	}

	// Both channels should receive batches
	receivedChan1 := false
	receivedChan2 := false

	for i := 0; i < 2; i++ {
		select {
		case batch := <-logChan1:
			assert.NotEmpty(t, batch, "Channel 1 batch should contain events")
			receivedChan1 = true
		case batch := <-logChan2:
			assert.NotEmpty(t, batch, "Channel 2 batch should contain events")
			receivedChan2 = true
		case <-time.After(3 * time.Second):
			t.Fatal("Timeout waiting for batches from multiple channels")
		}
	}

	assert.True(t, receivedChan1, "Channel 1 should have received batch")
	assert.True(t, receivedChan2, "Channel 2 should have received batch")
}

// TestLogger_SetChannel_WithCorrelationID tests channel logging preserves correlation IDs
func TestLogger_SetChannel_WithCorrelationID(t *testing.T) {
	logger := NewLogger()
	logChan := make(chan []models.LogEvent, 10)
	channelName := "test-channel-correlation"
	correlationID := "test-correlation-12345"

	// Register channel
	logger.SetChannel(channelName, logChan)
	defer logger.UnregisterChannel(channelName)

	// Create logger with correlation ID
	correlatedLogger := logger.WithCorrelationId(correlationID)

	// Log messages with correlation ID
	for i := 0; i < 5; i++ {
		correlatedLogger.Info().Msgf("Correlated message %d", i)
	}

	// Wait for batch and verify correlation ID
	select {
	case batch := <-logChan:
		require.NotEmpty(t, batch, "Batch should contain events")

		// All events should have the correlation ID
		for _, event := range batch {
			assert.Equal(t, correlationID, event.CorrelationID, "Event should have correct correlation ID")
			assert.Contains(t, event.Message, "Correlated message", "Message should match")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for correlated batch")
	}
}

// TestLogger_SetChannel_LevelFiltering tests that channel writers respect log levels (filter Debug, keep Info+)
func TestLogger_SetChannel_LevelFiltering(t *testing.T) {
	logger := NewLogger()
	logChan := make(chan []models.LogEvent, 10)
	channelName := "test-channel-level"

	// Register channel
	logger.SetChannel(channelName, logChan)
	defer logger.UnregisterChannel(channelName)

	// Get the writer and set level to Info (should filter Debug)
	writer := GetRegisteredWriter(channelName)
	require.NotNil(t, writer, "Channel writer should be registered")
	writer.WithLevel(log.InfoLevel)

	// Log messages at different levels
	logger.Debug().Msg("Debug message - should be filtered")
	logger.Info().Msg("Info message - should pass")
	logger.Warn().Msg("Warn message - should pass")
	logger.Error().Msg("Error message - should pass")

	// Add one more to trigger batch (total 4 logged, but Debug filtered)
	logger.Info().Msg("Extra info message")

	// Wait for batch
	select {
	case batch := <-logChan:
		require.NotEmpty(t, batch, "Batch should contain events")

		// Verify no Debug level messages
		for _, event := range batch {
			assert.NotEqual(t, log.DebugLevel, event.Level, "Debug messages should be filtered")
			assert.GreaterOrEqual(t, event.Level, log.InfoLevel, "Only Info+ messages should pass")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for filtered batch")
	}
}

// TestLogger_UnregisterChannel tests cleanup functionality
func TestLogger_UnregisterChannel(t *testing.T) {
	logger := NewLogger()
	logChan := make(chan []models.LogEvent, 10)
	channelName := "test-channel-unregister"

	// Register channel
	logger.SetChannel(channelName, logChan)

	// Verify writer is registered
	writer := GetRegisteredWriter(channelName)
	require.NotNil(t, writer, "Channel writer should be registered")

	// Log a message
	logger.Info().Msg("Message before unregister")

	// Unregister the channel
	logger.UnregisterChannel(channelName)

	// Verify writer is removed from registry
	writer = GetRegisteredWriter(channelName)
	assert.Nil(t, writer, "Channel writer should be unregistered")

	// Log another message (should not be sent to unregistered channel)
	logger.Info().Msg("Message after unregister")

	// The channel might receive one batch from before unregister, but not the second message
	select {
	case <-logChan:
		// This is fine - cleanup may allow one final batch
	case <-time.After(500 * time.Millisecond):
		// Also fine - channel properly cleaned up
	}

	// Verify no more messages come through
	select {
	case batch := <-logChan:
		// Check if this is an old batch or contains the "after unregister" message
		for _, event := range batch {
			assert.NotContains(t, event.Message, "Message after unregister", "Unregistered channel should not receive new messages")
		}
	case <-time.After(500 * time.Millisecond):
		// Expected - no messages should come through
	}
}

// TestLogger_SetChannel_ReplaceExisting tests replacing an existing channel writer
func TestLogger_SetChannel_ReplaceExisting(t *testing.T) {
	logger := NewLogger()
	logChan1 := make(chan []models.LogEvent, 10)
	logChan2 := make(chan []models.LogEvent, 10)
	channelName := "test-channel-replace"

	// Register first channel
	logger.SetChannel(channelName, logChan1)

	// Verify first writer is registered
	writer1 := GetRegisteredWriter(channelName)
	require.NotNil(t, writer1, "First channel writer should be registered")

	// Replace with second channel (same name)
	logger.SetChannel(channelName, logChan2)
	defer logger.UnregisterChannel(channelName)

	// Verify writer is still registered (but replaced)
	writer2 := GetRegisteredWriter(channelName)
	require.NotNil(t, writer2, "Second channel writer should be registered")

	// Log messages
	for i := 0; i < 5; i++ {
		logger.Info().Msgf("Replacement message %d", i)
	}

	// Only the second channel should receive messages
	receivedOnChan2 := false
	select {
	case batch := <-logChan2:
		require.NotEmpty(t, batch, "Channel 2 should receive batch")
		receivedOnChan2 = true
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for replacement channel batch")
	}

	assert.True(t, receivedOnChan2, "Replacement channel should receive messages")

	// Original channel should not receive new messages
	select {
	case <-logChan1:
		t.Error("Original channel should not receive new messages after replacement")
	case <-time.After(500 * time.Millisecond):
		// Expected - original channel should not receive messages
	}
}

// TestLogger_SetChannel_NilChannel tests error handling for nil channel (should panic)
func TestLogger_SetChannel_NilChannel(t *testing.T) {
	logger := NewLogger()
	channelName := "test-channel-nil"

	// Attempting to register nil channel should panic
	assert.Panics(t, func() {
		logger.SetChannel(channelName, nil)
	}, "SetChannel with nil channel should panic")
}

// TestLogger_SetChannelWithBuffer_InvalidParameters tests parameter validation with zero/negative values use defaults
func TestLogger_SetChannelWithBuffer_InvalidParameters(t *testing.T) {
	logger := NewLogger()
	logChan := make(chan []models.LogEvent, 10)
	channelName := "test-channel-invalid-params"

	// Register channel with invalid parameters (0 batch size, negative interval)
	// Should use defaults: batch size 5, flush interval 1 second
	logger.SetChannelWithBuffer(channelName, logChan, 0, -1*time.Second)
	defer logger.UnregisterChannel(channelName)

	// Verify writer is registered (should succeed with defaults)
	writer := GetRegisteredWriter(channelName)
	require.NotNil(t, writer, "Channel writer should be registered with default parameters")

	// Log messages to verify it works
	for i := 0; i < 5; i++ {
		logger.Info().Msgf("Invalid param message %d", i)
	}

	// Should receive batch (using default batch size of 5)
	select {
	case batch := <-logChan:
		assert.NotEmpty(t, batch, "Batch should be received with default parameters")
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for batch with default parameters")
	}
}

// TestLogger_SetChannel_ConcurrentWrites tests 10 goroutines each logging 5 messages, verify 50 total
func TestLogger_SetChannel_ConcurrentWrites(t *testing.T) {
	logger := NewLogger()
	logChan := make(chan []models.LogEvent, 100)
	channelName := "test-channel-concurrent"

	// Register channel with small batch size for faster flushing
	logger.SetChannelWithBuffer(channelName, logChan, 5, 100*time.Millisecond)
	defer logger.UnregisterChannel(channelName)

	// Verify writer is registered
	writer := GetRegisteredWriter(channelName)
	require.NotNil(t, writer, "Channel writer should be registered")

	const numGoroutines = 10
	const messagesPerGoroutine = 5
	const totalExpectedMessages = numGoroutines * messagesPerGoroutine

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent writers
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()
			for m := 0; m < messagesPerGoroutine; m++ {
				logger.Info().Msgf("Concurrent message from goroutine %d, message %d", goroutineID, m)
			}
		}(g)
	}

	// Wait for all goroutines to finish logging
	wg.Wait()

	// Collect all batches with timeout
	totalReceived := 0
	timeout := time.After(5 * time.Second)

	for totalReceived < totalExpectedMessages {
		select {
		case batch := <-logChan:
			totalReceived += len(batch)
			t.Logf("Received batch of %d events, total: %d/%d", len(batch), totalReceived, totalExpectedMessages)
		case <-timeout:
			t.Fatalf("Timeout waiting for all messages. Received %d/%d", totalReceived, totalExpectedMessages)
		}
	}

	// Verify we received all expected messages
	assert.Equal(t, totalExpectedMessages, totalReceived, "Should receive all 50 messages from concurrent writes")
}
