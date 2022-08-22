package ctxlog

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"time"
)

type printer struct {
	bufPool sync.Pool
	mapPool sync.Pool
	mu      sync.Mutex
	w       io.Writer
}

func newPrinter(w io.Writer) *printer {
	return &printer{
		bufPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		mapPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{}, 10)
			},
		},
		w: w,
	}
}

func (p *printer) print(cd *ctxdata, msg string) {
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()

	m := p.mapPool.Get().(map[string]interface{})
	defer func() {
		for k := range m {
			delete(m, k)
		}
		p.mapPool.Put(m)
	}()

	d := cd
	for {
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

				st, ok := f.value.(Stacker)
				if ok {
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

		if d.prev == nil {
			break
		}
		d = d.prev
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
