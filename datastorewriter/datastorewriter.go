package datastorewriter

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	consolewriter "github.com/ternarybob/arbor/consolewriter"

	bolt "go.etcd.io/bbolt"

	"github.com/rs/zerolog"
)

type DataStoreWriter struct {
	Out io.Writer
}

type LogEvent struct {
	Index         uint64    `json:"index"`
	Level         string    `json:"level"`
	Timestamp     time.Time `json:"time"`
	CorrelationID string    `json:"correlationid"`
	Prefix        string    `json:"prefix"`
	Message       string    `json:"message"`
	Error         string    `json:"error"`
}

const (
	CORRELATIONID_KEY = "CORRELATIONID"
	BUFFERLIMIT       = 1024
	FILEMODE          = 0600
	FILEPATH          = "./request.db"
)

var (
	loglevel    zerolog.Level  = zerolog.InfoLevel
	internallog zerolog.Logger = zerolog.New(&consolewriter.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger().Level(loglevel)
	db          *bolt.DB
)

func New() *DataStoreWriter {

	var (
		log zerolog.Logger = internallog.With().Str("prefix", "New").Logger()
		err error
	)

	if err = openDB(FILEPATH); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	return &DataStoreWriter{
		Out: os.Stdout,
	}

}

func (w *DataStoreWriter) Write(entry []byte) (int, error) {

	log := internallog.With().Str("prefix", "Write").Logger()
	ep := len(entry)

	if ep == 0 {
		return ep, nil
	}

	err := w.writeline(entry)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}

	return ep, nil
}

func (w *DataStoreWriter) writeline(event []byte) error {

	var (
		log zerolog.Logger = internallog.With().Str("prefix", "writeline").Logger()
		err error
	)

	if len(event) <= 0 {
		return fmt.Errorf("Event is Empty")
	}

	if err = openDB(FILEPATH); err != nil {
		return err
	}

	var logentry LogEvent

	if err := json.Unmarshal(event, &logentry); err != nil {

		log.Warn().Err(err).Msgf("Error:%s Event:%s", err.Error(), string(event))

		return err
	}

	if isEmpty(logentry.CorrelationID) {
		log.Warn().Err(err).Msgf("CorrelationID is empty -> no write to db")
		return nil
	}

	precnt := 0
	postcnt := 0

	err = db.Update(func(tx *bolt.Tx) (err error) {

		// b, err := tx.CreateBucketIfNotExists([]byte(logentry.CorrelationID))
		b, err := tx.CreateBucketIfNotExists([]byte(logentry.CorrelationID))
		if err != nil {
			return err
		}

		precnt = int(b.Sequence())

		id, _ := b.NextSequence()

		logentry.Index = id

		input, err := json.Marshal(logentry)
		if err != nil {
			log.Warn().Err(err).Msgf("Error when marshalling log entry event:%s", string(event))
			return err
		}

		err = b.Put(itob(id), input)

		postcnt = int(b.Sequence())

		return err

	})
	if err != nil {
		log.Warn().Err(err).Msg("")
		return nil
	}

	log.Trace().Msgf("CorrelationID:%s -> message:%s", logentry.CorrelationID, logentry.Message)
	log.Trace().Msgf("CorrelationID:%s -> added entries to db (%d->%d)", logentry.CorrelationID, precnt, postcnt)

	return nil

}

func GetEntries(correlationid string) (map[string]string, error) {

	var (
		log     zerolog.Logger    = internallog.With().Str("prefix", "GetEntries").Logger()
		entries map[string]string = make(map[string]string)
		err     error
	)

	if correlationid == "" {
		return entries, fmt.Errorf("requestid is nil")
	}

	if err = openDB(FILEPATH); err != nil {
		return entries, err
	}

	log.Debug().Msgf("Getting log entries correlationid:%s", correlationid)

	err = db.View(func(tx *bolt.Tx) error {

		bkt := tx.Bucket([]byte(correlationid))

		if bkt == nil {
			log.Debug().Msgf("No log entries found for correlationid:%s", correlationid)
			return nil
		}

		c := bkt.Cursor()

		log.Debug().Msgf("Bucket entries for requestid:%s cnt:%d", correlationid, c.Bucket().Sequence())

		for k, v := c.First(); k != nil; k, v = c.Next() {

			var logentry LogEvent

			if err := json.Unmarshal(v, &logentry); err != nil {

				log.Warn().Err(err).Msg("")

				continue
			}

			index := fmt.Sprintf("%03d", logentry.Index)

			entries[index] = logentry.format()
		}
		return nil

	})

	return entries, err
}

func (l *LogEvent) format() string {

	epoch := l.Timestamp.Format(time.Stamp)

	_level := Levels[parselevel(l.Level)]

	output := fmt.Sprintf("%s|%s", _level.ShortName, epoch)

	// Excluding as is redundant
	/*
		if l.CorrelationID != "" {
			output += fmt.Sprintf("|%s", l.CorrelationID)
		}
	*/

	if l.Prefix != "" {
		output += fmt.Sprintf("|%-65s", l.Prefix)
	}

	if l.Message != "" {
		output += fmt.Sprintf("|%s", l.Message)
	}

	if l.Error != "" {
		output += fmt.Sprintf("|%s", l.Error)
	}

	return output
}

func openDB(path string) error {

	var (
		log zerolog.Logger = internallog.With().Str("prefix", "openDB").Logger()
		err error
	)

	if db != nil {
		log.Debug().Msg("Datatabse is open")
		return nil
	}

	log.Trace().Msgf("Opening bolt.DB path:%s", path)

	if db, err = bolt.Open(path, FILEMODE, &bolt.Options{Timeout: 20 * time.Second}); err != nil {
		return err
	}

	log.Trace().Msg("BoltDB Opened")

	runtime.SetFinalizer(db, closeDB)

	return nil

}

func closeDB(db *bolt.DB) error {

	log := internallog.With().Str("prefix", "closeDB").Logger()

	log.Debug().Msg("Closing Database")

	if err := db.Close(); err != nil {
		log.Warn().Err(err).Msg("BoltDB did not close")
		return err
	}

	return nil
}

func itob(v uint64) []byte {

	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, uint64(v))

	return b

}

func isEmpty(a string) bool {

	return len(strings.TrimSpace(a)) == 0

}
