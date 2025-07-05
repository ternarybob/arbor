package writers

import "github.com/phuslu/log"

type IWriter interface {
	WithLevel(level log.Level) IWriter
	Write(p []byte) (n int, err error)
}
