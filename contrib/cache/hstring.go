package cache

import (
	"github.com/cespare/xxhash/v2"

	"github.com/byte4ever/conch/domain"
)

// HashableString is a string type that can be hashed into a unique Key.
type HashableString string

// Hash computes a unique Key for the HashableString instance.
// The Key combines results from MemHashString and xxhash.Sum64String.
func (h HashableString) Hash() domain.Key {
	return domain.Key{
		A: MemHashString(string(h)),
		B: xxhash.Sum64String(string(h)),
	}
}
