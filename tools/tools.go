package tools

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/liyanbing/go-cache/errors"
)

func CanConvertToNumber(in interface{}) bool {
	switch in.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case string:
		_, err := strconv.ParseFloat(in.(string), 64)
		return err == nil
	case []byte:
		_, err := strconv.ParseFloat(string(in.([]byte)), 64)
		return err == nil
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
	case []byte:
		return strconv.ParseFloat(string(input.([]byte)), 64)
	}
	return 0, errors.ErrInvalidValue
}

func ToString(input interface{}) (string, error) {
	switch input.(type) {
	case string:
		return input.(string), nil
	case []byte:
		return string(input.([]byte)), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", input), nil
	case bool:
		if input.(bool) {
			return "1", nil
		}
		return "0", nil
	default:
		value, err := json.Marshal(input)
		if err != nil {
			return "", err
		}
		return string(value), nil
	}
}
