package handler

import "github.com/romanzimoglyad/memcached/internal/model"

type Repository interface {
	Get(keys ...string) ([]model.Record, error)
	Set(record model.Record) error
	Delete(key string) error
}
