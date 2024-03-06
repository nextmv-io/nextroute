// © 2019-present nextmv.io inc

// package main holds the implementation of the nextroute template.
package main

import (
	"context"
	"log"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/check"
	"github.com/nextmv-io/nextroute/factory"
	"github.com/nextmv-io/nextroute/schema"
	"github.com/nextmv-io/sdk/run"
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
) (customOutput, error) {
	model, err := factory.NewModel(input, options.Model)
	if err != nil {
		return customOutput{}, err
	}

	solver, err := nextroute.NewParallelSolver(model)
	if err != nil {
		return customOutput{}, err
	}

	solutions, err := solver.Solve(ctx, options.Solve)
	if err != nil {
		return customOutput{}, err
	}
	last, err := solutions.Last()
	if err != nil {
		return customOutput{}, err
	}
	out := toOutput(last)

	return out, nil
}

type customOutput struct {
	Custom string  `json:"custom,omitempty"`
	Value  float64 `json:"value,omitempty"`
}

func toOutput(solution nextroute.Solution) customOutput {
	value := 0.0
	for _, t := range solution.Model().Objective().Terms() {
		value += solution.ObjectiveValue(t.Objective())
	}
	return customOutput{
		Custom: "hello world",
		Value:  value,
	}
}
