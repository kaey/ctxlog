package ctxlog

import (
	"fmt"
	"runtime"
)

// Stacker can be implemented by errors to include stack trace info in logs.
type Stacker interface {
	Stack() []uintptr
}

func stack(v Stacker) []string {
	frames := runtime.CallersFrames(v.Stack())
	st := make([]string, 0, len(v.Stack()))
	for {
		frame, more := frames.Next()
		st = append(st, fmt.Sprintf("%s:%d[%s]", frame.File, frame.Line, frame.Func.Name()))

		if !more {
			break
		}
	}

	return st
}
