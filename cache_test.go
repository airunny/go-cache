package go_cache

import (
	"context"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	redisCache "github.com/liyanbing/go-cache/cacher/redis"
	redis "gopkg.in/redis.v5"
)

var cache Cache

func TestMain(m *testing.M) {
	// redis
	redisCli := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	cache = redisCache.NewRedisCache(redisCli)
	cache.SetNamespace("test")

	// memory
	//cache = memory.NewMemoryCache(10)
	os.Exit(m.Run())
}

func TestFetchWithJson(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()

	cnt := int32(0)
	fetchFunc := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return &TempModel{
			Name: "peter",
			Age:  23,
			Id:   123123,
		}, nil
	}

	ret, err := FetchWithJson(ctx, cache, "json-key", time.Millisecond*100, fetchFunc, TempModel{})
	ast.Nil(err)
	ast.Equal("peter", ret.(*TempModel).Name)
	ast.EqualValues(23, ret.(*TempModel).Age)
	ast.EqualValues(123123, ret.(*TempModel).Id)
	ast.EqualValues(1, atomic.LoadInt32(&cnt))

	for i := 0; i < 10; i++ {
		// from cache
		ret, err = FetchWithJson(ctx, cache, "json-key", time.Millisecond*100, fetchFunc, TempModel{})
		ast.Nil(err)
		ast.Equal("peter", ret.(*TempModel).Name)
		ast.EqualValues(23, ret.(*TempModel).Age)
		ast.EqualValues(123123, ret.(*TempModel).Id)
		ast.EqualValues(1, atomic.LoadInt32(&cnt))
	}

	time.Sleep(100 * time.Millisecond)
	ret, err = FetchWithJson(ctx, cache, "json-key", time.Millisecond*100, fetchFunc, TempModel{})
	ast.Nil(err)
	ast.Equal("peter", ret.(*TempModel).Name)
	ast.EqualValues(23, ret.(*TempModel).Age)
	ast.EqualValues(123123, ret.(*TempModel).Id)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))

	// WithNoUseCache
	fetchFunc2 := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return &TempModel{
			Name: "mary",
			Age:  24,
			Id:   123,
		}, nil
	}
	_, err = FetchWithJson(ctx, cache, "json-key", time.Millisecond*100, fetchFunc2, TempModel{})
	ast.Nil(err)
	ast.Equal("peter", ret.(*TempModel).Name)
	ast.EqualValues(23, ret.(*TempModel).Age)
	ast.EqualValues(123123, ret.(*TempModel).Id)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))

	ctx2 := WithNoUseCache(ctx)
	ret, err = FetchWithJson(ctx2, cache, "json-key", time.Millisecond*100, fetchFunc2, TempModel{})
	ast.Nil(err)
	ast.Equal("mary", ret.(*TempModel).Name)
	ast.EqualValues(24, ret.(*TempModel).Age)
	ast.EqualValues(123, ret.(*TempModel).Id)
	ast.EqualValues(3, atomic.LoadInt32(&cnt))

	ret, err = FetchWithJson(ctx, cache, "json-key", time.Millisecond*100, fetchFunc2, TempModel{})
	ast.Nil(err)
	ast.Equal("mary", ret.(*TempModel).Name)
	ast.EqualValues(24, ret.(*TempModel).Age)
	ast.EqualValues(123, ret.(*TempModel).Id)
	ast.EqualValues(3, atomic.LoadInt32(&cnt))
}

