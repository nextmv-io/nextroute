package nextroute

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/nextmv-io/sdk/alns"
	"github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/run"
)

// NewSolver creates a new nextroute solver using the given model and options.
func NewSolver(
	model nextroute.Model,
	options nextroute.SolverOptions,
) (nextroute.Solver, error) {
	solver, err := NewSkeletonSolver(model)
	if err != nil {
		return nil, err
	}
	numberOfUnits, err := NewSolveParameter(
		options.Unplan.StartValue,
		options.Unplan.DeltaAfterIterations,
		options.Unplan.Delta,
		options.Unplan.MinValue,
		options.Unplan.MaxValue,
		options.Unplan.SnapBackAfterImprovement,
		options.Unplan.Zigzag,
	)
	if err != nil {
		return nil,
			fmt.Errorf("options.Unplan: %w", err)
	}
	unplan, err := NewSolveOperatorUnPlan(numberOfUnits)
	if err != nil {
		return nil, err
	}
	groupSize, err := NewSolveParameter(
		options.Plan.StartValue,
		options.Plan.DeltaAfterIterations,
		options.Plan.Delta,
		options.Plan.MinValue,
		options.Plan.MaxValue,
		options.Plan.SnapBackAfterImprovement,
		options.Plan.Zigzag,
	)
	if err != nil {
		return nil,
			fmt.Errorf("options.Plan: %w", err)
	}
	plan, err := NewSolveOperatorPlan(groupSize)
	if err != nil {
		return nil, err
	}
	maximumIterations, err := NewSolveParameter(
		options.Restart.StartValue,
		options.Restart.DeltaAfterIterations,
		options.Restart.Delta,
		options.Restart.MinValue,
		options.Restart.MaxValue,
		options.Restart.SnapBackAfterImprovement,
		options.Restart.Zigzag,
	)
	if err != nil {
		return nil,
			fmt.Errorf("options.Restart: %w", err)
	}
	restart, err := NewSolveOperatorRestart(maximumIterations)
	if err != nil {
		return nil, err
	}
	solver.AddSolveOperators(
		unplan,
		plan,
		restart,
	)
	solverWrapper := solverWrapperImpl{
		solver: solver,
	}
	return &solverWrapper, err
}

type solverWrapperImpl struct {
	solver nextroute.Solver
}

func (s *solverWrapperImpl) Solve(
	ctx context.Context,
	solveOptions nextroute.SolveOptions,
	startSolutions ...nextroute.Solution,
) (nextroute.SolutionChannel, error) {
	start := ctx.Value(run.Start).(time.Time)
	ctx, _ = context.WithDeadline(
		ctx,
		start.Add(solveOptions.Duration),
	)
	interpretedSolveOptions := nextroute.SolveOptions{
		Iterations: solveOptions.Iterations,
		Duration:   solveOptions.Duration,
	}
	if interpretedSolveOptions.Iterations == -1 {
		interpretedSolveOptions.Iterations = math.MaxInt
	}
	return s.solver.Solve(ctx, interpretedSolveOptions, startSolutions...)
}

func (s *solverWrapperImpl) Progression() []alns.ProgressionEntry {
	return s.solver.Progression()
}

func (s *solverWrapperImpl) AddSolveOperators(operators ...nextroute.SolveOperator) {
	s.solver.AddSolveOperators(operators...)
}

func (s *solverWrapperImpl) SolveEvents() nextroute.SolveEvents {
	return s.solver.SolveEvents()
}

func (s *solverWrapperImpl) Random() *rand.Rand {
	return s.solver.Random()
}

func (s *solverWrapperImpl) HasBestSolution() bool {
	return s.solver.HasBestSolution()
}

func (s *solverWrapperImpl) HasWorkSolution() bool {
	return s.solver.HasWorkSolution()
}

func (s *solverWrapperImpl) BestSolution() nextroute.Solution {
	return s.solver.BestSolution()
}

func (s *solverWrapperImpl) WorkSolution() nextroute.Solution {
	return s.solver.WorkSolution()
}

func (s *solverWrapperImpl) Model() nextroute.Model {
	return s.solver.Model()
}

func (s *solverWrapperImpl) Reset(
	solution nextroute.Solution,
	solveInformation nextroute.SolveInformation,
) {
	s.solver.Reset(solution, solveInformation)
}

func (s *solverWrapperImpl) SolveOperators() nextroute.SolveOperators {
	return s.solver.SolveOperators()
}

// DefaultSolverFactory creates a new SolverFactory.
func DefaultSolverFactory() nextroute.SolverFactory {
	return func(
		_ nextroute.ParallelSolveInformation,
		solution nextroute.Solution,
	) (nextroute.Solver, error) {
		nrPlanUnits := len(solution.Model().PlanUnits())

		unplanCount := 2
		maxUnplanCount := int(math.Max(2.0, 0.05*float64(nrPlanUnits)))

		options := nextroute.SolverOptions{
			Unplan: nextroute.IntParameterOptions{
				StartValue:               unplanCount,
				DeltaAfterIterations:     125,
				Delta:                    unplanCount,
				MinValue:                 unplanCount,
				MaxValue:                 maxUnplanCount,
				SnapBackAfterImprovement: true,
				Zigzag:                   true,
			},
			Plan: nextroute.IntParameterOptions{
				StartValue:               2,
				DeltaAfterIterations:     1000000000,
				Delta:                    0,
				MinValue:                 2,
				MaxValue:                 2,
				SnapBackAfterImprovement: true,
				Zigzag:                   true,
			},
		}

		solver, err := NewSkeletonSolver(solution.Model())
		if err != nil {
			return nil, err
		}
		numberOfUnits, err := NewSolveParameter(
			options.Unplan.StartValue,
			options.Unplan.DeltaAfterIterations,
			options.Unplan.Delta,
			options.Unplan.MinValue,
			options.Unplan.MaxValue,
			options.Unplan.SnapBackAfterImprovement,
			options.Unplan.Zigzag,
		)
		if err != nil {
			return nil,
				fmt.Errorf("options.Unplan: %w", err)
		}
		unplan, err := NewSolveOperatorUnPlan(numberOfUnits)
		if err != nil {
			return nil, err
		}
		groupSize, err := NewSolveParameter(
			options.Plan.StartValue,
			options.Plan.DeltaAfterIterations,
			options.Plan.Delta,
			options.Plan.MinValue,
			options.Plan.MaxValue,
			options.Plan.SnapBackAfterImprovement,
			options.Plan.Zigzag,
		)
		if err != nil {
			return nil,
				fmt.Errorf("options.Plan: %w", err)
		}
		plan, err := NewSolveOperatorPlan(groupSize)
		if err != nil {
			return nil, err
		}
		solver.AddSolveOperators(
			unplan,
			plan,
		)
		return solver, nil
	}
}
