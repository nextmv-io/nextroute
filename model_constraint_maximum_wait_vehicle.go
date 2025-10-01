// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// MaximumWaitVehicleConstraint is a constraint that limits the accumulated time
// a vehicle can wait at stops on its route. Wait is defined as the time between
// arriving at a location of stop and starting (to work),
// [SolutionStop.StartValue()] - [SolutionStop.ArrivalValue()].
type MaximumWaitVehicleConstraint struct {
	maxima VehicleTypeDurationExpression
}

// NewMaximumWaitVehicleConstraint returns a new MaximumWaitVehicleConstraint.
// The maximum wait constraint limits the accumulated time a vehicle can wait at
// stops on its route. Wait is defined as the time between arriving at a
// stop and starting to do whatever you need to do,
// [SolutionStop.StartValue()] - [SolutionStop.ArrivalValue()].
func NewMaximumWaitVehicleConstraint(
	maxima VehicleTypeDurationExpression,
) (*MaximumWaitVehicleConstraint, error) {
	if maxima == nil {
		return nil, fmt.Errorf("maxima must not be nil")
	}
	return &MaximumWaitVehicleConstraint{
		maxima: maxima,
	}, nil
}

type maximumWaitVehicleConstraintData struct {
	accumulatedWait float64
}

func (c *maximumWaitVehicleConstraintData) Copy() Copier {
	return &maximumWaitVehicleConstraintData{
		accumulatedWait: c.accumulatedWait,
	}
}

// String returns the name of the constraint.
func (l *MaximumWaitVehicleConstraint) String() string {
	return "maximum_vehicle_wait"
}

// EstimationCost returns the cost of estimating whether a move violates the
// constraint.
func (l *MaximumWaitVehicleConstraint) EstimationCost() Cost {
	return LinearStop
}

// Maximum returns the maximum expression which defines the maximum
// accumulated time a vehicle can wait on a route. Returns nil if not set.
func (l *MaximumWaitVehicleConstraint) Maximum() VehicleTypeDurationExpression {
	return l.maxima
}

// UpdateConstraintStopData updates the constraint data for the given stop.
func (l *MaximumWaitVehicleConstraint) UpdateConstraintStopData(
	solutionStop SolutionStop,
) (Copier, error) {
	if solutionStop.IsFirst() {
		// First stop, no waiting time - we immediately start driving.
		return &maximumWaitVehicleConstraintData{accumulatedWait: 0.0}, nil
	}

	previousData := solutionStop.Previous().ConstraintData(l).(*maximumWaitVehicleConstraintData)
	if previousData == nil {
		return nil, fmt.Errorf("no previous data found")
	}

	if solutionStop.IsLast() {
		// Last stop, no window to wait for - we immediately finish with data
		// from predecessor.
		return previousData, nil
	}

	wait := solutionStop.StartValue() - solutionStop.ArrivalValue()
	return &maximumWaitVehicleConstraintData{
		accumulatedWait: previousData.accumulatedWait + wait,
	}, nil
}

// EstimateIsViolated estimates if the constraint is violated by the given move.
func (l *MaximumWaitVehicleConstraint) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	stopPositionsCount := len(moveImpl.planUnit.solutionStopsImpl())
	vehicleType := vehicle.ModelVehicle().VehicleType()
	isDependentOnTime := vehicleType.TravelDurationExpression().IsDependentOnTime()

	maxWait := l.maxima.Value(vehicleType, nil, nil)

	generator := newSolutionStopGenerator(*moveImpl, false, true)
	defer generator.release()
	from, _ := generator.next()
	accumulatedWait := from.ConstraintData(l).(*maximumWaitVehicleConstraintData).accumulatedWait

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
		accumulatedWait += wait
		if accumulatedWait > maxWait {
			return true, NoPositionsHint()
		}

		from = to
	}

	return false, NoPositionsHint()
}

// DoesStopHaveViolations returns true if the stop has violations of the
// constraint.
func (l *MaximumWaitVehicleConstraint) DoesStopHaveViolations(solution SolutionStop) bool {
	stop := solution
	return stop.ConstraintData(l).(*maximumWaitVehicleConstraintData).accumulatedWait >
		l.maxima.Value(stop.vehicle().ModelVehicle().VehicleType(), nil, nil)
}

// IsTemporal returns true if the constraint is temporal.
func (l *MaximumWaitVehicleConstraint) IsTemporal() bool {
	return true
}
