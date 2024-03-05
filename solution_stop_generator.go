// © 2019-present nextmv.io inc
package nextroute

import (
	"slices"
	"sync"
)

// SolutionStopGenerator is an iterator of solution stops.
type SolutionStopGenerator interface {
	Next() SolutionStop
}

// NewSolutionStopGenerator return a solution stop iterator of a move.
// If startAtFirst is true, the first stop will be first stop of the vehicle.
// If endAtLast is true, the last stop will be the last stop of the vehicle.
//
// For example adding sequence A, B in a sequence 1 -> 2 -> 3 -> 4 -> 5 -> 6
// where A goes before 4 and B goes before 5 will generate the following
// solution stops: 3 -> A -> 4 -> B -> 5
// If startsAtFirst is true, the solution stops will start with 1:
// 1 -> 2 -> 3 -> A -> 4 -> B -> 5
// If endAtLast is also true, the solution stops will end with 6:
// 1 -> 2 -> 3 -> A -> 4 -> B -> 5 -> 6.
//
// For example:
//
//	   generator := NewSolutionStopGenerator(move, false, true)
//
//		  for solutionStop := generator.Next(); solutionStop != nil; solutionStop = generator.Next() {
//			  // Do something with solutionStop
//	   }
func NewSolutionStopGenerator(
	move SolutionMoveStops,
	startAtFirst bool,
	endAtLast bool,
) SolutionStopGenerator {
	nextStop := move.Vehicle().First()
	if !startAtFirst {
		nextStop = move.StopPositions()[0].Previous()
	}
	return &solutionStopGeneratorImpl{
		stopPositions:           slices.Clone(move.(*solutionMoveStopsImpl).stopPositions),
		startAtFirst:            startAtFirst,
		endAtLast:               endAtLast,
		nextStop:                nextStop.(solutionStopImpl),
		activeStopPositionIndex: 0,
	}
}

var solutionGeneratorPool = sync.Pool{
	New: func() interface{} { return new(solutionStopGeneratorImpl) },
}

func newSolutionStopGenerator(
	move solutionMoveStopsImpl,
	startAtFirst bool,
	endAtLast bool,
) *solutionStopGeneratorImpl {
	nextStop := move.vehicle().first()
	if !startAtFirst {
		nextStop = move.stopPositions[0].previous()
	}
	solutionStopGenerator := solutionGeneratorPool.Get().(*solutionStopGeneratorImpl)
	solutionStopGenerator.stopPositions = solutionStopGenerator.stopPositions[:0]
	solutionStopGenerator.stopPositions = append(solutionStopGenerator.stopPositions, move.stopPositions...)
	solutionStopGenerator.nextStop = nextStop
	solutionStopGenerator.startAtFirst = startAtFirst
	solutionStopGenerator.endAtLast = endAtLast
	solutionStopGenerator.activeStopPositionIndex = 0
	solutionStopGenerator.endReached = false
	return solutionStopGenerator
}

type solutionStopGeneratorImpl struct {
	nextStop                solutionStopImpl
	stopPositions           []stopPositionImpl
	activeStopPositionIndex int
	startAtFirst            bool
	endAtLast               bool
	endReached              bool
}

func (s *solutionStopGeneratorImpl) Next() SolutionStop {
	next, ok := s.next()
	if !ok {
		return nil
	}
	return next
}

func (s *solutionStopGeneratorImpl) release() {
	solutionGeneratorPool.Put(s)
}

func (s *solutionStopGeneratorImpl) next() (solutionStopImpl, bool) {
	if s.endReached {
		return solutionStopImpl{}, false
	}

	returnStop := s.nextStop

	if s.startAtFirst {
		if s.nextStop == s.stopPositions[s.activeStopPositionIndex].previous() {
			s.startAtFirst = false
			s.nextStop = s.stopPositions[s.activeStopPositionIndex].stop()
		} else {
			s.nextStop = s.nextStop.next()
		}
		return returnStop, true
	}

	if s.activeStopPositionIndex < len(s.stopPositions) {
		if s.nextStop == s.stopPositions[s.activeStopPositionIndex].stop() {
			s.nextStop = s.stopPositions[s.activeStopPositionIndex].next()
			s.activeStopPositionIndex++
			return returnStop, true
		}
		if s.nextStop == s.stopPositions[s.activeStopPositionIndex].previous() {
			s.nextStop = s.stopPositions[s.activeStopPositionIndex].stop()
			s.activeStopPositionIndex++
			return returnStop, true
		}
		if !s.nextStop.IsPlanned() {
			s.nextStop = s.stopPositions[s.activeStopPositionIndex-1].next()
		} else {
			if s.nextStop.IsLast() {
				s.endReached = true
			} else {
				s.nextStop = s.nextStop.next()
			}
		}

		return returnStop, true
	}

	if !s.nextStop.IsPlanned() {
		s.nextStop = s.stopPositions[s.activeStopPositionIndex-1].next()
		return returnStop, true
	}

	if s.endAtLast {
		if s.nextStop.IsLast() {
			s.endReached = true
			s.endAtLast = false
		} else {
			s.nextStop = s.nextStop.next()
		}
		return returnStop, true
	}

	s.endReached = true
	return returnStop, true
}
