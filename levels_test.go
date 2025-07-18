package arbor

import (
	"testing"

	"github.com/phuslu/log"
)

func TestLogLevel_Constants(t *testing.T) {
	// Test that our constants match phuslu/log constants
	testCases := []struct {
		arborLevel  LogLevel
		phusluLevel log.Level
		name        string
	}{
		{TraceLevel, log.TraceLevel, "Trace"},
		{DebugLevel, log.DebugLevel, "Debug"},
		{InfoLevel, log.InfoLevel, "Info"},
		{WarnLevel, log.WarnLevel, "Warn"},
		{ErrorLevel, log.ErrorLevel, "Error"},
		{FatalLevel, log.FatalLevel, "Fatal"},
		{PanicLevel, log.PanicLevel, "Panic"},
		{Disabled, 0, "Disabled"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if LogLevel(tc.phusluLevel) != tc.arborLevel {
				t.Errorf("Expected %s level %d, got %d", tc.name, tc.phusluLevel, tc.arborLevel)
			}
		})
	}
}

func TestParseLevelString_AllCases(t *testing.T) {
	testCases := []struct {
		input       string
		expected    log.Level
		expectError bool
	}{
		// Valid cases
		{"trace", log.TraceLevel, false},
		{"TRACE", log.TraceLevel, false},
		{"Trace", log.TraceLevel, false},
		{"debug", log.DebugLevel, false},
		{"DEBUG", log.DebugLevel, false},
		{"info", log.InfoLevel, false},
		{"INFO", log.InfoLevel, false},
		{"warn", log.WarnLevel, false},
		{"WARN", log.WarnLevel, false},
		{"warning", log.WarnLevel, false},
		{"WARNING", log.WarnLevel, false},
		{"error", log.ErrorLevel, false},
		{"ERROR", log.ErrorLevel, false},
		{"fatal", log.FatalLevel, false},
		{"FATAL", log.FatalLevel, false},
		{"panic", log.PanicLevel, false},
		{"PANIC", log.PanicLevel, false},
		{"disabled", log.PanicLevel + 1, false},
		{"DISABLED", log.PanicLevel + 1, false},
		{"off", log.PanicLevel + 1, false},
		{"OFF", log.PanicLevel + 1, false},

		// Invalid cases
		{"invalid", log.InfoLevel, true},
		{"unknown", log.InfoLevel, true},
		{"", log.InfoLevel, true},
		{"123", log.InfoLevel, true},
		{"trace123", log.InfoLevel, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := ParseLevelString(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s'", tc.input)
				}
				// For error cases, result should default to InfoLevel
				if result != tc.expected {
					t.Errorf("Expected default level %v for invalid input, got %v", tc.expected, result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("Expected level %v for input '%s', got %v", tc.expected, tc.input, result)
				}
			}
		})
	}
}

func TestParseLogLevel_AllCases(t *testing.T) {
	testCases := []struct {
		input    int
		expected log.Level
		name     string
	}{
		{int(TraceLevel), log.TraceLevel, "Trace"},
		{int(DebugLevel), log.DebugLevel, "Debug"},
		{int(InfoLevel), log.InfoLevel, "Info"},
		{int(WarnLevel), log.WarnLevel, "Warn"},
		{int(ErrorLevel), log.ErrorLevel, "Error"},
		{int(FatalLevel), log.FatalLevel, "Fatal"},
		{int(PanicLevel), log.PanicLevel, "Panic"},
		{int(Disabled), 0, "Disabled"},

		// Edge cases
		{-1, log.InfoLevel, "Negative"},
		{0, 0, "Zero"},
		{999, log.InfoLevel, "Large number"},
		{100, log.InfoLevel, "Unknown level"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseLogLevel(tc.input)
			if result != tc.expected {
				t.Errorf("Expected level %v for input %d, got %v", tc.expected, tc.input, result)
			}
		})
	}
}

func TestLogLevel_TypeDefinition(t *testing.T) {
	// Test that LogLevel is properly defined as uint32
	var level LogLevel = InfoLevel

	// Should be able to convert to int
	intLevel := int(level)
	if intLevel != int(log.InfoLevel) {
		t.Errorf("Expected int conversion to work, got %d", intLevel)
	}

	// Should be able to convert back
	backToLevel := LogLevel(intLevel)
	if backToLevel != level {
		t.Errorf("Expected round-trip conversion to work, got %v", backToLevel)
	}
}

func TestLogLevel_Comparison(t *testing.T) {
	// Test that log levels can be compared properly
	if TraceLevel >= DebugLevel {
		t.Error("TraceLevel should be less than DebugLevel")
	}

	if DebugLevel >= InfoLevel {
		t.Error("DebugLevel should be less than InfoLevel")
	}

	if InfoLevel >= WarnLevel {
		t.Error("InfoLevel should be less than WarnLevel")
	}

	if WarnLevel >= ErrorLevel {
		t.Error("WarnLevel should be less than ErrorLevel")
	}

	if ErrorLevel >= FatalLevel {
		t.Error("ErrorLevel should be less than FatalLevel")
	}

	if FatalLevel >= PanicLevel {
		t.Error("FatalLevel should be less than PanicLevel")
	}
}

func TestParseLevelString_CaseSensitivity(t *testing.T) {
	// Test that parsing is case-insensitive
	inputs := []string{"info", "INFO", "Info", "iNfO"}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			result, err := ParseLevelString(input)
			if err != nil {
				t.Errorf("Should not error for input '%s': %v", input, err)
			}
			if result != log.InfoLevel {
				t.Errorf("Expected InfoLevel for input '%s', got %v", input, result)
			}
		})
	}
}

func TestParseLevelString_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		input       string
		expectError bool
		description string
	}{
		{"", true, "empty string"},
		{" ", true, "whitespace only"},
		{"  info  ", true, "info with spaces"}, // This should fail as our function doesn't trim
		{"\t", true, "tab character"},
		{"\n", true, "newline character"},
		{"null", true, "null string"},
		{"undefined", true, "undefined string"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := ParseLevelString(tc.input)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s", tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.description, err)
			}
		})
	}
}

func TestLogLevel_DefaultBehavior(t *testing.T) {
	// Test default behavior for ParseLogLevel
	unknownInputs := []int{-999, 1000, 123456}

	for _, input := range unknownInputs {
		t.Run(string(rune(input)), func(t *testing.T) {
			result := ParseLogLevel(input)
			if result != log.InfoLevel {
				t.Errorf("Expected default InfoLevel for unknown input %d, got %v", input, result)
			}
		})
	}
}
