// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/nextmv-io/sdk/run"
)

// NewParallelSolver creates a new parallel solver.
func NewParallelSolver(
	model Model,
) (ParallelSolver, error) {
	parallelSolver, err := NewSkeletonParallelSolver(model)
	if err != nil {
		return nil, err
	}
	parallelSolver.SetSolverFactory(DefaultSolverFactory())
	parallelSolver.SetSolveOptionsFactory(DefaultSolveOptionsFactory())
	return &parallelSolverWrapperImpl{
		solver: parallelSolver,
	}, nil
}

// DefaultSolveOptionsFactory creates a new SolveOptionsFactory.
func DefaultSolveOptionsFactory() SolveOptionsFactory {
	return func(
		solveInformation ParallelSolveInformation,
	) (SolveOptions, error) {
		solveOptions := SolveOptions{
			Iterations: -1,
			Duration:   30 * time.Second,
		}
		solveOptions.Iterations = (1 + solveInformation.Random().Intn(10)) * 200
		return solveOptions, nil
	}
}

type parallelSolverWrapperImpl struct {
	solver ParallelSolver
}

func (p *parallelSolverWrapperImpl) ParallelSolveEvents() ParallelSolveEvents {
	return p.solver.ParallelSolveEvents()
}

func (p *parallelSolverWrapperImpl) Model() Model {
	return p.solver.Model()
}

func (p *parallelSolverWrapperImpl) SetSolverFactory(factory SolverFactory) {
	p.solver.SetSolverFactory(factory)
}

func (p *parallelSolverWrapperImpl) SetSolveOptionsFactory(factory SolveOptionsFactory) {
	p.solver.SetSolveOptionsFactory(factory)
}

func (p *parallelSolverWrapperImpl) SolveEvents() SolveEvents {
	return p.solver.SolveEvents()
}

func (p *parallelSolverWrapperImpl) Progression() []ProgressionEntry {
	return p.solver.Progression()
}

func (p *parallelSolverWrapperImpl) Solve(
	ctx context.Context,
	solveOptions ParallelSolveOptions,
	startSolutions ...Solution,
) (SolutionChannel, error) {
	start := ctx.Value(run.Start).(time.Time)
	ctx, _ = context.WithDeadline(
		ctx,
		start.Add(solveOptions.Duration),
	)

	interpretedParallelSolveOptions := ParallelSolveOptions{
		Iterations:           solveOptions.Iterations,
		Duration:             solveOptions.Duration,
		ParallelRuns:         solveOptions.ParallelRuns,
		StartSolutions:       solveOptions.StartSolutions,
		RunDeterministically: solveOptions.RunDeterministically,
	}

	if interpretedParallelSolveOptions.ParallelRuns == -1 {
		interpretedParallelSolveOptions.ParallelRuns = runtime.NumCPU()
	}

	if interpretedParallelSolveOptions.ParallelRuns > runtime.NumCPU() {
		interpretedParallelSolveOptions.ParallelRuns = runtime.NumCPU()
	}

	if interpretedParallelSolveOptions.Iterations == -1 {
		interpretedParallelSolveOptions.Iterations = math.MaxInt
	}

	if interpretedParallelSolveOptions.StartSolutions == -1 {
		interpretedParallelSolveOptions.StartSolutions = runtime.NumCPU()
	}

	initialSolutions := make(Solutions, interpretedParallelSolveOptions.StartSolutions)
	if interpretedParallelSolveOptions.StartSolutions > 0 {
		var wg sync.WaitGroup
		wg.Add(interpretedParallelSolveOptions.StartSolutions)
		solution, err := NewSolution(p.solver.Model())
		if err != nil {
			return nil, err
		}
		for idx := 0; idx < interpretedParallelSolveOptions.StartSolutions; idx++ {
			go func(idx int, sol Solution) {
				defer wg.Done()
				randomSolution, err := RandomSolutionConstruction(ctx, sol)
				if err != nil {
					panic(err)
				}
				initialSolutions[idx] = randomSolution
			}(idx, solution.Copy())
		}
		wg.Wait()
		startSolutions = append(startSolutions, initialSolutions...)
	}

	if len(startSolutions) == 0 {
		startSolution, err := NewSolution(p.solver.Model())
		if err != nil {
			return nil, err
		}
		startSolutions = append(startSolutions, startSolution)
	}

	return p.solver.Solve(
		ctx,
		interpretedParallelSolveOptions,
		startSolutions...,
	)
}
