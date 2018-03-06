// Package ctxlog is a json logger with context support.
package ctxlog

import (
	"context"
	"io"
	"io/ioutil"
	"os"
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
	printer    PrinterFunc
	ctxDataKey ctxKey
	fields     map[string]interface{}
	debug      bool
	stackField string
}

// New creates new Log instance. Output goes to stdout in json format by default. You can specify additional options.
func New(opts ...Option) *Log {
	l := new(Log)

	for _, opt := range opts {
		opt(l)
	}

	l.ctxDataKey = ctxKey("ctxlog.ctxdata")
	if l.printer == nil {
		l.printer = DefaultPrinter(os.Stdout)
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
			fields: copyFields(cd.fields),
			err:    cd.err,
		}
	}

	return &ctxData{
		debug:  l.debug,
		fields: copyFields(l.fields),
		err:    nil,
	}
}

func copyFields(from map[string]interface{}) map[string]interface{} {
	to := make(map[string]interface{}, len(from)+10)
	for k, v := range from {
		if err, ok := v.(error); ok {
			to[k] = err.Error()
			continue
		}
		to[k] = v
	}

	return to
}

func setFields(from, to map[string]interface{}) {
	for k, v := range from {
		if err, ok := v.(error); ok {
			to[k] = err.Error()
			continue
		}
		to[k] = v
	}
}

// Debug prints message with level=debug only if debug is enabled.
func (l *Log) Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {
	if l == nil {
		return
	}

	cd := l.getCtxData(ctx)
	if !cd.debug {
		return
	}
	if len(fields) > 0 {
		setFields(fields[0], cd.fields)
	}

	l.print(cd, levelDebug, msg)
}

// Info prints message msg with level=info.
func (l *Log) Info(ctx context.Context, msg string, fields ...map[string]interface{}) {
	if l == nil {
		return
	}

	cd := l.getCtxData(ctx)
	if len(fields) > 0 {
		setFields(fields[0], cd.fields)
	}
	l.print(cd, levelInfo, msg)
}

// Error prints message msg with level=error.
func (l *Log) Error(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	if l == nil {
		return
	}

	cd := l.getCtxData(ctx)
	if len(fields) > 0 {
		setFields(fields[0], cd.fields)
	}
	if err != nil {
		cd.err = err
	}
	l.print(cd, levelError, msg)
}

// Fatal prints message msg with level=fatal and calls os.Exit(1).
func (l *Log) Fatal(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	if l == nil {
		os.Stdout.WriteString(msg + "\n")
		os.Exit(1)
	}

	cd := l.getCtxData(ctx)
	if len(fields) > 0 {
		setFields(fields[0], cd.fields)
	}
	if err != nil {
		cd.err = err
	}
	l.print(cd, levelFatal, msg)
	os.Exit(1)
}

func (l *Log) print(cd *ctxData, level, msg string) {
	if cd.err != nil {
		if l.stackField != "" {
			st := stack(cd.err)
			if st != nil {
				cd.fields[l.stackField] = st
			}
		}
		cd.fields["error"] = cd.err.Error()
	}

	cd.fields["msg"] = msg
	cd.fields["level"] = level
	cd.fields["time"] = time.Now().UTC().Format(time.RFC3339Nano)

	l.printer(cd.fields)
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
	if l == nil {
		return ioutil.Discard
	}

	return &writer{
		l:   l,
		ctx: ctx,
	}
}
