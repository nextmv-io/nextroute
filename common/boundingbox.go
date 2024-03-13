// © 2019-present nextmv.io inc

package common

// BoundingBox contains information about a box.
type BoundingBox interface {
	// Maximum returns the maximum location of the bounding box. The right
	// lower corner of the bounding box.
	Maximum() Location
	// Minimum returns the minimum location of the bounding box. The left
	// upper corner of the bounding box.
	Minimum() Location
	// IsValid returns true if the bounding box is valid. A bounding box is
	// valid if the maximum location is greater than the minimum location.
	IsValid() bool
	// Width returns the width of the box.
	Width() Distance
	// Height returns the height of the box.
	Height() Distance
}

// NewInvalidBoundingBox returns an invalid bounding box.
func NewInvalidBoundingBox() BoundingBox {
	return boundingBox{
		maximum: NewInvalidLocation(),
		minimum: NewInvalidLocation(),
	}
}

// NewBoundingBox returns a bounding box for the given locations.
func NewBoundingBox(locations Locations) BoundingBox {
	if len(locations) == 0 || !locations[0].IsValid() {
		return NewInvalidBoundingBox()
	}

	minLatitude := locations[0].Latitude()
	maxLatitude := locations[0].Latitude()
	minLongitude := locations[0].Longitude()
	maxLongitude := locations[0].Longitude()

	for idx := 1; idx < len(locations); idx++ {
		if !locations[idx].IsValid() {
			return NewInvalidBoundingBox()
		}
		latitude := locations[idx].Latitude()
		longitude := locations[idx].Longitude()
		if minLatitude > latitude {
			minLatitude = latitude
		}
		if maxLatitude < latitude {
			maxLatitude = latitude
		}
		if minLongitude > longitude {
			minLongitude = longitude
		}
		if maxLongitude < longitude {
			maxLongitude = longitude
		}
	}
	maxLocation, _ := NewLocation(maxLongitude, maxLatitude)
	minLocation, _ := NewLocation(minLongitude, minLatitude)
	return boundingBox{
		maximum: maxLocation,
		minimum: minLocation,
	}
}

type boundingBox struct {
	maximum Location
	minimum Location
}

func (b boundingBox) Width() Distance {
	if !b.IsValid() {
		return NewDistance(0.0, Meters)
	}
	leftUpper, _ := NewLocation(b.minimum.Longitude(), b.minimum.Latitude())
	rightUpper, _ := NewLocation(b.maximum.Longitude(), b.minimum.Latitude())
	width, _ := Haversine(leftUpper, rightUpper)
	return width
}

func (b boundingBox) Height() Distance {
	if !b.IsValid() {
		return NewDistance(0.0, Meters)
	}
	leftUpper, _ := NewLocation(b.minimum.Longitude(), b.minimum.Latitude())
	leftLower, _ := NewLocation(b.minimum.Longitude(), b.maximum.Latitude())
	height, _ := Haversine(leftUpper, leftLower)
	return height
}

func (b boundingBox) IsValid() bool {
	return b.maximum.IsValid() &&
		b.minimum.IsValid() &&
		b.maximum.Longitude() >= b.minimum.Longitude() &&
		b.maximum.Latitude() >= b.minimum.Latitude()
}

func (b boundingBox) Maximum() Location {
	return b.maximum
}

func (b boundingBox) Minimum() Location {
	return b.minimum
}
