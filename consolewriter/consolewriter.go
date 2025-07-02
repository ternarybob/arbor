package consolewriter

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

type ConsoleWriter struct {
	Out io.Writer
}

type LogEvent struct {
	Level         string    `json:"level"`
	Timestamp     time.Time `json:"time"`
	Prefix        string    `json:"prefix"`
	CorrelationID string    `json:"correlationid"`
	Message       string    `json:"message"`
	Error         string    `json:"error"`
}

func New() *ConsoleWriter {

	return &ConsoleWriter{
		Out: os.Stdout,
	}

}

func (w *ConsoleWriter) Write(e []byte) (n int, err error) {

	n = len(e)
	if n <= 0 {
		return n, err
	}

	// fmt.Printf("%d", n)

	err = w.writeline(e)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (w *ConsoleWriter) writeline(event []byte) error {

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

func (w *ConsoleWriter) format(l *LogEvent, colour bool) string {

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
