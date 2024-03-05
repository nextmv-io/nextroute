// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addMinStopsObjective adds the min stops per vehicle objective to the
// Model.
func addMinStopsObjective(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	minStops := nextroute.NewVehicleTypeValueExpression("min_stops", 0)
	minStopsPenalty := nextroute.NewVehicleTypeValueExpression("min_stops_penalty", 0)
	present := false
	for v, vehicle := range input.Vehicles {
		if vehicle.MinStops == nil || *vehicle.MinStops == 0 {
			continue
		}
		if vehicle.MinStopsPenalty == nil || *vehicle.MinStopsPenalty == 0.0 {
			continue
		}
		err := minStops.SetValue(model.VehicleTypes()[v], float64(*vehicle.MinStops))
		if err != nil {
			return nil, err
		}
		err = minStopsPenalty.SetValue(model.VehicleTypes()[v], *vehicle.MinStopsPenalty)
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
		NewTerm(
			options.Objectives.MinStops,
			nextroute.NewMinStopsObjective(minStops, minStopsPenalty),
		)
	if err != nil {
		return nil, err
	}

	return model, nil
}
