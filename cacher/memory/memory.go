package memory

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/liyanbing/go-cache/errors"
)

func NewMemoryCache(max int32) *Memory {
	return &Memory{
		MaxEntries: max,
		quit:       make(chan struct{}, 1),
	}
}

type entry struct {
	value  interface{}
	expire int64
}

type Memory struct {
	cache      sync.Map
	quit       chan struct{}
	MaxEntries int32
	entriesNum int32
	namespace  string
}

func (m *Memory) SetNamespace(namespace string) {
	m.namespace = namespace
}

func (m *Memory) namespaceKey(key string) string {
	if m.namespace == "" {
		return key
	}
	return fmt.Sprintf("%v:%v", m.namespace, key)
}

func (m *Memory) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	key = m.namespaceKey(key)
	entriesNum := atomic.LoadInt32(&m.entriesNum)
	if m.MaxEntries > 0 && m.MaxEntries <= entriesNum {
		return nil
	}

	atomic.AddInt32(&m.entriesNum, 1)
	m.cache.Store(key, &entry{
		value:  value,
		expire: time.Now().Add(expiration).UnixNano(),
	})
	return nil
}

func (m *Memory) Get(_ context.Context, key string) (interface{}, error) {
	key = m.namespaceKey(key)
	value, ok := m.cache.Load(key)
	if !ok {
		return nil, errors.ErrEmptyCache
	}

	if m.checkAndDelete(key, value) {
		return nil, errors.ErrEmptyCache
	}
	return value.(*entry).value, nil
}

func (m *Memory) MGet(_ context.Context, keys ...string) ([]interface{}, error) {
	values := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		value, ok := m.cache.Load(m.namespaceKey(key))
		if ok {
			if !m.checkAndDelete(key, value) {
				values = append(values, value)
			}
		}
	}
	return values, nil
}

func (m *Memory) Remove(_ context.Context, key ...string) error {
	for _, value := range key {
		m.cache.Delete(m.namespaceKey(value))
		if m.MaxEntries > 0 {
			atomic.AddInt32(&m.entriesNum, -1)
		}
	}
	return nil
}

func (m *Memory) checkAndDelete(key string, value interface{}) bool {
	key = m.namespaceKey(key)
	data := value.(*entry)
	if time.Now().UnixNano() > data.expire {
		m.cache.Delete(key)
		if m.MaxEntries > 0 {
			atomic.AddInt32(&m.entriesNum, -1)
		}
		return true
	}
	return false
}

func (m *Memory) Run() {
	go func() {
		for {
			select {
			case <-time.After(time.Second):
			case <-m.quit:
				return
			}

			m.cache.Range(func(key, value interface{}) bool {
				m.checkAndDelete(key.(string), value)
				return true
			})
		}
	}()
}

func (m *Memory) Close() error {
	m.quit <- struct{}{}
	return nil
}
