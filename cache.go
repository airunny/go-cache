package go_cache

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"reflect"
	"time"

	"github.com/golang/groupcache/singleflight"
	"github.com/golang/protobuf/proto"

	"github.com/liyanbing/go-cache/errors"
)

type Cache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}

var (
	single = &singleflight.Group{}
)

type Fetcher func() (interface{}, error)

type Encoder func(interface{}) ([]byte, error)

type Decoder func(value []byte, model reflect.Type) (interface{}, error)

func FetchWithJson(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher, model interface{}) (interface{}, error) {
	return fetch(ctx, cache, key, expire, fetcher, typeFromModel(model), jsonEncode, jsonDecode)
}

func FetchWithString(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher) (string, error) {
	value, err := fetch(ctx, cache, key, expire, fetcher, nil, func(input interface{}) ([]byte, error) {
		return []byte(input.(string)), nil
	}, func(value []byte, _ reflect.Type) (interface{}, error) {
		return string(value), nil
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func FetchWithProtobuf(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher, model interface{}) (proto.Message, error) {
	value, err := fetch(ctx, cache, key, expire, fetcher, typeFromModel(model), protoEncode, protoDecode)
	if err != nil {
		return nil, err
	}
	return value.(proto.Message), nil
}

func fetch(
	ctx context.Context,
	cache Cache,
	key string,
	expire time.Duration,
	fetcher Fetcher,
	model reflect.Type,
	encoder Encoder,
	decoder Decoder) (interface{}, error) {

	do := func() (interface{}, error) {
		return single.Do(key, func() (interface{}, error) {
			value, err := fetcher()
			if err != nil {
				return nil, err
			}

			cacheData, err := encoder(value)
			if err != nil {
				return nil, err
			}

			err = cache.Set(ctx, key, cacheData, expire)
			if err != nil {
				log.Printf("set cache <%v,%v> Err:%v", key, value, err)
				err = nil
			}
			return value, nil
		})
	}

	if noUseCache(ctx) {
		return do()
	}

	cacheData, err := cache.Get(ctx, key)
	if err != nil && err != errors.ErrEmptyCache {
		return nil, err
	}

	if err == errors.ErrEmptyCache {
		return do()
	}
	return decoder(cacheData, model)
}

func typeFromModel(model interface{}) reflect.Type {
	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

func protoDecode(data []byte, model reflect.Type) (interface{}, error) {
	ret := reflect.New(model)
	err := proto.Unmarshal(data, ret.Interface().(proto.Message))
	if err != nil {
		return nil, err
	}
	return ret.Interface(), nil
}

func protoEncode(value interface{}) ([]byte, error) {
	return proto.Marshal(value.(proto.Message))
}

func jsonDecode(value []byte, model reflect.Type) (interface{}, error) {
	ret := reflect.New(model)
	err := json.NewDecoder(bytes.NewBuffer(value)).Decode(ret.Interface())
	if err != nil {
		return nil, err
	}
	return ret.Interface(), nil
}

func jsonEncode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}
