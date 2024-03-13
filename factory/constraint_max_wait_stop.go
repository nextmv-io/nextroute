// © 2019-present nextmv.io inc

package factory

import (
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	"github.com/nextmv-io/nextroute/schema"
)

// addMaximumWaitStopConstraint adds a MaximumWaitStopConstraint to the model.
func addMaximumWaitStopConstraint(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	maximumWaitPerStop := nextroute.NewStopDurationExpression("stop-wait-max", model.MaxDuration())

	maximumWaitPresent := addMaximumWaitStops(input, model, maximumWaitPerStop)

	alternateMaximumWaitPresent, err := addMaximumWaitAlternateStops(input, model, maximumWaitPerStop)
	if err != nil {
		return nil, err
	}

	if !maximumWaitPresent && !alternateMaximumWaitPresent {
		return model, nil
	}

	maxConstraint, err := nextroute.NewMaximumWaitStopConstraint(maximumWaitPerStop)
	if err != nil {
		return nil, err
	}

	err = model.AddConstraint(maxConstraint)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func addMaximumWaitStops(
	input schema.Input,
	model nextroute.Model,
	stopLimit nextroute.StopDurationExpression,
) bool {
	present := false
	modelStops := model.Stops()
	for s, inputStop := range input.Stops {
		if inputStop.MaxWait == nil {
			continue
		}
		present = true

		stopLimit.SetDuration(modelStops[s], time.Duration(*inputStop.MaxWait)*time.Second)
	}
	return present
}

func addMaximumWaitAlternateStops(
	input schema.Input,
	model nextroute.Model,
	stopLimit nextroute.StopDurationExpression,
) (bool, error) {
	if input.AlternateStops == nil {
		return false, nil
	}

	if common.AllTrue(
		*input.AlternateStops,
		func(stop schema.AlternateStop) bool {
			return stop.MaxWait == nil
		},
	) {
		return false, nil
	}

	data, err := getModelData(model)

	if err != nil {
		return false, err
	}

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

			if alternateInputStop.stop.MaxWait == nil {
				continue
			}

			stopLimit.SetDuration(stop, time.Duration(*alternateInputStop.stop.MaxWait)*time.Second)
		}
	}

	return true, nil
}
