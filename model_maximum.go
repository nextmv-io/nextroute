package nextroute

import (
	"fmt"
	"math"
)

// NewMaximum creates a new maximum construct which can be used as constraint
// or as objective.
func NewMaximum(
	expression ModelExpression,
	maximum VehicleTypeExpression,
) (Maximum, error) {
	return &maximumImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"maximum",
			ModelExpressions{expression},
		),
		maximum:       maximum,
		penaltyOffset: 0.0,
	}, nil
}

type maximumImpl struct {
	maximum VehicleTypeExpression
	deltas  []float64
	modelConstraintImpl
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
}

func (l *maximumImpl) PenaltyOffset() float64 {
	return l.penaltyOffset
}

func (l *maximumImpl) SetPenaltyOffset(penaltyOffset float64) error {
	if penaltyOffset < 0.0 {
		return fmt.Errorf(
			"maximum objective, penalty offset must be positive, it can not be %f",
			penaltyOffset,
		)
	}

	l.penaltyOffset = penaltyOffset

	return nil
}

func (l *maximumImpl) Lock(model Model) error {
	l.hasNegativeValues = l.Expression().HasNegativeValues()
	l.hasPositiveValues = l.Expression().HasPositiveValues()
	if _, ok := l.Expression().(ConstantExpression); ok {
		l.hasConstantExpression = true
	}
	if _, ok := l.Expression().(StopExpression); ok &&
		!l.hasNegativeValues {
		l.hasStopExpressionAndNoNegativeValues = true
	}
	l.resourceExpression = l.expressions[0]
	vehicleTypes := model.VehicleTypes()
	l.maximumByVehicleType = make([]float64, len(vehicleTypes))
	for _, vehicleType := range vehicleTypes {
		l.maximumByVehicleType[vehicleType.Index()] = l.maximum.Value(
			vehicleType,
			nil,
			nil,
		)
	}

	if !l.hasStopExpressionAndNoNegativeValues {
		return nil
	}

	planUnits := model.PlanStopsUnits()
	l.deltas = make([]float64, len(planUnits))
	for _, planUnit := range model.PlanStopsUnits() {
		delta := 0.0
		for _, stop := range planUnit.Stops() {
			delta += l.Expression().Value(nil, nil, stop)
		}
		l.deltas[planUnit.Index()] = delta
	}

	return nil
}

func (l *maximumImpl) String() string {
	return l.name
}

func (l *maximumImpl) ID() string {
	return l.name
}

func (l *maximumImpl) SetID(id string) {
	l.name = id
}

