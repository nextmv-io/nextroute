package factory

import (
	"github.com/nextmv-io/nextroute"
	sdkNextRoute "github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/nextroute/factory"
	"github.com/nextmv-io/sdk/nextroute/schema"
)

// addClusterConstraint adds a constraint which limits stops only to be added
// to the vehicle whose centroid is closest.
func addClusterConstraint(
	_ schema.Input,
	model sdkNextRoute.Model,
	_ factory.Options,
) (sdkNextRoute.Model, error) {
	cluster, err := nextroute.NewCluster()
	if err != nil {
		return model, err
	}
	err = model.AddConstraint(cluster)
	if err != nil {
		return model, err
	}
	return model, nil
}
