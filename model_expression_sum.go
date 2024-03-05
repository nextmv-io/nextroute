// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"slices"
	"strings"
)

// SumExpression is an expression that returns the sum of the values of the
// given expressions.
type SumExpression interface {
	ModelExpression

	// AddExpression adds an expression to the sum.
	AddExpression(expression ModelExpression)

	// Expressions returns the expressions that are part of the sum.
	Expressions() ModelExpressions
}

// NewSumExpression returns a new SumExpression.
func NewSumExpression(
	expressions ModelExpressions,
) SumExpression {
	name := "sum("
	for i, expression := range expressions {
		if i > 0 {
			name += ","
		}
		name += expression.Name()
	}
	name += ")"

	return &sumExpressionImpl{
		index:       NewModelExpressionIndex(),
		expressions: expressions,
		name:        name,
	}
}

type sumExpressionImpl struct {
	name        string
	expressions ModelExpressions
	index       int
}

func (n *sumExpressionImpl) HasNegativeValues() bool {
	return slices.ContainsFunc(n.expressions, func(expression ModelExpression) bool {
		return expression.HasNegativeValues()
	})
}

func (n *sumExpressionImpl) HasPositiveValues() bool {
	return slices.ContainsFunc(n.expressions, func(expression ModelExpression) bool {
		return expression.HasPositiveValues()
	})
}

func (n *sumExpressionImpl) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "sum[%v], ",
		n.index,
	)
	for _, expression := range n.expressions {
		fmt.Fprintf(&sb, " %v, ", expression)
	}
	return sb.String()
}

func (n *sumExpressionImpl) Index() int {
	return n.index
}

func (n *sumExpressionImpl) Name() string {
	return n.name
}

func (n *sumExpressionImpl) SetName(name string) {
	n.name = name
}

func (n *sumExpressionImpl) Value(
	vehicle ModelVehicleType,
	from ModelStop,
	to ModelStop,
) float64 {
	value := 0.0
	for _, expression := range n.expressions {
		value += expression.Value(vehicle, from, to)
	}
	return value
}

func (n *sumExpressionImpl) AddExpression(expression ModelExpression) {
	n.expressions = append(n.expressions, expression)
}

func (n *sumExpressionImpl) Expressions() ModelExpressions {
	expressions := make(ModelExpressions, len(n.expressions))
	copy(expressions, n.expressions)
	return expressions
}
