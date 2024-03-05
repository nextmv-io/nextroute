// © 2019-present nextmv.io inc

package nextroute

import "sync/atomic"

var expressionIndex uint32

// NewModelExpressionIndex returns the next unique expression index.
func NewModelExpressionIndex() int {
	return int(atomic.AddUint32(&expressionIndex, 1) - 1)
}

// ModelExpression is an expression that can be used in a model to define
// values for constraints and objectives. The expression is evaluated for
// each stop in the solution by invoking the Value() method. The value of
// the expression is then used in the constraints and objective.
type ModelExpression interface {
	// Index returns the unique index of the expression.
	Index() int

	// Name returns the name of the expression.
	Name() string

	// Value returns the value of the expression for the given vehicle type,
	// from stop and to stop.
	Value(ModelVehicleType, ModelStop, ModelStop) float64

	// HasNegativeValues returns true if the expression contains negative
	// values.
	HasNegativeValues() bool
	// HasPositiveValues returns true if the expression contains positive
	// values.
	HasPositiveValues() bool

	// SetName sets the name of the expression.
	SetName(string)
}

// ModelExpressions is a slice of ModelExpression.
type ModelExpressions []ModelExpression
