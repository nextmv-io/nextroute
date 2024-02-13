package factory

import (
	"github.com/nextmv-io/nextroute"
	sdkNextRoute "github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/nextroute/factory"
	"github.com/nextmv-io/sdk/nextroute/schema"
)

// addTravelDurationObjective adds the minimization of travel duration to the Model.
func addTravelDurationObjective(
	_ schema.Input,
	model sdkNextRoute.Model,
	options factory.Options,
) (sdkNextRoute.Model, error) {
	o := nextroute.NewTravelDurationObjective()
	_, err := model.Objective().NewTerm(options.Objectives.TravelDuration, o)
	if err != nil {
		return nil, err
	}

	return model, nil
}
