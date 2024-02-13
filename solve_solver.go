package nextroute

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/nextmv-io/sdk/alns"
	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/run"
)

// NewSkeletonSolver creates a new solver for the given model.
func NewSkeletonSolver(model nextroute.Model) (nextroute.Solver, error) {
	solver := &solveImpl{
		solveEvents: nextroute.NewSolveEvents(),
		model:       model,
	}
	solver.SolveEvents().NewBestSolution.Register(solver.OnImprovement)
	return solver, nil
}

type solveImpl struct {
	model          nextroute.Model
	solveEvents    nextroute.SolveEvents
	workSolution   nextroute.Solution
	bestSolution   nextroute.Solution
	random         *rand.Rand
	solveOperators nextroute.SolveOperators
	parameters     nextroute.SolveParameters
	progression    []alns.ProgressionEntry
}

func (s *solveImpl) OnImprovement(solveInformation nextroute.SolveInformation) {
	s.progression = append(s.progression, alns.ProgressionEntry{
		ElapsedSeconds: time.Since(solveInformation.Start()).Seconds(),
		Value:          solveInformation.Solver().BestSolution().Score(),
	})
}

func (s *solveImpl) Progression() []alns.ProgressionEntry {
	return common.DefensiveCopy(s.progression)
}

func (s *solveImpl) Model() nextroute.Model {
	return s.model
}

func (s *solveImpl) AddSolveOperators(solveOperators ...nextroute.SolveOperator) {
	for _, solveOperator := range solveOperators {
		s.solveOperators = append(s.solveOperators, solveOperator)
		for _, parameter := range solveOperator.Parameters() {
			s.register(parameter)
		}
		for _, st := range s.SolveOperators() {
			if interested, ok := st.(nextroute.InterestedInBetterSolution); ok {
				s.SolveEvents().NewBestSolution.Register(interested.OnBetterSolution)
			}
			if interested, ok := solveOperator.(nextroute.InterestedInStartSolve); ok {
				s.SolveEvents().Start.Register(interested.OnStartSolve)
			}
		}
	}
}

func (s *solveImpl) Random() *rand.Rand {
	return s.random
}

func (s *solveImpl) SolveEvents() nextroute.SolveEvents {
	return s.solveEvents
}

func (s *solveImpl) register(parameter nextroute.SolveParameter) {
	s.parameters = append(s.parameters, parameter)
}

func (s *solveImpl) SolveOperators() nextroute.SolveOperators {
	solveOperators := make(nextroute.SolveOperators, len(s.solveOperators))
	copy(solveOperators, s.solveOperators)
	return solveOperators
}

func (s *solveImpl) Reset(solution nextroute.Solution, solveInformation nextroute.SolveInformation) {
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

func (s *solveImpl) BestSolution() nextroute.Solution {
	return s.bestSolution
}

func (s *solveImpl) WorkSolution() nextroute.Solution {
	return s.workSolution
}

// invoke returns true if the best solution was improved.
func (s *solveImpl) invoke(
	ctx context.Context,
	solveOperator nextroute.SolveOperator,
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

func (s *solveImpl) newBestSolution(solution nextroute.Solution, solveInformation *solveInformationImpl) {
	s.bestSolution = solution.Copy()
	s.solveEvents.NewBestSolution.Trigger(solveInformation)
}

func (s *solveImpl) Solve(
	ctx context.Context,
	solveOptions nextroute.SolveOptions,
	startSolutions ...nextroute.Solution,
) (nextroute.SolutionChannel, error) {
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
		solveOperators: make(nextroute.SolveOperators, 0, len(s.solveOperators)),
		start:          start,
	}

	s.solveEvents.Start.Trigger(solveInformation)

	var err error

	// hard coding a size of 100 here. If we ever need to, we can make it
	// configurable.
	solutions := make(chan nextroute.Solution, 100)
	solutions <- s.bestSolution
	go func() {
		defer close(solutions)

	Loop:
		for iteration := 0; iteration < solveOptions.Iterations; iteration++ {
			solveInformation.iteration = iteration
			solveInformation.deltaScore = 0.0
			solveInformation.solveOperators = make(
				nextroute.SolveOperators,
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
