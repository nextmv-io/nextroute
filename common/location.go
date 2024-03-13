package common

import (
	"fmt"
)

// NewLocation creates a new Location. An error is returned if the longitude is
// not between (-180, 180) or the latitude is not between (-90, 90).
func NewLocation(longitude float64, latitude float64) (Location, error) {
	if longitude < -180 || longitude > 180 {
		return NewInvalidLocation(),
			fmt.Errorf("longitude %f must be between -180 and 180", longitude)
	}
	if latitude < -90 || latitude > 90 {
		return NewInvalidLocation(),
			fmt.Errorf("latitude %f must be between -90 and 90", latitude)
	}
	return location{
		longitude: longitude,
		latitude:  latitude,
		valid:     true,
	}, nil
}

// NewInvalidLocation creates a new invalid Location. Longitude and latitude
// are not important.
func NewInvalidLocation() Location {
	return location{
		valid: false,
	}
}

// Locations is a slice of Location.
type Locations []Location

// Unique returns a new slice of Locations with unique locations.
func (l Locations) Unique() Locations {
	unique := make(map[string]Location)
	for _, location := range l {
		// TODO: in Go 1.20 we don't need to use fmt.Sprintf here.
		// This can simply become unique[location] = struct{}{}
		unique[fmt.Sprintf("%v", location)] = location
	}
	result := make(Locations, 0, len(unique))
	for _, location := range unique {
		result = append(result, location)
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
	for l, location := range l {
		if !location.IsValid() {
			return NewInvalidLocation(),
				fmt.Errorf(
					"location %d (%f, %f) is invalid",
					l,
					location.Longitude(),
					location.Latitude(),
				)
		}
		lat += location.Latitude()
		lon += location.Longitude()
	}
	return NewLocation(lon/float64(len(l)), lat/float64(len(l)))
}

// Location represents a physical location on the earth.
type Location interface {
	// Longitude returns the longitude of the location.
	Longitude() float64
	// Latitude returns the latitude of the location.
	Latitude() float64
	// Equals returns true if the location is equal to the location given as an
	// argument.
	Equals(Location) bool
	// IsValid returns true if the location is valid. A location is valid if
	// the bounds of the longitude and latitude are correct.
	IsValid() bool
}

// Implements Location.
type location struct {
	longitude float64
	latitude  float64
	valid     bool
}

func (l location) String() string {
	return fmt.Sprintf(
		"{lat: %v,lon: %v}",
		l.latitude,
		l.longitude,
	)
}

func (l location) Longitude() float64 {
	return l.longitude
}

func (l location) Latitude() float64 {
	return l.latitude
}

func (l location) Equals(other Location) bool {
	return l.longitude == other.Longitude() && l.latitude == other.Latitude()
}

func (l location) IsValid() bool {
	return l.valid
}
