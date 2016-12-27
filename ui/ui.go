package ui

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

type UIPrinter struct {
	outWriter io.Writer
}

func NewUIPrinter(outWriter io.Writer) *UIPrinter {
	return &UIPrinter{
		outWriter: outWriter,
	}
}

func (p *UIPrinter) PrintDeprecationln(text string) {
	f := color.New(color.FgYellow).SprintfFunc()
	text = "WARNING: " + text
	fmt.Fprintln(p.outWriter, f(text))
}

func (p *UIPrinter) PrintErrorln(err error) {
	p.PrintErrorlnf("%v", err)
}

func (p *UIPrinter) PrintErrorlnf(text string, args ...interface{}) {
	f := color.New(color.FgRed).SprintfFunc()
	text = "ERROR: " + text
	fmt.Fprintln(p.outWriter, f(text, args...))
}
