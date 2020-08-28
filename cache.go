package go_cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/golang/groupcache/singleflight"
	"github.com/golang/protobuf/proto"

	"github.com/liyanbing/go-cache/errors"
	"github.com/liyanbing/go-cache/tools"
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
)

type Fetcher func() (interface{}, error)

type Cache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}

func FetchWithJson(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher, model interface{}) (interface{}, error) {
	return fetch(ctx, cache, key, expire, fetcher, jsonEncode, jsonDecode(model))
}

func FetchWithString(ctx context.Context, cache Cache, key string, expire time.Duration, fetcher Fetcher) (string, error) {
	value, err := fetch(ctx, cache, key, expire, fetcher, func(input interface{}) ([]byte, error) {
		return []byte(input.(string)), nil
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

type encoder func(interface{}) ([]byte, error)

type decoder func(value []byte) (interface{}, error)

func fetch(
	ctx context.Context,
	cache Cache,
	key string,
	expire time.Duration,
	fetcher Fetcher,
	e encoder,
	d decoder) (interface{}, error) {

	do := func() (interface{}, error) {
		return single.Do(key, func() (interface{}, error) {
			value, err := fetcher()
			if err != nil {
				return nil, err
			}

			cacheData, err := e(value)
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
	return d(cacheData)
}

func typeFromModel(model interface{}) reflect.Type {
	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

func protoDecode(model interface{}) decoder {
	return func(data []byte) (interface{}, error) {
		ret := reflect.New(typeFromModel(model))
		err := proto.Unmarshal(data, ret.Interface().(proto.Message))
		if err != nil {
			return nil, err
		}
		return ret.Interface(), nil
	}
}

func protoEncode(value interface{}) ([]byte, error) {
	mes, ok := value.(proto.Message)
	if !ok {
		return nil, errors.ErrInvalidValue
	}
	return proto.Marshal(mes)
}

func jsonDecode(model interface{}) decoder {
	return func(value []byte) (interface{}, error) {
		ret := reflect.New(typeFromModel(model))
		err := json.NewDecoder(bytes.NewBuffer(value)).Decode(ret.Interface())
		if err != nil {
			return nil, err
		}
		return ret.Interface(), nil
	}
}

func jsonEncode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}
