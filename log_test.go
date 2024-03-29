package ctxlog_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kaey/ctxlog"
)

func TestPrintInfo(t *testing.T) {
	buf := new(bytes.Buffer)
	log := ctxlog.New(buf, ctxlog.Value("foo", "bar"), ctxlog.Time(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
	ctx := context.Background()

	log.Print(ctx, "foo", ctxlog.Value("foo", "baz"))

	expected := `{"foo":"baz","msg":"foo","time":"2000-01-01T00:00:00Z"}` + "\n"
	got := buf.String()
	if expected != got {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestWithError(t *testing.T) {
	buf := new(bytes.Buffer)
	log := ctxlog.New(buf, ctxlog.Value("foo", "bar"), ctxlog.Time(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
	ctx := ctxlog.With(context.Background(), ctxlog.Error(fmt.Errorf("bar error")))

	log.Print(ctx, "foo")

	expected := `{"error":"bar error","foo":"bar","msg":"foo","time":"2000-01-01T00:00:00Z"}` + "\n"
	got := buf.String()
	if expected != got {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestEncoderError(t *testing.T) {
	buf := new(bytes.Buffer)
	log := ctxlog.New(buf, ctxlog.Value("foo", "bar"), ctxlog.Time(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
	ctx := context.Background()

	log.Print(ctx, "foo", ctxlog.Value("chan", make(chan struct{})))

	expected := `{"error":"json: unsupported type: chan struct {}","msg":"ctxlog: json encode error","orig_msg":"foo","time":"2000-01-01T00:00:00Z"}` + "\n"
	got := buf.String()
	if expected != got {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestNilLog(t *testing.T) {
	ctx := context.Background()
	var log *ctxlog.Log

	log.Print(ctx, "should not panic")
	log.Writer(ctx).Write([]byte("should not panic either"))
}
