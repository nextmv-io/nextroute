package nextroute

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nextmv-io/sdk/alns"
	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/run"
)

// Iterations is the key for the iterations performed.
const Iterations string = "iterations"

// The parallel solver will run multiple solver in parallel and return the best
// solution. The parallel solver will stop when the maximum duration is reached.
// The parallel solver will execute a single solver in a run for a given number
// of iterations and time before starting a new run defined by the
// DefaultSolveOptionsFactory. Everytime a run is finished
// the best solution is returned and new runs are started. A new run will be
// started with the global best found solution and with a solver defined by
// SolverFactory.

// NewSkeletonParallelSolver creates a new parallel solver.
func NewSkeletonParallelSolver(model nextroute.Model) (nextroute.ParallelSolver, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}
	parallelSolver := &parallelSolverImpl{
		parallelSolverObservedImpl: parallelSolverObservedImpl{
			observers: make([]ParallelSolverObserver, 0),
		},
		solveEvents: nextroute.NewSolveEvents(),
		model:       model,
	}

	return parallelSolver, nil
}

// newParallelSolveInformation is a factory for creating new solve information.
func newParallelSolveInformation(cycle, run int, random *rand.Rand) nextroute.ParallelSolveInformation {
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

// ParallelSolverObserver is the interface for observing the parallel solver.
type ParallelSolverObserver interface {
	// OnStart is called when the parallel solver is started.
	OnStart(
		solver nextroute.ParallelSolver,
		options nextroute.ParallelSolveOptions,
		parallelRuns int,
	)
	// OnNewRun is called when a new run is started.
	OnNewRun(
		solver nextroute.ParallelSolver,
	)
	// OnNewSolution is called when a new solution is found.
	OnNewSolution(
		solver nextroute.ParallelSolver,
		solution nextroute.Solution,
	)
}

type parallelSolverObservedImpl struct {
	observers []ParallelSolverObserver
}

func (o *parallelSolverObservedImpl) AddMetaSearchObserver(
	observer ParallelSolverObserver,
) {
	o.observers = append(o.observers, observer)
}

func (o *parallelSolverObservedImpl) OnStart(
	solver nextroute.ParallelSolver,
	options nextroute.ParallelSolveOptions,
	parallelRuns int,
) {
	for _, observer := range o.observers {
		observer.OnStart(solver, options, parallelRuns)
	}
}

func (o *parallelSolverObservedImpl) OnNewRun(
	solver nextroute.ParallelSolver,
) {
	for _, observer := range o.observers {
		observer.OnNewRun(solver)
	}
}

func (o *parallelSolverObservedImpl) OnNewSolution(
	solver nextroute.ParallelSolver,
	solution nextroute.Solution,
) {
	for _, observer := range o.observers {
		observer.OnNewSolution(solver, solution)
	}
}

type parallelSolverImpl struct {
	parallelSolverObservedImpl
	model               nextroute.Model
	progression         []alns.ProgressionEntry
	solveEvents         nextroute.SolveEvents
	solveOptionsFactory nextroute.SolveOptionsFactory
	solverFactory       nextroute.SolverFactory
}

func (s *parallelSolverImpl) Model() nextroute.Model {
	return s.model
}

func (s *parallelSolverImpl) Progression() []alns.ProgressionEntry {
	return common.DefensiveCopy(s.progression)
}

type solutionContainer struct {
	solution   nextroute.Solution
	iterations int
}

func (s *parallelSolverImpl) SetSolverFactory(
	solverFactory nextroute.SolverFactory,
) {
	s.solverFactory = solverFactory
}

func (s *parallelSolverImpl) SetSolveOptionsFactory(
	solveOptionsFactory nextroute.SolveOptionsFactory,
) {
	s.solveOptionsFactory = solveOptionsFactory
}

func (s *parallelSolverImpl) Solve(
	ctx context.Context,
	options nextroute.ParallelSolveOptions,
	startSolutions ...nextroute.Solution,
) (nextroute.SolutionChannel, error) {
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

	interpretedParallelSolveOptions := nextroute.ParallelSolveOptions{
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

	start := ctx.Value(run.Start).(time.Time)

	ctx, cancel := context.WithDeadline(
		ctx,
		start.Add(interpretedParallelSolveOptions.Duration),
	)

	solutions := make([]nextroute.Solution, len(startSolutions))
	copy(solutions, startSolutions)

	parallelRuns := interpretedParallelSolveOptions.ParallelRuns

	if parallelRuns < 1 {
		parallelRuns = runtime.NumCPU()
	}

	s.OnStart(s, options, parallelRuns)

	bestSolution := solutions[0]

	for _, solution := range solutions {
		if solution.Score() < bestSolution.Score() {
			bestSolution = solution
		}
	}

	bestSolution = bestSolution.Copy()

	parallelCount := make(chan struct{}, parallelRuns)

	syncResultChannel := make(chan solutionContainer)
	resultChannel := make(chan nextroute.Solution, 1)

	totalIterations := atomic.Int64{}

	reportBestSolution := func(solutionContainer solutionContainer) {
		resultChannel <- solutionContainer.solution
		s.progression = append(s.progression, alns.ProgressionEntry{
			ElapsedSeconds: time.Since(start).Seconds(),
			Value:          solutionContainer.solution.Score(),
			Iterations:     solutionContainer.iterations,
		})
	}

	reportBestSolution(solutionContainer{
		solution:   bestSolution,
		iterations: 0,
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

						solver.SolveEvents().Iterated.Register(func(_ nextroute.SolveInformation) {
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
						solutionChannel, err := solver.Solve(
							ctx,
							opt,
							solution,
						)
						if err != nil {
							panic(err)
						}
						for sol := range solutionChannel {
							syncResultChannel <- solutionContainer{
								solution:   sol,
								iterations: int(totalIterations.Load()),
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
			if dataMap, ok := ctx.Value(run.Data).(*sync.Map); ok {
				converted := int(totalIterations.Load())
				dataMap.Store(Iterations, converted)
			}
			close(resultChannel)
		}()
		for solverResult := range syncResultChannel {
			if solverResult.solution.Score() >= bestSolution.Score() {
				continue
			}

			bestSolution = solverResult.solution.Copy()

			reportBestSolution(solutionContainer{
				solution:   solverResult.solution.Copy(),
				iterations: solverResult.iterations,
			})
		}
	}()

	return resultChannel, nil
}

func (s *parallelSolverImpl) SolveEvents() nextroute.SolveEvents {
	return s.solveEvents
}

func (s *parallelSolverImpl) RegisterEvents(
	events nextroute.SolveEvents,
) {
	events.ContextDone.Register(func(info nextroute.SolveInformation) {
		s.solveEvents.ContextDone.Trigger(info)
	})
	events.Iterated.Register(func(info nextroute.SolveInformation) {
		s.solveEvents.Iterated.Trigger(info)
	})
	events.Iterating.Register(func(info nextroute.SolveInformation) {
		s.solveEvents.Iterating.Trigger(info)
	})
	events.OperatorExecuted.Register(func(info nextroute.SolveInformation) {
		s.solveEvents.OperatorExecuted.Trigger(info)
	})
	events.OperatorExecuting.Register(func(info nextroute.SolveInformation) {
		s.solveEvents.OperatorExecuting.Trigger(info)
	})
	events.NewBestSolution.Register(func(info nextroute.SolveInformation) {
		s.solveEvents.NewBestSolution.Trigger(info)
	})
	events.Start.Register(func(info nextroute.SolveInformation) {
		s.solveEvents.Start.Trigger(info)
	})
	events.Reset.Register(func(solution nextroute.Solution, info nextroute.SolveInformation) {
		s.solveEvents.Reset.Trigger(solution, info)
	})
	events.Done.Register(func(info nextroute.SolveInformation) {
		s.solveEvents.Done.Trigger(info)
	})
}
