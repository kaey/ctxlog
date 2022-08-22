package ctxlog_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kaey/ctxlog"
)

func ExampleLog_Print() {
	log := ctxlog.New(os.Stdout)
	ctx := context.Background()

	log.Print(ctx, "hello world", ctxlog.Time(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
	// Output: {"msg":"hello world","time":"2000-01-01T00:00:00Z"}
}

func ExampleLog_Print_error() {
	log := ctxlog.New(os.Stdout)
	ctx := context.Background()

	ctx = log.With(ctx, ctxlog.Value("foo", "bar"), ctxlog.Time(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
	log.Print(ctx, "hello again world", ctxlog.Error(fmt.Errorf("broken pipe")))
	// Output: {"error":"broken pipe","foo":"bar","msg":"hello again world","time":"2000-01-01T00:00:00Z"}
}
