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

type channelWriter struct {
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

func NewChannelWriter(config models.WriterConfiguration, bufferSize int, processor func(models.LogEvent) error) (IChannelWriter, error) {
	if processor == nil {
		return nil, errors.New("processor function cannot be nil")
	}

	if bufferSize <= 0 {
		bufferSize = 1000
	}

	return &channelWriter{
		config:     config,
		buffer:     make(chan models.LogEvent, bufferSize),
		bufferSize: bufferSize,
		done:       make(chan struct{}),
		processor:  processor,
		running:    false,
	}, nil
}

// newAsyncWriter creates and starts a channel writer with the given configuration and processor.
// This is a helper to reduce duplication in async writer factories (LogStoreWriter, ContextWriter).
// Returns an error if creation or starting fails - caller should handle appropriately.
func newAsyncWriter(config models.WriterConfiguration, bufferSize int, processor func(models.LogEvent) error) (IChannelWriter, error) {
	// Create channel writer
	writer, err := NewChannelWriter(config, bufferSize, processor)
	if err != nil {
		return nil, err
	}

	// Start the goroutine for backward compatibility
	if err := writer.Start(); err != nil {
		return nil, err
	}

	return writer, nil
}

func (cw *channelWriter) Start() error {
	cw.runningMux.Lock()
	defer cw.runningMux.Unlock()

	if cw.running {
		return errors.New("channel writer is already running")
	}

	cw.done = make(chan struct{})
	cw.wg.Add(1)
	go cw.processBuffer()
	cw.running = true

	return nil
}

func (cw *channelWriter) Stop() error {
	cw.runningMux.Lock()
	defer cw.runningMux.Unlock()

	if !cw.running {
		return nil
	}

	cw.running = false
	close(cw.done)
	cw.wg.Wait()

	return nil
}

func (cw *channelWriter) processBuffer() {
	defer cw.wg.Done()

	internalLog := common.NewLogger().WithContext("function", "channelWriter.processBuffer").GetLogger()

	for {
		select {
		case entry := <-cw.buffer:
			if err := cw.processor(entry); err != nil {
				internalLog.Warn().Err(err).Msg("Failed to process log entry")
			}
		case <-cw.done:
			for {
				select {
				case entry := <-cw.buffer:
					if err := cw.processor(entry); err != nil {
						internalLog.Warn().Err(err).Msg("Failed to process log entry during shutdown")
					}
				default:
					return
				}
			}
		}
	}
}

func (cw *channelWriter) Write(data []byte) (int, error) {
	if !cw.IsRunning() {
		return len(data), nil
	}

	internalLog := common.NewLogger().WithContext("function", "channelWriter.Write").GetLogger()

	n := len(data)
	if n == 0 {
		return n, nil
	}

	var logEvent models.LogEvent
	if err := json.Unmarshal(data, &logEvent); err != nil {
		return 0, err
	}

	cw.configMux.RLock()
	minLevel := cw.config.Level.ToLogLevel()
	cw.configMux.RUnlock()

	if logEvent.Level < minLevel {
		return n, nil
	}

	select {
	case cw.buffer <- logEvent:
		return n, nil
	default:
		internalLog.Warn().Msg("Channel writer buffer full, dropping entry")
		return n, nil
	}
}

func (cw *channelWriter) WithLevel(level log.Level) IWriter {
	cw.configMux.Lock()
	cw.config.Level = levels.FromLogLevel(level)
	cw.configMux.Unlock()
	return cw
}

func (cw *channelWriter) GetFilePath() string {
	return ""
}

func (cw *channelWriter) Close() error {
	cw.Stop()
	cw.closeOnce.Do(func() {
		close(cw.buffer)
	})
	return nil
}

func (cw *channelWriter) IsRunning() bool {
	cw.runningMux.RLock()
	defer cw.runningMux.RUnlock()
	return cw.running
}
