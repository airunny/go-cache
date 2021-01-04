package go_cache

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
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

type Fetcher func() (interface{}, error)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Remove(ctx context.Context, key string) error
}

func FetchWithJson(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher, model interface{}) (interface{}, error) {
	return fetch(ctx, cache, key, expire, fetcher, jsonEncode, jsonDecode(model))
}

func FetchWithString(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher) (string, error) {
	value, err := fetch(ctx, cache, key, expire, fetcher, func(input interface{}) ([]byte, error) {
		tep, ok := input.(string)
		if !ok {
			return nil, errors.ErrInvalidValue
		}
		return []byte(tep), nil
	}, func(value []byte) (interface{}, error) {
		return string(value), nil
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func FetchWithProtobuf(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher, model interface{}) (proto.Message, error) {
	value, err := fetch(ctx, cache, key, expire, fetcher, protoEncode, protoDecode(model))
	if err != nil {
		return nil, err
	}
	return value.(proto.Message), nil
}

func FetchWithNumber(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher) (float64, error) {
	value, err := fetch(ctx, cache, key, expire, fetcher, func(i interface{}) ([]byte, error) {
		if !tools.IsNumber(i) {
			return nil, errors.ErrInvalidValue
		}
		return []byte(fmt.Sprintf("%v", i)), nil
	}, func(value []byte) (interface{}, error) {
		return strconv.ParseFloat(string(value), 64)
	})
	if err != nil {
		return 0, err
	}
	return tools.ToFloat(value)
}

func FetchWithArray(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher, model interface{}) (interface{}, error) {
	return fetch(ctx, cache, key, expire, fetcher, func(i interface{}) ([]byte, error) {
		kind := reflect.TypeOf(i).Kind()
		if kind != reflect.Slice && kind != reflect.Array {
			return nil, errors.ErrInvalidValue
		}
		return jsonEncode(i)
	}, func(value []byte) (interface{}, error) {
		ret := reflect.New(reflect.MakeSlice(typeFromModel(model), 0, 0).Type())
		err := json.Unmarshal(value, ret.Interface())
		if err != nil {
			return nil, err
		}
		return ret.Elem().Interface(), nil
	})
}
