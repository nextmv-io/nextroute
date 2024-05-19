// © 2019-present nextmv.io inc

package nextroute

// ExpressionObjective is an objective that uses an expression to calculate an
// objective.
type ExpressionObjective interface {
	ModelObjective

	// Expression returns the expression that is used to calculate the
	// objective.
	Expression() ModelExpression
}

// NewExpressionObjective is the implementation of sdk.NewExpressionObjective.
func NewExpressionObjective(e ModelExpression) ExpressionObjective {
	return &expressionObjectiveImpl{
		expression: e,
		index:      NewModelExpressionIndex(),
	}
}

// expressionObjectiveImpl implements the ExpressionObjective
// interface.
type expressionObjectiveImpl struct {
	expression ModelExpression
	index      int
}

func (e *expressionObjectiveImpl) Expression() ModelExpression {
	return e.expression
}

func (e *expressionObjectiveImpl) Index() int {
	return e.index
}

func (e *expressionObjectiveImpl) Value(solution Solution) float64 {
	score := 0.0
	for _, r := range solution.Vehicles() {
		score += r.Last().CumulativeValue(e.expression)
	}
	return score
}

func (e *expressionObjectiveImpl) EstimateDeltaValue(
	move SolutionMoveStops,
) float64 {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()
	value := 0.0

	first := true
	var previousSolutionStop SolutionStop

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
	previousMove, _ := moveImpl.previous()
	currentValue := nextmove.CumulativeValue(e.expression) -
		previousMove.CumulativeValue(e.expression)

	return value - currentValue
}

func (e *expressionObjectiveImpl) ModelExpressions() ModelExpressions {
	return ModelExpressions{e.expression}
}

func (e *expressionObjectiveImpl) String() string {
	return "expression_objective"
}
