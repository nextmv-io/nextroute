// © 2019-present nextmv.io inc

package factory

import (
	"fmt"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/nextroute/schema"
)

// addAlternates adds the alternate stops to the Model.
func addAlternates(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	if input.AlternateStops == nil {
		return model, nil
	}

	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	alternateInputStops := make(map[string]alternateInputStop)
	for idx, alternate := range *input.AlternateStops {
		alternateInputStops[alternate.ID] = alternateInputStop{
			index: model.NumberOfStops() + idx,
			stop:  alternate,
		}
	}

	for _, vehicle := range input.Vehicles {
		if vehicle.AlternateStops == nil {
			continue
		}

		for _, alternateID := range *vehicle.AlternateStops {
			alternateInputStop, ok := alternateInputStops[alternateID]
			if !ok {
				return model, nmerror.NewInputDataError(fmt.Errorf("alternate stop %s on vehicle %s not found",
					alternateID,
					vehicle.ID,
				))
			}
			location, err := common.NewLocation(
				alternateInputStop.stop.Location.Lon,
				alternateInputStop.stop.Location.Lat,
			)
			if err != nil {
				return nil, err
			}

			stop, err := model.NewStop(location)

			if err != nil {
				return nil, err
			}

			stop.SetMeasureIndex(alternateInputStop.index)

			id := alternateStopID(alternateInputStop.stop.ID, vehicle)
			stop.SetID(alternateInputStop.stop.ID)
			stop.SetData(alternateInputStop)
			data.stopIDToIndex[id] = stop.Index()
		}
	}
	model.SetData(data)

	return model, nil
}

func alternateStopID(stopID string, vehicle schema.Vehicle) string {
	return fmt.Sprintf("alt_%s_%s_alt", vehicle.ID, stopID)
}

func alternateVehicleAttribute(idx int) string {
	return fmt.Sprintf("alt_%v_alt", idx)
}

type alternateInputStop struct {
	index int
	stop  schema.AlternateStop
}
