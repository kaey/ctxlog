package ctxlog

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
)

// PrinterFunc is called for every call to log Debug/Info/Debug/Fatal. It's job is to encode received fields and write them somewhere (see DefaultPrinter).
// Must be safe for concurrent use.
type PrinterFunc func(fields map[string]interface{})

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// DefaultPrinter prints fields in json format to w.
func DefaultPrinter(w io.Writer) PrinterFunc {
	var mu sync.Mutex
	return func(fields map[string]interface{}) {
		buf := bufPool.Get().(*bytes.Buffer)
		defer bufPool.Put(buf)
		buf.Reset()

		if err := json.NewEncoder(buf).Encode(fields); err != nil {
			encErr := map[string]interface{}{
				"time":     fields["time"],
				"error":    err.Error(),
				"msg":      "ctxlog: json encode error",
				"orig-msg": fields["msg"],
				"level":    levelError,
			}
			buf.Reset()
			if err := json.NewEncoder(buf).Encode(encErr); err != nil {
				panic(err)
			}
		}

		mu.Lock()
		_, _ = w.Write(buf.Bytes())
		mu.Unlock()
	}
}