func TestFetchWithJson2(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()

	type tmp struct {
		A []*TempModel `json:"a"`
	}

	cnt := int32(0)
	fetchFunc := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return &tmp{
			A: []*TempModel{
				{
					Name: "peter",
					Age:  23,
					Id:   123123,
				},
				{
					Name: "tome",
					Age:  33,
					Id:   123,
				},
			},
		}, nil
	}

	ret2, err := FetchWithJson(ctx, cache, "json-key1", time.Millisecond*100, fetchFunc, tmp{})
	ast.Nil(err)
	ret := ret2.(*tmp)
	ast.Equal(2, len(ret.A))
	ast.Equal("peter", ret.A[0].Name)
	ast.EqualValues(23, ret.A[0].Age)
	ast.EqualValues(123123, ret.A[0].Id)
	ast.Equal("tome", ret.A[1].Name)
	ast.EqualValues(33, ret.A[1].Age)
	ast.EqualValues(123, ret.A[1].Id)
	ast.EqualValues(1, atomic.LoadInt32(&cnt))

	for i := 0; i < 10; i++ {
		// from cache
		ret2, err = FetchWithJson(ctx, cache, "json-key1", time.Millisecond*100, fetchFunc, tmp{})
		ast.Nil(err)
		ret = ret2.(*tmp)
		ast.Equal(2, len(ret.A))
		ast.Equal("peter", ret.A[0].Name)
		ast.EqualValues(23, ret.A[0].Age)
		ast.EqualValues(123123, ret.A[0].Id)
		ast.Equal("tome", ret.A[1].Name)
		ast.EqualValues(33, ret.A[1].Age)
		ast.EqualValues(123, ret.A[1].Id)
		ast.EqualValues(1, atomic.LoadInt32(&cnt))
	}

	time.Sleep(100 * time.Millisecond)
	ret2, err = FetchWithJson(ctx, cache, "json-key1", time.Millisecond*100, fetchFunc, tmp{})
	ast.Nil(err)
	ret = ret2.(*tmp)
	ast.Equal(2, len(ret.A))
	ast.Equal("peter", ret.A[0].Name)
	ast.EqualValues(23, ret.A[0].Age)
	ast.EqualValues(123123, ret.A[0].Id)
	ast.Equal("tome", ret.A[1].Name)
	ast.EqualValues(33, ret.A[1].Age)
	ast.EqualValues(123, ret.A[1].Id)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))
}

func TestFetchWithString(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()

	cnt := int32(0)
	fetchFunc := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return "abc", nil
	}

	ret, err := FetchWithString(ctx, cache, "string-key", time.Millisecond*100, fetchFunc)
	ast.Nil(err)
	ast.Equal("abc", ret)
	ast.EqualValues(1, atomic.LoadInt32(&cnt))

	for i := 0; i < 10; i++ {
		// from cache
		ret, err = FetchWithString(ctx, cache, "string-key", time.Millisecond*100, fetchFunc)
		ast.Nil(err)
		ast.Equal("abc", ret)
		ast.EqualValues(1, atomic.LoadInt32(&cnt))
	}

	time.Sleep(100 * time.Millisecond)
	ret, err = FetchWithString(ctx, cache, "string-key", time.Millisecond*100, fetchFunc)
	ast.Nil(err)
	ast.Equal("abc", ret)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))

	// WithNoUseCache
	fetchFunc2 := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return "cba", nil
	}
	_, err = FetchWithString(ctx, cache, "string-key", time.Millisecond*100, fetchFunc2)
	ast.Nil(err)
	ast.Equal("abc", ret)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))

	ctx2 := WithNoUseCache(ctx)
	ret, err = FetchWithString(ctx2, cache, "string-key", time.Millisecond*100, fetchFunc2)
	ast.Nil(err)
	ast.Equal("cba", ret)
	ast.EqualValues(3, atomic.LoadInt32(&cnt))

	ret, err = FetchWithString(ctx, cache, "string-key", time.Millisecond*100, fetchFunc2)
	ast.Nil(err)
	ast.Equal("cba", ret)
	ast.EqualValues(3, atomic.LoadInt32(&cnt))
}

func TestFetchWithProtobuf(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()

	cnt := int32(0)
	fetchFunc := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return &TempModelPb{
			IsMember: true,
			ExpireAt: 101,
		}, nil
	}

	ret, err := FetchWithProtobuf(ctx, cache, "proto-key", time.Millisecond*100, fetchFunc, TempModelPb{})
	ast.Nil(err)
	ast.True(ret.(*TempModelPb).IsMember)
	ast.EqualValues(101, ret.(*TempModelPb).ExpireAt)
	ast.EqualValues(1, atomic.LoadInt32(&cnt))

	for i := 0; i < 10; i++ {
		// from cache
		ret, err = FetchWithProtobuf(ctx, cache, "proto-key", time.Millisecond*100, fetchFunc, TempModelPb{})
		ast.Nil(err)
		ast.True(ret.(*TempModelPb).IsMember)
		ast.EqualValues(101, ret.(*TempModelPb).ExpireAt)
		ast.EqualValues(1, atomic.LoadInt32(&cnt))
	}

	time.Sleep(100 * time.Millisecond)
	ret, err = FetchWithProtobuf(ctx, cache, "proto-key", time.Millisecond*100, fetchFunc, TempModelPb{})
	ast.Nil(err)
	ast.True(ret.(*TempModelPb).IsMember)
	ast.EqualValues(101, ret.(*TempModelPb).ExpireAt)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))
}

