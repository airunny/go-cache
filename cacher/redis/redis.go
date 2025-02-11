package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/liyanbing/go-cache/errors"

	redis "github.com/go-redis/redis/v8"
)

func NewRedisCache(cli *redis.Client) *Redis {
	return &Redis{
		cli: cli,
	}
}

type Redis struct {
	cli       *redis.Client
	namespace string
}

func (s *Redis) SetNamespace(namespace string) {
	s.namespace = namespace
}

func (s *Redis) namespaceKey(key string) string {
	if s.namespace == "" {
		return key
	}
	return fmt.Sprintf("%v:%v", s.namespace, key)
}

func (s *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	key = s.namespaceKey(key)
	return s.cli.Set(ctx, key, value, expiration).Err()
}

func (s *Redis) Get(ctx context.Context, key string) (interface{}, error) {
	key = s.namespaceKey(key)
	value, err := s.cli.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.ErrEmptyCache
		}
		return nil, err
	}
	return value, nil
}

func (s *Redis) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	newKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		newKeys = append(newKeys, s.namespaceKey(key))
	}

	value, err := s.cli.MGet(ctx, newKeys...).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.ErrEmptyCache
		}
		return nil, err
	}
	return value, nil
}

func (s *Redis) Remove(ctx context.Context, key ...string) error {
	keys := make([]string, 0, len(key))
	for _, value := range key {
		keys = append(keys, s.namespaceKey(value))
	}
	return s.cli.Del(ctx, keys...).Err()
}
