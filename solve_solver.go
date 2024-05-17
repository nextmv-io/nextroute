// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/nextmv-io/nextroute/common"
)

// IntParameterOptions are the options for an integer parameter.
type IntParameterOptions struct {
	StartValue               int  `json:"start_value"  usage:"start value"`
	DeltaAfterIterations     int  `json:"delta_after_iterations"  usage:"delta after each iterations"`
	Delta                    int  `json:"delta"  usage:"delta"`
	MinValue                 int  `json:"min_value"  usage:"min value of parameter"`
	MaxValue                 int  `json:"max_value"  usage:"max value of parameter"`
	SnapBackAfterImprovement bool `json:"snap_back_after_improvement"  usage:"snap back to start value after improvement of best solution"`
	Zigzag                   bool `json:"zigzag"  usage:"zigzag between min and max value lik a jig saw"`
}

// SolverOptions are the options for the solver and it's operators.
type SolverOptions struct {
	Unplan  IntParameterOptions `json:"unplan"  usage:"unplan parameter"`
	Plan    IntParameterOptions `json:"plan"  usage:"plan parameter"`
	Restart IntParameterOptions `json:"restart"  usage:"restart parameter"`
}

// SolveOptions holds the options for the solve process.
type SolveOptions struct {
	Iterations int           `json:"iterations"  usage:"maximum number of iterations, -1 assumes no limit" default:"-1"`
	Duration   time.Duration `json:"duration"  usage:"maximum duration of solver in seconds" default:"30s"`
}

// Solver is the interface for the Adaptive Local Neighborhood Search algorithm
// (ALNS) solver.
type Solver interface {
	Progressioner
	// AddSolveOperators adds a number of solve-operators to the solver.
	AddSolveOperators(...SolveOperator)

	// BestSolution returns the best solution found so far.
	BestSolution() Solution

	// HasBestSolution returns true if the solver has a best solution.
	HasBestSolution() bool
	// HasWorkSolution returns true if the solver has a work solution.
	HasWorkSolution() bool

	// Model returns the model used by the solver.
	Model() Model

	// Random returns the random number generator used by the solver.
	Random() *rand.Rand
	// Reset will reset the solver to use solution as work solution.
	Reset(solution Solution, solveInformation SolveInformation)

	// Solve starts the solving process using the given options. It returns the
	// solutions as a channel.
	Solve(context.Context, SolveOptions, ...Solution) (SolutionChannel, error)
	// SolveEvents returns the solve-events used by the solver.
	SolveEvents() SolveEvents
	// SolveOperators returns the solve-operators used by the solver.
	SolveOperators() SolveOperators

	// WorkSolution returns the current work solution.
	WorkSolution() Solution
}

// NewSkeletonSolver creates a new solver for the given model.
func NewSkeletonSolver(model Model) (Solver, error) {
	solver := &solveImpl{
		solveEvents: NewSolveEvents(),
		model:       model,
	}
	solver.SolveEvents().NewBestSolution.Register(solver.OnImprovement)
	return solver, nil
}

type solveImpl struct {
	model          Model
	solveEvents    SolveEvents
	workSolution   Solution
	bestSolution   Solution
	random         *rand.Rand
	solveOperators SolveOperators
	parameters     SolveParameters
	progression    []ProgressionEntry
}

func (s *solveImpl) OnImprovement(solveInformation SolveInformation) {
	s.progression = append(s.progression, ProgressionEntry{
		ElapsedSeconds: time.Since(solveInformation.Start()).Seconds(),
		Value:          solveInformation.Solver().BestSolution().Score(),
	})
}

func (s *solveImpl) Progression() []ProgressionEntry {
	return slices.Clone(s.progression)
}

func (s *solveImpl) Model() Model {
	return s.model
}

func (s *solveImpl) AddSolveOperators(solveOperators ...SolveOperator) {
	for _, solveOperator := range solveOperators {
		s.solveOperators = append(s.solveOperators, solveOperator)
		for _, parameter := range solveOperator.Parameters() {
			s.register(parameter)
		}
		for _, st := range s.SolveOperators() {
			if interested, ok := st.(InterestedInBetterSolution); ok {
				s.SolveEvents().NewBestSolution.Register(interested.OnBetterSolution)
			}
			if interested, ok := solveOperator.(InterestedInStartSolve); ok {
				s.SolveEvents().Start.Register(interested.OnStartSolve)
			}
		}
	}
}

func (s *solveImpl) Random() *rand.Rand {
	return s.random
}

func (s *solveImpl) SolveEvents() SolveEvents {
	return s.solveEvents
}

func (s *solveImpl) register(parameter SolveParameter) {
	s.parameters = append(s.parameters, parameter)
}

func (s *solveImpl) SolveOperators() SolveOperators {
	solveOperators := make(SolveOperators, len(s.solveOperators))
	copy(solveOperators, s.solveOperators)
	return solveOperators
}

