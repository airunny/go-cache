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

type Cache interface {
	// set global namespace
	SetNamespace(namespace string)
	// set value off key; auto delete from cache after expiration time
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// get value off key , return errors.ErrEmptyCache if not found key from cache
	Get(ctx context.Context, key string) (interface{}, error)
	// remove value by key
	Remove(ctx context.Context, key string) error
}

func FetchWithJson(ctx context.Context, cache Cache, key string, fetcher Fetcher, model interface{}) (interface{}, error) {
	return fetch(ctx, cache, key, fetcher, jsonEncode, jsonDecode(model))
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
	value, err := fetch(ctx, cache, key, fetcher, protoEncode, protoDecode(model))
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
