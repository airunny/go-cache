package redis

import (
	"context"
	"testing"
	"time"

	"github.com/liyanbing/go-cache/errors"
	"github.com/stretchr/testify/assert"

	redis "github.com/go-redis/redis/v8"
)

func TestNewRedisCache(t *testing.T) {
	// redis
	redisCli := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})

	cache := NewRedisCache(redisCli)
	cache.SetNamespace("test")

	// set name
	err := cache.Set(context.Background(), "name", "value", time.Second)
	assert.Nil(t, err)

	// get name
	value, err := cache.Get(context.Background(), "name")
	assert.Nil(t, err)
	assert.Equal(t, []byte("value"), value)

	time.Sleep(time.Second)

	// get name again
	value, err = cache.Get(context.Background(), "name")
	assert.Equal(t, errors.ErrEmptyCache, err)
	assert.Nil(t, value)

	// set name again
	err = cache.Set(context.Background(), "name", "value", time.Second)
	assert.Nil(t, err)

	// get name again
	value, err = cache.Get(context.Background(), "name")
	assert.Nil(t, err)
	assert.Equal(t, []byte("value"), value)

	// remove
	err = cache.Remove(context.Background(), "name")
	assert.Nil(t, err)

	// get again
	value, err = cache.Get(context.Background(), "name")
	assert.Equal(t, errors.ErrEmptyCache, err)
	assert.Nil(t, value)
}
