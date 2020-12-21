package go_cache

import (
	"bytes"
	"reflect"

	"github.com/golang/protobuf/proto"
)

type decoder func([]byte) (interface{}, error)

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

func jsonDecode(model interface{}) decoder {
	return func(data []byte) (interface{}, error) {
		ret := reflect.New(typeFromModel(model))
		err := json.NewDecoder(bytes.NewBuffer(data)).Decode(ret.Interface())
		if err != nil {
			return nil, err
		}
		return ret.Interface(), nil
	}
}
