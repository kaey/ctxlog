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

// Log levels.
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelError = "error"
	LevelFatal = "fatal"
)

type ctxKey string

// Log is a logging object which writes logs to os.Stdout in json format.
type Log struct {
	output   io.Writer
	outputMu sync.Mutex
	debug    *int32 // bool, accessed with sync/atomic.

	ctxKeyFields ctxKey
	ctxKeyDebug  ctxKey
}

type encodeError struct {
	Time    string `json:"time"`
	File    string `json:"file"`
	Error   string `json:"error"`
	Msg     string `json:"msg"`
	OrigMsg string `json:"orig-msg"`
	Level   string `json:"level"`
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
		return
	}

	atomic.StoreInt32(l.debug, 0)
}

// SetDebug returns new context with enabled/disabled debug messages. Overrides global debug flag.
func (l *Log) SetDebug(ctx context.Context, v bool) context.Context {
	if l == nil {
		return ctx
	}

	return context.WithValue(ctx, l.ctxKeyDebug, v)
}

// Debug prints message with level=debug only if debug is enabled.
func (l *Log) Debug(ctx context.Context, msg string) {
	if l == nil {
		return
	}

	debug, ok := ctx.Value(l.ctxKeyDebug).(bool)
	if ok {
		if !debug {
			return
		}
	} else {
		if atomic.LoadInt32(l.debug) == 0 {
			return
		}
	}

	l.print(ctx, LevelDebug, msg)
}

// Info prints message msg with level=info.
func (l *Log) Info(ctx context.Context, msg string) {
	if l == nil {
		return
	}

	l.print(ctx, LevelInfo, msg)
}

// Error prints message msg with level=error.
func (l *Log) Error(ctx context.Context, msg string) {
	if l == nil {
		return
	}

	l.print(ctx, LevelError, msg)
}

// Fatal prints message msg with level=fatal and calls os.Exit(1).
func (l *Log) Fatal(ctx context.Context, msg string) {
	if l == nil {
		os.Stdout.WriteString(msg)
		os.Exit(1)
	}

	l.print(ctx, LevelFatal, msg)
	os.Exit(1)
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func (l *Log) print(ctx context.Context, level, msg string) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := l.write(ctx, level, now, msg); err != nil {
		f := ""
		_, file, line, ok := runtime.Caller(2)
		if ok {
			f = fmt.Sprintf("%v:%v", file, line)
		}

		buf := bufPool.Get().(*bytes.Buffer)
		defer bufPool.Put(buf)
		buf.Reset()

		encErr := encodeError{
			Time:    now,
			Error:   err.Error(),
			File:    f,
			Msg:     "ctxlog: json encode error",
			OrigMsg: msg,
			Level:   LevelError,
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
		return err
	}

	l.outputMu.Lock()
	_, _ = l.output.Write(buf.Bytes())
	l.outputMu.Unlock()
	return nil
}

// WithFields returns new context with specified log fields added to it.
func (l *Log) WithFields(ctx context.Context, newFields map[string]interface{}) context.Context {
	if l == nil {
		return ctx
	}

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
	return l.WithFields(ctx, map[string]interface{}{key: value})
}

// WithError returns new context with error field added to it.
func (l *Log) WithError(ctx context.Context, err error) context.Context {
	return l.WithFields(ctx, map[string]interface{}{"error": err})
}
