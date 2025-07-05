package interfaces

import "io"

// IWriter defines the interface that all log writers must implement
type IWriter interface {
	// Write writes log data and returns the number of bytes written and any error
	// This embeds the standard io.Writer interface
	io.Writer

	// Close gracefully shuts down the writer and releases any resources
	// Writers should implement this to ensure proper cleanup
	Close() error
}

// ILevelWriter defines an extended interface for writers that support log level filtering
type ILevelWriter interface {
	IWriter

	// SetMinLevel sets the minimum log level that this writer will process
	// Logs below this level should be ignored by the writer
	SetMinLevel(level interface{}) error
}

// IBufferedWriter defines an interface for writers that support buffering
type IBufferedWriter interface {
	IWriter

	// Flush forces any buffered data to be written immediately
	Flush() error

	// SetBufferSize changes the buffer size for the writer
	SetBufferSize(size int) error
}

// IRotatableWriter defines an interface for writers that support log rotation
type IRotatableWriter interface {
	IWriter

	// Rotate manually triggers a log rotation
	Rotate() error

	// SetMaxFiles sets the maximum number of files to keep during rotation
	SetMaxFiles(maxFiles int)
}

// IFullFeaturedWriter combines all writer interfaces for maximum functionality
type IFullFeaturedWriter interface {
	IWriter
	ILevelWriter
	IBufferedWriter
	IRotatableWriter
}
