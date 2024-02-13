package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/nextroute"
)

// NewStopPosition returns a new StopPosition. Exposed for testing, should not
// be exposed in SDK.
func NewStopPosition(
	p nextroute.SolutionStop,
	s nextroute.SolutionStop,
	n nextroute.SolutionStop,
) nextroute.StopPosition {
	previous := p.(solutionStopImpl)
	stop := s.(solutionStopImpl)
	next := n.(solutionStopImpl)
	if stop.IsPlanned() {
		panic(fmt.Sprintf("stop %v is planned", stop))
	}
	if previous.IsPlanned() &&
		next.IsPlanned() {
		if previous.vehicle().index != next.vehicle().index {
			panic(
				fmt.Sprintf(
					"previous %v and next %v are planned but on different input",
					previous,
					next,
				),
			)
		}
		if previous.Position() >= next.Position() {
			panic(
				fmt.Sprintf(
					"previous %v and next %v are planned but previous is not before next",
					previous,
					next,
				),
			)
		}
	}
	return newStopPosition(previous, stop, next)
}

func newStopPosition(
	previous solutionStopImpl,
	stop solutionStopImpl,
	next solutionStopImpl,
) stopPositionImpl {
	return stopPositionImpl{
		previousStopIndex: previous.index,
		stopIndex:         stop.index,
		nextStopIndex:     next.index,
		solution:          stop.solution,
	}
}

type stopPositionImpl struct {
	solution          *solutionImpl
	previousStopIndex int
	stopIndex         int
	nextStopIndex     int
}

func (v stopPositionImpl) String() string {
	return fmt.Sprintf("stopPosition{%s[%v]->%s[%v]->%s[%v]",
		v.previous().ModelStop().ID(),
		v.previous().Index(),
		v.stop().ModelStop().ID(),
		v.stop().Index(),
		v.next().ModelStop().ID(),
		v.next().Index(),
	)
}

func (v stopPositionImpl) Previous() nextroute.SolutionStop {
	return v.solution.stopByIndexCache[v.previousStopIndex]
}

func (v stopPositionImpl) Next() nextroute.SolutionStop {
	return v.solution.stopByIndexCache[v.nextStopIndex]
}

func (v stopPositionImpl) Stop() nextroute.SolutionStop {
	return v.solution.stopByIndexCache[v.stopIndex]
}

func (v stopPositionImpl) previous() solutionStopImpl {
	return solutionStopImpl{
		index:    v.previousStopIndex,
		solution: v.solution,
	}
}

func (v stopPositionImpl) next() solutionStopImpl {
	return solutionStopImpl{
		index:    v.nextStopIndex,
		solution: v.solution,
	}
}

func (v stopPositionImpl) stop() solutionStopImpl {
	return solutionStopImpl{
		index:    v.stopIndex,
		solution: v.solution,
	}
}
