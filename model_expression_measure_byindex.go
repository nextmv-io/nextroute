package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/measure"
	"github.com/nextmv-io/sdk/nextroute"
)

// NewMeasureByIndexExpression returns a new MeasureByIndexExpression.
// A MeasureByIndexExpression is a ModelExpression that uses a measure.ByIndex to
// calculate the cost between two stops.
func NewMeasureByIndexExpression(measure measure.ByIndex) nextroute.ModelExpression {
	return &measureByIndexExpression{
		index:   NewModelExpressionIndex(),
		measure: measure,
		name:    "measure_by_index",
	}
}

type measureByIndexExpression struct {
	measure measure.ByIndex
	name    string
	index   int
}

func (m *measureByIndexExpression) HasNegativeValues() bool {
	return false
}

func (m *measureByIndexExpression) HasPositiveValues() bool {
	return true
}

func (m *measureByIndexExpression) String() string {
	return fmt.Sprintf("measure_by_index[%v]",
		m.index,
	)
}

func (m *measureByIndexExpression) Index() int {
	return m.index
}

func (m *measureByIndexExpression) Name() string {
	return m.name
}

func (m *measureByIndexExpression) SetName(n string) {
	m.name = n
}

func (m *measureByIndexExpression) Value(_ nextroute.ModelVehicleType, from, to nextroute.ModelStop) float64 {
	return m.measure.Cost(from.(*stopImpl).measureIndex, to.(*stopImpl).measureIndex)
}
