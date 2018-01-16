// Package ctxlog is a json logger with context support.
package ctxlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type ctxKey string

// Log
type Log struct {
	output io.Writer
	debug  *int32 // bool, accessed with sync/atomic.

	ctxKeyFields ctxKey
	ctxKeyDebug  ctxKey
}

type encodeError struct {
	Time    string `json:"time"`
	File    string `json:"file"`
	Error   string `json:"error"`
	Msg     string `json:"msg"`
	OrigMsg string `json:"orig-msg"`
}

// New creates new Log instance which prints messages to stderr.
func New() *Log {
	l := &Log{
		output: os.Stderr,
		debug:  new(int32),
	}
	l.ctxKeyFields = ctxKey(fmt.Sprintf("fields-%p", l))
	l.ctxKeyDebug = ctxKey(fmt.Sprintf("debug-%p", l))
	return l
}

// SetDebugGlobal globally enables/disables debug messages. Can be called concurrently with other functions.
func (l *Log) SetDebugGlobal(v bool) {
	if v {
		atomic.StoreInt32(l.debug, 1)
	}

	atomic.StoreInt32(l.debug, 0)
}

// SetDebug returns new context with enabled/disabled debug messages. Overrides global debug flag.
func (l *Log) SetDebug(ctx context.Context, v bool) context.Context {
	return context.WithValue(ctx, l.ctxKeyDebug, v)
}

// Debug prints message with level=debug only if debug is enabled.
func (l *Log) Debug(ctx context.Context, msg string) {
	debug, ok := ctx.Value(l.ctxKeyDebug).(bool)
	if !ok {
		if atomic.LoadInt32(l.debug) == 0 {
			return
		}
	} else if !debug {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	l.print(ctx, "debug", now, msg)
}

// Info prints message msg with level=info.
func (l *Log) Info(ctx context.Context, msg string) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	l.print(ctx, "info", now, msg)
}

// Error prints message msg with level=error.
func (l *Log) Error(ctx context.Context, msg string) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	l.print(ctx, "error", now, msg)
}

// Fatal prints message msg with level=fatal and calls os.Exit(1).
func (l *Log) Fatal(ctx context.Context, msg string) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	l.print(ctx, "fatal", now, msg)
	os.Exit(1)
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func (l *Log) print(ctx context.Context, level, timeStr, msg string) {
	fields, _ := ctx.Value(l.ctxKeyFields).(map[string]interface{})

	logFields := make(map[string]interface{}, len(fields)+2)
	for k, v := range fields {
		if err, ok := v.(error); ok {
			logFields[k] = err.Error()
			continue
		}
		logFields[k] = v
	}

	logFields["msg"] = msg
	logFields["time"] = timeStr
	logFields["level"] = level

	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()

	if err := json.NewEncoder(buf).Encode(logFields); err != nil {
		buf.Reset()

		f := ""
		_, file, line, ok := runtime.Caller(2)
		if ok {
			f = fmt.Sprintf("%v:%v", file, line)
		}
		encErr := encodeError{
			Time:    timeStr,
			Error:   err.Error(),
			File:    f,
			Msg:     "ctxlog: json encode error",
			OrigMsg: msg,
		}
		if err := json.NewEncoder(buf).Encode(encErr); err != nil {
			panic(err)
		}
	}

	_, _ = l.output.Write(buf.Bytes())
}

// AddFields returns new context with specified log fields added to it.
func (l *Log) AddFields(ctx context.Context, newFields map[string]interface{}) context.Context {
	var fields map[string]interface{}
	oldFields, ok := ctx.Value(l.ctxKeyFields).(map[string]interface{})
	if ok {
		fields = make(map[string]interface{}, len(oldFields)+len(newFields))
		for k, v := range oldFields {
			fields[k] = v
		}
	} else {
		fields = make(map[string]interface{}, len(newFields))
	}
	for k, v := range newFields {
		fields[k] = v
	}

	return context.WithValue(ctx, l.ctxKeyFields, fields)
}

// WithField returns new context with specified log field added to it.
func (l *Log) WithField(ctx context.Context, key string, value interface{}) context.Context {
	return l.AddFields(ctx, map[string]interface{}{key: value})
}

// WithField returns new context with error field added to it.
func (l *Log) WithError(ctx context.Context, err error) context.Context {
	return l.AddFields(ctx, map[string]interface{}{"error": err})
}
