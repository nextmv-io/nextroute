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

type options struct {
	Model  factory.Options                `json:"model,omitempty"`
	Solve  nextroute.ParallelSolveOptions `json:"solve,omitempty"`
	Format nextroute.FormatOptions        `json:"format,omitempty"`
	Check  check.Options                  `json:"check,omitempty"`
}

func main() {
	runner := run.CLI(solver)
	err := runner.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
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
	solver, err := nextroute.NewSkeletonSolver(model)
	if err != nil {
		return runSchema.Output{}, err
	}
	solver.AddSolveOperators(
		NewCustomUnPlanSearchOperator(),
		NewCustomPlanSearchOperator(),
	)

	parallelSolver, err := nextroute.NewSkeletonParallelSolver(model)
	if err != nil {
		return runSchema.Output{}, err
	}
	parallelSolver.SetSolverFactory(
		func(
			_ nextroute.ParallelSolveInformation,
			_ nextroute.Solution,
		) (nextroute.Solver, error) {
			solver, err := nextroute.NewSkeletonSolver(model)
			if err != nil {
				return nil, err
			}
			solver.AddSolveOperators(
				NewCustomUnPlanSearchOperator(),
				NewCustomPlanSearchOperator(),
			)
			return solver, nil
		},
	)

	parallelSolver.SetSolveOptionsFactory(
		func(
			_ nextroute.ParallelSolveInformation,
		) (nextroute.SolveOptions, error) {
			return nextroute.SolveOptions{
				Iterations: 1000,
				Duration:   10 * time.Minute,
			}, nil
		},
	)

	solutions, err := parallelSolver.Solve(ctx, options.Solve)
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
	return output, nil
}

type customUnplanImpl struct {
	nextroute.SolveOperator
}

// NewCustomUnPlanSearchOperator creates a new custom unplan search operator.
func NewCustomUnPlanSearchOperator() nextroute.SolveOperator {
	return &customUnplanImpl{
		nextroute.NewSolveOperator(
			1.0,
			true,
			nextroute.SolveParameters{},
		),
	}
}

func (d *customUnplanImpl) Execute(
	_ context.Context,
	runTimeInformation nextroute.SolveInformation,
) error {
	workSolution := runTimeInformation.
		Solver().
		WorkSolution()

	if workSolution.PlannedPlanUnits().Size() == 0 {
		return nil
	}

	randomPlannedPlanUnit := workSolution.PlannedPlanUnits().RandomElement()

	_, err := randomPlannedPlanUnit.UnPlan()

	if err != nil {
		return err
	}

	return nil
}

type customPlanImpl struct {
	nextroute.SolveOperator
}

// NewCustomPlanSearchOperator creates a new custom plan search operator.
func NewCustomPlanSearchOperator() nextroute.SolveOperator {
	return &customPlanImpl{
		SolveOperator: nextroute.NewSolveOperator(
			1.0,
			true,
			nextroute.SolveParameters{},
		),
	}
}

func (d *customPlanImpl) Execute(
	ctx context.Context,
	runTimeInformation nextroute.SolveInformation,
) error {
	workSolution := runTimeInformation.
		Solver().
		WorkSolution()

	unPlannedPlannedPlanUnits := workSolution.
		UnPlannedPlanUnits().
		SolutionPlanUnits()

	unPlannedPlannedPlanUnits = common.Shuffle(
		workSolution.Random(),
		unPlannedPlannedPlanUnits,
	)

	for _, unPlannedPlannedPlanUnit := range unPlannedPlannedPlanUnits {
		select {
		case <-ctx.Done():
		default:
			bestMove := workSolution.BestMove(
				ctx,
				unPlannedPlannedPlanUnit,
			)
			if bestMove.IsExecutable() {
				_, err := bestMove.Execute(ctx)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
