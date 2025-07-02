package cache

import (
	"github.com/byte4ever/conch/domain"
)

// HashableUint64 represents a hashable 64-bit unsigned integer type.
type HashableUint64 uint32

// Hash computes and returns a `Key` using the `HashableUint64` value.
func (h HashableUint64) Hash() domain.Key {
	return domain.Key{
		A: uint64(h),
		B: 0,
	}
}
