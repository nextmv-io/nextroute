// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/nextmv-io/sdk/run"
)

// NewSolver creates a new nextroute solver using the given model and options.
func NewSolver(
	model Model,
	options SolverOptions,
) (Solver, error) {
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
	unplanWeights, err := ParseUnPlanWeights(options.UnplanWeights)
	if err != nil {
		return nil,
			fmt.Errorf("options.UnplanWeights: %w", err)
	}
	unplan, err := NewSolveOperatorUnPlan(numberOfUnits, unplanWeights)
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
	solver Solver
}

func (s *solverWrapperImpl) Solve(
	ctx context.Context,
	solveOptions SolveOptions,
	startSolutions ...Solution,
) (SolutionChannel, error) {
	start := ctx.Value(run.Start).(time.Time)
	ctx, _ = context.WithDeadline(
		ctx,
		start.Add(solveOptions.Duration),
	)
	interpretedSolveOptions := SolveOptions{
		Iterations: solveOptions.Iterations,
		Duration:   solveOptions.Duration,
	}
	if interpretedSolveOptions.Iterations == -1 {
		interpretedSolveOptions.Iterations = math.MaxInt
	}
	return s.solver.Solve(ctx, interpretedSolveOptions, startSolutions...)
}

func (s *solverWrapperImpl) Progression() []ProgressionEntry {
	return s.solver.Progression()
}

func (s *solverWrapperImpl) AddSolveOperators(operators ...SolveOperator) {
	s.solver.AddSolveOperators(operators...)
}

func (s *solverWrapperImpl) SolveEvents() SolveEvents {
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

func (s *solverWrapperImpl) BestSolution() Solution {
	return s.solver.BestSolution()
}

func (s *solverWrapperImpl) WorkSolution() Solution {
	return s.solver.WorkSolution()
}

func (s *solverWrapperImpl) Model() Model {
	return s.solver.Model()
}

func (s *solverWrapperImpl) Reset(
	solution Solution,
	solveInformation SolveInformation,
) {
	s.solver.Reset(solution, solveInformation)
}

func (s *solverWrapperImpl) SolveOperators() SolveOperators {
	return s.solver.SolveOperators()
}

// DefaultSolverFactory creates a new SolverFactory.
func DefaultSolverFactory() SolverFactory {
	return func(
		info ParallelSolveInformation,
		solution Solution,
	) (Solver, error) {
		nrPlanUnits := len(solution.Model().PlanUnits())
		parallelOptions := info.Options()

		// Handle unplan units parameter - negative values indicate default
		unplanUnitsStartValue := parallelOptions.Solver.UnplanUnits.StartValue
		if unplanUnitsStartValue < 0 {
			unplanUnitsStartValue = 2
		}
		unplanUnitsDeltaAfter := parallelOptions.Solver.UnplanUnits.DeltaAfterIterations
		if unplanUnitsDeltaAfter < 0 {
			unplanUnitsDeltaAfter = 125
		}
		unplanUnitsDelta := parallelOptions.Solver.UnplanUnits.Delta
		if unplanUnitsDelta < 0 {
			unplanUnitsDelta = 2
		}
		unplanUnitsMinValue := parallelOptions.Solver.UnplanUnits.MinValue
		if unplanUnitsMinValue < 0 {
			unplanUnitsMinValue = 2
		}
		unplanUnitsMaxValue := parallelOptions.Solver.UnplanUnits.MaxValue
		if unplanUnitsMaxValue < 0 {
			unplanUnitsMaxValue = int(math.Max(2.0, 0.05*float64(nrPlanUnits)))
		}

		// Handle plan group size parameter - negative values indicate default
		planGroupSizeStartValue := parallelOptions.Solver.PlanGroupSize.StartValue
		if planGroupSizeStartValue < 0 {
			planGroupSizeStartValue = 2
		}
		planGroupSizeDeltaAfter := parallelOptions.Solver.PlanGroupSize.DeltaAfterIterations
		if planGroupSizeDeltaAfter < 0 {
			planGroupSizeDeltaAfter = 1000000000
		}
		planGroupSizeDelta := parallelOptions.Solver.PlanGroupSize.Delta
		if planGroupSizeDelta < 0 {
			planGroupSizeDelta = 0
		}
		planGroupSizeMinValue := parallelOptions.Solver.PlanGroupSize.MinValue
		if planGroupSizeMinValue < 0 {
			planGroupSizeMinValue = 2
		}
		planGroupSizeMaxValue := parallelOptions.Solver.PlanGroupSize.MaxValue
		if planGroupSizeMaxValue < 0 {
			planGroupSizeMaxValue = 2
		}

		options := SolverOptions{
			Unplan: IntParameterOptions{
				StartValue:               unplanUnitsStartValue,
				DeltaAfterIterations:     unplanUnitsDeltaAfter,
				Delta:                    unplanUnitsDelta,
				MinValue:                 unplanUnitsMinValue,
				MaxValue:                 unplanUnitsMaxValue,
				SnapBackAfterImprovement: parallelOptions.Solver.UnplanUnits.SnapBackAfterImprovement,
				Zigzag:                   parallelOptions.Solver.UnplanUnits.Zigzag,
			},
			UnplanWeights: parallelOptions.Solver.UnplanWeights,
			Plan: IntParameterOptions{
				StartValue:               planGroupSizeStartValue,
				DeltaAfterIterations:     planGroupSizeDeltaAfter,
				Delta:                    planGroupSizeDelta,
				MinValue:                 planGroupSizeMinValue,
				MaxValue:                 planGroupSizeMaxValue,
				SnapBackAfterImprovement: parallelOptions.Solver.PlanGroupSize.SnapBackAfterImprovement,
				Zigzag:                   parallelOptions.Solver.PlanGroupSize.Zigzag,
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
		unplanWeights, err := ParseUnPlanWeights(options.UnplanWeights)
		if err != nil {
			return nil,
				fmt.Errorf("options.UnplanWeights: %w", err)
		}
		unplan, err := NewSolveOperatorUnPlan(numberOfUnits, unplanWeights)
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
