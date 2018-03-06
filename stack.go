package ctxlog

import (
	"fmt"
	"runtime"
)

type StackTracer interface {
	StackTrace() []uintptr
}

func stack(err error) []string {
	e, ok := err.(StackTracer)
	if !ok {
		return nil
	}

	frames := runtime.CallersFrames(e.StackTrace())
	st := make([]string, 0, len(e.StackTrace()))
	for {
		frame, more := frames.Next()
		st = append(st, fmt.Sprintf("%s:%d[%s]", frame.File, frame.Line, frame.Func.Name()))

		if !more {
			break
		}
	}

	return st
}
