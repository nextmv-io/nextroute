// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nextmv-io/sdk/run"
)

// Iterations is the key for the iterations performed.
const Iterations string = "iterations"

// ParallelSolveOptions holds the options for the parallel solver.
type ParallelSolveOptions struct {
	Iterations           int           `json:"iterations"  usage:"maximum number of iterations, -1 assumes no limit; iterations are counted after start solutions are generated" default:"-1"`
	Duration             time.Duration `json:"duration" usage:"maximum duration of the solver" default:"5s"`
	ParallelRuns         int           `json:"parallel_runs" usage:"maximum number of parallel runs, -1 results in using all available resources" default:"-1"`
	StartSolutions       int           `json:"start_solutions" usage:"number of solutions to generate on top of those passed in; one solution generated with sweep algorithm, the rest generated randomly" default:"-1"`
	RunDeterministically bool          `json:"run_deterministically"  usage:"run the parallel solver deterministically"`
}

// ParallelSolver is the interface for parallel solver. The parallel solver will
// run multiple solver in parallel and return the best solution. The parallel
// solver will stop when the maximum duration is reached.
type ParallelSolver interface {
	Progressioner

	// Model returns the model of the solver.
	Model() Model

	// SetSolverFactory sets the factory for creating new solver.
	SetSolverFactory(SolverFactory)
	// SetSolveOptionsFactory sets the factory for creating new solve options.
	SetSolveOptionsFactory(SolveOptionsFactory)
	// Solve starts the solving process using the given options. It returns the
	// solutions as a channel.
	Solve(context.Context, ParallelSolveOptions, ...Solution) (SolutionChannel, error)
	// SolveEvents returns the solve-events used by the individual solver instances.
	SolveEvents() SolveEvents
	// ParallelSolveEvents returns the solve-events used by the parallel solver.
	ParallelSolveEvents() ParallelSolveEvents
}

// SolveOptionsFactory is a factory type for creating new solve options.
// This factory is used by the parallel solver to create new solve options for
// a new run of a solver.
type SolveOptionsFactory func(
	information ParallelSolveInformation,
) (SolveOptions, error)

// SolverFactory is a factory type for creating new solver. This factory is
// used by the parallel solver to create new solver for a new run.
type SolverFactory func(
	information ParallelSolveInformation,
	solution Solution) (Solver, error)

// ParallelSolveInformation holds the information about the current parallel
// solve run.
type ParallelSolveInformation interface {
	// Cycle returns the current cycle. A cycle is a set of parallel runs.
	// In each cycle multiple runs are executed in parallel. Cycle identifies
	// how often a new solver has been created and started with the best
	// solution of the previous runs.
	Cycle() int

	// Random returns the random number generator from the solution.
	Random() *rand.Rand
	// Run returns the current run. A run is a single solve run. In each cycle
	// multiple runs are executed in parallel. Run identifies a run.
	Run() int
}

// The parallel solver will run multiple solver in parallel and return the best
// solution. The parallel solver will stop when the maximum duration is reached.
// The parallel solver will execute a single solver in a run for a given number
// of iterations and time before starting a new run defined by the
// DefaultSolveOptionsFactory. Every time a run is finished
// the best solution is returned and new runs are started. A new run will be
// started with the global best found solution and with a solver defined by
// SolverFactory.

// NewSkeletonParallelSolver creates a new parallel solver.
func NewSkeletonParallelSolver(model Model) (ParallelSolver, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}
	parallelSolver := &parallelSolverImpl{
		solveEvents:         NewSolveEvents(),
		parallelSolveEvents: NewParallelSolveEvents(),
		model:               model,
	}

	return parallelSolver, nil
}

// newParallelSolveInformation is a factory for creating new solve information.
func newParallelSolveInformation(cycle, run int, random *rand.Rand) ParallelSolveInformation {
	return metaSolveInformationImpl{
		cycle:  cycle,
		run:    run,
		random: random,
	}
}

type metaSolveInformationImpl struct {
	random *rand.Rand
	cycle  int
	run    int
}

func (s metaSolveInformationImpl) Cycle() int {
	return s.cycle
}

func (s metaSolveInformationImpl) Run() int {
	return s.run
}

func (s metaSolveInformationImpl) Random() *rand.Rand {
	return s.random
}

type parallelSolverImpl struct {
	model               Model
	progression         []ProgressionEntry
	solveEvents         SolveEvents
	parallelSolveEvents ParallelSolveEvents
	solveOptionsFactory SolveOptionsFactory
	solverFactory       SolverFactory
}