func TestFetchWithNumber(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()

	cnt := int32(0)
	fetchFunc := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return int64(9887), nil
	}

	ret, err := FetchWithNumber(ctx, cache, "number-key", time.Millisecond*100, fetchFunc)
	ast.Nil(err)
	ast.EqualValues(float64(9887), ret)
	ast.EqualValues(1, atomic.LoadInt32(&cnt))

	for i := 0; i < 10; i++ {
		// from cache
		ret, err = FetchWithNumber(ctx, cache, "number-key", time.Millisecond*100, fetchFunc)
		ast.Nil(err)
		ast.EqualValues(float64(9887), ret)
		ast.EqualValues(1, atomic.LoadInt32(&cnt))
	}

	time.Sleep(100 * time.Millisecond)
	ret, err = FetchWithNumber(ctx, cache, "number-key", time.Millisecond*100, fetchFunc)
	ast.Nil(err)
	ast.EqualValues(float64(9887), ret)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))

	fetchFunc2 := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return 1024, nil
	}
	_, err = FetchWithNumber(ctx, cache, "number-key", time.Millisecond*100, fetchFunc2)
	ast.Nil(err)
	ast.EqualValues(float64(9887), ret)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))

	// WithNoUseCache
	ctx2 := WithNoUseCache(ctx)
	ret, err = FetchWithNumber(ctx2, cache, "number-key", time.Millisecond*100, fetchFunc2)
	ast.Nil(err)
	ast.EqualValues(float64(1024), ret)
	ast.EqualValues(3, atomic.LoadInt32(&cnt))

	ret, err = FetchWithNumber(ctx, cache, "number-key", time.Millisecond*100, fetchFunc2)
	ast.Nil(err)
	ast.EqualValues(float64(1024), ret)
	ast.EqualValues(3, atomic.LoadInt32(&cnt))
}

func TestFetchWithArray(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()

	cnt := int32(0)
	fetchFunc := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return []*TempModel{
			{
				Name: "peter",
				Age:  23,
				Id:   123123,
			},
			{
				Name: "tome",
				Age:  33,
				Id:   123,
			},
		}, nil
	}

	ret, err := FetchWithArray(ctx, cache, "array-key", time.Millisecond*100, fetchFunc, []*TempModel{})
	ast.Nil(err)
	value, ok := ret.([]*TempModel)
	for _, v := range value {
		ast.Condition(func() (success bool) {
			success = (v.Name == "peter" && v.Age == 23 && v.Id == 123123) ||
				(v.Name == "tome" && v.Age == 33 && v.Id == 123)
			return
		})
	}
	ast.Equal(2, len(value))
	ast.True(ok)

	// from cache
	for i := 0; i < 10; i++ {
		ret, err := FetchWithArray(ctx, cache, "array-key", time.Millisecond*100, fetchFunc, []*TempModel{})
		ast.Nil(err)
		value, ok := ret.([]*TempModel)
		for _, v := range value {
			ast.Condition(func() (success bool) {
				success = (v.Name == "peter" && v.Age == 23 && v.Id == 123123) ||
					(v.Name == "tome" && v.Age == 33 && v.Id == 123)
				return
			})
		}
		ast.Equal(2, len(value))
		ast.True(ok)
	}

	time.Sleep(100 * time.Millisecond)
	ret, err = FetchWithArray(ctx, cache, "array-key", time.Millisecond*100, fetchFunc, []*TempModel{})
	ast.Nil(err)
	value, ok = ret.([]*TempModel)
	for _, v := range value {
		ast.Condition(func() (success bool) {
			success = (v.Name == "peter" && v.Age == 23 && v.Id == 123123) ||
				(v.Name == "tome" && v.Age == 33 && v.Id == 123)
			return
		})
	}
	ast.Equal(2, len(value))
	ast.True(ok)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))

	fetchFunc2 := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return []*TempModel{
			{
				Name: "peter1",
				Age:  233,
				Id:   1231233,
			},
			{
				Name: "tome1",
				Age:  333,
				Id:   1233,
			},
		}, nil
	}
	_, err = FetchWithArray(ctx, cache, "array-key", time.Millisecond*100, fetchFunc2, []*TempModel{})
	ast.Nil(err)
	value, ok = ret.([]*TempModel)
	for _, v := range value {
		ast.Condition(func() (success bool) {
			success = (v.Name == "peter" && v.Age == 23 && v.Id == 123123) ||
				(v.Name == "tome" && v.Age == 33 && v.Id == 123)
			return
		})
	}
	ast.Equal(2, len(value))
	ast.True(ok)
	ast.EqualValues(2, atomic.LoadInt32(&cnt))

	// WithNoUseCache
	ctx2 := WithNoUseCache(ctx)
	ret, err = FetchWithArray(ctx2, cache, "array-key", time.Millisecond*100, fetchFunc2, []*TempModel{})
	ast.Nil(err)
	value, ok = ret.([]*TempModel)
	for _, v := range value {
		ast.Condition(func() (success bool) {
			success = (v.Name == "peter1" && v.Age == 233 && v.Id == 1231233) ||
				(v.Name == "tome1" && v.Age == 333 && v.Id == 1233)
			return
		})
	}
	ast.Equal(2, len(value))
	ast.True(ok)
	ast.EqualValues(3, atomic.LoadInt32(&cnt))

	ret, err = FetchWithArray(ctx, cache, "array-key", time.Millisecond*100, fetchFunc2, []*TempModel{})
	ast.Nil(err)
	value, ok = ret.([]*TempModel)
	for _, v := range value {
		ast.Condition(func() (success bool) {
			success = (v.Name == "peter1" && v.Age == 233 && v.Id == 1231233) ||
				(v.Name == "tome1" && v.Age == 333 && v.Id == 1233)
			return
		})
	}
	ast.Equal(2, len(value))
	ast.True(ok)
	ast.EqualValues(3, atomic.LoadInt32(&cnt))
}

