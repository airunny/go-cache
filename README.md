##### go-cache

###### 在具体的获取数据操作(fetcher)前加一层缓存操作


# go-cache

* 1、先根据key从cache中获取数据，如果不存在则从fetcher中获取数据（并发调用时只会有一个请求会调用fetcher,其他请求会复用这个请求返回的数据）

* 2、从fetcher获取到数据之后然后存储到cache中,然后返回从fetcher获取到的对象

* 3、如果cache中存在数据，则decode进model对象中返回（注意这时候返回的是model的指针）
 
tips 
  > 注意：如果对象在缓存中存在则一定返回的是对象指针，如果不存在返回的是fetcher返回的数据(为了统一fetcher最好也返回对象的指针)

## 安装 
go get github.com/liyanbing/go-cache

## 使用 
```go
package main

import (
	"encoding/json"
    "net/http"
    "context"

    redis_cacher "github.com/liyanbing/go-cache/cacher/redis"
	redis "gopkg.in/redis.v5"
    "github.com/liyanbing/go-cache"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Id   uint64 `json:"id"`
}

func GetUser() (interface{}, error) {
	return &User{
		Name: "peter",
		Age:  23,
		Id:   123123,
	}, nil
}


func main() {
	redisCli := redis.NewClient(&redis.Options{
    		Addr: "127.0.0.1:6379",
    		DB:   0,
    	})
	cache := redis_cacher.NewRedisCache(redisCli)
    ret, err := go-cache.FetchWithJson(ctx, cache, "json-key", time.Millisecond*100, fetchFunc, TempModel{})
}
```