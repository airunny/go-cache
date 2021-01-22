package errors

import "errors"

var (
	ErrEmptyCache        = errors.New("empty value")
	ErrInvalidValue      = errors.New("invalid value")
	ErrInvalidCacheValue = errors.New("value from cache should be []byte")
)
