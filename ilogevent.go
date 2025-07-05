package arbor

// ILogEvent represents a fluent interface for building log events
type ILogEvent interface {
	// String field methods
	Str(key, value string) ILogEvent
	
	// Error field method
	Err(err error) ILogEvent
	
	// Message methods
	Msg(message string)
	Msgf(format string, args ...interface{})
}
