package inmemory

import (
	"errors"
	"github.com/romanzimoglyad/memcached/internal/handler"
	"github.com/romanzimoglyad/memcached/internal/model"
	"sync"
)

var NotFound = errors.New("NotFound")

type Records struct {
	*sync.RWMutex
	m map[string]string
}

func NewRecords() *Records {
	return &Records{
		RWMutex: &sync.RWMutex{},
		m:       make(map[string]string),
	}
}

var _ handler.Repository = (*Records)(nil)

func (h *Records) Get(keys ...string) (out []model.Record, err error) {
	h.Lock()
	defer h.Unlock()
	for _, key := range keys {
		if value, ok := h.m[key]; ok {
			out = append(out, model.Record{
				Key:   key,
				Value: value,
			})
		}
	}
	return out, nil
}
func (h *Records) Set(record model.Record) error {
	h.Lock()
	defer h.Unlock()
	h.m[record.Key] = record.Value
	return nil
}

func (h *Records) Delete(key string) error {
	records, err := h.Get(key)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return NotFound
	}
	delete(h.m, key)
	return nil
}
