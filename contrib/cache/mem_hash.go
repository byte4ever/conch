package cache

import (
	"unsafe"
)

// MemHash computes a hash value for the given byte slice.
// It uses the runtime's memhash function for hashing.
// The input slice is converted into a stringStruct for processing.
// Returns a uint64 hash value.
func MemHash(data []byte) uint64 {
	ss := (*stringStruct)(unsafe.Pointer(&data))
	return uint64(memhash(ss.str, 0, uintptr(ss.len)))
}
