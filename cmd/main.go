// © 2019-present nextmv.io inc

// Package main allows you to run a nextroute solver from the command line.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/check"
	"github.com/nextmv-io/nextroute/factory"
	"github.com/nextmv-io/nextroute/schema"
	"github.com/nextmv-io/sdk/run"
	runSchema "github.com/nextmv-io/sdk/run/schema"
)

func main() {
	// If the only argument is 'version', print the version and exit.
	if len(os.Args) == 2 && strings.TrimLeft(os.Args[1], "-") == "version" {
		fmt.Println(nextroute.Version())
		return
	}
	// Continue with runner based execution.
	runner := run.CLI(solver)
	err := runner.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

type options struct {
	Model  factory.Options                `json:"model,omitempty"`
	Solve  nextroute.ParallelSolveOptions `json:"solve,omitempty"`
	Format nextroute.FormatOptions        `json:"format,omitempty"`
	Check  check.Options                  `json:"check,omitempty"`
}

type customOptions struct {
	MaxDuration *float64 `json:"max_duration,omitempty"`
}

// applyCustomOptions applies the extended custom options from the input to the
// actual options.
func applyCustomOptions(opts options, customOpts any) (options, error) {
	jOpts, err := json.Marshal(customOpts)
	if err != nil {
		return opts, err
	}
	var custom customOptions
	err = json.Unmarshal(jOpts, &custom)
	if err != nil {
		return opts, err
	}
	if custom.MaxDuration != nil {
		opts.Solve.Duration = time.Duration(*custom.MaxDuration * float64(time.Second))
	}
	return opts, nil
}

func solver(
	ctx context.Context,
	input schema.Input,
	options options,
) (runSchema.Output, error) {
	// Apply input embedded options, if any. This is used internally for
	// benchmarking and testing.
	if input.Options != nil {
		opts, err := applyCustomOptions(options, input.Options)
		if err != nil {
			return runSchema.Output{}, err
		}
		options = opts
	}

	// Create the model from the input and options.
	model, err := factory.NewModel(input, options.Model)
	if err != nil {
		return runSchema.Output{}, err
	}

	// Create the solver from the model.
	solver, err := nextroute.NewParallelSolver(model)
	if err != nil {
		return runSchema.Output{}, err
	}

	// Solve the model.
	solutions, err := solver.Solve(ctx, options.Solve)
	if err != nil {
		return runSchema.Output{}, err
	}

	// Get the last solution.
	// This call is blocking until the solver terminates. Alternatively,
	// solutions can be ranged over (see All() method).
	last, err := solutions.Last()
	if err != nil {
		return runSchema.Output{}, err
	}

	// Process the solution for output.
	output, err := check.Format(
		ctx,
		options,
		options.Check,
		solver,
		last,
	)
	if err != nil {
		return runSchema.Output{}, err
	}
	output.Statistics.Result.Custom = factory.DefaultCustomResultStatistics(last)

	return output, nil
}
