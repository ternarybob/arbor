// -----------------------------------------------------------------------
// Last Modified: Wednesday, 1st October 2025 4:20:00 pm
// Modified By: Bob McAllan
// -----------------------------------------------------------------------

package writers

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/phuslu/log"
	"github.com/ternarybob/arbor/common"
	"github.com/ternarybob/arbor/levels"
	"github.com/ternarybob/arbor/models"
)

// WebSocketClient represents a connected WebSocket client
type WebSocketClient interface {
	SendJSON(data interface{}) error
	Close() error
}

// websocketWriter polls the log store and broadcasts to WebSocket clients
type websocketWriter struct {
	store        ILogStore
	config       models.WriterConfiguration
	clients      map[string]WebSocketClient // clientID -> client
	clientsMux   sync.RWMutex
	lastSent     time.Time
	pollInterval time.Duration
	stopPoll     chan bool
}

// WebSocketWriter creates a new WebSocket writer that polls the log store
func WebSocketWriter(store ILogStore, config models.WriterConfiguration, pollInterval time.Duration) IWriter {
	internalLog := common.NewLogger().WithContext("function", "WebSocketWriter").GetLogger()

	if pollInterval == 0 {
		pollInterval = 500 * time.Millisecond // Default poll interval
	}

	wsw := &websocketWriter{
		store:        store,
		config:       config,
		clients:      make(map[string]WebSocketClient),
		pollInterval: pollInterval,
		lastSent:     time.Now(),
		stopPoll:     make(chan bool),
	}

	// Start polling for new logs
	go wsw.pollAndBroadcast()

	internalLog.Info().Msgf("WebSocket writer started with %v poll interval", pollInterval)

	return wsw
}

// AddClient registers a new WebSocket client
func (wsw *websocketWriter) AddClient(clientID string, client WebSocketClient) {
	wsw.clientsMux.Lock()
	wsw.clients[clientID] = client
	wsw.clientsMux.Unlock()

	internalLog := common.NewLogger().WithContext("function", "WebSocketWriter.AddClient").GetLogger()
	internalLog.Info().Msgf("WebSocket client added: %s (total: %d)", clientID, len(wsw.clients))
}

// RemoveClient unregisters a WebSocket client
func (wsw *websocketWriter) RemoveClient(clientID string) {
	wsw.clientsMux.Lock()
	if client, exists := wsw.clients[clientID]; exists {
		client.Close()
		delete(wsw.clients, clientID)
	}
	wsw.clientsMux.Unlock()

	internalLog := common.NewLogger().WithContext("function", "WebSocketWriter.RemoveClient").GetLogger()
	internalLog.Info().Msgf("WebSocket client removed: %s (total: %d)", clientID, len(wsw.clients))
}

// pollAndBroadcast periodically checks for new logs and broadcasts to clients
func (wsw *websocketWriter) pollAndBroadcast() {
	internalLog := common.NewLogger().WithContext("function", "WebSocketWriter.pollAndBroadcast").GetLogger()

	ticker := time.NewTicker(wsw.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wsw.broadcastNewLogs()
		case <-wsw.stopPoll:
			internalLog.Info().Msg("WebSocket polling stopped")
			return
		}
	}
}

// broadcastNewLogs sends new logs to all connected clients
func (wsw *websocketWriter) broadcastNewLogs() {
	internalLog := common.NewLogger().WithContext("function", "WebSocketWriter.broadcastNewLogs").GetLogger()

	// Get logs since last poll
	newLogs, err := wsw.store.GetSince(wsw.lastSent)
	if err != nil {
		internalLog.Warn().Err(err).Msg("Failed to retrieve new logs")
		return
	}

	if len(newLogs) == 0 {
		return // No new logs
	}

	wsw.lastSent = time.Now()

	// Broadcast to all connected clients
	wsw.clientsMux.RLock()
	clientCount := len(wsw.clients)
	clients := make(map[string]WebSocketClient, clientCount)
	for id, client := range wsw.clients {
		clients[id] = client
	}
	wsw.clientsMux.RUnlock()

	if clientCount == 0 {
		return // No clients connected
	}

	// Send to each client in goroutines
	for clientID, client := range clients {
		go func(id string, c WebSocketClient, logs []models.LogEvent) {
			if err := c.SendJSON(logs); err != nil {
				internalLog.Warn().Err(err).Msgf("Failed to send to client %s", id)
				// Remove failed client
				wsw.RemoveClient(id)
			}
		}(clientID, client, newLogs)
	}

	internalLog.Debug().Msgf("Broadcasted %d log entries to %d clients", len(newLogs), clientCount)
}

// GetLogsSince allows clients to request logs from a specific timestamp
func (wsw *websocketWriter) GetLogsSince(since time.Time) ([]models.LogEvent, error) {
	return wsw.store.GetSince(since)
}

// GetLogsByCorrelation allows clients to request logs for a specific correlation ID
func (wsw *websocketWriter) GetLogsByCorrelation(correlationID string) ([]models.LogEvent, error) {
	return wsw.store.GetByCorrelation(correlationID)
}

// Write is a no-op for WebSocket writer (reads from store, doesn't write)
func (wsw *websocketWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

// WithLevel sets the minimum log level
func (wsw *websocketWriter) WithLevel(level log.Level) IWriter {
	wsw.config.Level = levels.FromLogLevel(level)
	return wsw
}

// GetFilePath returns empty string as WebSocket writer doesn't write to files
func (wsw *websocketWriter) GetFilePath() string {
	return ""
}

// Close shuts down the WebSocket writer
func (wsw *websocketWriter) Close() error {
	close(wsw.stopPoll)

	// Close all clients
	wsw.clientsMux.Lock()
	for _, client := range wsw.clients {
		client.Close()
	}
	wsw.clients = make(map[string]WebSocketClient)
	wsw.clientsMux.Unlock()

	return nil
}

// SimpleWebSocketClient is a basic implementation for testing
type SimpleWebSocketClient struct {
	sendFunc  func(data interface{}) error
	closeFunc func() error
}

func NewSimpleWebSocketClient(sendFunc func(data interface{}) error, closeFunc func() error) WebSocketClient {
	return &SimpleWebSocketClient{
		sendFunc:  sendFunc,
		closeFunc: closeFunc,
	}
}

func (c *SimpleWebSocketClient) SendJSON(data interface{}) error {
	if c.sendFunc != nil {
		return c.sendFunc(data)
	}
	// Default: marshal to JSON (for testing)
	_, err := json.Marshal(data)
	return err
}

func (c *SimpleWebSocketClient) Close() error {
	if c.closeFunc != nil {
		return c.closeFunc()
	}
	return nil
}
