package tools

import (
	"strconv"

	"github.com/liyanbing/go-cache/errors"
)

func IsNumber(in interface{}) bool {
	switch in.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	}
	return false
}

func ToFloat(input interface{}) (float64, error) {
	switch input.(type) {
	case int:
		return float64(input.(int)), nil
	case uint:
		return float64(input.(uint)), nil
	case int8:
		return float64(input.(int8)), nil
	case uint8:
		return float64(input.(uint8)), nil
	case int16:
		return float64(input.(int16)), nil
	case uint16:
		return float64(input.(uint16)), nil
	case int32:
		return float64(input.(int32)), nil
	case uint32:
		return float64(input.(uint32)), nil
	case int64:
		return float64(input.(int64)), nil
	case uint64:
		return float64(input.(uint64)), nil
	case float32:
		return float64(input.(float32)), nil
	case float64:
		return input.(float64), nil
	case string:
		return strconv.ParseFloat(input.(string), 64)
	}
	return 0, errors.ErrInvalidValue
}
