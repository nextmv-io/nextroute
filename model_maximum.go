// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"math"
)

// Maximum can be used as a constraint or an objective that limits the maximum
// cumulative value can be assigned to a vehicle type. The maximum cumulative
// value is defined by the expression and the maximum value is defined by the
// maximum expression.
type Maximum struct {
	maximum VehicleTypeExpression
	deltas  []float64
	// hasNegativeValues is true if the expression has negative values.
	// This is used to optimize the estimation cost.
	hasNegativeValues bool
	// hasPositiveValues is true if the expression has positive values.
	// This is used to optimize the estimation cost.
	hasPositiveValues                    bool
	hasConstantExpression                bool
	hasStopExpressionAndNoNegativeValues bool
	resourceExpression                   ModelExpression
	maximumByVehicleType                 []float64
	penaltyOffset                        float64
	hasNoEffect                          []bool
	name                                 string
}

// NewMaximum creates a new maximum construct which can be used as constraint
// or as objective.
func NewMaximum(
	expression ModelExpression,
	maximum VehicleTypeExpression,
) (*Maximum, error) {
	return &Maximum{
		name:               "maximum",
		resourceExpression: expression,
		maximum:            maximum,
		penaltyOffset:      0.0,
	}, nil
}

// PenaltyOffset returns the penalty offset. Penalty offset is used to
// offset the penalty. The penalty offset is added to the penalty if there
// is at least one violation.
func (l *Maximum) PenaltyOffset() float64 {
	return l.penaltyOffset
}

// SetPenaltyOffset sets the penalty offset. Penalty offset is used to
// offset the penalty. The penalty offset is added to the penalty if there
// is at least one violation. The default penalty offset is 0.0 and it can
// be changed by this method and must be positive.
func (l *Maximum) SetPenaltyOffset(penaltyOffset float64) error {
	if penaltyOffset < 0.0 {
		return fmt.Errorf(
			"maximum objective, penalty offset must be positive, it can not be %f",
			penaltyOffset,
		)
	}

	l.penaltyOffset = penaltyOffset

	return nil
}

// Lock implements the Locker interface.
func (l *Maximum) Lock(model *Model) error {
	l.hasNegativeValues = l.Expression().HasNegativeValues()
	l.hasPositiveValues = l.Expression().HasPositiveValues()
	if _, ok := l.Expression().(ConstantExpression); ok {
		l.hasConstantExpression = true
	}
	if _, ok := l.Expression().(StopExpression); ok &&
		!l.hasNegativeValues {
		l.hasStopExpressionAndNoNegativeValues = true
	}
	vehicleTypes := model.VehicleTypes()
	l.maximumByVehicleType = make([]float64, len(vehicleTypes))
	for _, vehicleType := range vehicleTypes {
		l.maximumByVehicleType[vehicleType.Index()] = l.maximum.Value(
			vehicleType,
			nil,
			nil,
		)
	}

	planUnits := model.PlanStopsUnits()

	l.hasNoEffect = make([]bool, len(planUnits))

	if !l.hasStopExpressionAndNoNegativeValues {
		return nil
	}

	l.deltas = make([]float64, len(planUnits))

	for _, planUnit := range planUnits {
		delta := 0.0
		hasNoEffect := true
		for _, stop := range planUnit.Stops() {
			value := l.Expression().Value(nil, nil, stop)
			delta += value
			if nonZero(value) {
				hasNoEffect = false
			}
		}
		l.deltas[planUnit.Index()] = delta
		l.hasNoEffect[planUnit.Index()] = hasNoEffect
	}

	return nil
}

// String returns the name of the constraint.
func (l *Maximum) String() string {
	return l.name
}

// ID returns the id of the constraint.
func (l *Maximum) ID() string {
	return l.name
}

// SetID sets the id of the constraint.
func (l *Maximum) SetID(id string) {
	l.name = id
}

// EstimationCost returns the cost of estimating whether a move violates the
// constraint.
func (l *Maximum) EstimationCost() Cost {
	if l.hasNegativeValues && !l.hasPositiveValues {
		return Constant
	}

	if l.hasConstantExpression {
		return Constant
	}

	if l.hasStopExpressionAndNoNegativeValues {
		return Constant
	}

	return LinearStop
}

// Expression returns the expression which is used to calculate the
// cumulative value of each stop which is required to stay below the
// maximum value and above zero.
func (l *Maximum) Expression() ModelExpression {
	return l.resourceExpression
}

// ModelExpressions returns the expressions used by the constraint or objective.
func (l *Maximum) ModelExpressions() ModelExpressions {
	return ModelExpressions{l.resourceExpression}
}

