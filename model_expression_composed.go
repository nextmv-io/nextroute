package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/nextroute"
)

// NewComposedPerVehicleTypeExpression returns a new ComposedPerVehicleTypeExpression.
func NewComposedPerVehicleTypeExpression(
	defaultExpression nextroute.ModelExpression,
) nextroute.ComposedPerVehicleTypeExpression {
	i := NewModelExpressionIndex()
	return &composedPerVehicleTypeExpressionImpl{
		index:             i,
		defaultExpression: defaultExpression,
		name: fmt.Sprintf("composed_per_vehicle_type[%v] ",
			i,
		),
	}
}

type composedPerVehicleTypeExpressionImpl struct {
	defaultExpression nextroute.ModelExpression
	expressions       []nextroute.ModelExpression
	name              string
	index             int
}

func (t *composedPerVehicleTypeExpressionImpl) Get(vehicleType nextroute.ModelVehicleType) nextroute.ModelExpression {
	idx := vehicleType.Index()
	if idx >= 0 && idx < len(t.expressions) {
		if expression := t.expressions[idx]; expression != nil {
			return expression
		}
	}
	return t.defaultExpression
}

func (t *composedPerVehicleTypeExpressionImpl) Set(
	vehicleType nextroute.ModelVehicleType,
	expression nextroute.ModelExpression,
) {
	idx := vehicleType.Index()
	// we have to grow the slice in case the index is out of bounds
	if idx >= len(t.expressions) {
		newExpressions := make([]nextroute.ModelExpression, idx+1)
		copy(newExpressions, t.expressions)
		t.expressions = newExpressions
	}
	t.expressions[idx] = expression
}

func (t *composedPerVehicleTypeExpressionImpl) HasNegativeValues() bool {
	if t.defaultExpression.HasNegativeValues() {
		return true
	}
	for _, expression := range t.expressions {
		if expression.HasNegativeValues() {
			return true
		}
	}
	return false
}

func (t *composedPerVehicleTypeExpressionImpl) HasPositiveValues() bool {
	if t.defaultExpression.HasPositiveValues() {
		return true
	}
	for _, expression := range t.expressions {
		if expression.HasPositiveValues() {
			return true
		}
	}
	return false
}

func (t *composedPerVehicleTypeExpressionImpl) String() string {
	return t.Name()
}

func (t *composedPerVehicleTypeExpressionImpl) Index() int {
	return t.index
}

func (t *composedPerVehicleTypeExpressionImpl) Name() string {
	return t.name
}

func (t *composedPerVehicleTypeExpressionImpl) SetName(n string) {
	t.name = n
}

func (t *composedPerVehicleTypeExpressionImpl) DefaultExpression() nextroute.ModelExpression {
	return t.defaultExpression
}

func (t *composedPerVehicleTypeExpressionImpl) Value(
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) float64 {
	idx := vehicleType.Index()
	if idx >= 0 && idx < len(t.expressions) {
		if expression := t.expressions[idx]; expression != nil {
			return expression.Value(vehicleType, from, to)
		}
	}
	return t.defaultExpression.Value(vehicleType, from, to)
}
