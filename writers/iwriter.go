package writers

import "github.com/ternarybob/arbor/levels"

type IWriter interface {
	WithLevel(level levels.LogLevel) IWriter
	Write(p []byte) (n int, err error)
}
