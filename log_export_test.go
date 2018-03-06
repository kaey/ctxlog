package ctxlog

import (
	"context"
)

func (l *Log) Write(ctx context.Context, level, timeStr, msg string) error {
	return l.write(ctx, level, timeStr, msg)
}

type EncodeError = encodeError
