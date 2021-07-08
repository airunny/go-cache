package go_cache

import (
	"bytes"
	"reflect"

	"github.com/golang/protobuf/proto"
)

type Decoder func(interface{}) (interface{}, error)

func typeFromModel(model interface{}) reflect.Type {
	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

func ProtoDecode(model interface{}) Decoder {
	return func(data interface{}) (interface{}, error) {
		var byteData []byte
		switch data.(type) {
		case []byte:
			byteData = data.([]byte)
		case string:
			byteData = []byte(data.(string))
		default:
			return data, nil
		}

		ret := reflect.New(typeFromModel(model))
		err := proto.Unmarshal(byteData, ret.Interface().(proto.Message))
		if err != nil {
			return nil, err
		}
		return ret.Interface(), nil
	}
}

func JsonDecode(model interface{}) Decoder {
	return func(data interface{}) (interface{}, error) {
		var byteData []byte
		switch data.(type) {
		case []byte:
			byteData = data.([]byte)
		case string:
			byteData = []byte(data.(string))
		default:
			return data, nil
		}

		ret := reflect.New(typeFromModel(model))
		err := json.NewDecoder(bytes.NewBuffer(byteData)).Decode(ret.Interface())
		if err != nil {
			return nil, err
		}
		return ret.Interface(), nil
	}
}
