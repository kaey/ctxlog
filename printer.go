package ctxlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"
)

var bufPool sync.Pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var mapPool sync.Pool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 10)
	},
}

type printer struct {
	mu sync.Mutex
	w  io.Writer
}

func (p *printer) print(cd *ctxdata, msg string) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	m := mapPool.Get().(map[string]interface{})
	defer func() {
		for k := range m {
			delete(m, k)
		}
		mapPool.Put(m)
	}()

	for d := cd; d != nil; d = d.prev {
		for _, f := range d.fields {
			if _, exists := m[f.key]; exists {
				continue
			}

			switch f.key {
			case "error":
				err, ok := f.value.(error)
				if ok {
					m["error"] = err.Error()
				}

				var st Stacker
				if errors.As(err, &st) {
					m["error-stack"] = stack(st)
				}
			case "time":
				t, ok := f.value.(time.Time)
				if ok {
					m["time"] = t.UTC()
				}
			default:
				m[f.key] = f.value
			}
		}
	}

	m["msg"] = msg
	if _, exists := m["time"]; !exists {
		m["time"] = time.Now().UTC()
	}

	if err := json.NewEncoder(buf).Encode(m); err != nil {
		encErr := map[string]interface{}{
			"time":     m["time"],
			"error":    err.Error(),
			"msg":      "ctxlog: json encode error",
			"orig-msg": m["msg"],
		}
		buf.Reset()
		if err := json.NewEncoder(buf).Encode(encErr); err != nil {
			panic(err)
		}
	}

	p.mu.Lock()
	_, _ = p.w.Write(buf.Bytes())
	p.mu.Unlock()
}
