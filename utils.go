package go_cache

import (
	"context"
	"log"

	"github.com/liyanbing/go-cache/errors"
)

func fetch(
	ctx context.Context,
	cache Cache,
	key string,
	fetcher Fetcher,
	e encoder,
	d decoder) (interface{}, error) {

	do := func() (interface{}, error) {
		return single.Do(key, func() (interface{}, error) {
			value, expires, err := fetcher()
			if err != nil {
				return nil, err
			}

			cacheData, err := e(value)
			if err != nil {
				return nil, err
			}

			err = cache.Set(ctx, key, cacheData, expires)
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
