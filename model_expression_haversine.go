package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

// NewHaversineExpression returns a new HaversineExpression.
func NewHaversineExpression() nextroute.DistanceExpression {
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
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
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
	vehicle nextroute.ModelVehicleType,
	from nextroute.ModelStop,
	to nextroute.ModelStop,
) float64 {
	return haversineDistance(from.Location(), to.Location()).
		Value(vehicle.Model().DistanceUnit())
}
