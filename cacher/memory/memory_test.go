package memory

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/liyanbing/go-cache/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryCache(t *testing.T) {
	m := NewMemoryCache(10)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%v", i)
		err := m.Set(context.Background(), key, []byte(key), time.Second)
		assert.Nil(t, err)
	}

	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%v", i)
		value, err := m.Get(context.Background(), key)
		assert.Equal(t, errors.ErrEmptyCache, err)
		assert.Nil(t, value)
	}
}

func TestNewMemoryCache2(t *testing.T) {
	m := NewMemoryCache(10)
	wait := sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wait.Add(1)
		key := fmt.Sprintf("%v", i)
		go func(k string) {
			defer wait.Done()
			err := m.Set(context.Background(), k, []byte(key), time.Hour)
			assert.Nil(t, err)
		}(key)
	}
	time.Sleep(time.Second * 10)

	haveNum := int64(0)
	for i := 0; i < 1000; i++ {
		wait.Add(1)
		go func(k int) {
			defer wait.Done()
			key := fmt.Sprintf("%v", k)
			value, err := m.Get(context.Background(), key)
			if err == errors.ErrEmptyCache {
				assert.Nil(t, value)
			} else {
				atomic.AddInt64(&haveNum, 1)
				assert.NotNil(t, value, k)
			}
		}(i)
	}

	assert.Equal(t, int64(10), haveNum)
	wait.Wait()
}
