package arbor

import "time"

// ILogEvent represents a fluent interface for building log events
type ILogEvent interface {
	// String slice field method
	Strs(key string, values []string) ILogEvent

	// String field methods
	Str(key, value string) ILogEvent

	// Error field method
	Err(err error) ILogEvent

	// Message methods
	Msg(message string)
	Msgf(format string, args ...interface{})

	// Integer field method
	Int(key string, value int) ILogEvent

	// Int32 field method
	Int32(key string, value int32) ILogEvent

	// Int64 field method
	Int64(key string, value int64) ILogEvent

	// Float32 field method
	Float32(key string, value float32) ILogEvent

	// Duration field method
	Dur(key string, value time.Duration) ILogEvent

	// Float64 field method
	Float64(key string, value float64) ILogEvent

	// Bool field method
	Bool(key string, value bool) ILogEvent
}
