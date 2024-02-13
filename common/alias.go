// Package common contains common types and functions used by the nextroute.
package common

import (
	"math/rand"
	"time"

	c "github.com/nextmv-io/sdk/common"
)

// Alias is an interface that allows for sampling from a discrete
// distribution in O(1) time.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type Alias = c.Alias

// NewAlias creates a new Alias from the given weights.
// The weights must be positive and at least one weight must be given.
// The weights are normalized to sum to 1.
// NewAlias([]float64{1, 2, 3}) will return an Alias that will
// return 0 with probability 1/6, 1 with probability 1/3 and 2 with
// probability 1/2.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewAlias(weights []float64) (Alias, error) {
	return c.NewAlias(weights)
}

// NewLocation creates a new Location. An error is returned if the longitude is
// not between (-180, 180) or the latitude is not between (-90, 90).
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewLocation(longitude float64, latitude float64) (Location, error) {
	return c.NewLocation(longitude, latitude)
}

// NewInvalidLocation creates a new invalid Location. Longitude and latitude
// are not important.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewInvalidLocation() Location {
	return c.NewInvalidLocation()
}

// Locations is a slice of Location.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type Locations = c.Locations

// Location represents a physical location on the earth.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type Location = c.Location

// BoundingBox contains information about a box.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type BoundingBox = c.BoundingBox

// NewInvalidBoundingBox returns an invalid bounding box.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewInvalidBoundingBox() BoundingBox {
	return c.NewInvalidBoundingBox()
}

// NewBoundingBox returns a bounding box for the given locations.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewBoundingBox(locations Locations) BoundingBox {
	return c.NewBoundingBox(locations)
}

// DistanceUnit is the unit of distance.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type DistanceUnit = c.DistanceUnit

// Distance is a distance in a given unit.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type Distance = c.Distance

// NewDistance returns a new distance.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewDistance(
	value float64,
	unit DistanceUnit,
) Distance {
	return c.NewDistance(value, unit)
}

// NewDuration returns a new duration by unit.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewDuration(unit DurationUnit) time.Duration {
	return c.NewDuration(unit)
}

// DurationUnit is the unit of duration.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type DurationUnit = c.DurationUnit

// FastHaversine is a fast approximation of the haversine distance.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type FastHaversine = c.FastHaversine

// NewFastHaversine returns a new FastHaversine.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewFastHaversine(lat float64) FastHaversine {
	return c.NewFastHaversine(lat)
}

// Haversine calculates the distance between two locations using the
// Haversine formula. Haversine is a good approximation for short
// distances (up to a few hundred kilometers).
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Haversine(from, to Location) (Distance, error) {
	return c.Haversine(from, to)
}

// Intersect returns the intersection of two slices.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Intersect[T comparable](a []T, b []T) []T {
	return c.Intersect[T](a, b)
}

// NSmallest returns the n-smallest items in the slice items using the
// function f to determine the value of each item. If n is greater than
// the length of items, all items are returned.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NSmallest[T any](items []T, f func(T) float64, n int) []T {
	return c.NSmallest[T](items, f, n)
}

// Meters is the distance travelled by light in a vacuum in
// 1/299,792,458 seconds.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
const Meters = c.Meters

// Kilometers is 1000 meters.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
const Kilometers = c.Kilometers

// Miles is 1609.34 meters.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
const Miles = c.Miles

// KilometersPerHour is a speed unit of kilometers per hour.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
var KilometersPerHour = c.KilometersPerHour

// MetersPerSecond is a speed unit of meters per second.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
var MetersPerSecond = c.MetersPerSecond

// MilesPerHour is a speed unit of miles per hour.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
var MilesPerHour = c.MilesPerHour

