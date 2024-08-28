// © 2019-present nextmv.io inc

package main

import (
	"os"
	"testing"

	"github.com/nextmv-io/sdk/golden"
)

func TestMain(m *testing.M) {
	cleanUp()
	golden.CopyFile("../../cmd/main.go", "main.go")
	golden.Setup()
	code := m.Run()
	cleanUp()
	os.Exit(code)
}

// TestGolden executes a golden file test, where the .json input is fed and a
// .golden output is expected.
func TestGolden(t *testing.T) {
	golden.FileTests(
		t,
		"testdata",
		golden.Config{
			Args: []string{
				"-solve.duration", "10s",
				"-format.disable.progression",
				"-solve.parallelruns", "1",
				"-solve.iterations", "50",
				"-solve.rundeterministically",
				// for deterministic tests
				"-solve.startsolutions", "1",
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

func cleanUp() {
	keep := []string{
		"testdata",
		"main_test.go",
		"benchmark_test.go",
	}
	golden.Reset(keep)
}
