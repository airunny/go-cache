package memory

import (
	"context"
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
	value  []byte
	expire int64
}

type Memory struct {
	cache      sync.Map
	quit       chan struct{}
	MaxEntries int32
	entriesNum int32
}

func (m *Memory) Set(_ context.Context, key string, value []byte, expiration time.Duration) error {
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

func (m *Memory) Get(_ context.Context, key string) ([]byte, error) {
	value, ok := m.cache.Load(key)
	if !ok {
		return nil, errors.ErrEmptyCache
	}

	if m.checkAndDelete(key, value) {
		return nil, errors.ErrEmptyCache
	}
	return value.(*entry).value, nil
}

func (m *Memory) checkAndDelete(key string, value interface{}) bool {
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
