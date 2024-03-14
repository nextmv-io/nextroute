// package main holds the implementation of the nextroute template.
package main

import (
	"context"
	"log"

	"github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/nextroute/check"
	"github.com/nextmv-io/sdk/nextroute/factory"
	"github.com/nextmv-io/sdk/nextroute/schema"
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

	// solver, err := nextroute.NewParallelSolver(model)
	solver, err := createParallelSolver(model)
	if err != nil {
		return runSchema.Output{}, err
	}

	solutions, err := solver.Solve(ctx, options.Solve)
	if err != nil {
		return runSchema.Output{}, err
	}
	last := solutions.Last()

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

type customUnplanImpl struct {
	nextroute.SearchOperator
}

func NewCustomUnPlanSearchOperator() nextroute.SearchOperator {
	return &customUnplanImpl{
		SolveOperator: NewSolveOperator(
			1.0,
			false,
			SolveParameters{},
		),
	}
}

func (d *customUnplanImpl) Execute(
	ctx context.Context,
	runTimeInformation SolveInformation,
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
	nextroute.SearchOperator
}

func NewCustomPlanSearchOperator() nextroute.SearchOperator {
	return &customPlanImpl{
		SolveOperator: NewSolveOperator(
			1.0,
			false,
			SolveParameters{},
		),
	}
}

func (d *customPlanImpl) Execute(
	ctx context.Context,
	runTimeInformation SolveInformation,
) error {
	workSolution := runTimeInformation.
		Solver().
		WorkSolution()

	randomPlannedPlanUnit := workSolution.PlannedPlanUnits().RandomElement()

	bestMove := workSolution.BestMove(ctx, randomPlannedPlanUnit)

	if bestMove.IsExecutable() {
		_, err := bestMove.Execute(ctx)
		if err != nil {
			return err
		}
	}
	return nil
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

	// solver, err := nextroute.NewParallelSolver(model)
	solver, err := createParallelSolver(model)
	if err != nil {
		return runSchema.Output{}, err
	}

	solutions, err := solver.Solve(ctx, options.Solve)
	if err != nil {
		return runSchema.Output{}, err
	}
	last := solutions.Last()

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

func createParallelSolver(
	model nextroute.Model,
) (nextroute.ParallelSolver, error) {
	// Create a new parallel solver.
	parallelSolver, err := nextroute.NewSkeletonParallelSolver(
		model,
	)
	if err != nil {
		return nil, err
	}

	// Set the solver factory for the parallel solver. This factory is used to
	// create new solver instances for each cycle. The information contains data
	// about the current cycle and which solver of the n solvers is being
	// created. The solution is the best solution of the previous cycle (and
	// globally best).
	//
	// In this example we create identical solvers with custom operators, but
	// you can also create different solvers with different operators. There
	// is a random component in the operators, so the solvers will behave
	// differently.
	parallelSolver.SetSolverFactory(
		func(
			information nextroute.ParallelSolveInformation,
			solution nextroute.Solution,
		) (nextroute.Solver, error) {
			return createSolver(model)
		},
	)

	// The solve options factory is used to create new solve options for each
	// solver invocation in a cycle. The information contains data about the
	// current cycle and the solution of the previous cycle.
	parallelSolver.SetSolveOptionsFactory(
		nextroute.DefaultSolveOptionsFactory(),
	)

	return parallelSolver, nil
}

func createSolver(
	modelTable ModelTable,
) (nextroute.Solver, error) {
	// Create a new solver.
	solver, err := nextroute.NewSkeletonSolver(
		modelTable.Model(),
	)
	if err != nil {
		return nil, err
	}

	// Create and add the custom operators to the solver.
	unplanOperator, err := NewCustomUnPlanSearchOperator(
		modelTable,
		2,
	)
	if err != nil {
		return nil, err
	}

	planOperator, err := NewCustomPlanSearchOperator(
		modelTable,
	)
	if err != nil {
		return nil, err
	}

	solver.AddSolveOperators(
		unplanOperator,
		planOperator,
	)
	return solver, nil
}
