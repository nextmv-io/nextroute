// © 2019-present nextmv.io inc

package nextroute

// NewExpressionObjective is the implementation of sdk.NewExpressionObjective.
func NewExpressionObjective(e ModelExpression) *ExpressionObjective {
	return &ExpressionObjective{
		expression: e,
		index:      NewModelExpressionIndex(),
	}
}

// ExpressionObjective is an objective that uses an expression to calculate an
// objective.
type ExpressionObjective struct {
	expression ModelExpression
	index      int
}

// Expression returns the expression that is used to calculate the
// objective.
func (e *ExpressionObjective) Expression() ModelExpression {
	return e.expression
}

// Index returns the index of the objective.
func (e *ExpressionObjective) Index() int {
	return e.index
}

// Value implements the ModelObjective interface.
func (e *ExpressionObjective) Value(solution *Solution) float64 {
	score := 0.0
	for _, r := range solution.Vehicles() {
		score += r.Last().CumulativeValue(e.expression)
	}
	return score
}

// EstimateDeltaValue implements the ModelObjective interface.
func (e *ExpressionObjective) EstimateDeltaValue(
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

// ModelExpressions implements the RegisteredModelExpressions interface.
func (e *ExpressionObjective) ModelExpressions() ModelExpressions {
	return ModelExpressions{e.expression}
}

// String returns the string representation of the objective.
func (e *ExpressionObjective) String() string {
	return "expression_objective"
}
