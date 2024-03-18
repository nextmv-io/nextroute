// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"slices"
	"sync"
	"time"
)

// Filter filters a slice using a predicate function.
func Filter[T any](v []T, f func(T) bool) []T {
	r := make([]T, 0, len(v))
	for _, x := range v {
		if f(x) {
			r = append(r, x)
		}
	}
	return r
}

// Unique is a universal duplicate removal function for type instances in
// a slice that implement the comparable interface.
func Unique[T comparable](s []T) []T {
	inResult := make(map[T]bool)
	var result []T
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}

// UniqueDefined is a universal duplicate removal function for type instances in
// a slice that implement the comparable interface. The function f is used to
// extract the comparable value from the type instance.
func UniqueDefined[T any, I comparable](items []T, f func(T) I) []T {
	inResult := make(map[I]bool)
	var result []T
	for _, item := range items {
		if _, ok := inResult[f(item)]; !ok {
			inResult[f(item)] = true
			result = append(result, item)
		}
	}
	return result
}

// NotUnique returns the duplicate instances.
func NotUnique[T comparable](s []T) []T {
	inResult := make(map[T]bool)
	var result []T
	for _, str := range s {
		if _, ok := inResult[str]; ok {
			result = append(result, str)
		} else {
			inResult[str] = true
		}
	}
	return result
}

// NotUniqueDefined returns the instances for which f returns identical
// values.
func NotUniqueDefined[T any, I comparable](items []T, f func(T) I) []T {
	inResult := make(map[I]bool)
	var result []T
	for _, item := range items {
		if _, ok := inResult[f(item)]; ok {
			result = append(result, item)
		} else {
			inResult[f(item)] = true
		}
	}
	return result
}

// GroupBy groups the elements of a slice by a key function.
func GroupBy[T any, K comparable](s []T, f func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, instance := range s {
		key := f(instance)
		if _, ok := result[key]; !ok {
			result[key] = make([]T, 0)
		}
		result[key] = append(result[key], instance)
	}
	return result
}

// Map maps a slice of type T to a slice of type R using the function f.
func Map[T any, R any](v []T, f func(T) R) []R {
	r := make([]R, len(v))
	for idx, x := range v {
		r[idx] = f(x)
	}
	return r
}

// MapSlice maps a slice of type T to a slice of type R using the function f
// returning a slice of R.
func MapSlice[T any, R any](v []T, f func(T) []R) []R {
	r := make([]R, 0)
	for _, x := range v {
		r = append(r, f(x)...)
	}
	return r
}

// FindIndex returns the first index i satisfying predicate(s[i]).
func FindIndex[E any](s []E, predicate func(E) bool) int {
	for i, v := range s {
		if predicate(v) {
			return i
		}
	}
	return -1
}

// AllTrue returns true if all the given predicate evaluations are true.
func AllTrue[E any](s []E, predicate func(E) bool) bool {
	return All(s, true, predicate)
}

// AllFalse returns true if all the given predicate evaluations is false.
func AllFalse[E any](s []E, predicate func(E) bool) bool {
	return All(s, false, predicate)
}

// All returns true if all the given predicate evaluations evaluate to
// condition.
func All[E any](s []E, condition bool, predicate func(E) bool) bool {
	if len(s) == 0 {
		return false
	}
	for _, v := range s {
		if predicate(v) != condition {
			return false
		}
	}
	return true
}

// HasTrue returns true if any of the given predicate evaluations evaluate to
// true.
func HasTrue[E any](s []E, predicate func(E) bool) bool {
	return Has(s, true, predicate)
}

// HasFalse returns true if any of the given predicate evaluations evaluate to
// false.
func HasFalse[E any](s []E, predicate func(E) bool) bool {
	return Has(s, false, predicate)
}

// Has returns true if any of the given predicate evaluations is condition.
func Has[E any](s []E, condition bool, predicate func(E) bool) bool {
	for _, v := range s {
		if predicate(v) == condition {
			return true
		}
	}
	return false
}

// CopyMap copies all key/value pairs in source adding them to destination.
// If a key exists in both maps, the value in destination is overwritten.
func CopyMap[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](
	destination M1,
	source M2) {
	for k, v := range source {
		destination[k] = v
	}
}

// DefensiveCopy returns a defensive copy of a slice.
func DefensiveCopy[T any](v []T) []T {
	c := make([]T, len(v))
	copy(c, v)
	return c
}

// WithinTolerance returns true if a and b are within the given tolerance.
func WithinTolerance(a, b, tolerance float64) bool {
	if a == b {
		return true
	}
	d := math.Abs(a - b)
	if b == 0 {
		return d < tolerance
	}
	return (d / math.Abs(b)) < tolerance
}

// Truncate truncates a float64 to the given unit.
func Truncate(f float64, unit float64) float64 {
	bf := big.NewFloat(0).SetPrec(1000).SetFloat64(f)
	bu := big.NewFloat(0).SetPrec(1000).SetFloat64(unit)

	bf.Quo(bf, bu)

	i := big.NewInt(0)
	bf.Int(i)
	bf.SetInt(i)

	f, _ = bf.Mul(bf, bu).Float64()
	return f
}

