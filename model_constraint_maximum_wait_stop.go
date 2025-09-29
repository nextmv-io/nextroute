// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// MaximumWaitStopConstraint is a constraint that limits the time a vehicle can
// wait between two stops. Wait is defined as the time between arriving at a
// location of a stop and starting (to work),
// [SolutionStop.StartValue()] - [SolutionStop.ArrivalValue()].
type MaximumWaitStopConstraint struct {
	maxima StopDurationExpression
}

// NewMaximumWaitStopConstraint returns a new MaximumWaitStopConstraint. The
// maximum wait constraint for stops limits the time a vehicle can wait at a
// stop.  Wait is defined as the time between arriving at a
// stop and starting to do whatever you need to do,
// [SolutionStop.StartValue()] - [SolutionStop.ArrivalValue()].
func NewMaximumWaitStopConstraint(maxima StopDurationExpression) (
	*MaximumWaitStopConstraint,
	error,
) {
	if maxima == nil {
		return nil, fmt.Errorf("maxima must not be nil")
	}
	return &MaximumWaitStopConstraint{
		maxima: maxima,
	}, nil
}

// String returns the name of the constraint.
func (l *MaximumWaitStopConstraint) String() string {
	return "maximum_stop_wait"
}

// EstimationCost returns the cost of estimating whether a move violates the
// constraint.
func (l *MaximumWaitStopConstraint) EstimationCost() Cost {
	return LinearStop
}

// Maximum returns the maximum expression which defines the maximum time a
// vehicle can wait at a stop. Returns nil if not set.
func (l *MaximumWaitStopConstraint) Maximum() StopDurationExpression {
	return l.maxima
}

// EstimateIsViolated estimates whether the given move violates the constraint.
func (l *MaximumWaitStopConstraint) EstimateIsViolated(
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

		if wait > l.maxima.Value(nil, nil, to.ModelStop()) {
			return true, NoPositionsHint()
		}

		from = to
	}

	return false, NoPositionsHint()
}

// DoesStopHaveViolations returns true if the stop has violations of the
// constraint.
func (l *MaximumWaitStopConstraint) DoesStopHaveViolations(s SolutionStop) bool {
	stop := s
	return stop.StartValue()-stop.ArrivalValue() >
		l.maxima.Value(nil, nil, stop.ModelStop())
}

// IsTemporal returns true if the constraint is temporal.
func (l *MaximumWaitStopConstraint) IsTemporal() bool {
	return true
}
