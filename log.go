// Package ctxlog is a logger with context support.
package ctxlog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

type Log struct {
	printer *printer
	fields  []Field
}

func New(w io.Writer, fields ...Field) *Log {
	return &Log{
		printer: &printer{w: w},
		fields:  fields,
	}
}

// Print prints message msg with specified fields.
func (l *Log) Print(ctx context.Context, msg string, fields ...Field) {
	if l == nil {
		return
	}

	cd := l.newCtxData(ctx, fields)
	l.printer.print(cd, msg)
}

// With returns new context with specified fields added to it.
func (l *Log) With(ctx context.Context, fields ...Field) context.Context {
	if l == nil || len(fields) == 0 {
		return ctx
	}

	cd := l.newCtxData(ctx, fields)
	return context.WithValue(ctx, ctxkey, cd)
}

// Writer returns io.Writer which calls l.Print for every write to it.
func (l *Log) Writer(ctx context.Context) io.Writer {
	if l == nil {
		return io.Discard
	}

	return &writer{
		l:   l,
		ctx: ctx,
	}
}

type Field struct {
	key   string
	value interface{}
}

func Value(k string, v interface{}) Field {
	return Field{key: k, value: v}
}

func Hidden(v interface{}) Field {
	return Field{value: v}
}

func Error(err error) Field {
	return Field{key: "error", value: err}
}

func Time(t time.Time) Field {
	return Field{key: "time", value: t}
}

type ctxkeytype struct{}

var ctxkey = ctxkeytype{}

type ctxdata struct {
	prev   *ctxdata
	fields []Field
}

func (l *Log) newCtxData(ctx context.Context, fields []Field) *ctxdata {
	cd, ok := ctx.Value(ctxkey).(*ctxdata)
	if !ok {
		return &ctxdata{
			prev: &ctxdata{
				fields: l.fields,
			},
			fields: fields,
		}
	}

	return &ctxdata{
		prev:   cd,
		fields: fields,
	}
}

// Stacker can be implemented by errors to include stack trace info in logs.
// Use runtime.Callers to get pc slice.
type Stacker interface {
	Stack() (pc []uintptr)
}

func stack(v Stacker) []string {
	frames := runtime.CallersFrames(v.Stack())
	st := make([]string, 0, len(v.Stack()))
	for {
		frame, more := frames.Next()
		st = append(st, fmt.Sprintf("%s:%d[%s]", frame.File, frame.Line, frame.Func.Name()))

		if !more {
			break
		}
	}

	return st
}

type writer struct {
	l   *Log
	ctx context.Context
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.l.Print(w.ctx, string(bytes.TrimSpace(p)))
	return len(p), nil
}

var l *Log = New(os.Stdout)

// Print prints json line to stdout using msg and fields, as well as any fields stored in context.
func Print(ctx context.Context, msg string, fields ...Field) {
	l.Print(ctx, msg, fields...)
}

// With stores fields in context.
func With(ctx context.Context, fields ...Field) context.Context {
	return l.With(ctx, fields...)
}

// FieldFromContext returns field of specified type that was stored using With().
// It panics if such type was not found.
func FieldFromContext[T any](ctx context.Context) *T {
	cd, _ := ctx.Value(ctxkey).(*ctxdata)

	for d := cd; d != nil; d = d.prev {
		for _, f := range d.fields {
			v, ok := f.value.(*T)
			if !ok {
				continue
			}
			return v
		}
	}

	panic("ctxlog: value not found")
}
