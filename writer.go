package ctxlog

import (
	"bytes"
	"context"
)

type writer struct {
	l   *Log
	ctx context.Context
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.l.Info(w.ctx, string(bytes.TrimSpace(p)))
	return len(p), nil
}
