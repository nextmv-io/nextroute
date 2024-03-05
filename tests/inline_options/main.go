// © 2019-present nextmv.io inc
// package main holds the implementation of the nextroute template.
package main

import (
	"context"
	"log"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/check"
	"github.com/nextmv-io/nextroute/factory"
	"github.com/nextmv-io/nextroute/schema"
	"github.com/nextmv-io/sdk/run"
	runSchema "github.com/nextmv-io/sdk/run/schema"
)

func main() {
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

func solver(
	ctx context.Context,
	input schema.Input,
	options options,
) (runSchema.Output, error) {
	// Customize options in the code.
	options.Model.Objectives.TravelDuration = 0.5
	options.Model.Objectives.UnplannedPenalty = 0.3
	options.Model.Constraints.Disable.Capacity = true
	options.Model.Constraints.Disable.Attributes = true
	options.Model.Constraints.Disable.Precedence = true
	options.Solve.Duration = time.Duration(11) * time.Second
	options.Solve.Iterations = 51
	options.Format.Disable.Progression = true

	model, err := factory.NewModel(input, options.Model)
	if err != nil {
		return runSchema.Output{}, err
	}

	solver, err := nextroute.NewParallelSolver(model)
	if err != nil {
		return runSchema.Output{}, err
	}

	solutions, err := solver.Solve(ctx, options.Solve)
	if err != nil {
		return runSchema.Output{}, err
	}
	last := solutions.Last()

	output, err := check.Format(ctx, options, options.Check, solver, last)
	if err != nil {
		return runSchema.Output{}, err
	}
	output.Statistics.Result.Custom = factory.DefaultCustomResultStatistics(last)

	return output, nil
}
