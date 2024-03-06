// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

const defaultUnplannedPenaltyStop = 1_000_000
const defaultUnplannedPenaltyAlternateStop = 2_000_000

// addUnplannedObjective uses the unplanned penalty from the stops to
// create a new objective and add it to the model.
func addUnplannedObjective(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	if options.Objectives.UnplannedPenalty == 0 {
		return model, nil
	}

	unplannedPenalty := nextroute.NewStopExpression(
		"unplanned_penalty",
		defaultUnplannedPenaltyStop,
	)
	err := addUnplannedPenaltyStops(input, model, unplannedPenalty)
	if err != nil {
		return nil, err
	}

	err = addUnplannedPenaltyAlternateStops(input, model, unplannedPenalty)
	if err != nil {
		return nil, err
	}

	unplannedObjective := nextroute.NewUnPlannedObjective(unplannedPenalty)
	_, err = model.Objective().NewTerm(options.Objectives.UnplannedPenalty, unplannedObjective)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func addUnplannedPenaltyStops(
	input schema.Input,
	model nextroute.Model,
	unplannedPenaltyExpression nextroute.StopExpression,
) error {
	for s, inputStop := range input.Stops {
		if inputStop.UnplannedPenalty == nil {
			continue
		}
		stop, err := model.Stop(s)
		if err != nil {
			return err
		}
		err = unplannedPenaltyExpression.SetValue(stop, float64(*inputStop.UnplannedPenalty))
		if err != nil {
			return err
		}
	}
	return nil
}

func addUnplannedPenaltyAlternateStops(
	input schema.Input,
	model nextroute.Model,
	unplannedPenaltyExpression nextroute.StopExpression,
) error {
	if input.AlternateStops == nil {
		return nil
	}

	data, err := getModelData(model)

	if err != nil {
		return err
	}

	for _, vehicle := range input.Vehicles {
		if vehicle.AlternateStops == nil {
			continue
		}

		for _, alternateID := range *vehicle.AlternateStops {
			stop, err := model.Stop(data.stopIDToIndex[alternateStopID(alternateID, vehicle)])
			if err != nil {
				return err
			}

			alternateInputStop := stop.Data().(alternateInputStop)

			if alternateInputStop.stop.UnplannedPenalty == nil {
				err = unplannedPenaltyExpression.SetValue(stop, defaultUnplannedPenaltyAlternateStop)
				if err != nil {
					return err
				}
				continue
			}

			err = unplannedPenaltyExpression.SetValue(stop, float64(*alternateInputStop.stop.UnplannedPenalty))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
