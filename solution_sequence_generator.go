// © 2019-present nextmv.io inc

package nextroute

import (
	"math/rand"
	"slices"
	"sync/atomic"
)

// SequenceGeneratorChannel generates all possible sequences of solution stops
// for a given plan planUnit.
//
// If there are more possible sequences than the maximum number of sequences
// allowed by the model, the generator will stop generating sequences after
// the maximum number of sequences has been reached. The maximum number of
// sequences is set by the model's [Model.SequenceSampleSize] function.
// The sequences are generated in a random order.
//
// Example:
//
//	quit := make(chan struct{})
//	defer close(quit)
//	sequences := make([]SolutionStops, 0)
//	for solutionStops := range SequenceGeneratorChannel(solution.SolutionPlanUnit(planUnit), quit) {
//		sequences = append(sequences, solutionStops)
//	}
func SequenceGeneratorChannel(
	pu SolutionPlanUnit,
	quit <-chan struct{},
) chan SolutionStops {
	planUnit := pu.(*solutionPlanStopsUnitImpl)
	solution := planUnit.solution()
	maxSequences := int64(solution.Model().SequenceSampleSize())
	solutionStops := planUnit.SolutionStops()
	ch := make(chan SolutionStops)
	go func() {
		defer close(ch)
		switch planUnit.ModelPlanStopsUnit().NumberOfStops() {
		case 1:
			ch <- solutionStops
			return
		default:
			used := make([]bool, len(solutionStops))
			inDegree := map[int]int{}
			modelPlanUnit := planUnit.ModelPlanUnit().(*planMultipleStopsImpl)
			dag := modelPlanUnit.dag.(*directedAcyclicGraphImpl)
			for _, solutionStop := range solutionStops {
				inDegree[solutionStop.ModelStop().Index()] = 0
			}
			for _, arc := range dag.arcs {
				inDegree[arc.Destination().Index()]++
			}

			sequenceGenerator(
				solutionStops,
				make([]SolutionStop, 0, len(solutionStops)),
				used,
				inDegree,
				dag,
				solution.Random(),
				&maxSequences,
				func(solutionStops SolutionStops) {
					select {
					case <-quit:
						return
					case ch <- solutionStops:
					}
				},
				-1,
			)
		}
	}()

	return ch
}

func sequenceGenerator(
	stops, sequence SolutionStops,
	used []bool,
	inDegree map[int]int,
	dag DirectedAcyclicGraph,
	random *rand.Rand,
	maxSequences *int64,
	yield func(SolutionStops),
	directSuccessor int,
) {
	if len(sequence) == len(stops) {
		if atomic.AddInt64(maxSequences, -1) >= 0 {
			yield(slices.Clone(sequence))
		}
		return
	}

	stopOrder := random.Perm(len(stops))

	// we know the direct successor, so we move it to the front of the random
	// sequence
	if directSuccessor != -1 {
		for _, stopIdx := range stopOrder {
			if stops[stopIdx].Index() == directSuccessor {
				stopOrder = []int{stopIdx}
				break
			}
		}
	}
	isDirectSuccessor := directSuccessor != -1
	directSuccessor = -1

	for _, idx := range stopOrder {
		stop := stops[idx]
		if !used[idx] && inDegree[stop.ModelStop().Index()] == 0 {
			used[idx] = true
			outboundArcs := dag.OutboundArcs(stop.ModelStop())
			if len(outboundArcs) == 1 {
				arc := outboundArcs[0]
				inDegree[arc.Destination().Index()]--
				if dag.HasDirectArc(arc.Origin(), arc.Destination()) {
					directSuccessor = stop.Solution().SolutionStop(arc.Destination()).Index()
				}
			} else {
				outboundArcOrder := random.Perm(len(outboundArcs))
				for _, arcsIdx := range outboundArcOrder {
					arc := outboundArcs[arcsIdx]
					inDegree[arc.Destination().Index()]--
					if dag.HasDirectArc(arc.Origin(), arc.Destination()) {
						directSuccessor = stop.Solution().SolutionStop(arc.Destination()).Index()
					}
				}
			}
			sequenceGenerator(stops, append(sequence, stop), used, inDegree, dag, random, maxSequences, yield, directSuccessor)
			// reached the maximum number of sequences
			if *maxSequences == 0 {
				return
			}
			used[idx] = false
			for _, arc := range outboundArcs {
				inDegree[arc.Destination().Index()]++
			}
			if isDirectSuccessor {
				break
			}
		}
	}
}
