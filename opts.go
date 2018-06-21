package ctxlog

// Option is a func which alter behaviour of logger.
type Option func(log *Log)

// EnableDebug enables/disable debug messages globally.
func EnableDebug(v bool) Option {
	return func(log *Log) {
		log.debug = v
	}
}

// ErrorStack sets field name where to record stack trace if it is present in error. See Stacker and runtime.Callers().
func ErrorStack(field string) Option {
	return func(log *Log) {
		log.stackField = field
	}
}

// Fields sets fields which will be added to all messages.
func Fields(fields map[string]interface{}) Option {
	return func(log *Log) {
		newfields := make(map[string]interface{}, len(fields))
		setFields(fields, newfields)
		log.fields = newfields
	}
}

// Printer sets printer function. Default is json printer with os.Stdout.
func Printer(f PrinterFunc) Option {
	return func(log *Log) {
		log.printer = f
	}
}