// Maximum returns the maximum expression which defines the maximum
// cumulative value that can be assigned to a vehicle type.
func (l *Maximum) Maximum() VehicleTypeExpression {
	return l.maximum
}

// DoesStopHaveViolations returns true if the stop has violations of the
// constraint.
func (l *Maximum) DoesStopHaveViolations(s SolutionStop) bool {
	stop := s
	// We check if the cumulative value is below zero or above the maximum.
	// If there are stops with negative values, the cumulative value can be
	// below zero. Un-planning can result in a cumulative value below zero
	// therefore we need to check for this after un-planning.
	cumulativeValue := stop.CumulativeValue(l.Expression())

	maximum := l.maximum.Value(
		stop.vehicle().ModelVehicle().VehicleType(),
		nil,
		nil,
	)

	return outOf0Range(cumulativeValue, maximum)
}

// EstimateIsViolated estimates whether the given move violates the constraint.
func (l *Maximum) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	moveImpl := move.(*solutionMoveStopsImpl)

	if l.hasNoEffect[moveImpl.planUnit.modelPlanStopsUnit.Index()] {
		return false, NoPositionsHint()
	}

	// All contributions to the level are negative, no need to check
	// it will always be below the implied minimum level of zero.
	if l.hasNegativeValues && !l.hasPositiveValues {
		return true, SkipVehiclePositionsHint()
	}

	vehicle := moveImpl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()

	maximum := l.maximumByVehicleType[vehicleType.Index()]

	expression := l.resourceExpression

	if l.hasConstantExpression {
		value := expression.Value(nil, nil, nil)
		violated := outOf0Range(value, maximum)
		return violated, skipVehicleIfViolated(violated)
	}

	// All contributions to the level are positive, it is sufficient to check
	// if the delta level as a result of the move is not exceeding the maximum
	// level at the end of the vehicle. We can only do this if the expression
	// is a stop expression.
	if l.hasStopExpressionAndNoNegativeValues {
		cumulativeValue := vehicle.Last().CumulativeValue(expression)
		violated := greaterThan(
			cumulativeValue+l.deltas[moveImpl.planUnit.modelPlanStopsUnit.Index()],
			maximum,
		)
		return violated, skipVehicleIfViolated(violated)
	}

	generator := newSolutionStopGenerator(*moveImpl, false, false)
	defer generator.release()
	previousStop, _ := generator.next()
	previousModelStop := previousStop.ModelStop()

	level := previousStop.CumulativeValue(expression)

	for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
		modelStop := solutionStop.ModelStop()
		level += expression.Value(
			vehicleType,
			previousModelStop,
			modelStop,
		)

		if outOf0Range(level, maximum) {
			return true, NoPositionsHint()
		}
		previousStop = solutionStop
		previousModelStop = modelStop
	}

	if !l.hasNegativeValues {
		violated := level-previousStop.CumulativeValue(l.Expression())+
			vehicle.Last().CumulativeValue(l.Expression()) > maximum
		return violated, NoPositionsHint()
	}

	stop, _ := moveImpl.next()

	if stop.CumulativeValue(expression) != level {
		stop = stop.Next()

		for !stop.IsLast() {
			level += stop.Value(expression)

			if outOf0Range(level, maximum) {
				// TODO we can hint the move has to be past this stop
				return true, NoPositionsHint()
			}

			stop = stop.Next()
		}
	}

	return false, NoPositionsHint()
}

type maximumObjectiveDate struct {
	hasViolation bool
}

func (m *maximumObjectiveDate) Copy() Copier {
	return &maximumObjectiveDate{
		hasViolation: m.hasViolation,
	}
}

// UpdateObjectiveStopData updates the objective data for the stop.
func (l *Maximum) UpdateObjectiveStopData(
	solutionStop SolutionStop,
) (Copier, error) {
	if solutionStop.IsFirst() {
		return &maximumObjectiveDate{
			hasViolation: false,
		}, nil
	}

	hasViolation := solutionStop.Previous().ObjectiveData(l).(*maximumObjectiveDate).hasViolation

	if !hasViolation {
		value := solutionStop.CumulativeValue(l.resourceExpression)
		maximum := l.maximumByVehicleType[solutionStop.Vehicle().ModelVehicle().VehicleType().Index()]
		if outOf0Range(value, maximum) {
			hasViolation = true
		}
	}

	return &maximumObjectiveDate{
		hasViolation: hasViolation,
	}, nil
}

