package logviewer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/phuslu/log"
	"github.com/stretchr/testify/assert"
	"github.com/ternarybob/arbor/models"
)

func TestLogViewerService(t *testing.T) {
	// Setup temporary log directory
	tempDir, err := os.MkdirTemp("", "arbor-logviewer-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create some dummy log files
	file1Path := filepath.Join(tempDir, "test1.log")
	// Create a log entry
	entry1 := models.LogEvent{
		Level:     log.InfoLevel,
		Message:   "content1",
		Timestamp: time.Now(),
	}
	data1, _ := json.Marshal(entry1)
	err = os.WriteFile(file1Path, append(data1, '\n'), 0644)
	assert.NoError(t, err)

	// Ensure file1 is older
	oldTime := time.Now().Add(-1 * time.Hour)
	os.Chtimes(file1Path, oldTime, oldTime)

	file2Path := filepath.Join(tempDir, "test2.log")
	entry2 := models.LogEvent{
		Level:     log.ErrorLevel,
		Message:   "content2",
		Timestamp: time.Now(),
	}
	data2, _ := json.Marshal(entry2)
	err = os.WriteFile(file2Path, append(data2, '\n'), 0644)
	assert.NoError(t, err)

	service := NewService(tempDir)

	t.Run("ListLogFiles", func(t *testing.T) {
		files, err := service.ListLogFiles()
		assert.NoError(t, err)
		assert.Len(t, files, 2)

		// Check sorting (newest first)
		assert.Equal(t, "test2.log", files[0].Name)
		assert.Equal(t, "test1.log", files[1].Name)
	})

	t.Run("GetLogContent", func(t *testing.T) {
		entries, err := service.GetLogContent("test1.log", 0, nil)
		assert.NoError(t, err)
		assert.Len(t, entries, 1)
		assert.Equal(t, "content1", entries[0].Message)
	})

	t.Run("GetLogContent_Filter", func(t *testing.T) {
		// test2.log has Error level
		entries, err := service.GetLogContent("test2.log", 0, []string{"error"})
		assert.NoError(t, err)
		assert.Len(t, entries, 1)
		assert.Equal(t, "content2", entries[0].Message)

		// Filter mismatch
		entries, err = service.GetLogContent("test2.log", 0, []string{"info"})
		assert.NoError(t, err)
		assert.Len(t, entries, 0)
	})

	t.Run("GetLogContent_NotFound", func(t *testing.T) {
		_, err := service.GetLogContent("nonexistent.log", 0, nil)
		assert.Error(t, err)
		assert.Equal(t, "file not found", err.Error())
	})

	t.Run("GetLogContent_TraversalAttempt", func(t *testing.T) {
		_, err := service.GetLogContent("../outside.log", 0, nil)
		assert.Error(t, err)
		assert.Equal(t, "invalid file name", err.Error())
	})
}
