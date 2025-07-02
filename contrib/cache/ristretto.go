package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"

	"github.com/byte4ever/conch/domain"
)

// RistrettoKeyToHash converts a Ristretto key to two 64-bit hash values.
func RistrettoKeyToHash(key any) (uint64, uint64) {
	return key.(domain.Key).Values()
}

// RistrettoWrapper is a generic wrapper for managing a Ristretto cache.
// It associates a time-to-live (ttl) with cached objects.
// P is the key type implementing the conch.Hashable interface.
// R represents the type of values stored in the cache.
// The Cache field holds the underlying Ristretto cache instance.
// The ttl field specifies the duration objects are kept in the cache.
type RistrettoWrapper[P domain.Hashable, R any] struct {
	Cache *ristretto.Cache
	ttl   time.Duration
}

// WrapRistretto wraps a `ristretto.Cache` to provide additional functionality.
// It returns a `RistrettoWrapper` with specified key and value types and TTL.
func WrapRistretto[P domain.Hashable, R any](
	cache *ristretto.Cache,
	ttl time.Duration,
) *RistrettoWrapper[P, R] {
	return &RistrettoWrapper[P, R]{
		Cache: cache,
		ttl:   ttl,
	}
}

// Get retrieves a value from the cache using the provided key.
// Returns the value and a boolean indicating if the key was found.
func (w *RistrettoWrapper[P, R]) Get(
	_ context.Context,
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

// Store inserts a key-value pair into the cache with a specified TTL.
// The key is hashed using the Hash method from the Hashable interface.
// The value is stored in the underlying Ristretto cache.
func (w *RistrettoWrapper[P, R]) Store(
	_ context.Context,
	key P,
	value R,
) {
	w.Cache.SetWithTTL(key.Hash(), value, 1, w.ttl)
}
