package ctxlog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
		if err := log.Write(ctx, "info", "now", "foo"); err != nil {
			t.Errorf("unexpected error from print: %v", err)
		}

		expected := `{"level":"info","msg":"foo","time":"now"}` + "\n"
		got := buf.String()
		if expected != got {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	})

	t.Run("WithError", func(t *testing.T) {
		buf.Reset()
		ctx = log.WithError(ctx, fmt.Errorf("bar error"))
		if err := log.Write(ctx, "error", "now", "bar"); err != nil {
			t.Errorf("unexpected error from print: %v", err)
		}

		expected := `{"error":"bar error","level":"error","msg":"bar","time":"now"}` + "\n"
		got := buf.String()
		if expected != got {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	})

	t.Run("EncodeError", func(t *testing.T) {
		buf.Reset()
		ctx = log.WithField(ctx, "chan", make(chan string))
		if err := log.Write(ctx, "info", "now", "chan"); err == nil {
			t.Errorf("expected error from print, got nil")
		}
	})

	t.Run("EncodeErrorMarshal", func(t *testing.T) {
		buf.Reset()
		err := ctxlog.EncodeError{
			Time:    "now",
			Error:   "foo err",
			File:    "some file",
			Msg:     "encode error",
			OrigMsg: "original msg",
			Level:   ctxlog.LevelError,
		}

		if err := json.NewEncoder(buf).Encode(err); err != nil {
			t.Errorf("unexpected error from json.Encode: %v", err)
		}

		expected := `{"time":"now","file":"some file","error":"foo err","msg":"encode error","orig-msg":"original msg","level":"error"}` + "\n"
		got := buf.String()
		if expected != got {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	})
}

func TestNilLog(t *testing.T) {
	ctx := context.Background()
	var log *ctxlog.Log

	log.Debug(ctx, "should not panic")

	nctx := log.WithField(ctx, "foo", "bar")
	if ctx != nctx {
		t.Errorf("WithField called on nil logger should have returned original context")
	}
}
