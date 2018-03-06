package ctxlog

import (
	"fmt"
	"runtime"
)

// Stacker is an interface which can be implemented by errors to include stack trace info in logs.
type Stacker interface {
	Stack() []uintptr
}

func stack(err error) []string {
	e, ok := err.(Stacker)
	if !ok {
		return nil
	}

	frames := runtime.CallersFrames(e.Stack())
	st := make([]string, 0, len(e.Stack()))
	for {
		frame, more := frames.Next()
		st = append(st, fmt.Sprintf("%s:%d[%s]", frame.File, frame.Line, frame.Func.Name()))

		if !more {
			break
		}
	}

	return st
}
