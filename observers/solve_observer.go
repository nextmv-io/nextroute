// © 2019-present nextmv.io inc

package observers

import (
	"fmt"
	"os"
	"time"

	"github.com/nextmv-io/nextroute"
)

// SolveObserver is an observer for the solve process.
type SolveObserver interface {
	nextroute.SolutionObserver
	nextroute.SolutionUnPlanObserver

	// Register registers the solver to the solve observer.
	Register(solver nextroute.Solver) error

	// FileName returns the file name of the solve observer.
	FileName() string
}

// NewSolveObserver returns a new solve observer. The solve observer can be
// used to observe the solve process. The solve observer writes to the given
// file name.
//
// The solve observer writes the following columns:
// - The first column is the type of event. The type of event can be one of:
//
//   - `+` for a plan event.
//
//   - `-` for an unplan event.
//
//   - `~` for a failed event.
//
//   - `b` for a new best solution event.
//
//   - 'o' for the objective definition event.
//
//   - `r` for a reset event.
//
//   - The second column is the time since the solve process started in
//     nanoseconds.
//
//   - The third column is the step. The step is used to group events
//     together. The step is incremented for each unplan event.
//
//   - The next columns are dependent on the event type.
//
//   - For an objective definition event, the next columns are:
//
//   - The number of terms of the objective.
//
//   - For each term of the objective the factor and name of the objective.
//
//   - For a plan event, the next columns are:
//
//   - The previous stop ID.
//
//   - The stop ID.
//
//   - The next stop ID.
//
//   - For each term of the objective the score
//
//   - The score after planning.
//
//   - The estimated impact on the objective of planning.
//
//   - For an unplan event, the next columns are:
//
//   - The previous stop ID.
//
//   - The stop ID.
//
//   - The next stop ID.
//
//   - For each term of the objective the score
//
//   - The score after un-planning.
//
//   - For a failed event, the next column is the reason for the failure.
//
//   - For a reset event, the next columns are:
//
//   - The score of the work solution.
//
//   - The score of the solution resetting to.
//
//   - For a new best solution event, the next columns are:
//
//   - For each term of the objective the score
//
//   - Last column is the score of the new best solution
//
//     To use the observer add it to the model the following way:
//
//     solver, err := nextroute.NewSolver(model, solverOptions)
//     solveObserver, err := observers.NewSolveObserver("solve.log")
//     solveObserver.Register(solver)
func NewSolveObserver(fileName string) (SolveObserver, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	return &solveObserverImpl{
		fileName:     fileName,
		file:         file,
		start:        time.Now(),
		unplanBuffer: make([]string, 0, 4),
	}, nil
}

type solveObserverImpl struct {
	solver       nextroute.Solver
	fileName     string
	file         *os.File
	start        time.Time
	step         int
	unplanBuffer []string
}

func (s *solveObserverImpl) OnSolutionConstraintChecked(
	_ nextroute.ModelConstraint,
	_ bool,
) {
}

func (s *solveObserverImpl) OnStopConstraintChecked(
	_ nextroute.SolutionStop,
	_ nextroute.ModelConstraint,
	_ bool,
) {
}

func (s *solveObserverImpl) OnVehicleConstraintChecked(
	_ nextroute.SolutionVehicle,
	_ nextroute.ModelConstraint,
	_ bool) {
}

