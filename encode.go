package go_cache

import (
	"github.com/golang/protobuf/proto"
	"github.com/liyanbing/go-cache/errors"
)

type encoder func(interface{}) ([]byte, error)

func protoEncode(value interface{}) ([]byte, error) {
	mes, ok := value.(proto.Message)
	if !ok {
		return nil, errors.ErrInvalidValue
	}
	return proto.Marshal(mes)
}

func jsonEncode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}
