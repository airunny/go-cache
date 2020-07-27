package go_cache

import (
	"context"
)

var (
	noCache struct{}
)

func WithNoUseCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, noCache, struct{}{})
}

func noUseCache(ctx context.Context) bool {
	return ctx.Value(noCache) != nil
}
