package cache

import (
	"unsafe"
)

// memhash computes a hash value for the memory referenced by pointer p.
// h is the seed value, and s is the length of the memory block.
//
//go:noescape
//go:linkname memhash runtime.memhash
func memhash(p unsafe.Pointer, h, s uintptr) uintptr

// stringStruct is an internal representation of a string in memory.
// It mirrors the runtime string structure to access the underlying data pointer and length.
// This is used for efficient string hashing without additional allocations.
type stringStruct struct {
	str unsafe.Pointer // Pointer to the string data
	len int            // Length of the string
}

// MemHashString computes a hash for the given string using runtime.memhash.
// It leverages an internal stringStruct to access the string's pointer and length.
// Returns a uint64 hash value that can be used for various hashing purposes.
func MemHashString(str string) uint64 {
	ss := (*stringStruct)(unsafe.Pointer(&str))
	return uint64(memhash(ss.str, 0, uintptr(ss.len)))
}
