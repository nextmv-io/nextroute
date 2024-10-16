// © 2019-present nextmv.io inc

package main

import (
	"os"
	"testing"

	"github.com/nextmv-io/sdk/golden"
)

func TestMain(m *testing.M) {
	golden.Setup()
	code := m.Run()
	golden.Teardown()
	os.Exit(code)
}

// TestGolden executes a golden file test, where the .json input is fed and an
// output is expected.
func TestGolden(t *testing.T) {
	golden.FileTests(
		t,
		"input.json",
		golden.Config{
			Args: []string{
				"-solve.duration", "10s",
				"-format.disable.progression",
				"-solve.parallelruns", "1",
				"-solve.iterations", "10000",
				"-solve.rundeterministically",
				"-solve.startsolutions", "1",
				"-model.objectives.stopbalance", "1000.0",
			},
			TransientFields: []golden.TransientField{
				{Key: "$.version.sdk", Replacement: golden.StableVersion},
				{Key: "$.statistics.result.duration", Replacement: golden.StableFloat},
				{Key: "$.statistics.run.duration", Replacement: golden.StableFloat},
			},
			Thresholds: golden.Tresholds{
				Float: 0.01,
			},
		},
	)
}
