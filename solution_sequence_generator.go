// © 2019-present nextmv.io inc

package nextroute

import (
	"math/rand"
	"slices"
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
	ch := make(chan SolutionStops)
	go func() {
		defer close(ch)
		sequenceGeneratorSync(pu, func(solutionStops SolutionStops) {
			select {
			case <-quit:
				return
			case ch <- slices.Clone(solutionStops):
			}
		})
	}()

	return ch
}

func sequenceGeneratorSync(pu SolutionPlanUnit, yield func(SolutionStops)) {
	planUnit := pu.(*solutionPlanStopsUnitImpl)
	solutionStops := planUnit.solutionStops
	if planUnit.ModelPlanStopsUnit().NumberOfStops() == 1 {
		yield(planUnit.SolutionStops())
		return
	}
	solution := planUnit.solution()
	maxSequences := int64(solution.Model().SequenceSampleSize())
	nSolutionStops := len(solutionStops)
	used := make([]bool, nSolutionStops)
	inDegree := make(map[int]int, nSolutionStops)
	modelPlanUnit := planUnit.ModelPlanUnit().(*planMultipleStopsImpl)
	dag := modelPlanUnit.dag.(*directedAcyclicGraphImpl)
	for _, arc := range dag.arcs {
		inDegree[arc.Destination().Index()]++
	}

	recSequenceGenerator(
		solutionStops,
		make([]SolutionStop, 0, nSolutionStops),
		used,
		inDegree,
		dag,
		solution.Random(),
		&maxSequences,
		yield,
		-1,
	)
}

func recSequenceGenerator(
	stops []SolutionStop,
	sequence SolutionStops,
	used []bool,
	inDegree map[int]int,
	dag DirectedAcyclicGraph,
	random *rand.Rand,
	maxSequences *int64,
	yield func(SolutionStops),
	directSuccessor int,
) {
	nStops := len(stops)
	if *maxSequences == 0 {
		return
	}
	if len(sequence) == nStops {
		*maxSequences--
		if *maxSequences >= 0 {
			yield(sequence)
		}
		return
	}

	stopOrder := random.Perm(nStops)

	// we know the direct successor, so we move it to the front of the random
	// sequence
	if directSuccessor != -1 {
		for _, stopIdx := range stopOrder {
			if stops[stopIdx].Index() == directSuccessor {
				stopOrder = stopOrder[:1]
				stopOrder[0] = stopIdx
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
			recSequenceGenerator(
				stops, append(sequence, stop), used, inDegree, dag, random, maxSequences, yield, directSuccessor,
			)
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
