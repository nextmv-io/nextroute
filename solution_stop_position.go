// © 2019-present nextmv.io inc

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
		return StopPosition{}, fmt.Errorf("previous stop is nil")
	}
	if s == nil {
		return StopPosition{}, fmt.Errorf("stop is nil")
	}
	if n == nil {
		return StopPosition{}, fmt.Errorf("next stop is nil")
	}
	previous := p.(solutionStopImpl)
	stop := s.(solutionStopImpl)
	next := n.(solutionStopImpl)
	if previous.Solution() != stop.Solution() {
		return StopPosition{}, fmt.Errorf(
			"previous %v and stop %v are on different solutions",
			previous,
			stop,
		)
	}
	if stop.Solution() != next.Solution() {
		return StopPosition{}, fmt.Errorf(
			"stop %v and next %v are on different solutions",
			stop,
			next,
		)
	}
	if stop.IsPlanned() {
		return StopPosition{}, fmt.Errorf("stop %v is planned", stop)
	}
	if previous.IsPlanned() &&
		next.IsPlanned() {
		if previous.vehicle().index != next.vehicle().index {
			return StopPosition{}, fmt.Errorf(
				"previous %v and next %v are planned but on different vehicle",
				previous,
				next,
			)
		}
		if previous.Position() >= next.Position() {
			return StopPosition{}, fmt.Errorf(
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
) StopPosition {
	return StopPosition{
		previousStopIndex: previous.index,
		stopIndex:         stop.index,
		nextStopIndex:     next.index,
		solution:          stop.solution,
	}
}

func (v StopPosition) String() string {
	return fmt.Sprintf("stopPosition{%s[%v]->%s[%v]->%s[%v]",
		v.previous().ModelStop().ID(),
		v.previous().Index(),
		v.stop().ModelStop().ID(),
		v.stop().Index(),
		v.next().ModelStop().ID(),
		v.next().Index(),
	)
}

// Previous denotes the upcoming stop's previous stop if the associated move
// involving the stop position is executed. It's worth noting that
// the previous stop may not have been planned yet.
func (v StopPosition) Previous() SolutionStop {
	return v.solution.stopByIndexCache[v.previousStopIndex]
}

// Next denotes the upcoming stop's next stop if the associated move
// involving the stop position is executed. It's worth noting that
// the next stop may not have been planned yet.
func (v StopPosition) Next() SolutionStop {
	return v.solution.stopByIndexCache[v.nextStopIndex]
}

// Stop returns the stop which is not yet part of the solution. This stop
// is not planned yet if the move where the invoking stop position belongs
// to, has not been executed yet.
func (v StopPosition) Stop() SolutionStop {
	return v.solution.stopByIndexCache[v.stopIndex]
}

func (v StopPosition) previous() solutionStopImpl {
	return solutionStopImpl{
		index:    v.previousStopIndex,
		solution: v.solution,
	}
}

func (v StopPosition) next() solutionStopImpl {
	return solutionStopImpl{
		index:    v.nextStopIndex,
		solution: v.solution,
	}
}

func (v StopPosition) stop() solutionStopImpl {
	return solutionStopImpl{
		index:    v.stopIndex,
		solution: v.solution,
	}
}
