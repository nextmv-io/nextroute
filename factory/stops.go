// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	"github.com/nextmv-io/nextroute/schema"
)

// addStops adds the stops to the Model.
func addStops(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}
	for _, inputStop := range input.Stops {
		location, err := common.NewLocation(
			inputStop.Location.Lon,
			inputStop.Location.Lat,
		)
		if err != nil {
			return nil, err
		}

		stop, err := model.NewStop(location)
		if err != nil {
			return nil, err
		}

		stop.SetID(inputStop.ID)
		stop.SetData(inputStop)
		data.stopIDToIndex[inputStop.ID] = stop.Index()
	}
	model.SetData(data)

	return model, nil
}
