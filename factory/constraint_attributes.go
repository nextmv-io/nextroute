// © 2019-present nextmv.io inc

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

		err := constraint.SetStopAttributes(model.Stops()[s], *stop.CompatibilityAttributes)
		if err != nil {
			return nil, err
		}
		presentInStops = true
	}

	presentInVehicles := false
	for v, vehicle := range input.Vehicles {
		if vehicle.CompatibilityAttributes == nil {
			continue
		}

		err = constraint.SetVehicleTypeAttributes(model.VehicleTypes()[v], *vehicle.CompatibilityAttributes)
		if err != nil {
			return nil, err
		}
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