func (s *solveObserverImpl) Register(solver nextroute.Solver) error {
	if s.solver != nil {
		return fmt.Errorf("solve observer already registered to a solver")
	}
	s.solver = solver

	solver.SolveEvents().Start.Register(func(
		info nextroute.SolveInformation,
	) {
		solution := info.Solver().WorkSolution()
		score := ""
		for _, term := range solution.Model().Objective().Terms() {
			score += fmt.Sprintf(";%f", solution.ObjectiveValue(term.Objective()))
		}
		for _, solutionVehicle := range solution.Vehicles() {
			if solutionVehicle.IsEmpty() {
				break
			}
			previousID := solutionVehicle.First().ModelStop().ID()
			lastID := solutionVehicle.Last().ModelStop().ID()
			for _, solutionStop := range solutionVehicle.SolutionStops() {
				_, err := s.file.WriteString(
					fmt.Sprintf("s;%v;%v;%s;%s;%s;%s%s;%f;%f\n",
						time.Since(s.start).Nanoseconds(),
						s.step,
						solutionVehicle.ModelVehicle().ID(),
						previousID,
						solutionStop.ModelStop().ID(),
						lastID,
						score,
						solution.Score(),
						0.0,
					),
				)
				if err != nil {
					panic(err)
				}
				previousID = solutionStop.ModelStop().ID()
			}
			s.step++
		}
	})

	solver.SolveEvents().Reset.Register(
		func(
			solution nextroute.Solution,
			info nextroute.SolveInformation,
		) {
			_, err := s.file.WriteString(
				fmt.Sprintf("r;%v;%v;%v;%f\n",
					time.Since(s.start).Nanoseconds(),
					s.step,
					info.Solver().WorkSolution().Score(),
					solution.Score(),
				),
			)
			if err != nil {
				panic(err)
			}
			s.step++
		},
	)
	solver.SolveEvents().NewBestSolution.Register(
		func(
			info nextroute.SolveInformation,
		) {
			score := ""
			for _, term := range info.Solver().BestSolution().Model().Objective().Terms() {
				score += fmt.Sprintf(";%f", info.Solver().BestSolution().ObjectiveValue(term.Objective()))
			}
			_, err := s.file.WriteString(
				fmt.Sprintf("b;%v;%v%s;%f\n",
					time.Since(s.start).Nanoseconds(),
					s.step,
					score,
					info.Solver().BestSolution().Score(),
				),
			)
			if err != nil {
				panic(err)
			}
			s.step++
		},
	)

	solver.Model().RemoveSolutionObserver(s)
	solver.Model().AddSolutionObserver(s)

	solver.Model().RemoveSolutionUnPlanObserver(s)
	solver.Model().AddSolutionUnPlanObserver(s)

	solver.SolveEvents().Done.Register(func(_ nextroute.SolveInformation) {
		err := s.close()
		if err != nil {
			panic(err)
		}
	})

	_, err := s.file.WriteString(
		fmt.Sprintf(
			"o;%v;%v",
			time.Since(s.start).Nanoseconds(),
			len(solver.Model().Objective().Terms()),
		),
	)
	if err != nil {
		panic(err)
	}

	for _, term := range solver.Model().Objective().Terms() {
		_, err = s.file.WriteString(fmt.Sprintf(";%f;%v", term.Factor(), term.Objective()))

		if err != nil {
			panic(err)
		}
	}

	_, err = s.file.WriteString("\n")
	if err != nil {
		panic(err)
	}

	return nil
}

func (s *solveObserverImpl) FileName() string {
	return s.fileName
}

func (s *solveObserverImpl) close() error {
	s.solver.Model().RemoveSolutionObserver(s)
	s.solver.Model().RemoveSolutionUnPlanObserver(s)

	return s.file.Close()
}

func (s *solveObserverImpl) OnUnPlan(planUnit nextroute.SolutionPlanStopsUnit) {
	if s.solver.WorkSolution() != planUnit.Solution() {
		return
	}
	s.unplanBuffer = s.unplanBuffer[:0]
	for _, solutionStop := range planUnit.SolutionStops() {
		s.unplanBuffer = append(
			s.unplanBuffer,
			fmt.Sprintf("-;%v;%v;%s;%s;%s;%s",
				time.Since(s.start).Nanoseconds(),
				s.step,
				solutionStop.Vehicle().ModelVehicle().ID(),
				solutionStop.Previous().ModelStop().ID(),
				solutionStop.ModelStop().ID(),
				solutionStop.Next().ModelStop().ID(),
			),
		)
	}
	s.step++
}

func (s *solveObserverImpl) OnUnPlanFailed(
	planUnit nextroute.SolutionPlanStopsUnit,
) {
	if s.solver.WorkSolution() != planUnit.Solution() {
		return
	}

	_, err := s.file.WriteString(fmt.Sprintf("~;unplan failed: %v\n", planUnit))
	if err != nil {
		panic(err)
	}
}

func (s *solveObserverImpl) OnUnPlanSucceeded(
	solutionPlanStopsUnit nextroute.SolutionPlanStopsUnit,
) {
	if s.solver.WorkSolution() != solutionPlanStopsUnit.Solution() {
		return
	}

	solutionStops := solutionPlanStopsUnit.SolutionStops()
	if len(s.unplanBuffer) != len(solutionStops) {
		return
	}
	score := ""
	for _, term := range solutionPlanStopsUnit.Solution().Model().Objective().Terms() {
		score += fmt.Sprintf(";%f", solutionPlanStopsUnit.Solution().ObjectiveValue(term.Objective()))
	}
	for idx := range solutionPlanStopsUnit.SolutionStops() {
		_, err := s.file.WriteString(
			s.unplanBuffer[idx] +
				score +
				";" +
				fmt.Sprintf("%f", solutionPlanStopsUnit.Solution().Score()) +
				"\n",
		)
		if err != nil {
			panic(err)
		}
	}
}

