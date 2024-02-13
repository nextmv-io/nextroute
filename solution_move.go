package nextroute

import (
	"context"
	"math"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

// NotExecutableMove returns a constant new empty non-executable move.
var NotExecutableMove nextroute.SolutionMove = NewNotExecutableMove()

// NewNotExecutableMove creates a new empty non-executable move.
func NewNotExecutableMove() nextroute.SolutionMove {
	return newNotExecutableMove()
}

func newNotExecutableMove() nextroute.SolutionMove {
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

func (m solutionMoveImpl) PlanUnit() nextroute.SolutionPlanUnit {
	return nil
}

func (m solutionMoveImpl) TakeBest(that nextroute.SolutionMove) nextroute.SolutionMove {
	return that
}

func (m solutionMoveImpl) Value() float64 {
	return math.Inf(1)
}

func (m solutionMoveImpl) ValueSeen() int {
	return 0
}

func (m solutionMoveImpl) IncrementValueSeen(_ int) nextroute.SolutionMove {
	return m
}

// takeBestInPlace takes the best move and a candidate move.
// It tries to modify best in place, if the underlying type allows it.
// Otherwise it makes a deep copy of that and returns it.
func takeBestInPlace(best, that nextroute.SolutionMove) nextroute.SolutionMove {
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

// tryReplaceBy the move in src to dst by mutating dst if possible. Otherwise makes a copy.
func tryReplaceBy(dst, src nextroute.SolutionMove, newValueSeen int) nextroute.SolutionMove {
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
		m2.stopPositions = common.DefensiveCopy(m.stopPositions)
		m2.valueSeen = newValueSeen
		return &m2
	case solutionMoveUnitsImpl:
		m2 := m
		m2.moves = common.DefensiveCopy(m.moves)
		m2.valueSeen = newValueSeen
		return m2
	}
	return src.IncrementValueSeen(newValueSeen - src.ValueSeen())
}