func TestFetchWithArray2(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()

	cnt := int32(0)
	fetchFunc := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return []string{
			"1",
			"2",
			"3",
			"4",
			"5",
		}, nil
	}

	ret, err := FetchWithArray(ctx, cache, "array-key1", time.Millisecond*100, fetchFunc, []string{})
	ast.Nil(err)
	value, ok := ret.([]string)
	ast.Equal(strings.Join(value, ","), "1,2,3,4,5")
	ast.Equal(5, len(value))
	ast.True(ok)

	// from cache
	for i := 0; i < 10; i++ {
		ret, err := FetchWithArray(ctx, cache, "array-key1", time.Millisecond*100, fetchFunc, []string{})
		ast.Nil(err)
		value, ok := ret.([]string)
		ast.Equal(strings.Join(value, ","), "1,2,3,4,5")
		ast.Equal(5, len(value))
		ast.True(ok)
	}

	time.Sleep(100 * time.Millisecond)
	ret, err = FetchWithArray(ctx, cache, "array-key1", time.Millisecond*100, fetchFunc, []string{})
	ast.Nil(err)
	value, ok = ret.([]string)
	ast.Equal(strings.Join(value, ","), "1,2,3,4,5")
	ast.Equal(5, len(value))
	ast.True(ok)

	fetchFunc2 := func() (interface{}, error) {
		atomic.AddInt32(&cnt, 1)
		return []string{
			"11",
			"22",
			"33",
			"44",
			"55",
		}, nil
	}
	_, err = FetchWithArray(ctx, cache, "array-key1", time.Millisecond*100, fetchFunc2, []string{})
	ast.Nil(err)
	value, ok = ret.([]string)
	ast.Equal(strings.Join(value, ","), "1,2,3,4,5")
	ast.Equal(5, len(value))
	ast.True(ok)

	// WithNoUseCache
	ctx2 := WithNoUseCache(ctx)
	ret, err = FetchWithArray(ctx2, cache, "array-key1", time.Millisecond*100, fetchFunc2, []string{})
	ast.Nil(err)
	value, ok = ret.([]string)
	ast.Equal(strings.Join(value, ","), "11,22,33,44,55")
	ast.Equal(5, len(value))
	ast.True(ok)

	ret, err = FetchWithArray(ctx, cache, "array-key1", time.Millisecond*100, fetchFunc2, []string{})
	ast.Nil(err)
	value, ok = ret.([]string)
	ast.Equal(strings.Join(value, ","), "11,22,33,44,55")
	ast.Equal(5, len(value))
	ast.True(ok)
}

