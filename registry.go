package arbor

import (
	"sync"

	"github.com/ternarybob/arbor/writers"
)

const (
	WRITER_CONSOLE = "console"
	WRITER_FILE    = "file"
	WRITER_MEMORY  = "memory"
)

// WriterRegistry manages a collection of named writers with thread-safe access
// and implements the IWriterRegistry interface
type WriterRegistry struct {
	writers map[string]writers.IWriter
	mu      sync.RWMutex
}

// Ensure WriterRegistry implements IWriterRegistry
var _ IWriterRegistry = (*WriterRegistry)(nil)

// Global writer registry instance
var globalWriterRegistry = &WriterRegistry{
	writers: make(map[string]writers.IWriter),
}

// RegisterWriter registers a writer with the given name in the registry
func (wr *WriterRegistry) RegisterWriter(name string, writer writers.IWriter) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	wr.writers[name] = writer
}

// GetRegisteredWriter retrieves a writer by name from the registry
// Returns nil if the writer is not found
func (wr *WriterRegistry) GetRegisteredWriter(name string) writers.IWriter {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	return wr.writers[name]
}

// GetRegisteredMemoryWriter retrieves a memory writer by name from the registry
// Returns nil if the writer is not found or is not a memory writer
func (wr *WriterRegistry) GetRegisteredMemoryWriter(name string) writers.IMemoryWriter {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	writer := wr.writers[name]
	if memWriter, ok := writer.(writers.IMemoryWriter); ok {
		return memWriter
	}
	return nil
}

// GetRegisteredWriterNames returns a list of all registered writer names
func (wr *WriterRegistry) GetRegisteredWriterNames() []string {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	names := make([]string, 0, len(wr.writers))
	for name := range wr.writers {
		names = append(names, name)
	}
	return names
}

// UnregisterWriter removes a writer from the registry
func (wr *WriterRegistry) UnregisterWriter(name string) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	delete(wr.writers, name)
}

// GetWriterCount returns the number of registered writers
func (wr *WriterRegistry) GetWriterCount() int {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	return len(wr.writers)
}

// GetAllRegisteredWriters returns a copy of all registered writers
func (wr *WriterRegistry) GetAllRegisteredWriters() map[string]writers.IWriter {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	writersCopy := make(map[string]writers.IWriter)
	for name, writer := range wr.writers {
		writersCopy[name] = writer
	}
	return writersCopy
}

// RegisterWriter registers a writer with the given name in the global registry
func RegisterWriter(name string, writer writers.IWriter) {
	globalWriterRegistry.RegisterWriter(name, writer)
}

// GetRegisteredWriter retrieves a writer by name from the global registry
// Returns nil if the writer is not found
func GetRegisteredWriter(name string) writers.IWriter {
	return globalWriterRegistry.GetRegisteredWriter(name)
}

// GetRegisteredMemoryWriter retrieves a memory writer by name from the global registry
// Returns nil if the writer is not found or is not a memory writer
func GetRegisteredMemoryWriter(name string) writers.IMemoryWriter {
	return globalWriterRegistry.GetRegisteredMemoryWriter(name)
}

// GetRegisteredWriterNames returns a list of all registered writer names
func GetRegisteredWriterNames() []string {
	return globalWriterRegistry.GetRegisteredWriterNames()
}

// UnregisterWriter removes a writer from the global registry
func UnregisterWriter(name string) {
	globalWriterRegistry.UnregisterWriter(name)
}

// GetWriterCount returns the number of registered writers
func GetWriterCount() int {
	return globalWriterRegistry.GetWriterCount()
}

// GetAllRegisteredWriters returns a copy of all registered writers
func GetAllRegisteredWriters() map[string]writers.IWriter {
	return globalWriterRegistry.GetAllRegisteredWriters()
}
