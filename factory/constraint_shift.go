// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addVehicleEndTimeConstraint uses the latestEndConstraint of the model. It checks if
// the vehicle's shift end happens before what is already set in the
// latestEndConstraint (the constraint is created if it does not exist). If the
// shift end time is at an earlier time than what is already set, then the
// value is changed.
func addVehicleEndTimeConstraint(
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
		if inputVehicle.EndTime == nil {
			continue
		}

		vehicle := model.Vehicles()[v]
		if inputVehicle.EndTime.Before(latestEndExpression.Time(vehicle.Last())) {
			latestEndExpression.SetTime(vehicle.Last(), *inputVehicle.EndTime)
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
