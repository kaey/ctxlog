// Package ctxlog is a logger with context support.
package ctxlog

import (
	"context"
	"io"
	"io/ioutil"
	"os"
)

type ctxKey int

var ctxDataKey ctxKey

// Log is a logging object. Use New to create it.
type Log struct {
	printer *printer
	cos     []ContextOption
}

// New creates new Log instance. Prints to stderr with JSONPrinter by default.
func New(opts ...Option) *Log {
	l := new(Log)

	for _, opt := range opts {
		opt(l)
	}

	if l.printer == nil {
		l.printer = newPrinter(os.Stderr)
	}

	return l
}

// Print prints message msg with specified options.
func (l *Log) Print(ctx context.Context, msg string, cos ...ContextOption) {
	if l == nil {
		return
	}

	cd := l.newCtxData(ctx, cos)
	l.printer.print(cd, msg)
}

// With returns new context with specified ContextOptions added to it.
func (l *Log) With(ctx context.Context, cos ...ContextOption) context.Context {
	if l == nil || len(cos) == 0 {
		return ctx
	}

	cd := l.newCtxData(ctx, cos)
	return context.WithValue(ctx, ctxDataKey, cd)
}

// Writer returns io.Writer which calls l.Info for every write to it.
func (l *Log) Writer(ctx context.Context) io.Writer {
	if l == nil {
		return ioutil.Discard
	}

	return &writer{
		l:   l,
		ctx: ctx,
	}
}

type ctxData struct {
	prev *ctxData
	cos  []ContextOption
}

func (l *Log) newCtxData(ctx context.Context, cos []ContextOption) *ctxData {
	cd, ok := ctx.Value(ctxDataKey).(*ctxData)
	if !ok {
		return &ctxData{
			prev: &ctxData{
				cos: l.cos,
			},
			cos: cos,
		}
	}

	return &ctxData{
		prev: cd,
		cos:  cos,
	}
}
