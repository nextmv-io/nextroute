// © 2019-present nextmv.io inc

// package main holds the implementation of the nextroute template.
package main

import (
	"context"
	"log"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/check"
	"github.com/nextmv-io/nextroute/common"
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
	model, err := factory.NewModel(input, options.Model)
	if err != nil {
		return runSchema.Output{}, err
	}

	haversineExpression := nextroute.NewHaversineExpression()
	defaultDurationExpression := nextroute.NewTravelDurationExpression(
		haversineExpression,
		common.NewSpeed(10, common.MetersPerSecond),
	)

	slowDurationExpression := nextroute.NewScaledDurationExpression(defaultDurationExpression, 2.0)

	timeDependentExpression, err := nextroute.NewTimeDependentDurationExpression(
		model,
		defaultDurationExpression,
	)
	if err != nil {
		return runSchema.Output{}, err
	}

	s1 := time.Date(2023, 1, 1, 6, 30, 0, 0, time.UTC)
	e1 := time.Date(2023, 1, 1, 9, 30, 0, 0, time.UTC)

	err = timeDependentExpression.SetExpression(
		s1,
		e1,
		slowDurationExpression,
	)
	if err != nil {
		return runSchema.Output{}, err
	}

	s2 := time.Date(2023, 1, 1, 17, 30, 0, 0, time.UTC)
	e2 := time.Date(2023, 1, 1, 19, 30, 0, 0, time.UTC)

	err = timeDependentExpression.SetExpression(
		s2,
		e2,
		slowDurationExpression,
	)
	if err != nil {
		return runSchema.Output{}, err
	}

	for _, v := range model.Vehicles() {
		err := v.VehicleType().SetTravelDurationExpression(timeDependentExpression)
		if err != nil {
			return runSchema.Output{}, err
		}
	}

	solver, err := nextroute.NewParallelSolver(model)
	if err != nil {
		return runSchema.Output{}, err
	}

	solutions, err := solver.Solve(ctx, options.Solve)
	if err != nil {
		return runSchema.Output{}, err
	}
	last, err := solutions.Last()
	if err != nil {
		return runSchema.Output{}, err
	}

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
