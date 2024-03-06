// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addActivationPenaltyObjective adds the initialization cost (per vehicle)
// objective to the Model.
func addActivationPenaltyObjective(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	activationPenalty := nextroute.NewVehicleTypeValueExpression("activation_penalty", 0.0)
	present := false
	for v, vehicle := range input.Vehicles {
		if vehicle.ActivationPenalty == nil || *vehicle.ActivationPenalty == 0 {
			continue
		}
		err := activationPenalty.SetValue(model.VehicleTypes()[v], float64(*vehicle.ActivationPenalty))
		if err != nil {
			return nil, err
		}
		present = true
	}

	if !present {
		return model, nil
	}

	_, err := model.
		Objective().
		NewTerm(options.Objectives.VehicleActivationPenalty, nextroute.NewVehiclesObjective(activationPenalty))
	if err != nil {
		return nil, err
	}

	return model, nil
}
