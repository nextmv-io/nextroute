package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/sdk/nextroute/factory"
	"github.com/nextmv-io/sdk/nextroute/schema"
)

// addClusterConstraint adds a constraint which limits stops only to be added
// to the vehicle whose centroid is closest.
func addClusterConstraint(
	_ schema.Input,
	model nextroute.Model,
	_ factory.Options,
) (nextroute.Model, error) {
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
