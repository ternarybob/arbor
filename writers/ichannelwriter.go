package writers

type IChannelWriter interface {
	IWriter
	Start() error
	Stop() error
	IsRunning() bool
}
