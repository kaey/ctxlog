package ctxlog

import (
	"context"
	"io"
)

func (l *Log) Write(ctx context.Context, level, timeStr, msg string) error {
	return l.write(ctx, level, timeStr, msg)
}

func (l *Log) SetOutput(output io.Writer) {
	l.output = output
}

type EncodeError = encodeError
