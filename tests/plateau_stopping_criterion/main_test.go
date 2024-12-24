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

// TestPlateauIterations executes a golden file test applying the plateau
// stopping criterion using a number of iterations.
func TestPlateauIterations(t *testing.T) {
	config := golden.Config{
		GoldenExtension: ".iterations.golden",
		Args: []string{
			"-solve.duration", "10s",
			"-format.disable.progression",
			"-solve.parallelruns", "1",
			"-solve.rundeterministically",
			"-solve.startsolutions", "1",
			"-solve.plateau.iterations", "20",
		},
		TransientFields: []golden.TransientField{
			{Key: "$.version.sdk", Replacement: golden.StableVersion},
			{Key: "$.statistics.result.duration", Replacement: golden.StableFloat},
			{Key: "$.statistics.run.duration", Replacement: golden.StableFloat},
			{Key: ".solutions[0].check.duration_used", Replacement: golden.StableFloat},
		},
		Thresholds: golden.Tresholds{
			Float: 0.01,
			CustomThresholds: golden.CustomThresholds{
				Int: map[string]int{
					"$.statistics.run.iterations": 0,
				},
			},
		},
	}
	golden.FileTest(t, "input.json", config)
}

// TestPlateauDuration executes a golden file test applying the plateau
// stopping criterion using a duration.
func TestPlateauDuration(t *testing.T) {
	config := golden.Config{
		GoldenExtension: ".duration.golden",
		Args: []string{
			"-solve.duration", "10s",
			"-format.disable.progression",
			"-solve.parallelruns", "1",
			"-solve.rundeterministically",
			"-solve.startsolutions", "1",
			"-solve.plateau.duration", "0.5s",
		},
		TransientFields: []golden.TransientField{
			{Key: "$.version.sdk", Replacement: golden.StableVersion},
			{Key: "$.statistics.result.duration", Replacement: golden.StableFloat},
			{Key: "$.statistics.run.iterations", Replacement: golden.StableInt},
			{Key: ".solutions[0].check.duration_used", Replacement: golden.StableFloat},
		},
		OutputProcessConfig: golden.OutputProcessConfig{
			RoundingConfig: []golden.RoundingConfig{
				{Key: "$.statistics.run.duration", Precision: 1},
			},
		},
		Thresholds: golden.Tresholds{
			Float: 0.01,
			CustomThresholds: golden.CustomThresholds{
				Float: map[string]float64{
					"$.statistics.run.duration": 0.2,
				},
			},
		},
	}
	golden.FileTest(t, "input.json", config)
}
