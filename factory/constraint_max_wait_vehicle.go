// © 2019-present nextmv.io inc

package factory

import (
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addMaximumWaitVehicleConstraint adds a MaximumWaitVehicleConstraint to the
// model.
func addMaximumWaitVehicleConstraint(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	vehicleLimit := nextroute.NewVehicleTypeDurationExpression("vehicle-wait-max", model.MaxDuration())

	present := false

	// Add all maximum cumulative wait times of the vehicles.
	for _, vehicleType := range model.VehicleTypes() {
		maxWait := input.Vehicles[vehicleType.Index()].MaxWait
		if maxWait == nil {
			continue
		}
		present = true

		vehicleLimit.SetDuration(vehicleType, time.Duration(*maxWait)*time.Second)
	}

	if !present {
		return model, nil
	}

	// Create and then add constraint to model.
	maxConstraint, err := nextroute.NewMaximumWaitVehicleConstraint(vehicleLimit)
	if err != nil {
		return nil, err
	}

	err = model.AddConstraint(maxConstraint)
	if err != nil {
		return nil, err
	}

	return model, nil
}
