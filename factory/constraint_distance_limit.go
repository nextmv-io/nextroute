// © 2019-present nextmv.io inc

package factory

import (
	"fmt"
	"math"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	"github.com/nextmv-io/nextroute/schema"
)

// addDistanceLimitConstraint adds a distance limit for routes to the model.
func addDistanceLimitConstraint(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	composed := nextroute.NewComposedPerVehicleTypeExpression(
		nextroute.NewConstantExpression(
			"constant-route-distance",
			0,
		),
	)

	limit := nextroute.NewVehicleTypeDistanceExpression(
		"distanceLimit",
		common.NewDistance(math.MaxFloat64, common.Meters),
	)
	hasDistanceLimit := false
	for _, vehicleType := range model.VehicleTypes() {
		maxDistance := input.Vehicles[vehicleType.Index()].MaxDistance
		if maxDistance == nil {
			continue
		}

		hasDistanceLimit = true

		// Check if custom data is set properly.
		data, ok := vehicleType.Data().(vehicleTypeData)
		if !ok {
			return nil, fmt.Errorf(
				fmt.Sprintf("could not read custom data for vehicle %s",
					vehicleType.ID(),
				),
			)
		}

		// Get distance expression and set limit for the vehicle type.
		distanceExpression := data.DistanceExpression
		composed.Set(vehicleType, distanceExpression)
		err := limit.SetDistance(vehicleType, common.NewDistance(float64(*maxDistance), common.Meters))
		if err != nil {
			return nil, err
		}
	}

	if !hasDistanceLimit {
		return model, nil
	}

	// Create and then add constraint to model.
	maxConstraint, err := nextroute.NewMaximum(
		composed,
		limit,
	)
	if err != nil {
		return nil, err
	}
	maxConstraint.(nextroute.Identifier).SetID("distance_limit")

	err = model.AddConstraint(maxConstraint)
	if err != nil {
		return nil, err
	}

	return model, nil
}