const (
	// NanoSecond is 1/1,000,000,000 of a second.
	//
	// Deprecated: This package is deprecated and will be removed in the next major release.
	// Use [github.com/nextmv-io/sdk/common] instead.
	NanoSecond = c.NanoSecond
	// MicroSecond is 1/1,000,000 of a second.
	//
	// Deprecated: This package is deprecated and will be removed in the next major release.
	// Use [github.com/nextmv-io/sdk/common] instead.
	MicroSecond = c.MicroSecond
	// MilliSecond is 1/1,000 of a second.//
	// Deprecated: This package is deprecated and will be removed in the next major release.
	// Use [github.com/nextmv-io/sdk/common] instead.
	MilliSecond = c.MilliSecond
	// Second is the SI unit of time.
	//
	// Deprecated: This package is deprecated and will be removed in the next major release.
	// Use [github.com/nextmv-io/sdk/common] instead.
	Second = c.Second
	// Minute is 60 seconds.
	//
	// Deprecated: This package is deprecated and will be removed in the next major release.
	// Use [github.com/nextmv-io/sdk/common] instead.
	Minute = c.Minute
	// Hour is 60 minutes.
	//
	// Deprecated: This package is deprecated and will be removed in the next major release.
	// Use [github.com/nextmv-io/sdk/common] instead.
	Hour = c.Hour
	// Day is 24 hours.
	//
	// Deprecated: This package is deprecated and will be removed in the next major release.
	// Use [github.com/nextmv-io/sdk/common] instead.
	Day = c.Day
)

// Speed is the interface for a speed.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type Speed = c.Speed

// SpeedUnit represents a unit of speed.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type SpeedUnit = c.SpeedUnit

// NewSpeed creates a new speed.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewSpeed(
	distance float64,
	unit SpeedUnit,
) Speed {
	return c.NewSpeed(distance, unit)
}

// NewSpeedUnit returns a new speed unit.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewSpeedUnit(
	distanceUnit DistanceUnit,
	duration time.Duration,
) SpeedUnit {
	return c.NewSpeedUnit(distanceUnit, duration)
}

// Statistics describes statistics.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type Statistics = c.Statistics

// NewStatistics creates a new statistics object.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NewStatistics[T any](v []T, f func(T) float64) Statistics {
	return c.NewStatistics[T](v, f)
}

// Filter filters a slice using a predicate function.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Filter[T any](v []T, f func(T) bool) []T {
	return c.Filter[T](v, f)
}

// Unique is a universal duplicate removal function for type instances in
// a slice that implement the comparable interface.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Unique[T comparable](s []T) []T {
	return c.Unique[T](s)
}

// UniqueDefined is a universal duplicate removal function for type instances in
// a slice that implement the comparable interface. The function f is used to
// extract the comparable value from the type instance.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func UniqueDefined[T any, I comparable](items []T, f func(T) I) []T {
	return c.UniqueDefined[T, I](items, f)
}

// NotUnique returns the duplicate instances.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NotUnique[T comparable](s []T) []T {
	return c.NotUnique[T](s)
}

// NotUniqueDefined returns the instances for which f returns identical
// values.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func NotUniqueDefined[T any, I comparable](items []T, f func(T) I) []T {
	return c.NotUniqueDefined[T, I](items, f)
}

// GroupBy groups the elements of a slice by a key function.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func GroupBy[T any, K comparable](s []T, f func(T) K) map[K][]T {
	return c.GroupBy[T, K](s, f)
}

// Map maps a slice of type T to a slice of type R using the function f.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Map[T any, R any](v []T, f func(T) R) []R {
	return c.Map[T, R](v, f)
}

// MapSlice maps a slice of type T to a slice of type R using the function f
// returning a slice of R.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func MapSlice[T any, R any](v []T, f func(T) []R) []R {
	return c.MapSlice[T, R](v, f)
}

// FindIndex returns the first index i satisfying predicate(s[i]).
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func FindIndex[E any](s []E, predicate func(E) bool) int {
	return c.FindIndex[E](s, predicate)
}

// AllTrue returns true if all the given predicate evaluations are true.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func AllTrue[E any](s []E, predicate func(E) bool) bool {
	return c.AllTrue[E](s, predicate)
}

// AllFalse returns true if all the given predicate evaluations is false.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func AllFalse[E any](s []E, predicate func(E) bool) bool {
	return c.AllFalse[E](s, predicate)
}

// All returns true if all the given predicate evaluations evaluate to
// condition.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func All[E any](s []E, condition bool, predicate func(E) bool) bool {
	return c.All[E](s, condition, predicate)
}

// HasTrue returns true if any of the given predicate evaluations evaluate to
// true.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func HasTrue[E any](s []E, predicate func(E) bool) bool {
	return c.HasTrue[E](s, predicate)
}