func (s *solveObserverImpl) OnNewSolution(_ nextroute.Model) {
}

func (s *solveObserverImpl) OnNewSolutionCreated(solution nextroute.Solution) {
	score := ""
	for _, term := range solution.Model().Objective().Terms() {
		score += fmt.Sprintf(";%f", solution.ObjectiveValue(term.Objective()))
	}
	for _, solutionVehicle := range solution.Vehicles() {
		if solutionVehicle.IsEmpty() {
			break
		}
		previousID := solutionVehicle.First().ModelStop().ID()
		lastID := solutionVehicle.Last().ModelStop().ID()
		for _, solutionStop := range solutionVehicle.SolutionStops() {
			_, err := s.file.WriteString(
				fmt.Sprintf("s;%v;%v;%s;%s;%s;%s%s;%f;%f\n",
					time.Since(s.start).Nanoseconds(),
					s.step,
					solutionVehicle.ModelVehicle().ID(),
					previousID,
					solutionStop.ModelStop().ID(),
					lastID,
					score,
					solution.Score(),
					0.0,
				),
			)
			if err != nil {
				panic(err)
			}
			previousID = solutionStop.ModelStop().ID()
		}
		s.step++
	}
}

func (s *solveObserverImpl) OnCopySolution(_ nextroute.Solution) {
}

func (s *solveObserverImpl) OnCopiedSolution(_ nextroute.Solution) {
}

func (s *solveObserverImpl) OnCheckConstraint(
	_ nextroute.ModelConstraint,
	_ nextroute.CheckedAt,
) {
}

func (s *solveObserverImpl) OnCheckedConstraint(
	_ nextroute.ModelConstraint,
	_ bool,
) {
}

func (s *solveObserverImpl) OnEstimateIsViolated(
	_ nextroute.ModelConstraint,
) {
}

func (s *solveObserverImpl) OnEstimatedIsViolated(
	_ nextroute.SolutionMove,
	_ nextroute.ModelConstraint,
	_ bool,
	_ nextroute.StopPositionsHint,
) {
}

func (s *solveObserverImpl) OnEstimateDeltaObjectiveScore() {
}

func (s *solveObserverImpl) OnEstimatedDeltaObjectiveScore(
	_ float64,
) {
}

func (s *solveObserverImpl) OnBestMove(
	_ nextroute.Solution,
) {
}

func (s *solveObserverImpl) OnBestMoveFound(
	_ nextroute.SolutionMove,
) {
}

func (s *solveObserverImpl) OnPlan(
	_ nextroute.SolutionMove,
) {
}

func (s *solveObserverImpl) OnPlanFailed(move nextroute.SolutionMove, _ nextroute.ModelConstraint) {
	_, err := s.file.WriteString(fmt.Sprintf("~;plan failed: %v\n", move))
	if err != nil {
		panic(err)
	}
}

func (s *solveObserverImpl) OnPlanSucceeded(move nextroute.SolutionMove) {
	solutionMoveStops := move.(nextroute.SolutionMoveStops)
	solution := move.PlanUnit().Solution()

	if s.solver.WorkSolution() != solution {
		return
	}

	score := ""
	for _, term := range solution.Model().Objective().Terms() {
		score += fmt.Sprintf(";%f", solution.ObjectiveValue(term.Objective()))
	}

	stopPositions := solutionMoveStops.StopPositions()
	for idx, stopPosition := range stopPositions {
		nextID := stopPosition.Next().ModelStop().ID()
		for j := idx + 1; j < len(stopPositions); j++ {
			if nextID == stopPositions[j].Stop().ModelStop().ID() {
				nextID = stopPositions[j].Next().ModelStop().ID()
			}
		}
		_, err := s.file.WriteString(
			"+;" +
				fmt.Sprintf("%v", time.Since(s.start).Nanoseconds()) +
				";" +
				fmt.Sprintf("%v", s.step) +
				";" +
				solutionMoveStops.Vehicle().ModelVehicle().ID() +
				";" +
				stopPosition.Previous().ModelStop().ID() +
				";" +
				stopPosition.Stop().ModelStop().ID() +
				";" +
				nextID +
				score +
				";" +
				fmt.Sprintf("%f", solution.Score()) +
				";" +
				fmt.Sprintf("%f", move.Value()) +
				"\n",
		)
		if err != nil {
			panic(err)
		}
	}
	s.step++
}
