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

	// Add a custom customConstraint.
	customConstraint := customConstraint{}
	if err := model.AddConstraint(customConstraint); err != nil {
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

// customConstraint is a struct that allows to implement a custom constraint.
type customConstraint struct{}

// EstimateIsViolated returns true if the constraint is violated. If the
// constraint is violated, the solver needs a hint to determine if further
// moves should be generated for the vehicle.
func (c customConstraint) EstimateIsViolated(
	move nextroute.Move,
) (isViolated bool, stopPositionsHint nextroute.StopPositionsHint) {
	// If there are no stops planned on the vehicle, the constraint is not
	// violated.
	if move.Vehicle().IsEmpty() {
		return false, nextroute.NoPositionsHint()
	}

	// Get the type of the first stop in the vehicle that is not the vehicle's
	// starting location.
	stop := move.Vehicle().First().Next().ModelStop()
	customData := stop.Data().(schema.Stop).CustomData.(map[string]any)
	customType := customData["type"].(string)

	// Loop over all the stops that are part of the move. If the type of a stop
	// is different from the type of the first stop, the constraint is
	// violated.
	for _, stop := range move.PlanStopsUnit().ModelPlanStopsUnit().Stops() {
		customData := stop.Data().(schema.Stop).CustomData.(map[string]any)
		if customData["type"].(string) != customType {
			return true, nextroute.SkipVehiclePositionsHint()
		}
	}

	// If the constraint is not violated, the solver does not need a hint.
	return false, nextroute.NoPositionsHint()
}

// String returns the name of the constraint.
func (c customConstraint) String() string {
	return "my_custom_constraint"
}

// IsTemporal returns true if the constraint should be checked after all initial
// stops have been planned. It returns false if the constraint should be checked
// after each of the initial stops has been planned.
func (c customConstraint) IsTemporal() bool {
	return false
}