// HasFalse returns true if any of the given predicate evaluations evaluate to
// false.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func HasFalse[E any](s []E, predicate func(E) bool) bool {
	return c.HasFalse[E](s, predicate)
}

// Has returns true if any of the given predicate evaluations is condition.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Has[E any](s []E, condition bool, predicate func(E) bool) bool {
	return c.Has[E](s, condition, predicate)
}

// CopyMap copies all key/value pairs in source adding them to destination.
// If a key exists in both maps, the value in destination is overwritten.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func CopyMap[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](
	destination M1,
	source M2) {
	c.CopyMap[M1, M2, K, V](destination, source)
}

// DefensiveCopy returns a defensive copy of a slice.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func DefensiveCopy[T any](v []T) []T {
	return c.DefensiveCopy[T](v)
}

// WithinTolerance returns true if a and b are within the given tolerance.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func WithinTolerance(a, b, tolerance float64) bool {
	return c.WithinTolerance(a, b, tolerance)
}

// Truncate truncates a float64 to the given unit.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Truncate(f float64, unit float64) float64 {
	return c.Truncate(f, unit)
}

// DurationValue returns the value of a duration in the given time unit.
// Will panic if the time unit is zero.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func DurationValue(
	distance Distance,
	speed Speed,
	timeUnit time.Duration,
) float64 {
	return c.DurationValue(distance, speed, timeUnit)
}

// RandomElement returns a random element from the given slice. If the slice is
// empty, panic is raised. If source is nil, a new source is created using the
// current time.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func RandomElement[T any](
	source *rand.Rand,
	elements []T,
) T {
	return c.RandomElement[T](source, elements)
}

// RandomElements returns a slice of n random elements from the
// given slice. If n is greater than the length of the slice, all elements are
// returned. If n is less than or equal to zero, an empty slice is returned.
// If source is nil, a new source is created using the current time.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func RandomElements[T any](
	source *rand.Rand,
	elements []T,
	n int,
) []T {
	return c.RandomElements[T](source, elements, n)
}

// RandomElementIndices returns a slice of n random element indices from the
// given slice. If n is greater than the length of the slice, all indices are
// returned. If n is less than or equal to zero, an empty slice is returned.
// If source is nil, a new source is created using the current time.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func RandomElementIndices[T any](
	source *rand.Rand,
	elements []T,
	n int,
) []int {
	return c.RandomElementIndices[T](source, elements, n)
}

// RandomIndex returns a random index from the given size. If the index has
// already been used, a new index is generated. If source is nil, a new source
// is created using the current time.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func RandomIndex(source *rand.Rand, size int, indicesUsed map[int]bool) int {
	return c.RandomIndex(source, size, indicesUsed)
}

// Shuffle returns a shuffled copy of the given slice. If source is nil, a new
// source is created using the current time.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Shuffle[T any](source *rand.Rand, slice []T) []T {
	return c.Shuffle[T](source, slice)
}

// DefineLazy returns a Lazy[T] that will call the given function to
// calculate the value. The value is only calculated once, and the result
// is cached for subsequent calls.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func DefineLazy[T any](f func() T) c.Lazy[T] {
	return c.DefineLazy(f)
}

// Comparable is a type constraint for three types: int, int64, and string. By
// using this type constraint for a generic parameter, the parameter can be used
// as a map key and it can be sorted.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
type Comparable = c.Comparable

// Keys returns a slice of all values in the given map.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Keys[M ~map[K]V, K Comparable, V any](m M) []K {
	return c.Keys(m)
}

// Values returns a slice of all values in the given map.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Values[M ~map[K]V, K Comparable, V any](m M) []V {
	return c.Values(m)
}

// RangeMap ranges over a map in deterministic order by first sorting the
// keys. It provides a function which will be called for each key/value pair.
// The function should return true to stop iteration.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func RangeMap[M ~map[K]V, K Comparable, V any](
	m M,
	f func(key K, value V) bool,
) {
	c.RangeMap(m, f)
}

// Reverse reverses the given slice in place and returns it.
//
// Deprecated: This package is deprecated and will be removed in the next major release.
// Use [github.com/nextmv-io/sdk/common] instead.
func Reverse[T any](slice []T) []T {
	return c.Reverse(slice)
}
