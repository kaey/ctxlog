package ctxlog

import "io"

// Option is a func which alter behaviour of logger.
type Option func(log *Log)

// EnableDebug enables/disable debug messages globally.
func EnableDebug(v bool) Option {
	return func(log *Log) {
		log.debug = v
	}
}

// ErrorStackTrace sets field name where to record stack trace if it is present in error. See StackTracer and runtime.Callers().
func ErrorStackTrace(field string) Option {
	return func(log *Log) {
		log.stackTraceField = field
	}
}

// Fields sets fields which will be added to all messages.
func Fields(fields map[string]interface{}) Option {
	return func(log *Log) {
		log.fields = copyLogFields(fields)
	}
}

// Output sets output. Default is os.Stdout.
func Output(w io.Writer) Option {
	return func(log *Log) {
		log.output = w
	}
}
