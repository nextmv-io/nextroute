// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// MaximumWaitVehicleConstraint is a constraint that limits the accumulated time
// a vehicle can wait at stops on its route. Wait is defined as the time between
// arriving at a location of stop and starting (to work),
// [SolutionStop.StartValue()] - [SolutionStop.ArrivalValue()].
type MaximumWaitVehicleConstraint interface {
	ModelConstraint

	// Maximum returns the maximum expression which defines the maximum
	// accumulated time a vehicle can wait on a route. Returns nil if not set.
	Maximum() VehicleTypeDurationExpression
}

// NewMaximumWaitVehicleConstraint returns a new MaximumWaitVehicleConstraint.
// The maximum wait constraint limits the accumulated time a vehicle can wait at
// stops on its route. Wait is defined as the time between arriving at a
// stop and starting to do whatever you need to do,
// [SolutionStop.StartValue()] - [SolutionStop.ArrivalValue()].
func NewMaximumWaitVehicleConstraint(
	maxima VehicleTypeDurationExpression,
) (MaximumWaitVehicleConstraint, error) {
	if maxima == nil {
		return nil, fmt.Errorf("maxima must not be nil")
	}
	return &maximumWaitVehicleConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"maximum_vehicle_wait",
			ModelExpressions{},
		),
		maxima: maxima,
	}, nil
}

type maximumWaitVehicleConstraintImpl struct {
	maxima VehicleTypeDurationExpression
	modelConstraintImpl
}

type maximumWaitVehicleConstraintData struct {
	accumulatedWait float64
}

func (c *maximumWaitVehicleConstraintData) Copy() Copier {
	return &maximumWaitVehicleConstraintData{
		accumulatedWait: c.accumulatedWait,
	}
}

func (l *maximumWaitVehicleConstraintImpl) String() string {
	return l.name
}

func (l *maximumWaitVehicleConstraintImpl) EstimationCost() Cost {
	return LinearStop
}

func (l *maximumWaitVehicleConstraintImpl) Maximum() VehicleTypeDurationExpression {
	return l.maxima
}

func (l *maximumWaitVehicleConstraintImpl) UpdateConstraintStopData(
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

func (l *maximumWaitVehicleConstraintImpl) EstimateIsViolated(
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
			return true, constNoPositionsHint
		}

		from = to
	}

	return false, constNoPositionsHint
}

func (l *maximumWaitVehicleConstraintImpl) DoesStopHaveViolations(solution SolutionStop) bool {
	stop := solution
	return stop.ConstraintData(l).(*maximumWaitVehicleConstraintData).accumulatedWait >
		l.maxima.Value(stop.vehicle().ModelVehicle().VehicleType(), nil, nil)
}

func (l *maximumWaitVehicleConstraintImpl) IsTemporal() bool {
	return true
}
