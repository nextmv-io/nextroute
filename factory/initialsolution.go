// © 2019-present nextmv.io inc

package factory

import (
	"fmt"

	"github.com/nextmv-io/nextroute"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/nextroute/schema"
)

// addInitialSolution sets the initial solution.
func addInitialSolution(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	inputStopIDToModelStopIndex := map[string]int{}

	for idx, inputStop := range input.Stops {
		inputStopIDToModelStopIndex[inputStop.ID] = idx
	}

	modelStops := model.Stops()

	for idx, inputVehicle := range input.Vehicles {
		if inputVehicle.InitialStops == nil {
			continue
		}
		modelVehicle := model.Vehicles()[idx]

		for _, initialStop := range *inputVehicle.InitialStops {
			var modelStop nextroute.ModelStop

			if _, defined := inputStopIDToModelStopIndex[initialStop.ID]; defined {
				modelStop = modelStops[inputStopIDToModelStopIndex[initialStop.ID]]
			} else {
				modelStop, err = model.Stop(data.stopIDToIndex[alternateStopID(initialStop.ID, inputVehicle)])
				if err != nil {
					return nil, err
				}
			}

			if modelStop == nil {
				return nil, nmerror.NewInputDataError(fmt.Errorf("initial stop `%s` on vehicle `%s` not found, "+
					"stop must be defined in stops or alternate stops to be used as an initial stop",
					initialStop.ID,
					modelVehicle.ID(),
				))
			}
			fixed := initialStop.Fixed != nil && *initialStop.Fixed

			err := modelVehicle.AddStop(modelStop, fixed)

			if err != nil {
				return nil, err
			}
		}
	}

	return model, nil
}