// DurationValue returns the value of a duration in the given time unit.
// Will panic if the time unit is zero.
func DurationValue(
	distance Distance,
	speed Speed,
	timeUnit time.Duration,
) float64 {
	if timeUnit.Seconds() == 0 {
		panic(
			fmt.Errorf(
				"time unit is zero in duration calculation",
			),
		)
	}
	seconds := distance.Value(Meters) / speed.Value(MetersPerSecond)

	return seconds / timeUnit.Seconds()
}

// RandomElement returns a random element from the given slice. If the slice is
// empty, panic is raised. If source is nil, a new source is created using the
// current time.
func RandomElement[T any](
	source *rand.Rand,
	elements []T,
) T {
	if len(elements) == 0 {
		panic(fmt.Errorf("cannot select random element from empty slice"))
	}
	if source == nil {
		// math/rand is about 50 to 100 times faster than crypto/rand.
		// Also math/rand sequence is reproducible when given same seed. This is good for testing/debugging.
		// The rand use case here has no security concern.
		// G404 (CWE-338): Use of weak random number generator (math/rand instead of crypto/rand).
		/* #nosec */
		source = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return elements[source.Intn(len(elements))]
}

// RandomElements returns a slice of n random elements from the
// given slice. If n is greater than the length of the slice, all elements are
// returned. If n is less than or equal to zero, an empty slice is returned.
// If source is nil, a new source is created using the current time.
func RandomElements[T any](
	source *rand.Rand,
	elements []T,
	n int,
) []T {
	if source == nil {
		/* #nosec */
		source = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	if n <= 0 {
		return []T{}
	}
	if n >= len(elements) {
		return DefensiveCopy(elements)
	}
	result := make([]T, 0, n)
	indicesUsed := make(map[int]bool, 0)
	for i := 0; i < n; i++ {
		index := RandomIndex(
			source,
			len(elements),
			indicesUsed,
		)
		result = append(result, elements[index])
	}

	return result
}

// RandomElementIndices returns a slice of n random element indices from the
// given slice. If n is greater than the length of the slice, all indices are
// returned. If n is less than or equal to zero, an empty slice is returned.
// If source is nil, a new source is created using the current time.
func RandomElementIndices[T any](
	source *rand.Rand,
	elements []T,
	n int,
) []int {
	if source == nil {
		/* #nosec */
		source = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	if n <= 0 {
		return []int{}
	}
	if n >= len(elements) {
		result := make([]int, n)
		for i := 0; i < n; i++ {
			result[n] = i
		}
		return result
	}
	result := make([]int, 0, n)
	indicesUsed := make(map[int]bool, 0)
	for i := 0; i < n; i++ {
		index := RandomIndex(
			source,
			len(elements),
			indicesUsed,
		)
		result = append(result, index)
	}

	return result
}

// RandomIndex returns a random index from the given size. If the index has
// already been used, a new index is generated. If source is nil, a new source
// is created using the current time.
func RandomIndex(source *rand.Rand, size int, indicesUsed map[int]bool) int {
	if source == nil {
		/* #nosec */
		source = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	for {
		index := source.Intn(size)
		if used, ok := indicesUsed[index]; !ok || !used {
			indicesUsed[index] = true
			return index
		}
	}
}

// Shuffle returns a shuffled copy of the given slice. If source is nil, a new
// source is created using the current time.
func Shuffle[T any](source *rand.Rand, slice []T) []T {
	s := DefensiveCopy(slice)
	if source == nil {
		/* #nosec */
		source = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	for i := range slice {
		j := source.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// Lazy is a function that returns a value of type T. The value is only
// calculated once, and the result is cached for subsequent calls.
type Lazy[T any] func() T

// DefineLazy returns a Lazy[T] that will call the given function to
// calculate the value. The value is only calculated once, and the result
// is cached for subsequent calls.
func DefineLazy[T any](f func() T) Lazy[T] {
	var value T
	var once sync.Once
	return func() T {
		once.Do(func() {
			value = f()
			f = nil
		})
		return value
	}
}

// Keys returns a slice of all values in the given map.
func Keys[M ~map[K]V, K Comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	RangeMap(m, func(k K, _ V) bool {
		r = append(r, k)
		return false
	})
	return r
}

// Values returns a slice of all values in the given map.
func Values[M ~map[K]V, K Comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	RangeMap(m, func(_ K, v V) bool {
		r = append(r, v)
		return false
	})
	return r
}

// Comparable is a type constraint for three types: int, int64, and string. By
// using this type constraint for a generic parameter, the parameter can be used
// as a map key and it can be sorted.
type Comparable interface {
	int | int64 | string
}

// RangeMap ranges over a map in deterministic order by first sorting the
// keys. It provides a function which will be called for each key/value pair.
// The function should return true to stop iteration.
func RangeMap[M ~map[K]V, K Comparable, V any](
	m M,
	f func(key K, value V) bool,
) {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	for _, k := range keys {
		if f(k, m[k]) {
			break
		}
	}
}

// Reverse reverses the given slice in place and returns it.
func Reverse[T any](slice []T) []T {
	n := len(slice)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}
