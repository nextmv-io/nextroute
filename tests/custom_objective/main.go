// © 2019-present nextmv.io inc

// package main holds the implementation of the nextroute template.
package main

import (
	"context"
	"errors"
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

	// Check if there is custom data coming from the input.
	if input.CustomData != nil {
		// Unmarshal custom data that is part of the input into the custom
		// objective struct that we define below.
		customObjective, err := schema.ConvertCustomData[customObjective](input.CustomData)
		if err != nil {
			return runSchema.Output{}, errors.New("input.custom_data must be of type map[string]any")
		}

		// Add a custom objective. The term used is 1.0 for simplicity but it
		// can also be specified in the input.
		if _, err := model.Objective().NewTerm(1.0, &customObjective); err != nil {
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

	output, err := check.Format(ctx, options, options.Check, solver, last)
	if err != nil {
		return runSchema.Output{}, err
	}
	output.Statistics.Result.Custom = factory.DefaultCustomResultStatistics(last)

	return output, nil
}

// customObjective is a struct that allows to pass in custom data from the
// input and implement a new custom objective.
type customObjective struct {
	// BalancePenalty is the penalty that is multiplied by the difference
	// between the maximum number of stops and the minimum number of stops.
	BalancePenalty float64 `json:"balance_penalty,omitempty"`
}

// EstimateDeltaValue estimates the cost of planning the stops into a solution
// as defined in the move. The cost may imply an improvement in the objective's
// score.
func (c *customObjective) EstimateDeltaValue(
	move nextroute.Move,
) float64 {
	solution := move.Solution()
	// Calculate the maximum and minimum number of stops for the current
	// solution.
	maxNumStops := 0
	minNumStops := solution.Model().NumberOfStops()
	secondMinNumStops := solution.Model().NumberOfStops()

	// Loop over the vehicles of the solution to set the baseline of what are
	// the current maximum and minimum number of stops across all of them.
	for _, vehicle := range solution.Vehicles() {
		stops := vehicle.NumberOfStops()
		if stops > maxNumStops {
			maxNumStops = stops
		}

		if stops < minNumStops {
			minNumStops = stops
		}

		// In case the vehicle of the move is the one holding the minimum
		// number of stops, adding stops to it will make it no longer the
		// vehicle with the minimum number of stops. In this case, we need to
		// find the second minimum number of stops.
		if vehicle.Index() != move.Vehicle().Index() {
			if stops < secondMinNumStops {
				secondMinNumStops = stops
			}
		}
	}

	// If planning stops on the move's vehicle makes it have the maximum number
	// of stops, the delta can be calculated as the difference between the new
	// number of stops and the previous maximum number of stops.
	newNumStops := move.Vehicle().NumberOfStops() + len(move.StopPositions())
	if newNumStops > maxNumStops {
		return c.BalancePenalty * float64(newNumStops-maxNumStops)
	}

	// If the move's vehicle has the minimum number of stops, the delta can be
	// calculated as the difference between the new number of stops and the
	// previous minimum number of stops.
	if newNumStops > secondMinNumStops {
		minNumStops = secondMinNumStops
		return c.BalancePenalty * float64(secondMinNumStops-minNumStops)
	}

	// If the maximum and minimum number of stops do not change, there is no
	// delta.
	return 0.0
}

// Value returns the value of the custom objective's score, calculated for a
// given solution.
func (c *customObjective) Value(solution nextroute.Solution) float64 {
	maxNumStops := 0
	minNumStops := solution.Model().NumberOfStops()

	// Loop over the vehicles of the solution to calculate the maximum and
	// minimum number of stops across all of them.
	for _, vehicle := range solution.Vehicles() {
		stops := vehicle.NumberOfStops()
		if stops > maxNumStops {
			maxNumStops = stops
		}
		if stops < minNumStops {
			minNumStops = stops
		}
	}

	// This is the value of the solution for the given objective.
	return c.BalancePenalty * float64(maxNumStops-minNumStops)
}

// String implements the fmt.Stringer interface, allowing you to print the name
// of the objective in the output in a human-readable way.
func (c *customObjective) String() string {
	return "balance_penalty"
}
