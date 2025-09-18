// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"

	"github.com/nextmv-io/nextroute/common"
)

// NewHaversineExpression returns a new HaversineExpression.
func NewHaversineExpression() DistanceExpression {
	return &haversineExpression{
		index: NewModelExpressionIndex(),
		name:  "haversine",
	}
}

type haversineExpression struct {
	name  string
	index int
}

func (h *haversineExpression) HasNegativeValues() bool {
	return false
}

func (h *haversineExpression) HasPositiveValues() bool {
	return true
}

func (h *haversineExpression) String() string {
	return fmt.Sprintf("haversine[%v]",
		h.index,
	)
}

func (h *haversineExpression) Distance(
	vehicleType ModelVehicleType,
	from, to ModelStop,
) common.Distance {
	return common.NewDistance(h.Value(vehicleType, from, to), common.Meters)
}

func (h *haversineExpression) Index() int {
	return h.index
}

func (h *haversineExpression) Name() string {
	return h.name
}

func (h *haversineExpression) SetName(n string) {
	h.name = n
}

func (h *haversineExpression) Value(
	vehicle ModelVehicleType,
	from ModelStop,
	to ModelStop,
) float64 {
	fromLocation := from.Location()
	toLocation := to.Location()
	if !fromLocation.IsValid() || !toLocation.IsValid() {
		// if any location is invalid by convention we return a 0
		// distance
		return 0
	}
	return common.HaversineUnsafe(
		fromLocation,
		toLocation,
	).Value(vehicle.Model().DistanceUnit())
}
