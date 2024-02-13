package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/nextroute"
)

// NewTermExpression returns a new TermExpression.
func NewTermExpression(
	factor float64,
	expression nextroute.ModelExpression,
) nextroute.TermExpression {
	return &termExpression{
		index:      NewModelExpressionIndex(),
		expression: expression,
		factor:     factor,
		name:       fmt.Sprintf("%f * %s", factor, expression),
	}
}

type termExpression struct {
	expression nextroute.ModelExpression
	name       string
	index      int
	factor     float64
}

func (t *termExpression) HasNegativeValues() bool {
	if t.factor < 0 {
		return t.expression.HasPositiveValues()
	}
	return t.expression.HasNegativeValues()
}

func (t *termExpression) HasPositiveValues() bool {
	if t.factor < 0 {
		return t.expression.HasNegativeValues()
	}
	return t.expression.HasPositiveValues()
}

func (t *termExpression) String() string {
	return fmt.Sprintf("Term[%v] %v * %v",
		t.index,
		t.factor,
		t.expression,
	)
}

func (t *termExpression) Index() int {
	return t.index
}

func (t *termExpression) Name() string {
	return t.name
}

func (t *termExpression) SetName(n string) {
	t.name = n
}

func (t *termExpression) Factor() float64 {
	return t.factor
}

func (t *termExpression) Expression() nextroute.ModelExpression {
	return t.expression
}

func (t *termExpression) Value(
	vehicle nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) float64 {
	return t.factor * t.expression.Value(vehicle, from, to)
}
