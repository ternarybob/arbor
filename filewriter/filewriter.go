package filewriter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	loglevel    zerolog.Level  = zerolog.WarnLevel
	internallog zerolog.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(loglevel)
)

type WriteTask struct {
	data []byte
}

type FileWriter struct {
	mutex sync.Mutex
	file  *os.File
	queue chan WriteTask
	wg    sync.WaitGroup
	err   error
	Log   []string
}

type LogEvent struct {
	Level         string    `json:"level"`
	Timestamp     time.Time `json:"time"`
	Prefix        string    `json:"prefix"`
	CorrelationID string    `json:"correlationid"`
	Message       string    `json:"message"`
	Error         string    `json:"error"`
}

func New(file *os.File, bufferSize int) *FileWriter {

	// fileWriter, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)

	w := &FileWriter{
		file:  file,
		queue: make(chan WriteTask, bufferSize),
		wg:    sync.WaitGroup{},
	}
	w.wg.Add(1)
	go w.writeLoop()
	return w

}

// NewWithPath creates a new FileWriter with the specified file path, creating directories if needed
// It also implements file rotation based on the specified max number of log files.
func NewWithPath(filePath string, bufferSize, maxFiles int) (*FileWriter, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	fw := New(file, bufferSize)
	fw.rotateFiles(filePath, maxFiles)

	return fw, nil
}

// rotateFiles rotates the log files to ensure no more than maxFiles are kept
func (w *FileWriter) rotateFiles(filePath string, maxFiles int) {
	// File rotation logic here to delete older files when exceeding maxFiles
	files, err := filepath.Glob(filePath + "*")
	if err != nil {
		fmt.Println("Error fetching log files for rotation:", err)
		return
	}

	// Ensure files are sorted, oldest first
	sort.Strings(files)

	// Remove old log files
	for len(files) > maxFiles {
		os.Remove(files[0])
		files = files[1:]
	}
}

func (w *FileWriter) Write(e []byte) (n int, err error) {

	n = len(e)
	if n <= 0 {
		return n, err
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.err != nil {
		return // Silently ignore write if already in error state
	}

	select {
	case w.queue <- WriteTask{data: e}:
	default:
		fmt.Println("Write queue full, data dropped:", string(e))
	}

	return n, nil
}

func (w *FileWriter) writeline(event []byte) error {

	log := internallog.With().Str("prefix", "writeline").Logger()

	if len(event) <= 0 {
		log.Warn().Msg("Entry is Empty")
		return fmt.Errorf("Entry is Empty")
	}

	var logentry LogEvent

	if err := json.Unmarshal(event, &logentry); err != nil {

		log.Warn().Err(err).Msgf("error:%s entry:%s", err.Error(), string(event))

		return err
	}

	_, err := fmt.Printf("%s\n", w.format(&logentry, true))
	if err != nil {
		return err
	}

	return nil
}

// Close waits for pending writes and closes the writer
func (w *FileWriter) Close() error {
	close(w.queue)
	w.wg.Wait()
	return w.err
}

// writeLoop continuously processes tasks from the queue and writes to the file
func (w *FileWriter) writeLoop() {
	defer w.wg.Done()
	for task := range w.queue {
		_, err := w.file.Write(task.data)

		if err != nil {
			fmt.Println("Write error:", err)
			w.mutex.Lock()
			w.err = err // Update error state if desired for future writes
			w.mutex.Unlock()
			// No retry logic, continue processing remaining tasks
		}
	}
}

func (w *FileWriter) format(l *LogEvent, colour bool) string {

	timestamp := l.Timestamp.Format(time.Stamp)

	output := fmt.Sprintf("%s|%s", levelprint(l.Level, colour), timestamp)

	if l.CorrelationID != "" {
		output += fmt.Sprintf("|%s", l.CorrelationID)
	}

	if l.Prefix != "" {
		output += fmt.Sprintf("|%-55s", l.Prefix)
	}

	if l.Message != "" {
		output += fmt.Sprintf("|%s", l.Message)
	}

	if l.Error != "" {
		output += fmt.Sprintf("|%s", l.Error)
	}

	return output
}
