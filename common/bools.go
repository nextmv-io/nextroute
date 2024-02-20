// Package common provides common functionality for nextroute.
package common

// Bools is a data structure for storing a sequence of booleans.
// Use NewBools to create a new instance.
type Bools []uint64

// NewBools creates a data structure for storing a sequence of booleans.
// It is optimized for memory usage by requiring roughly only 1 bit per boolean.
// Even though Get and Set run in constant time in a few nanoseconds, they are still slower.
// A single Get is around 2x slower (0.6 ns vs 0.3) and Set is around 7x slower (2.2 ns vs 0.3).
//
// The size argument is the number of booleans to store. It must not be negative.
// The initial value of each boolean is set to initial.
func NewBools(size int, initial bool) Bools {
	if size < 0 {
		panic("size must not be negative")
	}
	// We model the sequence of booleans as a sequence of 64-bit integers.
	// Each integer can store 64 booleans - one bool for each bit.
	values := make([]uint64, size/64+1)
	if initial {
		for i := 0; i < len(values); i++ {
			values[i] = ^uint64(0)
		}
	}
	return Bools(values)
}

// Set sets the ith bit to v. It panics if i is out of bounds.
func (b Bools) Set(i int, v bool) {
	var k uint64 = 1 << (i % 64)
	if v {
		// set the kth bit to 1
		b[i/64] |= k
	} else {
		// set kth bit to 0
		b[i/64] &= ^k
	}
}

// Get returns the value of the ith bit. It panics if i is out of bounds.
func (b Bools) Get(i int) bool {
	return b[i/64]&(1<<(i%64)) != 0
}
