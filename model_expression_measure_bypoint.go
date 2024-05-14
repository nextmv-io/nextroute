// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/measure"
)

// NewMeasureByPointExpression returns a new MeasureByPointExpression.
// A MeasureByPointExpression is a ModelExpression that uses a measure.ByPoint to
// calculate the cost between two stops.
func NewMeasureByPointExpression(measure measure.ByPoint) ModelExpression {
	return &measureByPointExpression{
		index:   NewModelExpressionIndex(),
		measure: measure,
		name:    "measure_by_point",
	}
}

type measureByPointExpression struct {
	measure measure.ByPoint
	name    string
	index   int
}

func (m *measureByPointExpression) HasNegativeValues() bool {
	return false
}

func (m *measureByPointExpression) HasPositiveValues() bool {
	return true
}

func (m *measureByPointExpression) String() string {
	return fmt.Sprintf("measure_by_point[%v]",
		m.index,
	)
}

func (m *measureByPointExpression) Index() int {
	return m.index
}

func (m *measureByPointExpression) Name() string {
	return m.name
}

func (m *measureByPointExpression) SetName(n string) {
	m.name = n
}

func (m *measureByPointExpression) Value(_ ModelVehicleType, from, to ModelStop) float64 {
	locFrom, locTo := from.Location(), to.Location()
	value := m.measure.Cost(
		measure.Point{locFrom.Longitude(), locFrom.Latitude()},
		measure.Point{locTo.Longitude(), locTo.Latitude()},
	)
	return value
}
