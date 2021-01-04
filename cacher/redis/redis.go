package redis

import (
	"context"
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
	cli *redis.Client
}

func (s *Redis) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.cli.Set(key, value, expiration).Err()
}

func (s *Redis) Get(_ context.Context, key string) ([]byte, error) {
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
	return s.cli.Del(key).Err()
}
