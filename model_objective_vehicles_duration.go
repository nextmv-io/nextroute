package nextroute

import (
	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

// NewVehiclesDurationObjective returns a new VehiclesDurationObjective.
func NewVehiclesDurationObjective() nextroute.VehiclesDurationObjective {
	return &vehiclesDurationObjectiveImpl{}
}

func (t *vehiclesDurationObjectiveImpl) Lock(model nextroute.Model) error {
	t.canIncurWaitingTime = common.HasTrue(model.Stops(), func(stop nextroute.ModelStop) bool {
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
	t.vehicleTypesByIndex = make([]nextroute.ModelVehicleType, len(vehicleTypes))
	for _, vehicle := range model.Vehicles() {
		t.vehicleTypesByIndex[vehicle.Index()] = vehicle.VehicleType()
	}
	return nil
}

type vehiclesDurationObjectiveImpl struct {
	isDependentOnTimeByVehicleType []bool
	vehicleTypesByIndex            []nextroute.ModelVehicleType
	canIncurWaitingTime            bool
}

func (t *vehiclesDurationObjectiveImpl) ModelExpressions() nextroute.ModelExpressions {
	return nextroute.ModelExpressions{}
}

func (t *vehiclesDurationObjectiveImpl) EstimateDeltaValue(
	move nextroute.SolutionMoveStops,
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
	previousStop := vehicle.first()

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
		return end - vehicle.last().EndValue()
	}

	for solutionStop := nextmove.next(); !solutionStop.IsLast(); solutionStop = solutionStop.next() {
		_, _, _, end = vehicleType.TemporalValues(
			end,
			solutionStop.previous().ModelStop(),
			solutionStop.ModelStop(),
		)
		tempEnd := solutionStop.EndValue()

		if tempEnd >= end {
			return 0.0
		}
	}

	last := vehicle.last()
	_, _, _, end = vehicleType.TemporalValues(
		end,
		last.previous().ModelStop(),
		last.ModelStop(),
	)

	return end - last.EndValue()
}

func (t *vehiclesDurationObjectiveImpl) Value(
	solution nextroute.Solution,
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
