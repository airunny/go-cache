package go_cache

import (
	"context"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/liyanbing/go-cache/cacher/lru"
	"github.com/liyanbing/go-cache/cacher/memory"
	"github.com/redis/go-redis/v9"

	redisCache "github.com/liyanbing/go-cache/cacher/redis"
)

type Bridge interface {
	Cache
	FetchWithJson(ctx context.Context, key string, fetcher Fetcher, model interface{}) (interface{}, error)
	FetchWithString(ctx context.Context, key string, fetcher Fetcher) (string, error)
	FetchWithProtobuf(ctx context.Context, key string, fetcher Fetcher, model interface{}) (proto.Message, error)
	FetchWithNumber(ctx context.Context, key string, fetcher Fetcher) (float64, error)
	FetchWithArray(ctx context.Context, key string, fetcher Fetcher, model interface{}) (interface{}, error)
	FetchWithIncludeKeys(ctx context.Context, output CacheValueOutput, empty EmptyCache, dec Decoder, otherKeys ...string) error
	FetchWithKeys(ctx context.Context, keys ...string) ([]interface{}, error)
}

type cacheType int8

const (
	cacheTypeRedis cacheType = iota
	cacheTypeMemory
	cacheTypeLRU
	cacheTypeCustom
)

type option struct {
	cacheType        cacheType
	redisCli         *redis.Client
	cache            Cache
	memoryMaxEntries int32
	lruMaxEntries    int
}

func defaultOption() option {
	return option{
		cacheType:     cacheTypeLRU,
		lruMaxEntries: 100,
	}
}

type BridgeOption func(option)

func WithRedis(cli *redis.Client) BridgeOption {
	return func(o option) {
		o.cacheType = cacheTypeRedis
		o.redisCli = cli
	}
}

func WithMemory(maxEntries int32) BridgeOption {
	return func(o option) {
		o.cacheType = cacheTypeMemory
		o.memoryMaxEntries = maxEntries
	}
}

func WithLRU(maxEntries int32) BridgeOption {
	return func(o option) {
		o.cacheType = cacheTypeLRU

	}
}

func WithCache(cache Cache) BridgeOption {
	return func(o option) {
		o.cacheType = cacheTypeCustom
		o.cache = cache
	}
}

var (
	_ Bridge = (*bridger)(nil)
)

type bridger struct {
	cacheType cacheType
	Cache
	redisCli redis.Client
}

func NewBridge(opts ...BridgeOption) Bridge {
	o := defaultOption()
	for _, opt := range opts {
		opt(o)
	}

	switch o.cacheType {
	case cacheTypeRedis:
		if o.redisCli == nil {
			log.Fatal("empty redis client")
		}
		o.cache = redisCache.NewRedisCache(o.redisCli)
	case cacheTypeMemory:
		if o.memoryMaxEntries <= 0 {
			o.memoryMaxEntries = 100
		}
		o.cache = memory.NewMemoryCache(o.memoryMaxEntries)
	case cacheTypeLRU:
		if o.lruMaxEntries <= 0 {
			o.lruMaxEntries = 100
		}
		o.cache = lru.NewLRU(o.lruMaxEntries)
	case cacheTypeCustom:
		if o.cache == nil {
			log.Fatal("empty cache")
		}
	}

	return &bridger{
		Cache: o.cache,
	}
}

func (c *bridger) FetchWithJson(ctx context.Context, key string, fetcher Fetcher, model interface{}) (interface{}, error) {
	return FetchWithJson(ctx, c.Cache, key, fetcher, model)
}

func (c *bridger) FetchWithString(ctx context.Context, key string, fetcher Fetcher) (string, error) {
	return FetchWithString(ctx, c.Cache, key, fetcher)
}

func (c *bridger) FetchWithProtobuf(ctx context.Context, key string, fetcher Fetcher, model interface{}) (proto.Message, error) {
	return FetchWithProtobuf(ctx, c.Cache, key, fetcher, model)
}

func (c *bridger) FetchWithNumber(ctx context.Context, key string, fetcher Fetcher) (float64, error) {
	return FetchWithNumber(ctx, c.Cache, key, fetcher)
}

func (c *bridger) FetchWithArray(ctx context.Context, key string, fetcher Fetcher, model interface{}) (interface{}, error) {
	return FetchWithArray(ctx, c.Cache, key, fetcher, model)
}

func (c *bridger) FetchWithIncludeKeys(ctx context.Context, output CacheValueOutput, empty EmptyCache, dec Decoder, otherKeys ...string) error {
	return FetchWithIncludeKeys(ctx, c.Cache, output, empty, dec, otherKeys...)
}

func (c *bridger) FetchWithKeys(ctx context.Context, keys ...string) ([]interface{}, error) {
	return FetchWithKeys(ctx, c.Cache, keys...)
}
