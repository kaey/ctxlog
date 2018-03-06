// Package ctxlog is a json logger with context support.
package ctxlog

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// Log levels.
const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelError = "error"
	levelFatal = "fatal"
)

type ctxKey string

// Log is a logging object. Use New to create it.
type Log struct {
	output          io.Writer
	outputMu        sync.Mutex
	ctxDataKey      ctxKey
	fields          map[string]interface{}
	debug           bool
	stackTraceField string
}

// New creates new Log instance. Output goes to stdout by default. You can specify additional options.
func New(opts ...Option) *Log {
	l := new(Log)

	for _, opt := range opts {
		opt(l)
	}

	l.ctxDataKey = ctxKey("ctxlog.ctxdata")
	if l.output == nil {
		l.output = os.Stdout
	}

	return l
}

type ctxData struct {
	debug  bool
	fields map[string]interface{}
	err    error
}

func (l *Log) getCtxData(ctx context.Context) *ctxData {
	cd, ok := ctx.Value(l.ctxDataKey).(*ctxData)
	if ok {
		return &ctxData{
			debug:  cd.debug,
			fields: copyLogFields(cd.fields),
			err:    cd.err,
		}
	}

	return &ctxData{
		debug:  l.debug,
		fields: copyLogFields(l.fields),
		err:    nil,
	}
}

func copyLogFields(m map[string]interface{}) map[string]interface{} {
	r := make(map[string]interface{}, len(m)+5)
	for k, v := range m {
		if err, ok := v.(error); ok {
			r[k] = err.Error()
			continue
		}
		r[k] = v
	}

	return r
}

// Debug prints message with level=debug only if debug is enabled.
func (l *Log) Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {
	if l == nil {
		return
	}

	if len(fields) > 0 {
		ctx = l.WithFields(ctx, fields[0])
	}
	l.print(ctx, levelDebug, msg)
}

// Info prints message msg with level=info.
func (l *Log) Info(ctx context.Context, msg string, fields ...map[string]interface{}) {
	if l == nil {
		return
	}

	if len(fields) > 0 {
		ctx = l.WithFields(ctx, fields[0])
	}
	l.print(ctx, levelInfo, msg)
}

// Error prints message msg with level=error.
func (l *Log) Error(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	if l == nil {
		return
	}

	if len(fields) > 0 {
		ctx = l.WithFields(ctx, fields[0])
	}
	if err != nil {
		ctx = l.WithError(ctx, err)
	}
	l.print(ctx, levelError, msg)
}

// Fatal prints message msg with level=fatal and calls os.Exit(1).
func (l *Log) Fatal(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	if l == nil {
		os.Stdout.WriteString(msg)
		os.Exit(1)
	}

	if len(fields) > 0 {
		ctx = l.WithFields(ctx, fields[0])
	}
	if err != nil {
		ctx = l.WithError(ctx, err)
	}
	l.print(ctx, levelFatal, msg)
	os.Exit(1)
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type encodeError struct {
	Time    string `json:"time"`
	Error   string `json:"error"`
	Msg     string `json:"msg"`
	OrigMsg string `json:"orig-msg"`
	Level   string `json:"level"`
}

func (l *Log) print(ctx context.Context, level, msg string) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := l.write(ctx, level, now, msg); err != nil {
		buf := bufPool.Get().(*bytes.Buffer)
		defer bufPool.Put(buf)
		buf.Reset()

		encErr := encodeError{
			Time:    now,
			Error:   err.Error(),
			Msg:     "ctxlog: json encode error",
			OrigMsg: msg,
			Level:   levelError,
		}
		if err := json.NewEncoder(buf).Encode(encErr); err != nil {
			panic(err)
		}

		l.outputMu.Lock()
		_, _ = l.output.Write(buf.Bytes())
		l.outputMu.Unlock()
	}
}

func (l *Log) write(ctx context.Context, level, timeStr, msg string) error {
	cd := l.getCtxData(ctx)

	if level == levelDebug && !cd.debug {
		return nil
	}

	if cd.err != nil {
		if l.stackTraceField != "" {
			st := stack(cd.err)
			if st != nil {
				cd.fields[l.stackTraceField] = st
			}
		}
		cd.fields["error"] = cd.err.Error()
	}

	cd.fields["msg"] = msg
	cd.fields["time"] = timeStr
	cd.fields["level"] = level

	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()

	if err := json.NewEncoder(buf).Encode(cd.fields); err != nil {
		return err
	}

	l.outputMu.Lock()
	_, _ = l.output.Write(buf.Bytes())
	l.outputMu.Unlock()
	return nil
}

// WithFields returns new context with specified log fields added to it.
func (l *Log) WithFields(ctx context.Context, fields map[string]interface{}) context.Context {
	if l == nil {
		return ctx
	}

	if len(fields) == 0 {
		return ctx
	}

	cd := l.getCtxData(ctx)
	for k, v := range fields {
		cd.fields[k] = v
	}

	return context.WithValue(ctx, l.ctxDataKey, cd)
}

// WithField returns new context with specified log field added to it.
func (l *Log) WithField(ctx context.Context, key string, value interface{}) context.Context {
	if l == nil {
		return ctx
	}

	cd := l.getCtxData(ctx)
	cd.fields[key] = value

	return context.WithValue(ctx, l.ctxDataKey, cd)
}

// WithError returns new context with error field added to it.
func (l *Log) WithError(ctx context.Context, err error) context.Context {
	if l == nil {
		return ctx
	}

	if err == nil {
		return ctx
	}

	cd := l.getCtxData(ctx)
	cd.err = err

	return context.WithValue(ctx, l.ctxDataKey, cd)
}

// WithDebug returns new context with enabled/disabled debug messages. Overrides global option.
func (l *Log) WithDebug(ctx context.Context, v bool) context.Context {
	if l == nil {
		return ctx
	}

	cd := l.getCtxData(ctx)
	cd.debug = v

	return context.WithValue(ctx, l.ctxDataKey, cd)
}

// Writer returns io.Writer which logs data written to it in ctxlog format.
func (l *Log) Writer(ctx context.Context) io.Writer {
	return &writer{
		l:   l,
		ctx: ctx,
	}
}
