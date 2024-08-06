// © 2019-present nextmv.io inc

package factory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	"github.com/nextmv-io/nextroute/schema"
	runSchema "github.com/nextmv-io/sdk/run/schema"
)

// Format formats a solution in a basic format using the [schema.Output] to
// format a solution.
func Format(
	ctx context.Context,
	options any,
	progressioner nextroute.Progressioner,
	solutions ...nextroute.Solution,
) runSchema.Output {
	return nextroute.Format(
		ctx,
		options,
		progressioner,
		func(solution nextroute.Solution) any {
			return ToSolutionOutput(solution)
		},
		solutions...,
	)
}

// toSolutionOutputStops converts a solution plan unit to a slice of
// [schema.StopOutput].
func toSolutionOutputStops(solutionPlanUnit nextroute.SolutionPlanUnit) []schema.StopOutput {
	switch v := solutionPlanUnit.(type) {
	case nextroute.SolutionPlanStopsUnit:
		return common.Map(
			v.SolutionStops(),
			func(s nextroute.SolutionStop) schema.StopOutput {
				return toStopOutput(s.ModelStop())
			},
		)
	case nextroute.SolutionPlanUnitsUnit:
		if v.ModelPlanUnitsUnit().PlanAll() {
			return common.MapSlice(
				v.SolutionPlanUnits(),
				toSolutionOutputStops,
			)
		}
	}
	return []schema.StopOutput{}
}

// ToSolutionOutput converts a solution to a [schema.SolutionOutput].
func ToSolutionOutput(solution nextroute.Solution) schema.SolutionOutput {
	unplannedStops := common.MapSlice(
		solution.UnPlannedPlanUnits().SolutionPlanUnits(),
		toSolutionOutputStops,
	)
	sort.SliceStable(unplannedStops, func(i, j int) bool {
		return unplannedStops[i].ID < unplannedStops[j].ID
	})

	return schema.SolutionOutput{
		Unplanned: unplannedStops,
		Vehicles: common.Map(
			solution.Vehicles(),
			toVehicleOutput,
		),
		Objective: toObjectiveOutput(solution),
	}
}

func toStopOutput(modelStop nextroute.ModelStop) schema.StopOutput {
	var customData any
	if inputStop, ok := modelStop.Data().(schema.Stop); ok {
		customData = inputStop.CustomData
	}
	return schema.StopOutput{
		ID: modelStop.ID(),
		Location: schema.Location{
			Lon: modelStop.Location().Longitude(),
			Lat: modelStop.Location().Latitude(),
		},
		CustomData: customData,
	}
}

func toPlannedStopOutput(solutionStop nextroute.SolutionStop) schema.PlannedStopOutput {
	timezoneLocation := solutionStop.
		Vehicle().
		ModelVehicle().
		Start().
		Location()

	plannedStopOutput := schema.PlannedStopOutput{
		Stop:                     toStopOutput(solutionStop.ModelStop()),
		TravelDuration:           int(solutionStop.TravelDuration().Seconds()),
		CumulativeTravelDuration: int(solutionStop.CumulativeTravelDuration().Seconds()),
		Duration:                 int(solutionStop.End().Sub(solutionStop.Start()).Seconds()),
		WaitingDuration:          int(solutionStop.Start().Sub(solutionStop.Arrival()).Seconds()),
	}

	arrival := solutionStop.Arrival().In(timezoneLocation)
	end := solutionStop.End().In(timezoneLocation)
	start := solutionStop.Start().In(timezoneLocation)

	if solutionStop.Vehicle().First().Start() !=
		solutionStop.Vehicle().ModelVehicle().Model().Epoch() {
		plannedStopOutput.ArrivalTime = &arrival
		plannedStopOutput.EndTime = &end
		plannedStopOutput.StartTime = &start
	}

	if inputStop, ok := solutionStop.ModelStop().Data().(schema.Stop); ok {
		if inputStop.TargetArrivalTime != nil {
			targetArrivalTime := inputStop.TargetArrivalTime.In(timezoneLocation)
			plannedStopOutput.TargetArrivalTime = &targetArrivalTime
		}

		if inputStop.EarlyArrivalTimePenalty != nil && inputStop.TargetArrivalTime != nil {
			plannedStopOutput.EarlyArrivalDuration =
				int(math.Max(inputStop.TargetArrivalTime.Sub(arrival).Seconds(), 0.0))
		}

		if inputStop.LateArrivalTimePenalty != nil && inputStop.TargetArrivalTime != nil {
			plannedStopOutput.LateArrivalDuration =
				int(math.Max(arrival.Sub(*inputStop.TargetArrivalTime).Seconds(), 0.0))
		}

		mixItems := make(map[string]nextroute.MixItem)
		for _, constraint := range solutionStop.Vehicle().ModelVehicle().Model().Constraints() {
			if noMixConstraint, ok := constraint.(nextroute.NoMixConstraint); ok {
				mixItems[strings.TrimPrefix(noMixConstraint.ID(), "no_mix_")] = noMixConstraint.Value(solutionStop)
			}
		}
		if len(mixItems) > 0 {
			plannedStopOutput.MixItems = mixItems
		}
	}

	hasTravelDistance := solutionStop.Previous().ModelStop().Location().IsValid() &&
		solutionStop.ModelStop().Location().IsValid()
	if data, ok := solutionStop.Vehicle().ModelVehicle().VehicleType().Data().(vehicleTypeData); ok && hasTravelDistance {
		distance := data.DistanceExpression.Value(
			solutionStop.Vehicle().ModelVehicle().VehicleType(),
			solutionStop.Previous().ModelStop(),
			solutionStop.ModelStop(),
		)
		plannedStopOutput.TravelDistance = int(distance)
	}

	return plannedStopOutput
}

