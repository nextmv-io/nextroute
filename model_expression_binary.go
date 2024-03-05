// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// BinaryFunction is a function that takes two float64 values and returns a
// float64 value.
type BinaryFunction func(float64, float64) float64

// BinaryExpression is an expression that takes two expressions as input and
// returns a value.
type BinaryExpression interface {
	ModelExpression
	// Left returns the left expression.
	Left() ModelExpression
	// Right returns the right expression.
	Right() ModelExpression
}

// NewOperatorExpression returns a new BinaryExpression that uses the given
// operator function.
func NewOperatorExpression(
	name string,
	left ModelExpression,
	right ModelExpression,
	operator BinaryFunction,
) BinaryExpression {
	return &binaryExpression{
		index:    NewModelExpressionIndex(),
		left:     left,
		right:    right,
		operator: operator,
		name:     name,
	}
}

type binaryExpression struct {
	left     ModelExpression
	right    ModelExpression
	operator BinaryFunction
	name     string
	index    int
}

func (b *binaryExpression) HasNegativeValues() bool {
	return b.left.HasNegativeValues() || b.right.HasNegativeValues()
}

func (b *binaryExpression) HasPositiveValues() bool {
	return b.left.HasPositiveValues() || b.right.HasPositiveValues()
}

func (b *binaryExpression) Index() int {
	return b.index
}

func (b *binaryExpression) String() string {
	return fmt.Sprintf("Binary[%v] '%v' left: %v, right: %v",
		b.index,
		b.name,
		b.left,
		b.right,
	)
}

func (b *binaryExpression) Name() string {
	return fmt.Sprintf("%s(%s,%s)",
		b.name,
		b.left.Name(),
		b.right.Name(),
	)
}

func (b *binaryExpression) SetName(n string) {
	b.name = n
}

func (b *binaryExpression) Left() ModelExpression {
	return b.left
}

func (b *binaryExpression) Right() ModelExpression {
	return b.right
}

func (b *binaryExpression) Value(
	vehicle ModelVehicleType,
	from, to ModelStop,
) float64 {
	return b.operator(
		b.left.Value(vehicle, from, to),
		b.right.Value(vehicle, from, to),
	)
}
