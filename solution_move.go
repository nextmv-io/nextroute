// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"math"
	"slices"
)

// SolutionMove is a move in a solution.
type SolutionMove interface {
	// Execute executes the move. Returns true if the move was executed
	// successfully, false if the move was not executed successfully. A
	// move is not successful if it did not result in a change in the
	// solution without violating any hard constraints. A move can be
	// marked executable even if it is not successful in executing.
	Execute(context.Context) (bool, error)

	// IsExecutable returns true if the move is executable, false if the
	// move is not executable. A move is executable if the estimates believe
	// the move will result in a change in the solution without violating
	// any hard constraints.
	IsExecutable() bool

	// IsImprovement returns true if the move is estimated to be executable and
	// the move has a an estimated delta objective value less than zero, false
	// if the move is not executable or the move has a value of zero or greater
	// than zero.
	IsImprovement() bool

	// PlanUnit returns the [SolutionPlanUnit] that is affected by the move.
	PlanUnit() SolutionPlanUnit

	// TakeBest returns the best move between the given move and the
	// current move. The best move is the move with the lowest score. If
	// the scores are equal, a random uniform distribution is used to
	// determine the move to use.
	TakeBest(that SolutionMove) SolutionMove

	// Value returns the score of the move. The score is the difference
	// between the score of the solution before the move and the score of
	// the solution after the move. The score is based on the estimates and
	// the actual score of the solution after the move should be retrieved
	// using Solution.Score after the move has been executed.
	Value() float64

	// ValueSeen returns the number of times the value of this move has been
	// seen by the estimates. A tie-breaker is a mechanism used to resolve
	// situations where multiple moves have the same value. In cases where the
	// same value is seen multiple times, a tie-breaker is applied to ensure
	// that each option has an equal chance of being selected.
	ValueSeen() int

	// IncrementValueSeen increments the number of times the value of this move
	// has been seen by the estimates and returns the move. A tie-breaker is a
	// mechanism used to resolve situations where multiple moves have the same
	// value. In cases where the same value is seen multiple times, a
	// tie-breaker is applied to ensure that each option has an equal chance of
	// being selected.
	IncrementValueSeen(inc int) SolutionMove
}

// SolutionMoves is a slice of SolutionMove.
type SolutionMoves []SolutionMove

// NotExecutableMove returns a constant new empty non-executable move.
var NotExecutableMove = NewNotExecutableMove()

// NewNotExecutableMove creates a new empty non-executable move.
func NewNotExecutableMove() SolutionMove {
	return newNotExecutableMove()
}

func newNotExecutableMove() SolutionMove {
	return solutionMoveImpl{}
}

type solutionMoveImpl struct {
}

func (m solutionMoveImpl) Execute(_ context.Context) (bool, error) {
	return false, nil
}

func (m solutionMoveImpl) IsExecutable() bool {
	return false
}

func (m solutionMoveImpl) IsImprovement() bool {
	return false
}

func (m solutionMoveImpl) PlanUnit() SolutionPlanUnit {
	return nil
}

func (m solutionMoveImpl) TakeBest(that SolutionMove) SolutionMove {
	return that
}

func (m solutionMoveImpl) Value() float64 {
	return math.Inf(1)
}

func (m solutionMoveImpl) ValueSeen() int {
	return 0
}

func (m solutionMoveImpl) IncrementValueSeen(_ int) SolutionMove {
	return m
}

// takeBestInPlace takes the best move and a candidate move.
// It tries to modify best in place, if the underlying type allows it.
// Otherwise, it makes a deep copy of that and returns it.
func takeBestInPlace(best, that SolutionMove) SolutionMove {
	if !that.IsExecutable() {
		return best
	}
	if !best.IsExecutable() {
		return tryReplaceBy(best, that, that.ValueSeen())
	}
	if best.Value() < that.Value() {
		return best
	}
	if best.Value() > that.Value() {
		return tryReplaceBy(best, that, that.ValueSeen())
	}
	if best.PlanUnit().Solution().Random().Intn(best.ValueSeen()+that.ValueSeen()) == 0 {
		switch b := best.(type) {
		case *solutionMoveStopsImpl:
			b.valueSeen++
			return best
		default:
			return best.IncrementValueSeen(1)
		}
	}
	return tryReplaceBy(best, that, best.ValueSeen()+that.ValueSeen())
}

// tryReplaceBy the move in src to dst by mutating dst if possible. Otherwise,
// makes a copy.
func tryReplaceBy(dst, src SolutionMove, newValueSeen int) SolutionMove {
	// first we try to mutate in place
	if dstImpl, ok := dst.(*solutionMoveStopsImpl); ok {
		if srcImpl, ok := src.(*solutionMoveStopsImpl); ok {
			dstImpl.replaceBy(srcImpl, newValueSeen)
			return dst
		}
	}
	// otherwise we create a copy
	switch m := src.(type) {
	case *solutionMoveStopsImpl:
		m2 := *m
		m2.stopPositions = slices.Clone(m.stopPositions)
		m2.valueSeen = newValueSeen
		return &m2
	case solutionMoveUnitsImpl:
		m2 := m
		m2.moves = slices.Clone(m.moves)
		m2.valueSeen = newValueSeen
		return m2
	}
	return src.IncrementValueSeen(newValueSeen - src.ValueSeen())
}
