// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addClusterConstraint adds a constraint which limits stops only to be added
// to the vehicle whose centroid is closest.
func addClusterConstraint(
	_ schema.Input,
	model nextroute.Model,
	_ Options,
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
