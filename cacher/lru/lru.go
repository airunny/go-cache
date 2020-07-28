package lru

import (
	"context"
	"time"

	"github.com/golang/groupcache/lru"

	"github.com/liyanbing/go-cache/errors"
)

type LRU struct {
	cache *lru.Cache
}

func NewLRU(max int) *LRU {
	return &LRU{
		cache: lru.New(max),
	}
}

func (s *LRU) Set(_ context.Context, key string, value []byte, expiration time.Duration) error {
	s.cache.Add(lru.Key(key), value)
	return nil
}

func (s *LRU) Get(_ context.Context, key string) ([]byte, error) {
	value, ok := s.cache.Get(lru.Key(key))
	if !ok {
		return nil, errors.ErrEmptyCache
	}
	return value.([]byte), nil
}
