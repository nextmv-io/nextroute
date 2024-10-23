// © 2019-present nextmv.io inc

package main

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/nextmv-io/sdk/golden"
)

const pythonFile = "main.py"

var pythonFileDestination = path.Join("..", "..", pythonFile)

func TestMain(m *testing.M) {
	// Move the python file to the `src` so that the import path in that file
	// is resolved.
	input, err := os.ReadFile(pythonFile)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(pythonFileDestination, input, 0644)
	if err != nil {
		panic(err)
	}

	// Compile the Go binary that is needed for this test.
	cmd := exec.Command(
		"go", "build",
		"-o", path.Join("..", "..", "nextroute", "bin", "nextroute.exe"),
		path.Join("..", "..", "..", "cmd", "main.go"),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	// Run the tests.
	code := m.Run()

	// Clean up the python file.
	err = os.Remove(pythonFileDestination)
	if err != nil {
		panic(err)
	}

	os.Exit(code)
}

func TestPythonSolveGolden(t *testing.T) {
	// These golden file tests are based on the original Go golden file tests.
	// It uses the `./tests/golden` directory (relative to the root of the
	// project) as a data source. It executes a Python script that uses the
	// Nextmv Python SDK to load options and read/write JSON files.
	golden.FileTests(
		t,
		path.Join("..", "..", "..", "tests", "golden", "testdata"),
		golden.Config{
			Args: []string{
				"-solve_duration", "10",
				// for deterministic tests
				"-format_disable_progression", "true",
				"-solve_parallelruns", "1",
				"-solve_iterations", "50",
				"-solve_rundeterministically", "true",
				"-solve_startsolutions", "1",
			},
			TransientFields: []golden.TransientField{
				{Key: "$.statistics.result.duration", Replacement: golden.StableFloat},
				{Key: "$.statistics.run.duration", Replacement: golden.StableFloat},
				{Key: "$.statistics.result.value", Replacement: golden.StableFloat},
				{Key: "$.options.nextmv.output", Replacement: "output.json"},
				{Key: "$.options.nextmv.input", Replacement: "input.json"},
				{Key: "$.statistics.result.custom.max_travel_duration", Replacement: golden.StableFloat},
				{Key: "$.statistics.result.custom.min_travel_duration", Replacement: golden.StableFloat},
				{Key: "$.statistics.result.custom.max_duration", Replacement: golden.StableFloat},
				{Key: "$.statistics.result.custom.min_duration", Replacement: golden.StableFloat},
			},
			Thresholds: golden.Tresholds{
				Float: 0.01,
			},
			ExecutionConfig: &golden.ExecutionConfig{
				Command:    "python3",
				Args:       []string{pythonFileDestination},
				InputFlag:  "-input",
				OutputFlag: "-output",
			},
		},
	)
}
