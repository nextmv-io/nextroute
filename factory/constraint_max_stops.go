package factory

import (
	"math"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/sdk/nextroute/factory"
	"github.com/nextmv-io/sdk/nextroute/schema"
)

// addMaximumStopsConstraint adds a MaximumStopsConstraint to the model.
func addMaximumStopsConstraint(
	input schema.Input,
	model nextroute.Model,
	_ factory.Options,
) (nextroute.Model, error) {
	limit := nextroute.NewVehicleTypeValueExpression(
		"stopsLimit",
		math.MaxFloat64,
	)

	present := false
	for _, vehicleType := range model.VehicleTypes() {
		maxStops := input.Vehicles[vehicleType.Index()].MaxStops
		if maxStops == nil {
			continue
		}

		present = true

		limit.SetValue(vehicleType, float64(*maxStops))
	}

	if !present {
		return model, nil
	}

	// Create and then add constraint to model.
	maxConstraint, err := nextroute.NewMaximumStopsConstraint(limit)
	if err != nil {
		return nil, err
	}

	err = model.AddConstraint(maxConstraint)
	if err != nil {
		return nil, err
	}

	return model, nil
}
