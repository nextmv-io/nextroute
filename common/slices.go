// © 2019-present nextmv.io inc

package common

// CopySliceFrom cuts of a slice from `alloc` and copies the data from `data`
// into it. It returns the new slice and the remaining slice of `alloc`. This
// can be used in places where we allocate once and copy multiple times.
func CopySliceFrom[T any](alloc []T, data []T) ([]T, []T) {
	n := len(data)
	newData, alloc := alloc[:n], alloc[n:]
	copy(newData, data)
	return newData, alloc
}
