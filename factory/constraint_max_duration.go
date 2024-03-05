// © 2019-present nextmv.io inc

package factory

import (
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addMaximumDurationConstraint uses the latestEndConstraint of the model. It
// checks if, when adding the maximum duration to the vehicle's start time, the
// end time happens before what is already set in the latestEndConstraint (the
// constraint is created if it does not exist). If this end time is at an
// earlier time than what is already set, then the value is changed.
func addMaximumDurationConstraint(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	latestEndExpression, model, err := latestEndExpression(model)
	if err != nil {
		return nil, err
	}

	present := false
	for v, inputVehicle := range input.Vehicles {
		if inputVehicle.MaxDuration == nil {
			continue
		}

		vehicle := model.Vehicles()[v]
		end := vehicle.Start().Add(time.Duration(*inputVehicle.MaxDuration) * time.Second)
		if end.Before(latestEndExpression.Time(vehicle.Last())) {
			latestEndExpression.SetTime(vehicle.Last(), end)
		}

		present = true
	}

	if !present {
		return model, nil
	}

	model, err = addLatestEndConstraint(model, latestEndExpression)
	if err != nil {
		return nil, err
	}

	return model, nil
}
