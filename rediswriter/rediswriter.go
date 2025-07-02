package rediswriter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	consolewriter "github.com/ternarybob/arbor/consolewriter"

	"github.com/ternarybob/funktion"
	"github.com/ternarybob/satus"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type redisWriter struct {
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
)

var (
	cfg         *satus.AppConfig = satus.GetAppConfig()
	internallog zerolog.Logger   = zerolog.New(&consolewriter.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger().Level(zerolog.WarnLevel)
	rdb         *redis.Client
)

func init() {

	var (
		ctx context.Context = context.Background()
		log zerolog.Logger  = internallog.With().Str("prefix", "init").Logger()
	)

	if len(cfg.Connections) <= 0 {
		log.Info().Msgf("Redis connection not configured -> returning")
		return
	}

	dataconfig, err := getDataConnection("redis", "redis")
	if err != nil {
		panic(err)
	}

	if funktion.IsEmpty(dataconfig.Name) {
		log.Info().Msgf("Redis connection name isEmpty -> returning")
		return
	}

	log.Info().Msgf("Redis connection found name:%s type:%s hosts:%s", dataconfig.Name, dataconfig.Type, strings.Join(dataconfig.Hosts, ","))

	options := &redis.Options{
		Addr:     dataconfig.Hosts[0],
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	rdb = redis.NewClient(options)

	if err := rdb.Ping(ctx).Err(); err != nil {
		internallog.Fatal().Err(err).Msg("Ping Error")
	}

}

func New() *redisWriter {

	var (
		ctx context.Context = context.Background()
		log zerolog.Logger  = internallog.With().Str("prefix", "New").Logger()
	)

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("Ping Error")
	}

	return &redisWriter{

		Out: os.Stdout,
	}

}

func (w *redisWriter) Write(entry []byte) (int, error) {

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

func (w *redisWriter) writeline(event []byte) error {

	var (
		log zerolog.Logger  = internallog.With().Str("prefix", "writeline").Logger()
		ctx context.Context = context.Background()
		err error
	)

	if len(event) <= 0 {
		return fmt.Errorf("Event is Empty")
	}

	var logevent LogEvent

	if err := json.Unmarshal(event, &logevent); err != nil {
		log.Warn().Err(err).Msgf("Error:%s Event:%s", err.Error(), string(event))
		return err
	}

	if funktion.IsEmpty(logevent.CorrelationID) {
		log.Warn().Err(err).Msgf("CorrelationID is empty -> no write to db")
		return nil
	}

	input, err := json.Marshal(logevent)
	if err != nil {
		log.Warn().Err(err).Msgf("Error when marshalling log entry event:%s", string(event))
		return err
	}

	if err := rdb.LPush(ctx, logevent.CorrelationID, input).Err(); err != nil {
		log.Warn().Err(err).Msgf("HSet Error")
		return err
	}

	log.Trace().Msgf("CorrelationID:%s -> message:%s", logevent.CorrelationID, logevent.Message)

	return nil

}

func GetEntries(correlationid string) (map[string]string, error) {

	var (
		log    zerolog.Logger    = internallog.With().Str("prefix", "GetEntries").Logger()
		ctx    context.Context   = context.Background()
		output map[string]string = make(map[string]string)
	)

	if rdb == nil {
		log.Warn().Msgf("redis connection not available/configured")
		return output, nil
	}

	if correlationid == "" {
		return output, fmt.Errorf("requestid is nil")
	}

	log.Debug().Msgf("Getting log entries correlationid:%s", correlationid)

	cntr := 0

	for rdb.LLen(ctx, correlationid).Val() > 0 {

		logevent := &LogEvent{}

		event, err := rdb.RPop(ctx, correlationid).Bytes()
		if err != nil {
			log.Warn().Err(err).Msgf("HSet Error")
			return output, err
		}

		if err := json.Unmarshal(event, &logevent); err != nil {
			log.Warn().Err(err).Msg("")
			continue
		}

		index := fmt.Sprintf("%03d", cntr)

		output[index] = logevent.format()

		cntr++

	}

	return output, nil
}

func (l *LogEvent) format() string {

	timestamp := l.Timestamp.Format(time.Stamp)

	_level := Levels[parselevel(l.Level)]

	output := fmt.Sprintf("%s|%s", _level.ShortName, timestamp)

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

func getDataConnection(connectionType string, connectionName string) (satus.DataConfig, error) {

	var (
		output satus.DataConfig
		log    zerolog.Logger = internallog.With().Str("prefix", "getDataConnection").Logger()
	)

	output, err := cfg.GetScopedDataConnectionbyType(connectionType)
	if err != nil {
		log.Warn().Err(err).Msgf("Database (%s) connection '%s' not found in config.yml", connectionType, connectionName)
		return output, err
	}

	return output, nil

}
