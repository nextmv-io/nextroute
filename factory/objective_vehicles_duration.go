// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addVehiclesDurationObjective adds the minimization of the sum of vehicles
// duration to the model.
func addVehiclesDurationObjective(
	_ schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	o := nextroute.NewVehiclesDurationObjective()
	_, err := model.Objective().NewTerm(options.Objectives.VehiclesDuration, o)
	if err != nil {
		return nil, err
	}

	return model, nil
}
