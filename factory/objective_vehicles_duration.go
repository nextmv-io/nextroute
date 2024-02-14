package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/sdk/nextroute/factory"
	"github.com/nextmv-io/sdk/nextroute/schema"
)

// addVehiclesDurationObjective adds the minimization of the sum of vehicles
// duration to the model.
func addVehiclesDurationObjective(
	_ schema.Input,
	model nextroute.Model,
	options factory.Options,
) (nextroute.Model, error) {
	o := nextroute.NewVehiclesDurationObjective()
	_, err := model.Objective().NewTerm(options.Objectives.VehiclesDuration, o)
	if err != nil {
		return nil, err
	}

	return model, nil
}