// EstimateDeltaValue estimates the delta value of the objective as a result
// of the move.
func (l *Maximum) EstimateDeltaValue(
	move SolutionMoveStops,
) (deltaValue float64) {
	moveImpl := move.(*solutionMoveStopsImpl)

	if l.hasNoEffect[moveImpl.planUnit.modelPlanStopsUnit.Index()] {
		return 0.0
	}

	vehicle := moveImpl.vehicle()

	hasViolation := vehicle.Last().ObjectiveData(l).(*maximumObjectiveDate).hasViolation

	vehicleType := vehicle.ModelVehicle().VehicleType()
	maximum := l.maximumByVehicleType[vehicleType.Index()]

	if l.hasConstantExpression {
		value := l.resourceExpression.Value(nil, nil, nil)
		if greaterThan(value, maximum) {
			return value - maximum + l.penaltyOffset
		}
		if lessThanZero(value) {
			return math.Abs(value) + l.penaltyOffset
		}
		return 0.0
	}

	// All contributions to the level are positive, it is sufficient to check
	// if the delta level as a result of the move is not exceeding the maximum
	// level at the end of the vehicle. We can only do this if the expression
	// is a stop expression.
	if l.hasStopExpressionAndNoNegativeValues {
		cumulativeValue := vehicle.Last().CumulativeValue(l.resourceExpression)

		returnValue := 0.0
		excess := cumulativeValue + l.deltas[moveImpl.planUnit.modelPlanStopsUnit.Index()] - maximum
		if greaterThanZero(excess) {
			if !hasViolation {
				returnValue += l.penaltyOffset
			}
			returnValue += excess
		}
		return returnValue
	}

	estimateDeltaValue := 0.0

	generator := newSolutionStopGenerator(*moveImpl, false, true)
	defer generator.release()
	previousStop, _ := generator.next()

	level := previousStop.CumulativeValue(l.resourceExpression)

	for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
		modelStop := solutionStop.ModelStop()

		level += l.resourceExpression.Value(
			vehicleType,
			previousStop.ModelStop(),
			modelStop,
		)

		if outOf0Range(level, maximum) {
			deltaViolation := level - maximum
			if solutionStop.IsPlanned() {
				deltaViolation -= solutionStop.CumulativeValue(l.resourceExpression)
			}
			if greaterThanZero(deltaViolation) {
				estimateDeltaValue += deltaViolation
				if !hasViolation {
					estimateDeltaValue += l.penaltyOffset
					hasViolation = true
				}
			}
		}

		if solutionStop == moveImpl.Next() {
			if level <= solutionStop.CumulativeValue(l.resourceExpression) {
				break
			}
		}

		previousStop = solutionStop
	}

	return estimateDeltaValue
}

// Value computes value of a solution.
func (l *Maximum) Value(
	solution *Solution,
) (value float64) {
	score := 0.0
	for _, vehicle := range solution.vehicles {
		vehicleType := vehicle.ModelVehicle().VehicleType()
		maximum := l.maximumByVehicleType[vehicleType.Index()]

		if l.hasStopExpressionAndNoNegativeValues {
			cumulativeValue := vehicle.Last().CumulativeValue(l.resourceExpression)
			excess := cumulativeValue - maximum
			if greaterThanZero(excess) {
				score += excess + l.penaltyOffset
			}
			continue
		}
		vehicleScore := 0.0
		solutionStop := vehicle.First()
		for {
			excess := solutionStop.CumulativeValue(l.resourceExpression) - maximum
			if greaterThanZero(excess) {
				vehicleScore += excess
			}
			if solutionStop.IsLast() {
				break
			}
			solutionStop = solutionStop.Next()
		}
		if greaterThanZero(vehicleScore) {
			vehicleScore += l.penaltyOffset
		}
		score += vehicleScore
	}

	return score
}

const (
	// epsilon is used to compare floating point numbers for equality. This is
	// used only by the maximum constraint for now.
	epsilon float64 = 1e-12
)

// greaterThanZero returns true if value is greater than zero considering a
// small epsilon.
func greaterThanZero(value float64) bool {
	return value > epsilon
}

// lessThanZero returns true if value is less than zero considering a small
// epsilon.
func lessThanZero(value float64) bool {
	return value < -epsilon
}

// nonZero returns true if value is not zero considering a small epsilon.
func nonZero(value float64) bool {
	return math.Abs(value) > epsilon
}

// outOf0Range returns true if value is less than zero or greater than max
// considering a small epsilon.
func outOf0Range(value, maximum float64) bool {
	return value < -epsilon || value > maximum+epsilon
}

// greaterThan returns true if value is greater than threshold considering a
// small epsilon.
func greaterThan(value, threshold float64) bool {
	return value > threshold+epsilon
}
