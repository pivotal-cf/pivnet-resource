package logger

import (
	"fmt"
	"io"
)

//go:generate counterfeiter . Logger

type Logger interface {
	Debugf(format string, a ...interface{}) (n int, err error)
}

type logger struct {
	sink io.Writer
}

func NewLogger(sink io.Writer) Logger {
	return &logger{
		sink: sink,
	}
}

func (l logger) Debugf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(l.sink, format, a...)
}
