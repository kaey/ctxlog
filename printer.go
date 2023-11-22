package ctxlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var bufPool sync.Pool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

var mapPool sync.Pool = sync.Pool{
	New: func() any {
		return make(map[string]any, 10)
	},
}

func (l *Log) print(cd *ctxdata, msg string) {
	m := mapPool.Get().(map[string]any)
	defer func() {
		clear(m)
		mapPool.Put(m)
	}()

	handleFields := func(fs []Field) {
		for _, f := range fs {
			if f.key == "" {
				continue
			}
			if _, exists := m[f.key]; exists {
				continue
			}

			switch f.key {
			case "error":
				err, ok := f.val.(error)
				if ok {
					m["error"] = err.Error()
				}

				var st Stacker
				if errors.As(err, &st) {
					m["error_stack"] = stack(st)
				}
			case "time":
				t, ok := f.val.(time.Time)
				if ok {
					m["time"] = t.UTC()
				}
			default:
				m[f.key] = f.val
			}
		}
	}

	for d := cd; d != nil; d = d.prev {
		handleFields(d.fields)
	}
	handleFields(l.fields)

	m["msg"] = msg
	if _, ok := m["time"].(time.Time); !ok {
		m["time"] = time.Now().UTC()
	}

	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	if err := json.NewEncoder(buf).Encode(m); err != nil {
		t := m["time"].(time.Time)
		encErr := map[string]string{
			"time":     t.Format(time.RFC3339),
			"error":    err.Error(),
			"msg":      "ctxlog: json encode error",
			"orig_msg": msg,
		}
		buf.Reset()
		if err := json.NewEncoder(buf).Encode(encErr); err != nil {
			panic(err)
		}
	}

	l.mu.Lock()
	_, _ = buf.WriteTo(l.w)
	l.mu.Unlock()
}
