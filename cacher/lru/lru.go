package lru

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/liyanbing/go-cache/errors"
)

type LRU struct {
	cache     *lru.Cache
	namespace string
}

func NewLRU(max int) *LRU {
	return &LRU{
		cache: lru.New(max),
	}
}

func (s *LRU) SetNamespace(namespace string) {
	s.namespace = namespace
}

func (s *LRU) namespaceKey(key string) string {
	if s.namespace == "" {
		return key
	}
	return fmt.Sprintf("%v:%v", s.namespace, key)
}

func (s *LRU) Set(_ context.Context, key string, value []byte, expiration time.Duration) error {
	key = s.namespaceKey(key)
	s.cache.Add(lru.Key(key), value)
	return nil
}

func (s *LRU) Get(_ context.Context, key string) ([]byte, error) {
	key = s.namespaceKey(key)
	value, ok := s.cache.Get(lru.Key(key))
	if !ok {
		return nil, errors.ErrEmptyCache
	}
	return value.([]byte), nil
}

func (s *LRU) Remove(_ context.Context, key string) error {
	key = s.namespaceKey(key)
	s.cache.Remove(lru.Key(key))
	return nil
}
