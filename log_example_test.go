package ctxlog_test

import (
	"context"
	"fmt"

	"github.com/kaey/ctxlog"
)

func Example() {
	log := ctxlog.New()
	ctx := context.Background()

	log.Info(ctx, "hello world")
	// Prints: {"level":"info","msg":"hello world","time":"2018-01-16T13:01:35.623558838Z"}

	ctx = log.WithField(ctx, "foo", "bar")
	err := fmt.Errorf("broken pipe")
	log.Error(ctx, "hello again world", err)
	// Prints: {"foo":"bar","level":"error","msg":"hello again world","time":"2018-01-16T13:01:35.62356885Z"}
}
