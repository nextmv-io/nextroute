// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// MaximumWaitStopConstraint is a constraint that limits the time a vehicle can
// wait between two stops. Wait is defined as the time between arriving at a
// location of a stop and starting (to work),
// [SolutionStop.StartValue()] - [SolutionStop.ArrivalValue()].
type MaximumWaitStopConstraint interface {
	ModelConstraint

	// Maximum returns the maximum expression which defines the maximum time a
	// vehicle can wait at a stop. Returns nil if not set.
	Maximum() StopDurationExpression
}

// NewMaximumWaitStopConstraint returns a new MaximumWaitStopConstraint. The
// maximum wait constraint for stops limits the time a vehicle can wait at a
// stop.  Wait is defined as the time between arriving at a
// stop and starting to do whatever you need to do,
// [SolutionStop.StartValue()] - [SolutionStop.ArrivalValue()].
func NewMaximumWaitStopConstraint(maxima StopDurationExpression) (
	MaximumWaitStopConstraint,
	error,
) {
	if maxima == nil {
		return nil, fmt.Errorf("maxima must not be nil")
	}
	return &maximumWaitStopConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"maximum_stop_wait",
			ModelExpressions{},
		),
		maxima: maxima,
	}, nil
}

type maximumWaitStopConstraintImpl struct {
	maxima StopDurationExpression
	modelConstraintImpl
}

func (l *maximumWaitStopConstraintImpl) String() string {
	return l.name
}

func (l *maximumWaitStopConstraintImpl) EstimationCost() Cost {
	return LinearStop
}

func (l *maximumWaitStopConstraintImpl) Maximum() StopDurationExpression {
	return l.maxima
}

func (l *maximumWaitStopConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	solutionMoveStops := move.(*solutionMoveStopsImpl)

	vehicle := solutionMoveStops.vehicle()
	stopPositionsCount := len(solutionMoveStops.planUnit.solutionStopsImpl())
	vehicleType := vehicle.ModelVehicle().VehicleType()
	isDependentOnTime := vehicleType.TravelDurationExpression().IsDependentOnTime()

	generator := newSolutionStopGenerator(*solutionMoveStops, false, true)
	defer generator.release()
	from, _ := generator.next()
	previousEnd := from.EndValue()

	for to, ok := generator.next(); ok; to, ok = generator.next() {
		var arrival, start float64

		_, arrival, start, previousEnd = vehicleType.TemporalValues(
			previousEnd,
			from.ModelStop(),
			to.ModelStop(),
		)

		if !to.IsPlanned() {
			stopPositionsCount--
		}

		if !isDependentOnTime &&
			stopPositionsCount == 0 &&
			to.IsPlanned() &&
			arrival == to.ArrivalValue() {
			break
		}

		wait := start - arrival

		if wait > l.maxima.Value(nil, nil, to.modelStop()) {
			return true, constNoPositionsHint
		}

		from = to
	}

	return false, constNoPositionsHint
}

func (l *maximumWaitStopConstraintImpl) DoesStopHaveViolations(s SolutionStop) bool {
	stop := s
	return stop.StartValue()-stop.ArrivalValue() >
		l.maxima.Value(nil, nil, stop.modelStop())
}

func (l *maximumWaitStopConstraintImpl) IsTemporal() bool {
	return true
}
