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

func (s *LRU) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	key = s.namespaceKey(key)
	s.cache.Add(lru.Key(key), value)
	return nil
}

func (s *LRU) Get(_ context.Context, key string) (interface{}, error) {
	key = s.namespaceKey(key)
	value, ok := s.cache.Get(lru.Key(key))
	if !ok {
		return nil, errors.ErrEmptyCache
	}
	return value, nil
}

func (s *LRU) MGet(_ context.Context, keys ...string) ([]interface{}, error) {
	values := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		value, ok := s.cache.Get(lru.Key(key))
		if ok {
			values = append(values, value)
		}
	}
	return values, nil
}

func (s *LRU) Remove(_ context.Context, key ...string) error {
	for _, value := range key {
		s.cache.Remove(lru.Key(s.namespaceKey(value)))
	}
	return nil
}
