package ctxlog_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/kaey/ctxlog"
)

func newTestEncoder(f map[string]interface{}) ctxlog.PrinterFunc {
	return func(fields map[string]interface{}) {
		for k, v := range fields {
			f[k] = v
		}
	}
}

func TestPrintInfo(t *testing.T) {
	got := make(map[string]interface{})
	log := ctxlog.New(ctxlog.Printer(newTestEncoder(got)))
	ctx := context.Background()

	log.Info(ctx, "foo")
	got["time"] = "now"
	expected := map[string]interface{}{
		"level": "info",
		"msg":   "foo",
		"time":  "now",
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestWithError(t *testing.T) {
	got := make(map[string]interface{})
	log := ctxlog.New(ctxlog.Printer(newTestEncoder(got)))
	ctx := log.WithError(context.Background(), fmt.Errorf("bar error"))

	log.Info(ctx, "foo")
	got["time"] = "now"
	expected := map[string]interface{}{
		"level": "info",
		"msg":   "foo",
		"time":  "now",
		"error": "bar error",
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestPrinter(t *testing.T) {
	buf := new(bytes.Buffer)
	printer := ctxlog.DefaultPrinter(buf)

	fields := map[string]interface{}{
		"error": "some error",
		"level": "info",
		"msg":   "foo",
		"time":  "now",
	}

	printer(fields)
	expected := `{"error":"some error","level":"info","msg":"foo","time":"now"}` + "\n"
	got := buf.String()

	if expected != got {
		t.Errorf("expected: %v, got: %v", expected, got)
	}

	buf.Reset()
	fields["chan"] = make(chan string)
	printer(fields)
	expected = `{"error":"json: unsupported type: chan string","level":"error","msg":"ctxlog: json encode error","orig-msg":"foo","time":"now"}` + "\n"
	got = buf.String()

	if expected != got {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestNilLog(t *testing.T) {
	ctx := context.Background()
	var log *ctxlog.Log

	log.Debug(ctx, "should not panic")
	log.Info(ctx, "should not panic")
	log.Error(ctx, "should not panic", fmt.Errorf("some err"))

	if w := log.Writer(ctx); w != ioutil.Discard {
		t.Errorf("expected discard writer, got %v", w)
	}

	nctx := log.WithField(ctx, "foo", "bar")
	if ctx != nctx {
		t.Errorf("WithField called on nil logger should have returned original context")
	}
}
