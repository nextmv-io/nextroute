package nextroute

import (
	"fmt"
	"strings"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

// NewSumExpression returns a new SumExpression.
func NewSumExpression(
	expressions nextroute.ModelExpressions,
) nextroute.SumExpression {
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
	expressions nextroute.ModelExpressions
	index       int
}

func (n *sumExpressionImpl) HasNegativeValues() bool {
	return common.HasTrue(n.expressions, func(expression nextroute.ModelExpression) bool {
		return expression.HasNegativeValues()
	})
}

func (n *sumExpressionImpl) HasPositiveValues() bool {
	return common.HasTrue(n.expressions, func(expression nextroute.ModelExpression) bool {
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
	vehicle nextroute.ModelVehicleType,
	from nextroute.ModelStop,
	to nextroute.ModelStop,
) float64 {
	value := 0.0
	for _, expression := range n.expressions {
		value += expression.Value(vehicle, from, to)
	}
	return value
}

func (n *sumExpressionImpl) AddExpression(expression nextroute.ModelExpression) {
	n.expressions = append(n.expressions, expression)
}

func (n *sumExpressionImpl) Expressions() nextroute.ModelExpressions {
	expressions := make(nextroute.ModelExpressions, len(n.expressions))
	copy(expressions, n.expressions)
	return expressions
}
