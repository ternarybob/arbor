package arbor

import "github.com/ternarybob/arbor/writers"

// NewWriterRegistry creates a new instance of WriterRegistry
func NewWriterRegistry() IWriterRegistry {
	return &WriterRegistry{
		writers: make(map[string]writers.IWriter),
	}
}

// IWriterRegistry defines the interface for managing a collection of named writers
// with thread-safe access operations
type IWriterRegistry interface {
	// RegisterWriter registers a writer with the given name in the registry
	RegisterWriter(name string, writer writers.IWriter)

	// GetRegisteredWriter retrieves a writer by name from the registry
	// Returns nil if the writer is not found
	GetRegisteredWriter(name string) writers.IWriter

	// GetRegisteredMemoryWriter retrieves a memory writer by name from the registry
	// Returns nil if the writer is not found or is not a memory writer
	GetRegisteredMemoryWriter(name string) writers.IMemoryWriter

	// GetRegisteredWriterNames returns a list of all registered writer names
	GetRegisteredWriterNames() []string

	// UnregisterWriter removes a writer from the registry
	UnregisterWriter(name string)

	// GetWriterCount returns the number of registered writers
	GetWriterCount() int

	// GetAllRegisteredWriters returns a copy of all registered writers
	GetAllRegisteredWriters() map[string]writers.IWriter
}
