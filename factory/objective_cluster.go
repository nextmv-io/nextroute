// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addClusterObjective adds an objective which prefers clustered routes.
func addClusterObjective(
	_ schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	cluster, err := nextroute.NewCluster()
	if err != nil {
		return model, err
	}
	if _, err = model.Objective().NewTerm(options.Objectives.Cluster, cluster); err != nil {
		return nil, err
	}
	return model, nil
}