type TempModelPb struct {
	IsMember bool  `protobuf:"varint,1,opt,name=is_member,json=isMember" json:"is_member,omitempty"`
	ExpireAt int64 `protobuf:"varint,2,opt,name=expire_at,json=expireAt" json:"expire_at,omitempty"`
}

func (m *TempModelPb) Reset()                    { *m = TempModelPb{} }
func (m *TempModelPb) String() string            { return proto.CompactTextString(m) }
func (*TempModelPb) ProtoMessage()               {}
func (*TempModelPb) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

var fileDescriptor0 = []byte{
	// 582 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x55, 0xdb, 0x6e, 0xda, 0x40,
	0x10, 0x0d, 0xa1, 0x09, 0x64, 0x12, 0x52, 0xb3, 0xa5, 0x29, 0x71, 0x5a, 0x35, 0xf2, 0x53, 0x54,
	0xb5, 0xa8, 0x4d, 0xfb, 0x03, 0x10, 0x22, 0x44, 0xa5, 0xa8, 0xc8, 0x20, 0x55, 0xea, 0x0b, 0x32,
	0xde, 0x11, 0x59, 0x09, 0xbc, 0x96, 0xbd, 0x46, 0xe1, 0xdb, 0xfa, 0x73, 0xd5, 0xae, 0xd7, 0x31,
	0xbe, 0x95, 0xf6, 0x8d, 0x99, 0x73, 0xe6, 0xcc, 0xce, 0xc5, 0x03, 0x9c, 0xf9, 0x62, 0xc9, 0x57,
	0xb4, 0xe7, 0x07, 0x5c, 0x70, 0x72, 0x1c, 0x5b, 0xd6, 0x17, 0xe8, 0x4c, 0x23, 0xd7, 0x45, 0xa4,
	0x83, 0x68, 0xfb, 0x80, 0xeb, 0x05, 0x06, 0xfd, 0x60, 0x19, 0x92, 0x4b, 0x68, 0x3a, 0x11, 0x65,
	0x62, 0xce, 0x68, 0xb7, 0x76, 0x5d, 0xbb, 0x79, 0x61, 0x37, 0x94, 0x3d, 0xa6, 0xd6, 0x45, 0x31,
	0xc4, 0xc6, 0xd0, 0xb7, 0x7a, 0xd0, 0x1e, 0xa1, 0x18, 0x7b, 0x1b, 0x26, 0xf0, 0x8e, 0x53, 0x4c,
	0x74, 0x16, 0xd1, 0x16, 0x83, 0x1d, 0x1d, 0x65, 0x8f, 0xa9, 0xf5, 0x2d, 0xc7, 0x97, 0x22, 0xe4,
	0x3d, 0x9c, 0x32, 0xe5, 0x99, 0xbb, 0x9c, 0x62, 0xf7, 0xf0, 0xba, 0x76, 0x73, 0x62, 0x03, 0x7b,
	0x26, 0x59, 0x13, 0x20, 0x03, 0xe6, 0xd1, 0x7f, 0x4e, 0xb3, 0x5f, 0xb1, 0x93, 0x57, 0x54, 0xd5,
	0x7c, 0x86, 0x57, 0x23, 0x14, 0x71, 0x79, 0x53, 0xe1, 0x88, 0x28, 0xdc, 0x57, 0xcf, 0x8f, 0x42,
	0x84, 0xaa, 0xe8, 0x0a, 0x4e, 0x58, 0x38, 0x5f, 0x2b, 0xb7, 0x0a, 0x69, 0xda, 0x4d, 0x16, 0xc6,
	0x34, 0x09, 0xe2, 0x93, 0xcf, 0x02, 0x9c, 0x3b, 0x42, 0x3d, 0xad, 0x6e, 0x37, 0x63, 0x47, 0x5f,
	0x58, 0x9f, 0xc0, 0x98, 0x04, 0x58, 0x98, 0x4b, 0x55, 0xfe, 0x1c, 0x5d, 0x25, 0xff, 0xcb, 0x18,
	0x7f, 0x01, 0xf9, 0xe9, 0xac, 0x56, 0x28, 0xee, 0x9f, 0xdc, 0x47, 0xc7, 0x5b, 0xc6, 0x8d, 0xfc,
	0x08, 0x75, 0xb1, 0xf5, 0x15, 0xf7, 0xfc, 0xd6, 0xec, 0xe9, 0x9d, 0xc9, 0x12, 0x67, 0x5b, 0x1f,
	0x6d, 0x49, 0x23, 0x17, 0x70, 0x1c, 0xf2, 0x28, 0x70, 0x51, 0xbf, 0x5d, 0x5b, 0xb2, 0xa5, 0xd9,
	0x90, 0x64, 0x41, 0x62, 0xef, 0x10, 0x7d, 0x1e, 0x32, 0xb1, 0xaf, 0xa0, 0xfb, 0x1c, 0x5f, 0x55,
	0xd4, 0x81, 0x23, 0x97, 0x33, 0x2f, 0x54, 0xe4, 0xba, 0x1d, 0x1b, 0xc4, 0x84, 0x26, 0x65, 0xce,
	0x9a, 0x7b, 0x34, 0x4c, 0xda, 0x98, 0xd8, 0xd6, 0x10, 0x5e, 0x4f, 0xa2, 0xc0, 0x7d, 0x74, 0x42,
	0x1c, 0x6c, 0x87, 0xb1, 0x37, 0x49, 0xcd, 0x03, 0x9a, 0x49, 0xad, 0xec, 0x31, 0x25, 0x06, 0xd4,
	0xbd, 0x68, 0xad, 0xa5, 0xe4, 0x4f, 0xeb, 0x4d, 0x89, 0x8a, 0xaa, 0xaa, 0x0f, 0x24, 0x05, 0xee,
	0x38, 0xf3, 0xfe, 0x5f, 0xbb, 0x93, 0x97, 0x90, 0xc2, 0x1f, 0x46, 0xf9, 0x26, 0xca, 0xbe, 0x93,
	0x53, 0x68, 0x8c, 0xbd, 0x8d, 0xb3, 0x62, 0xd4, 0x38, 0x20, 0x6d, 0x68, 0x49, 0xfa, 0x8c, 0xeb,
	0x07, 0x19, 0x35, 0xe9, 0xd2, 0xc6, 0x8c, 0x4b, 0xcc, 0x38, 0xbc, 0xfd, 0x7d, 0x04, 0x8d, 0x29,
	0x06, 0x1b, 0xe6, 0x22, 0x19, 0x41, 0x2b, 0xd3, 0x53, 0x72, 0x99, 0x9d, 0xf1, 0xce, 0x68, 0xcc,
	0x72, 0x48, 0x15, 0x7d, 0x40, 0xbe, 0xc3, 0x79, 0xf6, 0x75, 0xa4, 0x62, 0x5b, 0x94, 0x54, 0x05,
	0x96, 0x6a, 0x65, 0xeb, 0x4f, 0xb5, 0x8a, 0xad, 0x35, 0x2b, 0x30, 0xad, 0x35, 0x85, 0x76, 0x61,
	0x4e, 0xe4, 0x5d, 0x31, 0x64, 0x67, 0x11, 0xcc, 0x6a, 0x58, 0x8b, 0x0e, 0xe1, 0x6c, 0xf7, 0xd3,
	0x22, 0xdd, 0xe7, 0x80, 0xdc, 0xf7, 0x69, 0x96, 0x22, 0x5a, 0x65, 0x02, 0x46, 0xfe, 0x70, 0x92,
	0xb7, 0x09, 0xbf, 0xec, 0x0a, 0x9b, 0x95, 0xa8, 0x56, 0x7c, 0x80, 0x97, 0xb9, 0x93, 0x43, 0xae,
	0x92, 0x90, 0x92, 0xeb, 0x65, 0x56, 0x81, 0x5a, 0x6e, 0x04, 0xad, 0xcc, 0x45, 0x4e, 0x97, 0xa3,
	0x70, 0xd8, 0xcd, 0x72, 0x28, 0x1d, 0x68, 0xf6, 0xa4, 0xa6, 0x03, 0x2d, 0x1e, 0x6f, 0xb3, 0x02,
	0x8b, 0xb5, 0x16, 0xc7, 0xea, 0x0f, 0xeb, 0xeb, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xfc, 0x99,
	0xe9, 0x37, 0xc0, 0x06, 0x00, 0x00,
}

type TempModel struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Id   uint64 `json:"id"`
}
