package cache

import (
	"context"

	"github.com/byte4ever/conch/domain"
)

// WrapProcessorFunc wraps a ProcessorFunc with caching functionality.
// It retrieves results from the cache if available, otherwise executes the
// ProcessorFunc, caching the result if no error occurs.
// The cache is indexed based on the parameters of the Processable elements.
func WrapProcessorFunc[
	T domain.Processable[Param, Result],
	Param domain.Hashable,
	Result any](
	cache domain.Cache[Param, Result],
	f domain.ProcessorFunc[T, Param, Result],
) domain.ProcessorFunc[T, Param, Result] {
	return func(ctx context.Context, elem T) {
		params := elem.GetParam()

		if value, found := cache.Get(ctx, params); found {
			elem.SetValue(value)
			return
		}

		f(ctx, elem)

		if elem.GetError() == nil {
			cache.Store(ctx, params, elem.GetValue())
		}
	}
}
