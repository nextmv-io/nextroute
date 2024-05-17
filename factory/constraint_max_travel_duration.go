// © 2019-present nextmv.io inc

package factory

import (
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addMaximumTravelDurationConstraint
// TODO
func addMaximumTravelDurationConstraint(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	maximumWaitPerVehicle := nextroute.NewVehicleTypeDurationExpression("vehicle-travel-max", model.MaxDuration())
	maximumTravelPresent := addMaximumTravelDurationVehicles(input, model, maximumWaitPerVehicle)

	if maximumTravelPresent {
		cnstr, err := nextroute.NewMaximumDurationConstraint(maximumWaitPerVehicle)
		if err != nil {
			return nil, err
		}
		err = model.AddConstraint(cnstr)
		if err != nil {
			return nil, err
		}
	}

	return model, nil
}

func addMaximumTravelDurationVehicles(
	input schema.Input,
	model nextroute.Model,
	vehicleLimit nextroute.VehicleTypeDurationExpression,
) bool {
	present := false
	modelVehicles := model.Vehicles()
	for v, inputVehicle := range input.Vehicles {
		if inputVehicle.MaxTravelDuration == nil {
			continue
		}
		present = true

		vehicleLimit.SetDuration(modelVehicles[v].VehicleType(), time.Duration(*inputVehicle.MaxTravelDuration)*time.Second)
	}
	return present
}
