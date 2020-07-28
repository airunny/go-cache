package errors

import "errors"

var (
	ErrEmptyCache   = errors.New("empty value")
	ErrInvalidValue = errors.New("invalid value")
)
