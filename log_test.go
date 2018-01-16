package ctxlog_test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/kaey/ctxlog"
)

func TestLog(t *testing.T) {
	ctx := context.Background()
	log := ctxlog.New()
	buf := new(bytes.Buffer)
	log.SetOutput(buf)

	t.Run("PrintInfo", func(t *testing.T) {
		buf.Reset()
		log.Print(ctx, "info", "now", "foo")

		expected := `{"level":"info","msg":"foo","time":"now"}` + "\n"
		got := buf.String()
		if expected != got {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	})

	t.Run("WithError", func(t *testing.T) {
		buf.Reset()
		ctx = log.WithError(ctx, fmt.Errorf("bar error"))
		log.Print(ctx, "error", "now", "bar")

		expected := `{"error":"bar error","level":"error","msg":"bar","time":"now"}` + "\n"
		got := buf.String()
		if expected != got {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	})

	t.Run("EncodeError", func(t *testing.T) {
		buf.Reset()
		ctx = log.WithField(ctx, "chan", make(chan string))
		log.Print(ctx, "info", "now", "chan")

		// Log will also contain file name and line where log was called, this field is cut here.
		expected := `{"time":"now","error":"json: unsupported type: chan string","msg":"ctxlog: json encode error","orig-msg":"chan"}` + "\n"
		got := regexp.MustCompile(`"file":".*?",`).ReplaceAllString(buf.String(), "")
		if expected != got {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	})
}
