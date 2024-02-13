package factory

import (
	"github.com/nextmv-io/sdk/common"
	sdkNextRoute "github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/nextroute/factory"
	"github.com/nextmv-io/sdk/nextroute/schema"
)

// addStops adds the stops to the Model.
func addStops(
	input schema.Input,
	model sdkNextRoute.Model,
	_ factory.Options,
) (sdkNextRoute.Model, error) {
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
