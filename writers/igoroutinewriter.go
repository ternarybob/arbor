package writers

type IGoroutineWriter interface {
	IWriter
	Start() error
	Stop() error
	IsRunning() bool
}
