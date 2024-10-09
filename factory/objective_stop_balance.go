// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addStopBalanceObjective adds the stop balance objective to the model.
func addStopBalanceObjective(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	balance := nextroute.NewStopBalanceObjective()
	if _, err := model.Objective().NewTerm(options.Objectives.StopBalance, balance); err != nil {
		return nil, err
	}
	return model, nil
}
