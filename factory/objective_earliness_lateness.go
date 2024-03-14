// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	"github.com/nextmv-io/nextroute/schema"
)

// addEarlinessObjective adds an earliness penalty (per stop) objective to the
// Model.
func addEarlinessObjective(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	targetTimeExpression, model, err := targetTimeExpression(model)
	if err != nil {
		return nil, err
	}

	factorExpression := nextroute.NewStopExpression(
		"earliness_penalty_factor",
		0.0,
	)

	stopsHaveTargets, err := addEarlinessTargetStops(
		input,
		model,
		factorExpression,
		targetTimeExpression,
	)
	if err != nil {
		return nil, err
	}

	alternateStopsHaveTargets, err := addEarlinessTargetsAlternateStops(
		input,
		model,
		factorExpression,
		targetTimeExpression,
	)
	if err != nil {
		return nil, err
	}

	if !stopsHaveTargets && !alternateStopsHaveTargets {
		return model, nil
	}

	earlinessObjective, err := nextroute.NewEarlinessObjective(
		targetTimeExpression,
		factorExpression,
		nextroute.OnArrival,
	)
	if err != nil {
		return nil, err
	}

	_, err = model.Objective().NewTerm(options.Objectives.EarlyArrivalPenalty, earlinessObjective)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func addEarlinessTargetStops(
	input schema.Input,
	model nextroute.Model,
	factorExpression nextroute.StopExpression,
	targetTimeExpression nextroute.StopTimeExpression,
) (bool, error) {
	present := false
	for index, inputStop := range input.Stops {
		if inputStop.TargetArrivalTime == nil ||
			inputStop.EarlyArrivalTimePenalty == nil ||
			*inputStop.EarlyArrivalTimePenalty == 0.0 {
			continue
		}

		present = true
		stop, err := model.Stop(index)
		if err != nil {
			return false, err
		}

		err = factorExpression.SetValue(stop, *inputStop.EarlyArrivalTimePenalty)
		if err != nil {
			return false, err
		}
		targetTimeExpression.SetTime(stop, *inputStop.TargetArrivalTime)
	}

	return present, nil
}

func addEarlinessTargetsAlternateStops(
	input schema.Input,
	model nextroute.Model,
	factorExpression nextroute.StopExpression,
	targetTimeExpression nextroute.StopTimeExpression,
) (bool, error) {
	if input.AlternateStops == nil {
		return false, nil
	}

	if common.AllTrue(
		*input.AlternateStops,
		func(stop schema.AlternateStop) bool {
			return stop.TargetArrivalTime == nil ||
				stop.EarlyArrivalTimePenalty == nil ||
				*stop.EarlyArrivalTimePenalty == 0.0
		},
	) {
		return false, nil
	}

	data, err := getModelData(model)
	if err != nil {
		return false, err
	}

	hasEarlinessTarget := false

	for _, vehicle := range input.Vehicles {
		if vehicle.AlternateStops == nil {
			continue
		}

		for _, alternateID := range *vehicle.AlternateStops {
			stop, err := model.Stop(data.stopIDToIndex[alternateStopID(alternateID, vehicle)])
			if err != nil {
				return false, err
			}

			alternateInputStop := stop.Data().(alternateInputStop)

			if alternateInputStop.stop.TargetArrivalTime == nil ||
				alternateInputStop.stop.EarlyArrivalTimePenalty == nil ||
				*alternateInputStop.stop.EarlyArrivalTimePenalty == 0.0 {
				continue
			}
			hasEarlinessTarget = true
			err = factorExpression.SetValue(stop, *alternateInputStop.stop.EarlyArrivalTimePenalty)
			if err != nil {
				return false, err
			}
			targetTimeExpression.SetTime(stop, *alternateInputStop.stop.TargetArrivalTime)
		}
	}

	return hasEarlinessTarget, nil
}

// addLatenessObjective adds a lateness penalty (per stop) objective to the
// Model.
func addLatenessObjective(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	if options.Objectives.LateArrivalPenalty == 0.0 {
		return model, nil
	}

	targetTimeExpression, model, err := targetTimeExpression(model)
	if err != nil {
		return nil, err
	}

	latenessObjective, err := nextroute.NewLatestArrival(targetTimeExpression)
	if err != nil {
		return nil, err
	}

	stopsHaveTargets, err := addLatenessTargetStops(input, model, latenessObjective)

	if err != nil {
		return nil, err
	}

	alternateStopsHaveTargets, err := addLatenessTargetsAlternateStops(input, model, latenessObjective)
	if err != nil {
		return nil, err
	}

	if !stopsHaveTargets && !alternateStopsHaveTargets {
		return model, nil
	}

	_, err = model.Objective().NewTerm(options.Objectives.LateArrivalPenalty, latenessObjective)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func addLatenessTargetStops(
	input schema.Input,
	model nextroute.Model,
	objective nextroute.LatestArrival,
) (bool, error) {
	hasTargets := false

	for index, inputStop := range input.Stops {
		if inputStop.TargetArrivalTime == nil ||
			inputStop.LateArrivalTimePenalty == nil ||
			*inputStop.LateArrivalTimePenalty == 0.0 {
			continue
		}
		stop, err := model.Stop(index)
		if err != nil {
			return false, err
		}
		objective.Latest().SetTime(stop, *inputStop.TargetArrivalTime)
		err = objective.SetFactor(*inputStop.LateArrivalTimePenalty, stop)
		if err != nil {
			return false, err
		}

		hasTargets = true
	}

	return hasTargets, nil
}

func addLatenessTargetsAlternateStops(
	input schema.Input,
	model nextroute.Model,
	objective nextroute.LatestArrival,
) (bool, error) {
	if input.AlternateStops == nil {
		return false, nil
	}

	if common.AllTrue(
		*input.AlternateStops,
		func(stop schema.AlternateStop) bool {
			return stop.TargetArrivalTime == nil ||
				stop.LateArrivalTimePenalty == nil ||
				*stop.LateArrivalTimePenalty == 0.0
		},
	) {
		return false, nil
	}

	data, err := getModelData(model)
	if err != nil {
		return false, err
	}

	hasTarget := false

	for _, vehicle := range input.Vehicles {
		if vehicle.AlternateStops == nil {
			continue
		}

		for _, alternateID := range *vehicle.AlternateStops {
			stop, err := model.Stop(data.stopIDToIndex[alternateStopID(alternateID, vehicle)])
			if err != nil {
				return false, err
			}

			alternateInputStop := stop.Data().(alternateInputStop)

			if alternateInputStop.stop.TargetArrivalTime == nil ||
				alternateInputStop.stop.LateArrivalTimePenalty == nil ||
				*alternateInputStop.stop.LateArrivalTimePenalty == 0.0 {
				continue
			}
			hasTarget = true
			objective.Latest().SetTime(stop, *alternateInputStop.stop.TargetArrivalTime)
			err = objective.SetFactor(*alternateInputStop.stop.LateArrivalTimePenalty, stop)
			if err != nil {
				return false, err
			}
		}
	}

	return hasTarget, nil
}
