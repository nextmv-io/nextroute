package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addAttributesConstraint adds the attributes constraint to the model.
func addAttributesConstraint(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	constraint, err := nextroute.NewAttributesConstraint()
	if err != nil {
		return nil, err
	}

	presentInStops := false
	for s, stop := range input.Stops {
		if stop.CompatibilityAttributes == nil {
			continue
		}

		constraint.SetStopAttributes(model.Stops()[s], *stop.CompatibilityAttributes)
		presentInStops = true
	}

	presentInVehicles := false
	for v, vehicle := range input.Vehicles {
		if vehicle.CompatibilityAttributes == nil {
			continue
		}

		constraint.SetVehicleTypeAttributes(model.VehicleTypes()[v], *vehicle.CompatibilityAttributes)
		presentInVehicles = true
	}

	if !presentInStops && !presentInVehicles {
		return model, nil
	}

	err = model.AddConstraint(constraint)
	if err != nil {
		return nil, err
	}

	return model, nil
}
