// Package main allows you to run a nextroute solver from the command line
// without the need of compiling plugins.
package main

import (
	"context"
	"log"

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
	Check  check.Options
	Model  factory.Options
	Solve  nextroute.ParallelSolveOptions
	Format nextroute.FormatOptions
}

func solver(ctx context.Context,
	input schema.Input,
	options options,
) (runSchema.Output, error) {
	model, err := factory.NewModel(input, options.Model)
	if err != nil {
		return runSchema.Output{}, err
	}

	parallelSolver, err := nextroute.NewParallelSolver(model)
	if err != nil {
		return runSchema.Output{}, err
	}

	solutions, err := parallelSolver.Solve(ctx, options.Solve)
	if err != nil {
		return runSchema.Output{}, err
	}

	return check.Format(
		ctx,
		options,
		options.Check,
		parallelSolver,
		solutions.Last(),
	)
}
