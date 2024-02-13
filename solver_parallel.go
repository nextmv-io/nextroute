package nextroute

import (
	"context"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/nextmv-io/sdk/alns"
	"github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/run"
)

// NewParallelSolver creates a new parallel solver.
func NewParallelSolver(
	model nextroute.Model,
) (nextroute.ParallelSolver, error) {
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
func DefaultSolveOptionsFactory() nextroute.SolveOptionsFactory {
	return func(
		solveInformation nextroute.ParallelSolveInformation,
	) (nextroute.SolveOptions, error) {
		solveOptions := nextroute.SolveOptions{
			Iterations: -1,
			Duration:   30 * time.Second,
		}
		solveOptions.Iterations = (1 + solveInformation.Random().Intn(10)) * 200
		return solveOptions, nil
	}
}

type parallelSolverWrapperImpl struct {
	solver nextroute.ParallelSolver
}

func (p *parallelSolverWrapperImpl) Model() nextroute.Model {
	return p.solver.Model()
}

func (p *parallelSolverWrapperImpl) SetSolverFactory(factory nextroute.SolverFactory) {
	p.solver.SetSolverFactory(factory)
}

func (p *parallelSolverWrapperImpl) SetSolveOptionsFactory(factory nextroute.SolveOptionsFactory) {
	p.solver.SetSolveOptionsFactory(factory)
}

func (p *parallelSolverWrapperImpl) SolveEvents() nextroute.SolveEvents {
	return p.solver.SolveEvents()
}

func (p *parallelSolverWrapperImpl) Progression() []alns.ProgressionEntry {
	return p.solver.Progression()
}

func (p *parallelSolverWrapperImpl) Solve(
	ctx context.Context,
	solveOptions nextroute.ParallelSolveOptions,
	startSolutions ...nextroute.Solution,
) (nextroute.SolutionChannel, error) {
	start := ctx.Value(run.Start).(time.Time)
	ctx, _ = context.WithDeadline(
		ctx,
		start.Add(solveOptions.Duration),
	)

	interpretedParallelSolveOptions := nextroute.ParallelSolveOptions{
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

	initialSolutions := make(nextroute.Solutions, interpretedParallelSolveOptions.StartSolutions)
	if interpretedParallelSolveOptions.StartSolutions > 0 {
		var wg sync.WaitGroup
		wg.Add(interpretedParallelSolveOptions.StartSolutions)
		solution, err := NewSolution(p.solver.Model())
		if err != nil {
			return nil, err
		}
		for idx := 0; idx < interpretedParallelSolveOptions.StartSolutions; idx++ {
			go func(idx int, sol nextroute.Solution) {
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
