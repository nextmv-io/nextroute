// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"math"
)

// NewLocation creates a new Location. An error is returned if the longitude is
// not between (-180, 180) or the latitude is not between (-90, 90).
func NewLocation(longitude float64, latitude float64) (Location, error) {
	if !isValidLongitude(longitude) {
		return NewInvalidLocation(),
			fmt.Errorf("longitude %f must be between -180 and 180", longitude)
	}
	if !isValidLatitude(latitude) {
		return NewInvalidLocation(),
			fmt.Errorf("latitude %f must be between -90 and 90", latitude)
	}
	return Location{
		longitude: longitude,
		latitude:  latitude,
		valid:     true,
	}, nil
}

// NewInvalidLocation creates a new invalid Location. Longitude and latitude
// are not important.
func NewInvalidLocation() Location {
	return Location{
		longitude: math.NaN(),
		latitude:  math.NaN(),
		valid:     false,
	}
}

// Locations is a slice of Location.
type Locations []Location

// Unique returns a new slice of Locations with unique locations.
func (l Locations) Unique() Locations {
	unique := make(map[Location]struct{}, len(l))
	for _, location := range l {
		unique[location] = struct{}{}
	}
	result := make(Locations, len(unique))
	i := 0
	for location := range unique {
		result[i] = location
		i++
	}
	return result
}

// Centroid returns the centroid of the locations. If locations is empty, the
// centroid will be an invalid location.
func (l Locations) Centroid() (Location, error) {
	if len(l) == 0 {
		return NewInvalidLocation(), nil
	}
	lat := 0.0
	lon := 0.0
	for _, location := range l {
		// invalid locations are encoded as NaN, which will propagate
		// so we can avoid a check here.
		lat += location.Latitude()
		lon += location.Longitude()
	}
	n := float64(len(l))
	loc, err := NewLocation(lon/n, lat/n)
	if err != nil {
		return NewInvalidLocation(), err
	}
	return loc, nil
}

// Location represents a location on earth.
type Location struct {
	longitude float64
	latitude  float64
	valid     bool
}

// String returns a string representation of the location.
func (l Location) String() string {
	return fmt.Sprintf(
		"{lat: %v,lon: %v}",
		l.latitude,
		l.longitude,
	)
}

// Longitude returns the longitude of the location.
func (l Location) Longitude() float64 {
	return l.longitude
}

// Latitude returns the latitude of the location.
func (l Location) Latitude() float64 {
	return l.latitude
}

// Equals returns true if the invoking location is equal to the other location.
func (l Location) Equals(other Location) bool {
	return l.longitude == other.Longitude() && l.latitude == other.Latitude()
}

// IsValid returns true if the location is valid. A location is valid if
// the bounds of the longitude and latitude are correct.
func (l Location) IsValid() bool {
	return l.valid
}

func isValidLongitude(longitude float64) bool {
	return longitude >= -180 && longitude <= 180
}

func isValidLatitude(latitude float64) bool {
	return latitude >= -90 && latitude <= 90
}
