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
		w: w,
	}
}

func (p *printer) print(cd *ctxData, msg string) {
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer p.bufPool.Put(buf)
	buf.Reset()

	m := make(map[string]interface{}, 10)

	d := cd
	for {
		for _, co := range d.cos {
			if _, exists := m[co.key]; exists {
				continue
			}

			switch co.key {
			case "error":
				err, ok := co.value.(error)
				if ok {
					m["error"] = err.Error()
				}

				st, ok := co.value.(Stacker)
				if ok {
					m["error-stack"] = stack(st)
				}
			case "time":
				t, ok := co.value.(time.Time)
				if ok {
					m["time"] = t.UTC()
				}
			default:
				m[co.key] = co.value
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
