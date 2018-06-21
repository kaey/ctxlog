package ctxlog

import "time"

type ContextOption struct {
	key   string
	value interface{}
}

func Field(k string, v interface{}) ContextOption {
	return ContextOption{key: k, value: v}
}

func Error(err error) ContextOption {
	return ContextOption{key: "error", value: err}
}

func Time(t time.Time) ContextOption {
	return ContextOption{key: "time", value: t}
}
