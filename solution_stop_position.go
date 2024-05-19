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
	previous := p
	stop := s
	next := n
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
	previous SolutionStop,
	stop SolutionStop,
	next SolutionStop,
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
		v.Previous().ModelStop().ID(),
		v.Previous().Index(),
		v.Stop().ModelStop().ID(),
		v.Stop().Index(),
		v.Next().ModelStop().ID(),
		v.Next().Index(),
	)
}

// Previous denotes the upcoming stop's previous stop if the associated move
// involving the stop position is executed. It's worth noting that
// the previous stop may not have been planned yet.
func (v StopPosition) Previous() SolutionStop {
	return SolutionStop{
		index:    v.previousStopIndex,
		solution: v.solution,
	}
}

// Next denotes the upcoming stop's next stop if the associated move
// involving the stop position is executed. It's worth noting that
// the next stop may not have been planned yet.
func (v StopPosition) Next() SolutionStop {
	return SolutionStop{
		index:    v.nextStopIndex,
		solution: v.solution,
	}
}

// Stop returns the stop which is not yet part of the solution. This stop
// is not planned yet if the move where the invoking stop position belongs
// to, has not been executed yet.
func (v StopPosition) Stop() SolutionStop {
	return SolutionStop{
		index:    v.stopIndex,
		solution: v.solution,
	}
}
