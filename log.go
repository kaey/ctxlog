// Package ctxlog is a logger with context support.
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

type ctxKey int

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

	if l.printer == nil {
		l.printer = DefaultPrinter(os.Stdout)
	}

	return l
}

// Debug prints message with level=debug only if debug is enabled.
func (l *Log) Debug(ctx context.Context, msg string, fieldsVar ...map[string]interface{}) {
	if l == nil {
		return
	}

	cd := l.copyCtxData(ctx, true, fieldsVar, nPrintFields)
	if cd == nil {
		return
	}
	l.print(cd, levelDebug, msg)
}

// Info prints message msg with level=info.
func (l *Log) Info(ctx context.Context, msg string, fieldsVar ...map[string]interface{}) {
	if l == nil {
		return
	}

	cd := l.copyCtxData(ctx, false, fieldsVar, nPrintFields)
	l.print(cd, levelInfo, msg)
}

// Error prints message msg with level=error.
func (l *Log) Error(ctx context.Context, msg string, err error, fieldsVar ...map[string]interface{}) {
	if l == nil {
		return
	}

	cd := l.copyCtxData(ctx, false, fieldsVar, nPrintFields)
	cd.err = err
	l.print(cd, levelError, msg)
}

// Fatal prints message msg with level=fatal and calls os.Exit(1).
func (l *Log) Fatal(ctx context.Context, msg string, err error, fieldsVar ...map[string]interface{}) {
	if l == nil {
		os.Stdout.WriteString(msg + "\n")
		os.Exit(1)
	}

	cd := l.copyCtxData(ctx, false, fieldsVar, nPrintFields)
	cd.err = err
	l.print(cd, levelFatal, msg)
	os.Exit(1)
}

const nPrintFields = 4 // error + msg + level + time

func (l *Log) print(cd *ctxData, level, msg string) {
	if cd.err != nil {
		cd.fields["error"] = cd.err.Error()

		if l.stackField != "" {
			st := stack(cd.err)
			if st != nil {
				cd.fields[l.stackField] = st
			}
		}
	}

	cd.fields["msg"] = msg
	cd.fields["level"] = level
	cd.fields["time"] = time.Now().UTC()

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

	fieldsVar := []map[string]interface{}{fields}
	cd := l.copyCtxData(ctx, false, fieldsVar, 0)
	return context.WithValue(ctx, l.ctxDataKey, cd)
}

// WithField returns new context with specified log field added to it.
func (l *Log) WithField(ctx context.Context, key string, value interface{}) context.Context {
	if l == nil {
		return ctx
	}

	cd := l.copyCtxData(ctx, false, nil, 1)
	setField(cd.fields, key, value)
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

	cd := l.copyCtxData(ctx, false, nil, 0)
	cd.err = err
	return context.WithValue(ctx, l.ctxDataKey, cd)
}

// WithDebug returns new context with enabled/disabled debug messages. Overrides global option.
func (l *Log) WithDebug(ctx context.Context, v bool) context.Context {
	if l == nil {
		return ctx
	}

	cd := l.copyCtxData(ctx, false, nil, 0)
	cd.debug = v
	return context.WithValue(ctx, l.ctxDataKey, cd)
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
	debug  bool
	fields map[string]interface{}
	err    error
}

func (l *Log) copyCtxData(ctx context.Context, debugLevel bool, fieldsVar []map[string]interface{}, prealloc int) *ctxData {
	cd, ok := ctx.Value(l.ctxDataKey).(*ctxData)
	if !ok {
		cd = &ctxData{
			debug:  l.debug,
			fields: l.fields,
			err:    nil,
		}
	}

	if debugLevel && !cd.debug {
		return nil
	}

	var fields map[string]interface{}
	if len(fieldsVar) > 0 {
		fields = fieldsVar[0]
	}

	newfields := make(map[string]interface{}, len(cd.fields)+len(fields)+prealloc)
	for k, v := range cd.fields {
		newfields[k] = v
	}
	setFields(fields, newfields)

	return &ctxData{
		debug:  cd.debug,
		fields: newfields,
		err:    cd.err,
	}
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

func setField(m map[string]interface{}, k string, v interface{}) {
	if err, ok := v.(error); ok {
		m[k] = err.Error()
		return
	}
	m[k] = v
}
