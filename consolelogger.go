package arbor

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"

	consolewriter "github.com/ternarybob/arbor/consolewriter"
	"github.com/ternarybob/arbor/memorywriter"
	"github.com/ternarybob/arbor/filewriter"
	"github.com/ternarybob/arbor/ginwriter"

	"github.com/ternarybob/satus"

	"github.com/labstack/echo/v4"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/rs/zerolog"
)

const (
	CORRELATION_ID_KEY string = "correlationid"
	LOGGERCONTEXT_KEY  string = "consolelogger"
	WRITER_CONSOLE     string = "writerconsole"
	WRITER_DATA        string = "writerdata"
	WRITER_REDIS       string = "writerredis"
	WRITER_ARRAY       string = "writerarray"
)

var (
	internallog zerolog.Logger = zerolog.New(&consolewriter.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger().Level(zerolog.WarnLevel)
	copieropts  copier.Option  = copier.Option{IgnoreEmpty: true, DeepCopy: false}
)

type consolelogger struct {
	logger  zerolog.Logger
	writers map[string]io.Writer
}

func ConsoleLogger() IConsoleLogger {

	var (
		cfg          *satus.AppConfig     = satus.GetAppConfig()
		namedwriters map[string]io.Writer = make(map[string]io.Writer)
		writers      []io.Writer
	)

	loglevel, err := ParseLevel(cfg.Service.LogLevel)
	if err != nil {
		loglevel = InfoLevel
	}

	// zerolog.SetGlobalLevel(loglevel)
	internallog.Trace().Msgf("Setting cfg.Service.LogLevel:%s loglevel:%s", cfg.Service.LogLevel, loglevel.String())

	// Add Writers
	namedwriters[WRITER_CONSOLE] = consolewriter.New()

	for k, v := range namedwriters {

		internallog.Trace().Msgf("Adding Writer name:%s type:%s", k, reflect.TypeOf(v))

		writers = append(writers, namedwriters[k])

	}

	mw := io.MultiWriter(writers...)

	return &consolelogger{
		logger:  zerolog.New(mw).With().Timestamp().Logger().Level(loglevel),
		writers: namedwriters,
	}

}

// func (d *consolelogger) WithRequestContext(ctx *gin.Context) IConsoleLogger {
func (d *consolelogger) WithRequestContext(ctx echo.Context) IConsoleLogger {

	var (
		internallog = internallog.With().Str("prefix", "WithRequestContext").Logger()
		writers     []io.Writer
	)

	if ctx == nil {
		panic(fmt.Errorf("Context not available | nil"))
	}

	correlationid := ctx.Response().Header().Get(echo.HeaderXRequestID)

	if isEmpty(correlationid) {
		internallog.Warn().Msgf("Correlation Key Not Available -> New Logger NOT created")
		return d
	}

	if len(d.writers) == 0 {
		d.writers = make(map[string]io.Writer)
	}

	// Add Writers
	if _, ok := d.writers[WRITER_CONSOLE]; !ok {
		d.writers[WRITER_CONSOLE] = consolewriter.New()
	}

	if _, ok := d.writers[WRITER_DATA]; !ok {
		d.writers[WRITER_DATA] = memorywriter.New()
	}

	// Add to mulit writer
	for k, v := range d.writers {
		internallog.Debug().Msgf("Adding Writer name:%s type:%s", k, reflect.TypeOf(v))
		writers = append(writers, d.writers[k])
	}

	mw := io.MultiWriter(writers...)

	currentlevel := d.GetLevel()

	o := &consolelogger{
		logger:  zerolog.New(mw).With().Timestamp().Str("correlationid", correlationid).Logger().Level(currentlevel),
		writers: d.writers,
	}

	return o

}

func (d *consolelogger) WithWriter(name string, writer io.Writer) IConsoleLogger {

	var (
		writers []io.Writer
	)

	if len(d.writers) == 0 {
		d.writers = make(map[string]io.Writer)
	}

	// Ensure Default Writer
	if _, ok := d.writers[WRITER_CONSOLE]; !ok {
		d.writers[WRITER_CONSOLE] = consolewriter.New()
	}

	// Add Writer
	d.writers[name] = writer

	// Add to mulit writer
	for k, v := range d.writers {
		internallog.Debug().Msgf("Adding Writer name:%s type:%s", k, reflect.TypeOf(v))
		writers = append(writers, v)
	}

	mw := io.MultiWriter(writers...)

	currentlevel := d.GetLevel()

	o := &consolelogger{
		logger:  zerolog.New(mw).With().Timestamp().Logger().Level(currentlevel),
		writers: d.writers,
	}

	return o

}

func (d *consolelogger) WithCorrelationId(correlationid string) IConsoleLogger {

	var (
		writers []io.Writer
	)

	if isEmpty(correlationid) {

		uuid, err := uuid.NewRandom()
		if err != nil {
			internallog.Warn().Err(err).Msg("")
		}

		correlationid = uuid.String()
	}

	internallog.Trace().Msgf("Adding CorrelationId -> CorrelationId:%s", correlationid)

	// Ensure writers map exists
	if len(d.writers) == 0 {
		d.writers = make(map[string]io.Writer)
	}

	// Add console writer if not present
	if _, ok := d.writers[WRITER_CONSOLE]; !ok {
		d.writers[WRITER_CONSOLE] = consolewriter.New()
	}

	// Add memory writer for correlation ID logging
	if _, ok := d.writers[WRITER_DATA]; !ok {
		d.writers[WRITER_DATA] = memorywriter.New()
		internallog.Debug().Msgf("Added memory writer for correlation ID logging")
	}

	// Build multi-writer from all writers
	for k, v := range d.writers {
		internallog.Debug().Msgf("Adding Writer name:%s type:%s", k, reflect.TypeOf(v))
		writers = append(writers, v)
	}

	mw := io.MultiWriter(writers...)
	currentLevel := d.GetLevel()

	return &consolelogger{
		logger:  zerolog.New(mw).With().Timestamp().Str("correlationid", correlationid).Logger().Level(currentLevel),
		writers: d.writers,
	}

}

func (d *consolelogger) GetLevel() Level {

	return d.logger.GetLevel()

}

func (d *consolelogger) WithPrefix(value string) IConsoleLogger {

	d.logger = d.logger.With().Str("prefix", value).Logger()

	return d

}

func (d *consolelogger) WithLevel(lvl Level) IConsoleLogger {

	var (
		output = &consolelogger{}
	)

	if err := copier.CopyWithOption(&output, &d, copieropts); err != nil {
		internallog.Warn().Err(err).Msgf("Unable to copy existing service -> reverted to inital")
		return d
	}

	output.logger = output.logger.Level(lvl)

	return output

}

func (d *consolelogger) WithFunction() IConsoleLogger {

	functionName := d.getFunctionName()

	d.logger = d.logger.With().Timestamp().Str("prefix", functionName).Logger()

	return d

}

func (d *consolelogger) WithContext(key string, value string) IConsoleLogger {

	d.logger = d.logger.With().Timestamp().Str(key, value).Logger()

	return d

}

func (d *consolelogger) WithCorrelationid(value string) {

	d.logger = d.logger.With().Str(CORRELATION_ID_KEY, value).Logger()

}

func (d *consolelogger) getFunctionName() string {

	// Assuming called from within the package runtime.Caller(2)
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	return fn.Name()

}

func (d *consolelogger) GetLogger() *zerolog.Logger {

	var (
		output = &consolelogger{}
	)

	if err := copier.CopyWithOption(&output, &d, copieropts); err != nil {
		internallog.Warn().Err(err).Msgf("Unable to copy existing service -> reverted to inital")
		return &d.logger
	}

	return &output.logger

}

func (d *consolelogger) WithFileWriterPath(name string, filePath string, bufferSize int) (IConsoleLogger, error) {
	// Create file writer with directory creation
	fileWriter, err := filewriter.NewWithPath(filePath, bufferSize)
	if err != nil {
		return nil, err
	}

	// Use the existing WithWriter method
	return d.WithWriter(name, fileWriter), nil
}

func (d *consolelogger) GinWriter() io.Writer {
	// Import ginwriter for this to work
	return &ginwriter.GinWriter{
		Out:   os.Stdout,
		Level: d.GetLevel(),
	}
}

func (d *consolelogger) GetMemoryLogs(correlationid string, minLevel Level) (map[string]string, error) {
	return memorywriter.GetEntriesWithLevel(correlationid, minLevel)
}
