package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/liyanbing/go-cache/errors"

	redis "gopkg.in/redis.v5"
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

func (s *Redis) Set(_ context.Context, key string, value []byte, expiration time.Duration) error {
	key = s.namespaceKey(key)
	return s.cli.Set(key, value, expiration).Err()
}

func (s *Redis) Get(_ context.Context, key string) ([]byte, error) {
	key = s.namespaceKey(key)
	value, err := s.cli.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.ErrEmptyCache
		}
		return nil, err
	}
	return value, nil
}

func (s *Redis) Remove(_ context.Context, key string) error {
	key = s.namespaceKey(key)
	return s.cli.Del(key).Err()
}
