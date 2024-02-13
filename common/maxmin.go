package common

// Max returns the larger of a and b.
func Max[T number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Min returns the smaller of a and b.
func Min[T number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint |
		~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}
