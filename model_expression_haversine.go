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
	if !from.Location().IsValid() || !to.Location().IsValid() {
		return common.NewDistance(0.0, common.Meters)
	}
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
	if !from.Location().IsValid() || !to.Location().IsValid() {
		return 0.0
	}
	return haversineDistance(from.Location(), to.Location()).
		Value(vehicle.Model().DistanceUnit())
}
