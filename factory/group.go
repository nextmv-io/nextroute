// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addGroupInformation adds information to the Model data, when stops
// have to be grouped together on the same vehicle but not necessarily
// in a particular order.
func addGroupInformation(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	if input.StopGroups == nil || len(*input.StopGroups) == 0 {
		return model, nil
	}

	groups := make([]group, len(*input.StopGroups))

	for index, stopGroup := range *input.StopGroups {
		groups[index] = group{
			stops: map[string]struct{}{},
		}
		for _, stopID := range stopGroup {
			groups[index].stops[stopID] = struct{}{}
		}
	}

	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	data.groups = groups

	model.SetData(data)

	return model, nil
}