func (s *solveImpl) Reset(solution Solution, solveInformation SolveInformation) {
	s.solveEvents.Reset.Trigger(solution, solveInformation)
	s.workSolution = solution.Copy()
	if s.workSolution.Score() < s.bestSolution.Score() {
		solveInfoImpl := solveInformation.(*solveInformationImpl)
		solveInfoImpl.deltaScore = s.workSolution.Score() - s.bestSolution.Score()
		s.newBestSolution(s.workSolution, solveInfoImpl)
	}
}

func (s *solveImpl) HasBestSolution() bool {
	return s.bestSolution != nil
}

func (s *solveImpl) HasWorkSolution() bool {
	return s.workSolution != nil
}

func (s *solveImpl) BestSolution() Solution {
	return s.bestSolution
}

func (s *solveImpl) WorkSolution() Solution {
	return s.workSolution
}

// invoke returns true if the best solution was improved.
func (s *solveImpl) invoke(
	ctx context.Context,
	solveOperator SolveOperator,
	solveInformation *solveInformationImpl,
) (bool, error) {
	// Check if the solve-operator should be executed.
	if s.Random().Float64() > solveOperator.Probability() {
		return false, nil
	}

	solveInformation.solveOperators = append(
		solveInformation.solveOperators,
		solveOperator,
	)

	s.solveEvents.OperatorExecuting.Trigger(solveInformation)
	err := solveOperator.Execute(ctx, solveInformation)
	if err != nil {
		return false,
			fmt.Errorf("%T: %w", solveOperator, err)
	}
	s.solveEvents.OperatorExecuted.Trigger(solveInformation)

	if !solveOperator.CanResultInImprovement() {
		return false, nil
	}

	delta := s.workSolution.Score() - s.bestSolution.Score()
	if delta >= 0.0 {
		return false, nil
	}

	solveInformation.deltaScore += delta

	s.newBestSolution(s.workSolution, solveInformation)

	return true, nil
}

func (s *solveImpl) newBestSolution(solution Solution, solveInformation *solveInformationImpl) {
	s.bestSolution = solution.Copy()
	s.solveEvents.NewBestSolution.Trigger(solveInformation)
}

func (s *solveImpl) Solve(
	ctx context.Context,
	solveOptions SolveOptions,
	startSolutions ...Solution,
) (SolutionChannel, error) {
	if len(startSolutions) == 0 {
		startSolution, err := NewSolution(s.model)
		if err != nil {
			return nil, err
		}
		startSolutions = append(startSolutions, startSolution)
	}
	if len(startSolutions) > 1 {
		return nil, fmt.Errorf("only one start solution allowed")
	}
	if len(s.solveOperators) == 0 {
		return nil, fmt.Errorf("solver is empty, no solve operators provided")
	}

	if !common.HasTrue(s.solveOperators, func(operator SolveOperator) bool {
		return operator.CanResultInImprovement()
	}) {
		return nil, fmt.Errorf(
			"no solve operator can result in improvement," +
				" enable CanResultInImprovement on at least one solve operator",
		)
	}

	newWorkSolution := startSolutions[0].Copy()
	s.bestSolution = startSolutions[0].Copy()
	s.workSolution = newWorkSolution
	s.random = rand.New(rand.NewSource(newWorkSolution.Random().Int63()))

	start := time.Now()

	ctx, cancel := context.WithDeadline(
		ctx,
		start.Add(solveOptions.Duration),
	)

	solveInformation := &solveInformationImpl{
		iteration:      0,
		solver:         s,
		solveOperators: make(SolveOperators, 0, len(s.solveOperators)),
		start:          start,
	}

	s.solveEvents.Start.Trigger(solveInformation)

	var err error

	// hard coding a size of 100 here. If we ever need to, we can make it
	// configurable.
	solutions := make(chan SolutionInfo, 100)
	solutions <- SolutionInfo{
		Solution: s.bestSolution,
		Error:    nil,
	}
	go func() {
		defer func() {
			close(solutions)
			cancel()
		}()

	Loop:
		for iteration := 0; iteration < solveOptions.Iterations; iteration++ {
			solveInformation.iteration = iteration
			solveInformation.deltaScore = 0.0
			// we do not clear the elements of solveOperators as they are
			// stable across iterations. We do not risk a memory leak here.
			solveInformation.solveOperators = solveInformation.solveOperators[:0]
			s.solveEvents.Iterating.Trigger(solveInformation)
			for _, solveOperator := range s.solveOperators {
				select {
				case <-ctx.Done():
					s.solveEvents.ContextDone.Trigger(solveInformation)
					break Loop
				default:
					improved, e := s.invoke(ctx, solveOperator, solveInformation)
					if e != nil {
						solutions <- SolutionInfo{
							Solution: nil,
							Error:    e,
						}
						break Loop
					}
					if improved {
						solutions <- SolutionInfo{
							Solution: s.bestSolution,
							Error:    nil,
						}
					}
				}
			}

			for _, parameter := range s.parameters {
				parameter.Update(solveInformation)
			}
			s.solveEvents.Iterated.Trigger(solveInformation)
		}
		s.solveEvents.Done.Trigger(solveInformation)
	}()

	return solutions, err
}
