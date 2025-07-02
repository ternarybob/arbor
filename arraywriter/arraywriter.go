package arraywriter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	loglevel    zerolog.Level  = zerolog.WarnLevel
	internallog zerolog.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(loglevel)
)

type ArrayWriter struct {
	Out io.Writer
	Log []string
}

type LogEvent struct {
	Level         string    `json:"level"`
	Timestamp     time.Time `json:"time"`
	Prefix        string    `json:"prefix"`
	CorrelationID string    `json:"correlationid"`
	Message       string    `json:"message"`
	Error         string    `json:"error"`
}

func New() *ArrayWriter {

	return &ArrayWriter{
		Out: os.Stdout,
	}

}

func (w *ArrayWriter) Write(e []byte) (n int, err error) {

	log := internallog.With().Str("prefix", "writeline").Logger()

	n = len(e)
	if n <= 0 {
		return n, err
	}

	err = w.writeline(e)
	if err != nil {
		log.Warn().Err(err).Msg("")
		return n, err
	}

	return n, nil
}

func (w *ArrayWriter) writeline(event []byte) error {

	if len(event) <= 0 {
		return fmt.Errorf("Entry is Empty")
	}

	var logentry LogEvent

	if err := json.Unmarshal(event, &logentry); err != nil {
		return err
	}

	w.Log = append(w.Log, w.format(&logentry))

	return nil
}

func (w *ArrayWriter) format(l *LogEvent) string {

	timestamp := l.Timestamp.Format(time.Stamp)
	level := Levels[parselevel(l.Level)]

	output := fmt.Sprintf("%s|%s", level.ShortName, timestamp)

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
