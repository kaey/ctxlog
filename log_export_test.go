package ctxlog

import (
	"context"
	"io"
)

func (l *Log) Print(ctx context.Context, level, timeStr, msg string) {
	l.print(ctx, level, timeStr, msg)
}

func (l *Log) SetOutput(output io.Writer) {
	l.output = output
}
