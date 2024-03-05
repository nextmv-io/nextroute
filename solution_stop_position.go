package nextroute

import (
	"fmt"
)

// NewStopPosition returns a new StopPosition.
func NewStopPosition(
	p SolutionStop,
	s SolutionStop,
	n SolutionStop,
) (StopPosition, error) {
	if p == nil {
		return nil, fmt.Errorf("previous stop is nil")
	}
	if s == nil {
		return nil, fmt.Errorf("stop is nil")
	}
	if n == nil {
		return nil, fmt.Errorf("next stop is nil")
	}
	previous := p.(solutionStopImpl)
	stop := s.(solutionStopImpl)
	next := n.(solutionStopImpl)
	if previous.Solution() != stop.Solution() {
		return nil, fmt.Errorf(
			"previous %v and stop %v are on different solutions",
			previous,
			stop,
		)
	}
	if stop.Solution() != next.Solution() {
		return nil, fmt.Errorf(
			"stop %v and next %v are on different solutions",
			stop,
			next,
		)
	}
	if stop.IsPlanned() {
		return nil, fmt.Errorf("stop %v is planned", stop)
	}
	if previous.IsPlanned() &&
		next.IsPlanned() {
		if previous.vehicle().index != next.vehicle().index {
			return nil, fmt.Errorf(
				"previous %v and next %v are planned but on different vehicle",
				previous,
				next,
			)
		}
		if previous.Position() >= next.Position() {
			return nil, fmt.Errorf(
				"previous %v and next %v are planned but previous is not before next",
				previous,
				next,
			)
		}
	}
	return newStopPosition(previous, stop, next), nil
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

func (v stopPositionImpl) Previous() SolutionStop {
	return v.solution.stopByIndexCache[v.previousStopIndex]
}

func (v stopPositionImpl) Next() SolutionStop {
	return v.solution.stopByIndexCache[v.nextStopIndex]
}

func (v stopPositionImpl) Stop() SolutionStop {
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