func (s *parallelSolverImpl) ParallelSolveEvents() ParallelSolveEvents {
	return s.parallelSolveEvents
}

func (s *parallelSolverImpl) SolveEvents() SolveEvents {
	return s.solveEvents
}

func (s *parallelSolverImpl) Model() Model {
	return s.model
}

func (s *parallelSolverImpl) Progression() []ProgressionEntry {
	return slices.Clone(s.progression)
}

type solutionContainer struct {
	Solution   Solution
	Error      error
	Iterations int
}

func (s *parallelSolverImpl) SetSolverFactory(
	solverFactory SolverFactory,
) {
	s.solverFactory = solverFactory
}

func (s *parallelSolverImpl) SetSolveOptionsFactory(
	solveOptionsFactory SolveOptionsFactory,
) {
	s.solveOptionsFactory = solveOptionsFactory
}

func (s *parallelSolverImpl) Solve(
	ctx context.Context,
	options ParallelSolveOptions,
	startSolutions ...Solution,
) (SolutionChannel, error) {
	// TODO: check options
	if s.solveOptionsFactory == nil {
		return nil,
			fmt.Errorf(
				"parallel solver, solve options factory cannot be nil," +
					" define an options factory with SetSolveOptionsFactory",
			)
	}
	if s.solverFactory == nil {
		return nil,
			fmt.Errorf(
				"parallel solver, solver factory cannot be nil," +
					" define a solver factory with SetSolverFactory",
			)
	}

	interpretedParallelSolveOptions := ParallelSolveOptions{
		Iterations:           options.Iterations,
		Duration:             options.Duration,
		ParallelRuns:         options.ParallelRuns,
		StartSolutions:       options.StartSolutions,
		RunDeterministically: options.RunDeterministically,
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

	if len(startSolutions) == 0 {
		solution, err := NewSolution(s.model)
		if err != nil {
			return nil, err
		}
		startSolutions = append(startSolutions, solution)
	}

	for idx, solution := range startSolutions {
		if solution == nil {
			return nil, fmt.Errorf("start solution cannot be nil")
		}
		if solution.Model() != s.model {
			return nil, fmt.Errorf("start solution at index %v it's model does not match solver model", idx)
		}
	}

	start := time.Now()

	if ctx.Value(run.Start) != nil {
		start = ctx.Value(run.Start).(time.Time)
	}

	ctx, cancel := context.WithDeadline(
		ctx,
		start.Add(interpretedParallelSolveOptions.Duration),
	)

	solutions := make([]Solution, len(startSolutions))
	copy(solutions, startSolutions)

	parallelRuns := interpretedParallelSolveOptions.ParallelRuns

	if parallelRuns < 1 {
		parallelRuns = runtime.NumCPU()
	}

	s.ParallelSolveEvents().Start.Trigger(
		s,
		options,
		parallelRuns,
	)

	bestSolution := solutions[0]

	for _, solution := range solutions {
		if solution.Score() < bestSolution.Score() {
			bestSolution = solution
		}
	}

	bestSolution = bestSolution.Copy()

	parallelCount := make(chan struct{}, parallelRuns)

	syncResultChannel := make(chan solutionContainer)
	resultChannel := make(chan SolutionInfo, 1)

	totalIterations := atomic.Int64{}

	reportBestSolution := func(solutionContainer solutionContainer) {
		resultChannel <- SolutionInfo{
			Solution: solutionContainer.Solution,
			Error:    solutionContainer.Error,
		}
		if solutionContainer.Solution != nil {
			s.progression = append(s.progression, ProgressionEntry{
				ElapsedSeconds: time.Since(start).Seconds(),
				Value:          solutionContainer.Solution.Score(),
				Iterations:     solutionContainer.Iterations,
			})
		}
	}

	reportBestSolution(solutionContainer{
		Solution:   bestSolution,
		Error:      nil,
		Iterations: 0,
	})

	var solutionsMutex sync.Mutex
	iterationsLeft := atomic.Int64{}
	iterationsLeft.Store(int64(interpretedParallelSolveOptions.Iterations))
	var waitGroup sync.WaitGroup

	go func() {
		defer close(syncResultChannel)
		runCount := 0
	Loop:
		for {
			for i := 0; i < parallelRuns; i++ {
				select {
				case <-ctx.Done():
					waitGroup.Wait()
					break Loop
				default:
					parallelCount <- struct{}{}
					waitGroup.Add(1)
					runCount++
					go func(r int) {
						defer func() {
							<-parallelCount
							waitGroup.Done()
						}()

						solution := bestSolution.Copy()

						if len(solutions) > 0 {
							solutionsMutex.Lock()
							if len(solutions) > 0 {
								solution = solutions[len(solutions)-1]
								solutions = solutions[:len(solutions)-1]
							}
							solutionsMutex.Unlock()
						}

						cycle := (r-1)/parallelRuns + 1

						metaSolveInformation := newParallelSolveInformation(
							cycle,
							r,
							solution.Random(),
						)

						solver, err := s.solverFactory(
							metaSolveInformation,
							solution,
						)
						if err != nil {
							panic(err)
						}

						s.RegisterEvents(solver.SolveEvents())

						solver.SolveEvents().Iterated.Register(func(_ SolveInformation) {
							if totalIterations.Add(1) >= int64(interpretedParallelSolveOptions.Iterations) {
								cancel()
							}
						})

						opt, err := s.solveOptionsFactory(
							metaSolveInformation,
						)
						if err != nil {
							panic(err)
						}

						updatedIterations := iterationsLeft.Add(int64(opt.Iterations) * -1)
						if updatedIterations+int64(opt.Iterations) <= 0 {
							<-ctx.Done()
							return
						}
						if updatedIterations < 0 {
							opt.Iterations = int(updatedIterations + int64(opt.Iterations))
						}

						s.ParallelSolveEvents().StartSolver.Trigger(
							metaSolveInformation,
							solver,
							opt,
							solution,
						)

						solutionChannel, err := solver.Solve(
							ctx,
							opt,
							solution,
						)
						if err != nil {
							panic(err)
						}
						for sol := range solutionChannel {
							if sol.Solution != nil {
								s.ParallelSolveEvents().NewSolution.Trigger(
									metaSolveInformation,
									sol.Solution,
								)
							}

							syncResultChannel <- solutionContainer{
								Solution:   sol,
								Error:      sol.Error,
								Iterations: int(totalIterations.Load()),
							}
						}
					}(runCount)
				}
			}
			if interpretedParallelSolveOptions.RunDeterministically {
				waitGroup.Wait()
			}
		}
	}()

	go func() {
		defer func() {
			iterations := int(totalIterations.Load())
			if dataMap, ok := ctx.Value(run.Data).(*sync.Map); ok {
				converted := iterations
				dataMap.Store(Iterations, converted)
			}
			close(resultChannel)

			s.ParallelSolveEvents().End.Trigger(s, iterations, bestSolution)
		}()
		for solverResult := range syncResultChannel {
			if solverResult.Error != nil {
				reportBestSolution(solutionContainer{
					Solution:   nil,
					Error:      solverResult.Error,
					Iterations: solverResult.Iterations,
				})
				cancel()
				continue
			}

			if solverResult.Solution.Score() >= bestSolution.Score() {
				continue
			}

			bestSolution = solverResult.Solution.Copy()

			reportBestSolution(solutionContainer{
				Solution:   solverResult.Solution.Copy(),
				Error:      solverResult.Error,
				Iterations: solverResult.Iterations,
			})
		}
	}()

	return resultChannel, nil
}

func (s *parallelSolverImpl) RegisterEvents(
	events SolveEvents,
) {
	events.ContextDone.Register(func(info SolveInformation) {
		s.solveEvents.ContextDone.Trigger(info)
	})
	events.Iterated.Register(func(info SolveInformation) {
		s.solveEvents.Iterated.Trigger(info)
	})
	events.Iterating.Register(func(info SolveInformation) {
		s.solveEvents.Iterating.Trigger(info)
	})
	events.OperatorExecuted.Register(func(info SolveInformation) {
		s.solveEvents.OperatorExecuted.Trigger(info)
	})
	events.OperatorExecuting.Register(func(info SolveInformation) {
		s.solveEvents.OperatorExecuting.Trigger(info)
	})
	events.NewBestSolution.Register(func(info SolveInformation) {
		s.solveEvents.NewBestSolution.Trigger(info)
	})
	events.Start.Register(func(info SolveInformation) {
		s.solveEvents.Start.Trigger(info)
	})
	events.Reset.Register(func(solution Solution, info SolveInformation) {
		s.solveEvents.Reset.Trigger(solution, info)
	})
	events.Done.Register(func(info SolveInformation) {
		s.solveEvents.Done.Trigger(info)
	})
}