func toVehicleOutput(vehicle nextroute.SolutionVehicle) schema.VehicleOutput {
	solutionStops := common.Filter(
		vehicle.SolutionStops(),
		func(solutionStop nextroute.SolutionStop) bool {
			return solutionStop.ModelStop().Location().IsValid()
		},
	)

	route := common.Map(
		solutionStops,
		toPlannedStopOutput,
	)

	routeTravelDistance := 0
	routeStopsDuration := 0
	for idx, stop := range route {
		routeTravelDistance += stop.TravelDistance
		routeStopsDuration += stop.Duration

		route[idx].CumulativeTravelDistance = routeTravelDistance
	}

	vehicleOutput := schema.VehicleOutput{
		ID:                  vehicle.ModelVehicle().ID(),
		Route:               route,
		RouteDuration:       int(vehicle.Duration().Seconds()),
		RouteTravelDuration: int(vehicle.Last().CumulativeTravelDuration().Seconds()),
		RouteTravelDistance: routeTravelDistance,
		RouteStopsDuration:  routeStopsDuration,
	}

	if inputVehicle, ok := vehicle.ModelVehicle().Data().(schema.Vehicle); ok {
		if inputVehicle.CustomData != nil {
			vehicleOutput.CustomData = inputVehicle.CustomData
		}
		if inputVehicle.AlternateStops != nil {
			model := vehicle.ModelVehicle().Model()
			data, err := getModelData(model)

			if err != nil {
				return schema.VehicleOutput{
					ID: fmt.Sprintf("error in outputting vehicle: %v", err),
				}
			}

			alternateStops := make([]string, 0)

			for _, alternateID := range *inputVehicle.AlternateStops {
				stop, err := model.Stop(data.stopIDToIndex[alternateStopID(alternateID, inputVehicle)])
				if err != nil {
					return schema.VehicleOutput{
						ID: fmt.Sprintf("error in outputting vehicle: %v", err),
					}
				}
				if vehicle.First().Solution().SolutionStop(stop).IsPlanned() {
					alternateStops = append(alternateStops, stop.ID())
				}
			}

			vehicleOutput.AlternateStops = &alternateStops
		}
	}

	vehicleOutput.RouteWaitingDuration = vehicleOutput.RouteDuration -
		vehicleOutput.RouteTravelDuration - vehicleOutput.RouteStopsDuration

	return vehicleOutput
}

func toObjectiveOutput(solution nextroute.Solution) schema.ObjectiveOutput {
	return schema.ObjectiveOutput{
		Name: fmt.Sprintf("%v", solution.Model().Objective()),
		Objectives: common.Map(
			solution.Model().Objective().Terms(),
			func(modelObjectiveTerm nextroute.ModelObjectiveTerm) schema.ObjectiveOutput {
				return schema.ObjectiveOutput{
					Name:   fmt.Sprintf("%v", modelObjectiveTerm.Objective()),
					Factor: modelObjectiveTerm.Factor(),
					Base:   solution.ObjectiveValue(modelObjectiveTerm.Objective()) / modelObjectiveTerm.Factor(),
					Value:  solution.ObjectiveValue(modelObjectiveTerm.Objective()),
				}
			},
		),
		Value: solution.ObjectiveValue(solution.Model().Objective()),
	}
}

// DefaultCustomResultStatistics creates default custom statistics for a given
// solution.
func DefaultCustomResultStatistics(solution nextroute.Solution) schema.CustomResultStatistics {
	vehicleCount := 0
	maxTravelDuration := 0
	minTravelDuration := math.MaxInt64
	maxDuration := 0
	minDuration := math.MaxInt64
	maxStops := 0
	minStops := math.MaxInt64
	for _, vehicle := range solution.Vehicles() {
		if vehicle.IsEmpty() {
			continue
		}

		vehicleCount++
		duration := vehicle.Duration().Seconds()
		if int(duration) > maxDuration {
			maxDuration = int(duration)
		}
		if int(duration) < minDuration {
			minDuration = int(duration)
		}

		travelDuration := int(vehicle.Last().CumulativeTravelDuration().Seconds())
		if travelDuration > maxTravelDuration {
			maxTravelDuration = travelDuration
		}
		if travelDuration < minTravelDuration {
			minTravelDuration = travelDuration
		}

		stops := vehicle.NumberOfStops()
		if stops > maxStops {
			maxStops = stops
		}
		if stops < minStops {
			minStops = stops
		}
	}

	unplannedStops := common.MapSlice(
		solution.UnPlannedPlanUnits().SolutionPlanUnits(),
		toSolutionOutputStops,
	)

	return schema.CustomResultStatistics{
		ActivatedVehicles: vehicleCount,
		UnplannedStops:    len(unplannedStops),
		MaxTravelDuration: maxTravelDuration,
		MaxDuration:       maxDuration,
		MinTravelDuration: minTravelDuration,
		MinDuration:       minDuration,
		MaxStopsInVehicle: maxStops,
		MinStopsInVehicle: minStops,
	}
}
