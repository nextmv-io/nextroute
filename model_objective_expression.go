package nextroute

import (
	"github.com/nextmv-io/sdk/nextroute"
)

// NewExpressionObjective is the implementation of sdk.NewExpressionObjective.
func NewExpressionObjective(e nextroute.ModelExpression) nextroute.ExpressionObjective {
	return &expressionObjectiveImpl{
		expression: e,
		index:      NewModelExpressionIndex(),
	}
}

// expressionObjectiveImpl implements the nextroute.ExpressionObjective
// interface.
type expressionObjectiveImpl struct {
	expression nextroute.ModelExpression
	index      int
}

func (e *expressionObjectiveImpl) Expression() nextroute.ModelExpression {
	return e.expression
}

func (e *expressionObjectiveImpl) Index() int {
	return e.index
}

func (e *expressionObjectiveImpl) InternalValue(solution *solutionImpl) float64 {
	score := 0.0
	for _, r := range solution.vehicles {
		score += r.last().CumulativeValue(e.expression)
	}
	return score
}

func (e *expressionObjectiveImpl) Value(solution nextroute.Solution) float64 {
	return e.InternalValue(solution.(*solutionImpl))
}

func (e *expressionObjectiveImpl) EstimateDeltaValue(
	move nextroute.SolutionMoveStops,
) float64 {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()
	value := 0.0

	first := true
	var previousSolutionStop solutionStopImpl

	generator := newSolutionStopGenerator(*moveImpl, false, false)
	defer generator.release()

	for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
		if first {
			previousSolutionStop = solutionStop
			first = false
			continue
		}

		value += e.expression.Value(
			vehicleType,
			previousSolutionStop.ModelStop(),
			solutionStop.ModelStop(),
		)
		previousSolutionStop = solutionStop
	}

	nextmove, _ := moveImpl.next()
	previousmove, _ := moveImpl.previous()
	currentValue := nextmove.CumulativeValue(e.expression) -
		previousmove.CumulativeValue(e.expression)

	return value - currentValue
}

func (e *expressionObjectiveImpl) ModelExpressions() nextroute.ModelExpressions {
	return nextroute.ModelExpressions{e.expression}
}

func (e *expressionObjectiveImpl) String() string {
	return "expression_objective"
}
