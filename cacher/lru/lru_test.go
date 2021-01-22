package lru

import (
	"context"
	"testing"

	"github.com/liyanbing/go-cache/errors"
	"github.com/stretchr/testify/assert"
)

func TestLRU_Get(t *testing.T) {
	instance := NewLRU(2)
	// set global namespace
	instance.SetNamespace("test")

	_, err := instance.Get(context.Background(), "name")
	assert.Equal(t, errors.ErrEmptyCache, err)

	// set name value
	err = instance.Set(context.Background(), "name", "value", 0)
	assert.Nil(t, err)

	// get name value
	value, err := instance.Get(context.Background(), "name")
	assert.Nil(t, err)
	assert.Equal(t, "value", value)

	// set name1 value1
	err = instance.Set(context.Background(), "name1", "value1", 0)
	assert.Nil(t, err)
	// get name
	value, err = instance.Get(context.Background(), "name")
	assert.Nil(t, err)
	assert.Equal(t, "value", value)

	// get name1
	value, err = instance.Get(context.Background(), "name1")
	assert.Nil(t, err)
	assert.Equal(t, "value1", value)

	// get name
	value, err = instance.Get(context.Background(), "name")
	assert.Nil(t, err)
	assert.Equal(t, "value", value)

	// set name2 value2
	err = instance.Set(context.Background(), "name2", "value2", 0)
	assert.Nil(t, err)

	// get name2
	value, err = instance.Get(context.Background(), "name2")
	assert.Nil(t, err)
	assert.Equal(t, "value2", value)

	// get name
	value, err = instance.Get(context.Background(), "name")
	assert.Nil(t, err)
	assert.Equal(t, "value", value)

	// get name1
	value, err = instance.Get(context.Background(), "name1")
	assert.Equal(t, errors.ErrEmptyCache, err)

	// remove name1
	err = instance.Remove(context.Background(), "name1")
	assert.Nil(t, err)

	// get name1
	value, err = instance.Get(context.Background(), "name1")
	assert.Equal(t, errors.ErrEmptyCache, err)
}
