// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addDistanceObjective adds the minimization of travel distance to the Model.
func addDistanceObjective(
	_ schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	o := nextroute.NewDistanceObjective()
	_, err := model.Objective().NewTerm(options.Objectives.Distance, o)
	if err != nil {
		return nil, err
	}
	return model, nil
}
