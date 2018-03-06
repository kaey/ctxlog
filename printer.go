package ctxlog

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
)

type PrinterFunc func(fields map[string]interface{})

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

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
