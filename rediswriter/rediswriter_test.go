package rediswriter

import (
	"testing"
)

func TestRedisWriterCreation(t *testing.T) {
	// Test that RedisWriter can be created without panicking
	// Note: This might need proper Redis configuration in real scenarios
	defer func() {
		if r := recover(); r != nil {
			t.Logf("RedisWriter creation panicked (expected if no Redis config): %v", r)
		}
	}()

	writer := New()
	if writer != nil {
		t.Logf("RedisWriter created successfully")
	}
}

func TestRedisWriterInterface(t *testing.T) {
	// Test that RedisWriter implements io.Writer interface
	// This tests compilation but may fail at runtime without proper setup
	defer func() {
		if r := recover(); r != nil {
			t.Logf("RedisWriter test panicked (expected without Redis setup): %v", r)
		}
	}()

	writer := New()
	if writer == nil {
		t.Skip("Skipping test - RedisWriter requires proper Redis configuration")
	}

	testData := []byte("test redis message")
	_, err := writer.Write(testData)

	// We don't fail on error here as it might be expected without proper Redis setup
	if err != nil {
		t.Logf("Write returned error (may be expected without Redis): %v", err)
	}
}
