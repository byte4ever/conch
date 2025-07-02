package domain

// Key represents a composite of two 64-bit unsigned integers.
type Key struct {
	A, B uint64
}

// Values returns the two uint64 components of the Key: A and B.
func (k Key) Values() (uint64, uint64) {
	return k.A, k.B
}
