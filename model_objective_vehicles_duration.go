// © 2019-present nextmv.io inc

package nextroute

import "slices"

// VehiclesDurationObjective is an objective that uses the vehicle duration as an
// objective.
type VehiclesDurationObjective interface {
	ModelObjective
}

// NewVehiclesDurationObjective returns a new VehiclesDurationObjective.
func NewVehiclesDurationObjective() VehiclesDurationObjective {
	return &vehiclesDurationObjectiveImpl{}
}

func (t *vehiclesDurationObjectiveImpl) Lock(model Model) error {
	t.canIncurWaitingTime = slices.ContainsFunc(model.Stops(), func(stop ModelStop) bool {
		return stop.(*stopImpl).canIncurWaitingTime()
	})
	vehicleTypes := model.VehicleTypes()
	t.isDependentOnTimeByVehicleType = make([]bool, len(vehicleTypes))
	for _, vehicleType := range model.VehicleTypes() {
		t.isDependentOnTimeByVehicleType[vehicleType.Index()] = vehicleType.
			TravelDurationExpression().
			IsDependentOnTime()
	}
	// caching the vehicle type by index for performance
	t.vehicleTypesByIndex = make([]ModelVehicleType, len(vehicleTypes))
	for _, vehicle := range model.Vehicles() {
		t.vehicleTypesByIndex[vehicle.Index()] = vehicle.VehicleType()
	}
	return nil
}

type vehiclesDurationObjectiveImpl struct {
	isDependentOnTimeByVehicleType []bool
	vehicleTypesByIndex            []ModelVehicleType
	canIncurWaitingTime            bool
}

func (t *vehiclesDurationObjectiveImpl) ModelExpressions() ModelExpressions {
	return ModelExpressions{}
}

func (t *vehiclesDurationObjectiveImpl) EstimateDeltaValue(
	move SolutionMoveStops,
) float64 {
	solutionMoveStops := move.(*solutionMoveStopsImpl)
	vehicle := solutionMoveStops.vehicle()
	vehicleType := t.vehicleTypesByIndex[vehicle.index]

	isDependentOnTime := t.isDependentOnTimeByVehicleType[vehicleType.Index()]
	if !isDependentOnTime && !t.canIncurWaitingTime {
		return solutionMoveStops.deltaStopTravelDurationValue(vehicleType)
	}

	first := true
	end := 0.0
	previousStop := vehicle.First()

	generator := newSolutionStopGenerator(*solutionMoveStops, false, isDependentOnTime)
	defer generator.release()

	for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
		if first {
			previousStop = solutionStop
			end = solutionStop.EndValue()
			first = false
			continue
		}

		_, _, _, end = vehicleType.TemporalValues(
			end,
			previousStop.ModelStop(),
			solutionStop.ModelStop(),
		)

		previousStop = solutionStop
	}

	nextmove, _ := solutionMoveStops.next()

	if nextmove.IsLast() || isDependentOnTime {
		return end - vehicle.Last().EndValue()
	}

	for solutionStop := nextmove.Next(); !solutionStop.IsLast(); solutionStop = solutionStop.Next() {
		_, _, _, end = vehicleType.TemporalValues(
			end,
			solutionStop.Previous().ModelStop(),
			solutionStop.ModelStop(),
		)
		tempEnd := solutionStop.EndValue()

		if tempEnd >= end {
			return 0.0
		}
	}

	last := vehicle.Last()
	_, _, _, end = vehicleType.TemporalValues(
		end,
		last.Previous().ModelStop(),
		last.ModelStop(),
	)

	return end - last.EndValue()
}

func (t *vehiclesDurationObjectiveImpl) Value(
	solution Solution,
) float64 {
	solutionImp := solution.(*solutionImpl)
	score := 0.0
	for _, r := range solutionImp.vehicles {
		score += r.DurationValue()
	}
	return score
}

func (t *vehiclesDurationObjectiveImpl) String() string {
	return "vehicles_duration"
}
