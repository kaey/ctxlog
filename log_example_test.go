package ctxlog_test

import (
	"context"

	"github.com/kaey/ctxlog"
)

func Example() {
	log := ctxlog.New()
	ctx := context.Background()

	log.Info(ctx, "hello world")
	// Prints: {"level":"info","msg":"hello world","time":"2018-01-16T13:01:35.623558838Z"}

	ctx = log.WithField(ctx, "foo", "bar")
	log.Error(ctx, "hello again world")
	// Prints: {"foo":"bar","level":"error","msg":"hello again world","time":"2018-01-16T13:01:35.62356885Z"}
}
