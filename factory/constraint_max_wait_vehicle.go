package factory

import (
	"time"

	"github.com/nextmv-io/nextroute"
	sdkNextRoute "github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/nextroute/factory"
	"github.com/nextmv-io/sdk/nextroute/schema"
)

// addMaximumWaitVehicleConstraint adds a MaximumWaitVehicleConstraint to the
// model.
func addMaximumWaitVehicleConstraint(
	input schema.Input,
	model sdkNextRoute.Model,
	_ factory.Options,
) (sdkNextRoute.Model, error) {
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
