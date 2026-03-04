package arbor

import (
	"testing"
	"time"

	"github.com/ternarybob/arbor/models"
	"github.com/ternarybob/arbor/writers"
)

func TestLogger_WithLogStore_RegistersWriter(t *testing.T) {
	// Create an InMemoryLogStore to act as the external ILogStore
	store, err := writers.NewInMemoryLogStore(models.WriterConfiguration{})
	if err != nil {
		t.Fatalf("Failed to create test log store: %v", err)
	}
	defer store.Close()

	// Create logger with the log store
	config := models.WriterConfiguration{
		Type: models.LogWriterTypeLogStore,
	}
	_ = NewLogger().WithLogStore(store, config)

	// Verify the writer was registered
	writer := GetRegisteredWriter(WRITER_LOGSTORE)
	if writer == nil {
		t.Fatal("Expected WRITER_LOGSTORE to be registered")
	}
}

func TestLogger_WithLogStore_LogsFlowToStore(t *testing.T) {
	store, err := writers.NewInMemoryLogStore(models.WriterConfiguration{})
	if err != nil {
		t.Fatalf("Failed to create test log store: %v", err)
	}
	defer store.Close()

	config := models.WriterConfiguration{
		Type: models.LogWriterTypeLogStore,
	}
	logger := NewLogger().
		WithLogStore(store, config).
		WithCorrelationId("logstore-test-1")

	// Log messages
	logger.Info().Msg("hello from logstore")
	logger.Warn().Msg("warning from logstore")

	// Allow async writes to flush
	time.Sleep(100 * time.Millisecond)

	// Verify entries landed in the store
	recent, err := store.GetRecent(10)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}

	if len(recent) < 2 {
		t.Errorf("Expected at least 2 entries in store, got %d", len(recent))
	}
}

func TestLogger_WithLogStore_CorrelationQuery(t *testing.T) {
	store, err := writers.NewInMemoryLogStore(models.WriterConfiguration{})
	if err != nil {
		t.Fatalf("Failed to create test log store: %v", err)
	}
	defer store.Close()

	config := models.WriterConfiguration{
		Type: models.LogWriterTypeLogStore,
	}
	logger := NewLogger().WithLogStore(store, config)

	cid := "corr-logstore-42"
	logger.WithCorrelationId(cid).Info().Msg("correlated message")
	logger.WithCorrelationId("other-id").Info().Msg("different correlation")

	time.Sleep(100 * time.Millisecond)

	entries, err := store.GetByCorrelation(cid)
	if err != nil {
		t.Fatalf("GetByCorrelation failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for correlation %s, got %d", cid, len(entries))
	}
}
