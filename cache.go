package go_cache

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/golang/groupcache/singleflight"
	"github.com/golang/protobuf/proto"
	"github.com/liyanbing/go-cache/errors"
	"github.com/liyanbing/go-cache/tools"

	jsonIter "github.com/json-iterator/go"
)

/**
 * 在具体的获取数据操作(fetcher)前加一层缓存操作
 * 1、先根据key从cache中获取数据，如果不存在则从fetcher中获取数据（并发调用时只会有一个请求会调用fetcher,其他请求会复用这个请求返回的数据）
 * 2、从fetcher获取到数据之后然后存储到cache中,然后返回从fetcher获取到的对象
 * 3、如果cache中存在数据，则decode进model对象中返回（注意这时候返回的是model的指针）
 * 注意：如果对象在缓存中存在则一定返回的是对象指针，如果不存在返回的是fetcher返回的数据(为了统一fetcher最好也返回对象的指针)
 */

var (
	single = &singleflight.Group{}
	json   = jsonIter.ConfigCompatibleWithStandardLibrary
)

/**
 * value: 需要存储在缓存中的值
 * expiration: value 过期时间
 * err：错误
 */
type Fetcher func() (value interface{}, expiration time.Duration, err error)

type EmptyCache func(outerKey string)

type CacheValueOutput func(value interface{}) error

type Cache interface {
	// set global namespace
	SetNamespace(namespace string)
	// set value of key; auto delete from bridger after expiration time
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// get value of key , return errors.ErrEmptyCache if not found key from bridger
	Get(ctx context.Context, key string) (interface{}, error)
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	// remove value by key
	Remove(ctx context.Context, key ...string) error
}

func FetchWithJson(ctx context.Context, cache Cache, key string, fetcher Fetcher, model interface{}) (interface{}, error) {
	return fetch(ctx, cache, key, fetcher, jsonEncode, JsonDecode(model))
}

func FetchWithString(ctx context.Context, cache Cache, key string, fetcher Fetcher) (string, error) {
	value, err := fetch(ctx, cache, key, fetcher, func(input interface{}) ([]byte, error) {
		var data []byte
		switch input.(type) {
		case string:
			data = []byte(input.(string))
		case []byte:
			data = input.([]byte)
		}
		return data, nil
	}, func(value interface{}) (interface{}, error) {
		return tools.ToString(value)
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func FetchWithProtobuf(ctx context.Context, cache Cache, key string, fetcher Fetcher, model interface{}) (proto.Message, error) {
	value, err := fetch(ctx, cache, key, fetcher, protoEncode, ProtoDecode(model))
	if err != nil {
		return nil, err
	}
	return value.(proto.Message), nil
}

func FetchWithNumber(ctx context.Context, cache Cache, key string, fetcher Fetcher) (float64, error) {
	value, err := fetch(ctx, cache, key, fetcher, func(i interface{}) ([]byte, error) {
		if !tools.CanConvertToNumber(i) {
			return nil, errors.ErrInvalidValue
		}
		return []byte(fmt.Sprintf("%v", i)), nil
	}, func(value interface{}) (interface{}, error) {
		return value, nil
	})
	if err != nil {
		return 0, err
	}
	return tools.ToFloat(value)
}

func FetchWithArray(ctx context.Context, cache Cache, key string, fetcher Fetcher, model interface{}) (interface{}, error) {
	return fetch(ctx, cache, key, fetcher, func(i interface{}) ([]byte, error) {
		kind := reflect.TypeOf(i).Kind()
		if kind != reflect.Slice && kind != reflect.Array {
			return nil, errors.ErrInvalidValue
		}
		return jsonEncode(i)
	}, func(value interface{}) (interface{}, error) {
		dataValue, ok := value.([]byte)
		if !ok {
			return nil, errors.ErrInvalidCacheValue
		}

		ret := reflect.New(reflect.MakeSlice(typeFromModel(model), 0, 0).Type())
		err := json.Unmarshal(dataValue, ret.Interface())
		if err != nil {
			return nil, err
		}
		return ret.Elem().Interface(), nil
	})
}

// 批量获取otherKeys的缓存数据，如果缓存中不存在则会通过fetcher获取不存在缓存中的数据，通过fetcher获取到的数据不会加入缓存
func FetchWithIncludeKeys(ctx context.Context, cache Cache, output CacheValueOutput, empty EmptyCache, dec Decoder, otherKeys ...string) error {
	for _, key := range otherKeys {
		cachedValue, err := cache.Get(ctx, key)
		if err == errors.ErrEmptyCache {
			empty(key)
			continue
		}
		if err != nil {
			return err
		}

		value, err := dec(cachedValue)
		if err != nil {
			return err
		}

		err = output(value)
		if err != nil {
			return err
		}
	}
	return nil
}

func FetchWithKeys(ctx context.Context, cache Cache, keys ...string) ([]interface{}, error) {
	return cache.MGet(ctx, keys...)
}
