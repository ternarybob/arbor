package writers

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

type goroutineWriter struct {
	config     models.WriterConfiguration
	configMux  sync.RWMutex
	buffer     chan models.LogEvent
	bufferSize int
	done       chan struct{}
	processor  func(models.LogEvent) error
	running    bool
	runningMux sync.RWMutex
	wg         sync.WaitGroup
	closeOnce  sync.Once
}

func NewGoroutineWriter(config models.WriterConfiguration, bufferSize int, processor func(models.LogEvent) error) (IGoroutineWriter, error) {
	if processor == nil {
		return nil, errors.New("processor function cannot be nil")
	}

	if bufferSize <= 0 {
		bufferSize = 1000
	}

	return &goroutineWriter{
		config:     config,
		buffer:     make(chan models.LogEvent, bufferSize),
		bufferSize: bufferSize,
		done:       make(chan struct{}),
		processor:  processor,
		running:    false,
	}, nil
}

// newAsyncWriter creates and starts a goroutine writer with the given configuration and processor.
// This is a helper to reduce duplication in async writer factories (LogStoreWriter, ContextWriter).
// Returns an error if creation or starting fails - caller should handle appropriately.
func newAsyncWriter(config models.WriterConfiguration, bufferSize int, processor func(models.LogEvent) error) (IGoroutineWriter, error) {
	// Create goroutine writer
	writer, err := NewGoroutineWriter(config, bufferSize, processor)
	if err != nil {
		return nil, err
	}

	// Start the goroutine for backward compatibility
	if err := writer.Start(); err != nil {
		return nil, err
	}

	return writer, nil
}

func (gw *goroutineWriter) Start() error {
	gw.runningMux.Lock()
	defer gw.runningMux.Unlock()

	if gw.running {
		return errors.New("goroutine writer is already running")
	}

	gw.done = make(chan struct{})
	gw.wg.Add(1)
	go gw.processBuffer()
	gw.running = true

	return nil
}

func (gw *goroutineWriter) Stop() error {
	gw.runningMux.Lock()
	defer gw.runningMux.Unlock()

	if !gw.running {
		return nil
	}

	gw.running = false
	close(gw.done)
	gw.wg.Wait()

	return nil
}

func (gw *goroutineWriter) processBuffer() {
	defer gw.wg.Done()

	internalLog := common.NewLogger().WithContext("function", "goroutineWriter.processBuffer").GetLogger()

	for {
		select {
		case entry := <-gw.buffer:
			if err := gw.processor(entry); err != nil {
				internalLog.Warn().Err(err).Msg("Failed to process log entry")
			}
		case <-gw.done:
			for {
				select {
				case entry := <-gw.buffer:
					if err := gw.processor(entry); err != nil {
						internalLog.Warn().Err(err).Msg("Failed to process log entry during shutdown")
					}
				default:
					return
				}
			}
		}
	}
}

func (gw *goroutineWriter) Write(data []byte) (int, error) {
	if !gw.IsRunning() {
		return len(data), nil
	}

	internalLog := common.NewLogger().WithContext("function", "goroutineWriter.Write").GetLogger()

	n := len(data)
	if n == 0 {
		return n, nil
	}

	var logEvent models.LogEvent
	if err := json.Unmarshal(data, &logEvent); err != nil {
		return 0, err
	}

	gw.configMux.RLock()
	minLevel := gw.config.Level.ToLogLevel()
	gw.configMux.RUnlock()

	if logEvent.Level < minLevel {
		return n, nil
	}

	select {
	case gw.buffer <- logEvent:
		return n, nil
	default:
		internalLog.Warn().Msg("Goroutine writer buffer full, dropping entry")
		return n, nil
	}
}

func (gw *goroutineWriter) WithLevel(level log.Level) IWriter {
	gw.configMux.Lock()
	gw.config.Level = levels.FromLogLevel(level)
	gw.configMux.Unlock()
	return gw
}

func (gw *goroutineWriter) GetFilePath() string {
	return ""
}

func (gw *goroutineWriter) Close() error {
	gw.Stop()
	gw.closeOnce.Do(func() {
		close(gw.buffer)
	})
	return nil
}

func (gw *goroutineWriter) IsRunning() bool {
	gw.runningMux.RLock()
	defer gw.runningMux.RUnlock()
	return gw.running
}
