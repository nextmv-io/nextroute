// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// ComposedPerVehicleTypeExpression is an expression that uses an expression for
// each vehicle type.
type ComposedPerVehicleTypeExpression interface {
	ModelExpression

	// DefaultExpression returns the default expression that is used if no
	// expression is defined for a specific vehicle type.
	DefaultExpression() ModelExpression

	// Get returns the expression that is defined for the given vehicle type. If
	// no expression is defined for the given vehicle type, the default
	// expression is returned.
	Get(vehicleType ModelVehicleType) ModelExpression
	// Set sets the expression for the given vehicle type.
	Set(vehicleType ModelVehicleType, expression ModelExpression)
}

// NewComposedPerVehicleTypeExpression returns a new ComposedPerVehicleTypeExpression.
func NewComposedPerVehicleTypeExpression(
	defaultExpression ModelExpression,
) ComposedPerVehicleTypeExpression {
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
	defaultExpression ModelExpression
	expressions       []ModelExpression
	name              string
	index             int
}

func (t *composedPerVehicleTypeExpressionImpl) Get(vehicleType ModelVehicleType) ModelExpression {
	idx := vehicleType.Index()
	if idx >= 0 && idx < len(t.expressions) {
		if expression := t.expressions[idx]; expression != nil {
			return expression
		}
	}
	return t.defaultExpression
}

func (t *composedPerVehicleTypeExpressionImpl) Set(
	vehicleType ModelVehicleType,
	expression ModelExpression,
) {
	idx := vehicleType.Index()
	// we have to grow the slice in case the index is out of bounds
	if idx >= len(t.expressions) {
		newExpressions := make([]ModelExpression, idx+1)
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

func (t *composedPerVehicleTypeExpressionImpl) DefaultExpression() ModelExpression {
	return t.defaultExpression
}

func (t *composedPerVehicleTypeExpressionImpl) Value(
	vehicleType ModelVehicleType,
	from, to ModelStop,
) float64 {
	idx := vehicleType.Index()
	if idx >= 0 && idx < len(t.expressions) {
		if expression := t.expressions[idx]; expression != nil {
			return expression.Value(vehicleType, from, to)
		}
	}
	return t.defaultExpression.Value(vehicleType, from, to)
}
