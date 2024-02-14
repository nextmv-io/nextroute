package nextroute

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/nextmv-io/sdk/alns"
	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/run"
)

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
	progression    []alns.ProgressionEntry
}

func (s *solveImpl) OnImprovement(solveInformation SolveInformation) {
	s.progression = append(s.progression, alns.ProgressionEntry{
		ElapsedSeconds: time.Since(solveInformation.Start()).Seconds(),
		Value:          solveInformation.Solver().BestSolution().Score(),
	})
}

func (s *solveImpl) Progression() []alns.ProgressionEntry {
	return common.DefensiveCopy(s.progression)
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
	newWorkSolution := startSolutions[0].Copy()
	s.bestSolution = startSolutions[0].Copy()
	s.workSolution = newWorkSolution
	s.random = rand.New(rand.NewSource(newWorkSolution.Random().Int63()))

	start := ctx.Value(run.Start).(time.Time)

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
	solutions := make(chan Solution, 100)
	solutions <- s.bestSolution
	go func() {
		defer close(solutions)

	Loop:
		for iteration := 0; iteration < solveOptions.Iterations; iteration++ {
			solveInformation.iteration = iteration
			solveInformation.deltaScore = 0.0
			solveInformation.solveOperators = make(
				SolveOperators,
				0,
				len(s.solveOperators),
			)

			s.solveEvents.Iterating.Trigger(solveInformation)
			for _, solveOperator := range s.SolveOperators() {
				select {
				case <-ctx.Done():
					s.solveEvents.ContextDone.Trigger(solveInformation)
					break Loop
				default:
					improved, e := s.invoke(ctx, solveOperator, solveInformation)
					if e != nil {
						err = fmt.Errorf("%w", e)
						break Loop
					}
					if improved {
						solutions <- s.bestSolution
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
