package nextroute

import (
	"fmt"
)

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