func (l *maximumImpl) EstimationCost() Cost {
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

func (l *maximumImpl) Expression() ModelExpression {
	return l.expressions[0]
}

func (l *maximumImpl) Maximum() VehicleTypeExpression {
	return l.maximum
}

func (l *maximumImpl) DoesStopHaveViolations(s SolutionStop) bool {
	stop := s.(solutionStopImpl)
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

	return cumulativeValue > maximum || cumulativeValue < 0.0
}

func (l *maximumImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	// All contributions to the level are negative, no need to check
	// it will always be below the implied minimum level of zero.
	if l.hasNegativeValues && !l.hasPositiveValues {
		return true, constSkipVehiclePositionsHint
	}

	moveImpl := move.(*solutionMoveStopsImpl)

	vehicle := moveImpl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()

	maximum := l.maximumByVehicleType[vehicleType.Index()]

	expression := l.resourceExpression

	if l.hasConstantExpression {
		value := expression.Value(nil, nil, nil)
		if value > maximum || value < 0 {
			return true, constSkipVehiclePositionsHint
		}
		return false, constNoPositionsHint
	}

	// All contributions to the level are positive, it is sufficient to check
	// if the delta level as a result of the move is not exceeding the maximum
	// level at the end of the vehicle. We can only do this if the expression
	// is a stop expression.
	if l.hasStopExpressionAndNoNegativeValues {
		cumulativeValue := vehicle.last().CumulativeValue(expression)

		if cumulativeValue+l.deltas[moveImpl.planUnit.modelPlanStopsUnit.Index()] > maximum {
			return true, constSkipVehiclePositionsHint
		}

		return false, constNoPositionsHint
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

		if level > maximum || level < 0 {
			return true, constNoPositionsHint
		}
		previousStop = solutionStop
		previousModelStop = modelStop
	}

	if !l.hasNegativeValues {
		violated := level-previousStop.CumulativeValue(l.Expression())+
			vehicle.last().CumulativeValue(l.Expression()) > maximum
		return violated, constNoPositionsHint
	}

	stop, _ := moveImpl.next()

	if stop.CumulativeValue(expression) < level {
		stop = stop.next()

		for !stop.IsLast() {
			level += stop.Value(expression)

			if level > maximum || level < 0 {
				// TODO we can hint the move has to be past this stop
				return true, constNoPositionsHint
			}

			stop = stop.next()
		}
	}

	return false, constNoPositionsHint
}

type maximumObjectiveDate struct {
	hasViolation bool
}

func (m *maximumObjectiveDate) Copy() Copier {
	return &maximumObjectiveDate{
		hasViolation: m.hasViolation,
	}
}

func (l *maximumImpl) UpdateObjectiveStopData(
	solutionStop SolutionStop,
) (Copier, error) {
	if solutionStop.IsFirst() {
		return &maximumObjectiveDate{
			hasViolation: false,
		}, nil
	}
	hasViolation := solutionStop.Previous().ObjectiveData(l).(*maximumObjectiveDate).hasViolation

	if !hasViolation {
		maximum := l.maximumByVehicleType[solutionStop.Vehicle().ModelVehicle().VehicleType().Index()]
		value := solutionStop.CumulativeValue(l.resourceExpression)
		if value > maximum || value < 0 {
			hasViolation = true
		}
	}
	return &maximumObjectiveDate{
		hasViolation: hasViolation,
	}, nil
}

func (l *maximumImpl) EstimateDeltaValue(
	move SolutionMoveStops,
) (deltaValue float64) {
	moveImpl := move.(*solutionMoveStopsImpl)

	vehicle := moveImpl.vehicle()

	hasViolation := vehicle.last().ObjectiveData(l).(*maximumObjectiveDate).hasViolation

	vehicleType := vehicle.ModelVehicle().VehicleType()
	maximum := l.maximumByVehicleType[vehicleType.Index()]

	if l.hasConstantExpression {
		value := l.resourceExpression.Value(nil, nil, nil)
		if value > maximum {
			return value - maximum + l.penaltyOffset
		}
		if value < 0 {
			return math.Abs(value) + l.penaltyOffset
		}
		return 0.0
	}

	// All contributions to the level are positive, it is sufficient to check
	// if the delta level as a result of the move is not exceeding the maximum
	// level at the end of the vehicle. We can only do this if the expression
	// is a stop expression.
	if l.hasStopExpressionAndNoNegativeValues {
		cumulativeValue := vehicle.last().CumulativeValue(l.resourceExpression)

		returnValue := 0.0
		excess := cumulativeValue + l.deltas[moveImpl.planUnit.modelPlanStopsUnit.Index()] - maximum
		if excess > 0 {
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

		if level > maximum || level < 0 {
			deltaViolation := level - maximum
			if solutionStop.IsPlanned() {
				deltaViolation -= solutionStop.CumulativeValue(l.resourceExpression)
			}
			if deltaViolation > 0. {
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

func (l *maximumImpl) Value(
	solution Solution,
) (value float64) {
	solutionImp := solution.(*solutionImpl)

	score := 0.0

	for _, vehicle := range solutionImp.vehicles {
		vehicleType := vehicle.ModelVehicle().VehicleType()
		maximum := l.maximumByVehicleType[vehicleType.Index()]

		if l.hasStopExpressionAndNoNegativeValues {
			cumulativeValue := vehicle.last().CumulativeValue(l.resourceExpression)
			excess := cumulativeValue - maximum
			if excess > 0 {
				score += excess
			}
			continue
		}

		for _, solutionStop := range vehicle.SolutionStops() {
			solutionStop.CumulativeValue(l.resourceExpression)
			excess := solutionStop.CumulativeValue(l.resourceExpression) - maximum
			if excess > 0 {
				score += excess
			}
		}
	}

	if score > 0 {
		score += l.penaltyOffset
	}

	return score
}
