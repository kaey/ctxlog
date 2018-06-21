package ctxlog

import "io"

// Option is a func which alter behaviour of logger.
type Option func(log *Log)

// Fields sets fields which will be added to all messages.
func ContextOptions(cos ...ContextOption) Option {
	return func(l *Log) {
		ncos := make([]ContextOption, len(cos))
		copy(ncos, cos)
		l.cos = ncos
	}
}

// Printer sets printer function. Default is json printer with os.Stdout.
func Output(w io.Writer) Option {
	return func(l *Log) {
		l.printer = newPrinter(w)
	}
}
