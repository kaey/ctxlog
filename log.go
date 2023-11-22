// Package ctxlog is a logger with context support.
package ctxlog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
)

var log *Log

func Global(l *Log) {
	log = l
}

// Print prints json line with Global logger using msg and fields, as well as any fields stored in context.
func Print(ctx context.Context, msg string, fields ...Field) {
	log.Print(ctx, msg, fields...)
}

// Writer returns io.Writer for Global logger which calls l.Print for every write to it.
func Writer(ctx context.Context) io.Writer {
	return log.Writer(ctx)
}

// With returns new context with specified fields added to it.
func With(ctx context.Context, fields ...Field) context.Context {
	if len(fields) == 0 {
		return ctx
	}

	cd, _ := ctx.Value(ctxkey).(*ctxdata)
	return context.WithValue(ctx, ctxkey, &ctxdata{prev: cd, fields: fields})
}

type Log struct {
	fields []Field

	mu sync.Mutex
	w  io.Writer
}

func New(w io.Writer, fields ...Field) *Log {
	return &Log{
		fields: fields,
		w:      w,
	}
}

// Print prints message msg with specified fields.
func (l *Log) Print(ctx context.Context, msg string, fields ...Field) {
	if l == nil {
		return
	}

	cd, _ := ctx.Value(ctxkey).(*ctxdata)
	l.print(&ctxdata{prev: cd, fields: fields}, msg)
}

// Writer returns io.Writer which calls l.Print for every write to it.
func (l *Log) Writer(ctx context.Context) io.Writer {
	return &writer{
		l:   l,
		ctx: ctx,
	}
}

type writer struct {
	l   *Log
	ctx context.Context
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.l.Print(w.ctx, string(bytes.TrimSpace(p)))
	return len(p), nil
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
			return st
		}
	}
}

type Field struct {
	key string
	val any
}

func Value(k string, v any) Field {
	return Field{key: k, val: v}
}

func Error(err error) Field {
	return Field{key: "error", val: err}
}

func Time(t time.Time) Field {
	return Field{key: "time", val: t}
}

type ctxkeytype struct{}

var ctxkey = ctxkeytype{}

type ctxdata struct {
	prev   *ctxdata
	fields []Field
}
