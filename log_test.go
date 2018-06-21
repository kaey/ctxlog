package ctxlog_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/kaey/ctxlog"
)

func TestPrintInfo(t *testing.T) {
	buf := new(bytes.Buffer)
	log := ctxlog.New(
		ctxlog.ContextOptions(
			ctxlog.Field("foo", "bar"),
			ctxlog.Time(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
		),
		ctxlog.Output(buf),
	)
	ctx := context.Background()

	log.Print(ctx, "foo")

	expected := `{"foo":"bar","msg":"foo","time":"2000-01-01T00:00:00Z"}` + "\n"
	got := buf.String()
	if expected != got {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestWithError(t *testing.T) {
	buf := new(bytes.Buffer)
	log := ctxlog.New(
		ctxlog.ContextOptions(
			ctxlog.Field("foo", "bar"),
			ctxlog.Time(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
		),
		ctxlog.Output(buf),
	)
	ctx := log.With(context.Background(), ctxlog.Error(fmt.Errorf("bar error")))

	log.Print(ctx, "foo")

	expected := `{"error":"bar error","foo":"bar","msg":"foo","time":"2000-01-01T00:00:00Z"}` + "\n"
	got := buf.String()
	if expected != got {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestEncoderError(t *testing.T) {
	buf := new(bytes.Buffer)
	log := ctxlog.New(
		ctxlog.ContextOptions(
			ctxlog.Field("foo", "bar"),
			ctxlog.Time(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
		),
		ctxlog.Output(buf),
	)
	ctx := context.Background()

	log.Print(ctx, "foo", ctxlog.Field("chan", make(chan struct{})))

	expected := `{"error":"json: unsupported type: chan struct {}","msg":"ctxlog: json encode error","orig-msg":"foo","time":"2000-01-01T00:00:00Z"}` + "\n"
	got := buf.String()
	if expected != got {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestNilLog(t *testing.T) {
	ctx := context.Background()
	var log *ctxlog.Log

	log.Print(ctx, "should not panic")

	if w := log.Writer(ctx); w != ioutil.Discard {
		t.Errorf("expected discard writer, got %v", w)
	}

	nctx := log.With(ctx, ctxlog.Field("foo", "bar"))
	if ctx != nctx {
		t.Errorf("With called on nil logger should have returned original context")
	}
}
