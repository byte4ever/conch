package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"

	"github.com/byte4ever/conch"
)

func RistrettoKeyToHash(key interface{}) (uint64, uint64) {
	return key.(conch.Key).Values()
}

type RistrettoWrapper[P conch.Hashable, R interface{}] struct {
	Cache *ristretto.Cache
	ttl   time.Duration
}

func WrapRistretto[P conch.Hashable, R interface{}](
	cache *ristretto.Cache,
	ttl time.Duration,
) *RistrettoWrapper[P, R] {
	return &RistrettoWrapper[P, R]{
		Cache: cache,
		ttl:   ttl,
	}
}

func (w *RistrettoWrapper[P, R]) Get(
	ctx context.Context,
	key P,
) (R, bool) {
	var zeroR R

	cacheVal, found := w.Cache.Get(key.Hash())
	if !found {
		return zeroR, false
	}

	val, ok := cacheVal.(R)
	if !ok {
		panic("invalid type in cache")
	}

	return val, true
}

func (w *RistrettoWrapper[P, R]) Store(
	_ context.Context,
	key P,
	value R,
) {
	w.Cache.SetWithTTL(key.Hash(), value, 1, w.ttl)
}
