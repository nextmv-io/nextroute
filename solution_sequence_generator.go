package nextroute

import (
	"math/rand"
	"sync/atomic"

	"github.com/nextmv-io/sdk/common"
)

// SequenceGeneratorChannel generates all possible sequences of solution stops
// for a given plan planUnit.
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
			counter := map[int]int{}
			modelPlanUnit := planUnit.ModelPlanUnit().(*planMultipleStopsImpl)
			dag := modelPlanUnit.dag.(*directedAcyclicGraphImpl)
			for _, arc := range dag.arcs {
				counter[arc.Destination().Index()]++
			}

			sequenceGenerator(
				solutionStops,
				make([]SolutionStop, 0, len(solutionStops)),
				used,
				counter,
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
			)
		}
	}()

	return ch
}

func sequenceGenerator(
	stops, sequence SolutionStops,
	used []bool,
	counter map[int]int,
	dag DirectedAcyclicGraph,
	random *rand.Rand,
	maxSequences *int64,
	yield func(SolutionStops),
) {
	if len(sequence) == len(stops) {
		if atomic.AddInt64(maxSequences, -1) >= 0 {
			yield(common.DefensiveCopy(sequence))
		}
		return
	}

	stopOrder := random.Perm(len(stops))

	for _, idx := range stopOrder {
		stop := stops[idx]
		if !used[idx] && counter[stop.ModelStop().Index()] == 0 {
			used[idx] = true
			outboundArcs := dag.OutboundArcs(stop.ModelStop())
			if len(outboundArcs) == 1 {
				arc := outboundArcs[0]
				counter[arc.Destination().Index()]--
			} else {
				outboundArcOrder := random.Perm(len(outboundArcs))
				for _, arcsIdx := range outboundArcOrder {
					arc := outboundArcs[arcsIdx]
					counter[arc.Destination().Index()]--
				}
			}

			sequenceGenerator(stops, append(sequence, stop), used, counter, dag, random, maxSequences, yield)
			// reached the maximum number of sequences
			if *maxSequences == 0 {
				return
			}
			used[idx] = false
			for _, arc := range outboundArcs {
				counter[arc.Destination().Index()]++
			}
		}
	}
}
