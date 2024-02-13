package factory

import (
	"github.com/nextmv-io/nextroute"
	sdkNextRoute "github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/nextroute/factory"
	"github.com/nextmv-io/sdk/nextroute/schema"
)

// addClusterObjective adds an objective which prefers clustered routes.
func addClusterObjective(
	_ schema.Input,
	model sdkNextRoute.Model,
	options factory.Options,
) (sdkNextRoute.Model, error) {
	cluster, err := nextroute.NewCluster()
	if err != nil {
		return model, err
	}
	if _, err = model.Objective().NewTerm(options.Objectives.Cluster, cluster); err != nil {
		return nil, err
	}
	return model, nil
}
