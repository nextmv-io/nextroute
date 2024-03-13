// © 2019-present nextmv.io inc

package nextroute

import (
	"math/rand"
	"slices"

	"github.com/nextmv-io/nextroute/common"
)

// ImmutableSolutionPlanUnitCollection is a collection of solution plan
// units.
type ImmutableSolutionPlanUnitCollection interface {
	// Iterator returns a channel that can be used to iterate over the solution
	// plan units in the collection.
	// If you break out of the for loop before the channel is closed,
	// the goroutine launched by the Iterator() method will be blocked forever,
	// waiting to send the next element on the channel. This can lead to a
	// goroutine leak and potentially exhaust the system resources. Therefore,
	// it is recommended to always use the following pattern:
	//    iter := collection.Iterator()
	//    for {
	//        element, ok := <-iter
	//        if !ok {
	//            break
	//        }
	//        // do something with element, potentially break out of the loop
	//    }
	//    close(iter)
	Iterator(quit <-chan struct{}) <-chan SolutionPlanUnit
	// RandomDraw returns a random sample of n different solution plan units.
	RandomDraw(n int) SolutionPlanUnits
	// RandomElement returns a random solution plan unit.
	RandomElement() SolutionPlanUnit
	// Size return the number of solution plan units in the collection.
	Size() int
	// SolutionPlanUnit returns the solution plan units in the collection
	// which correspond to the given model plan unit. If no such solution
	// plan unit is found, nil is returned.
	SolutionPlanUnit(modelPlanUnit ModelPlanUnit) SolutionPlanUnit
	// SolutionPlanUnits returns the solution plan units in the collection.
	// The returned slice is a defensive copy of the internal slice, so
	// modifying it will not affect the collection.
	SolutionPlanUnits() SolutionPlanUnits
}

// SolutionPlanUnitCollection is a collection of solution plan units.
type SolutionPlanUnitCollection interface {
	ImmutableSolutionPlanUnitCollection
	// Add adds a [SolutionPlanUnit] to the collection.
	Add(solutionPlanUnit SolutionPlanUnit)
	// Remove removes a [SolutionPlanUnit] from the collection.
	Remove(solutionPlanUnit SolutionPlanUnit)
}

// NewSolutionPlanUnitCollection returns a new SolutionPlanUnitCollection.
func NewSolutionPlanUnitCollection(
	random *rand.Rand,
	planUnits SolutionPlanUnits,
) SolutionPlanUnitCollection {
	p := solutionPlanUnitCollectionImpl{
		solutionPlanUnitCollectionBaseImpl: solutionPlanUnitCollectionBaseImpl{
			random:            random,
			solutionPlanUnits: slices.Clone(planUnits),
			indices:           make(map[int]int, len(planUnits)), // TODO: can this be an int slice?
		},
	}
	for i, planUnit := range planUnits {
		p.indices[planUnit.ModelPlanUnit().Index()] = i
	}
	return &p
}

type solutionPlanUnitCollectionImpl struct {
	solutionPlanUnitCollectionBaseImpl
}

func (s *solutionPlanUnitCollectionImpl) Add(solutionPlanUnit SolutionPlanUnit) {
	s.add(solutionPlanUnit)
}

func (s *solutionPlanUnitCollectionImpl) Remove(solutionPlanUnit SolutionPlanUnit) {
	s.remove(solutionPlanUnit)
}

func newSolutionPlanUnitCollectionBaseImpl(
	random *rand.Rand,
	initialCapacity int,
) solutionPlanUnitCollectionBaseImpl {
	return solutionPlanUnitCollectionBaseImpl{
		random:            random,
		solutionPlanUnits: make(SolutionPlanUnits, 0, initialCapacity),
		indices:           make(map[int]int, initialCapacity),
	}
}

type solutionPlanUnitCollectionBaseImpl struct {
	random            *rand.Rand
	indices           map[int]int
	solutionPlanUnits SolutionPlanUnits
}

func (p *solutionPlanUnitCollectionBaseImpl) SolutionPlanUnits() SolutionPlanUnits {
	return slices.Clone(p.solutionPlanUnits)
}

func (p *solutionPlanUnitCollectionBaseImpl) RandomElement() SolutionPlanUnit {
	return common.RandomElement(p.random, p.solutionPlanUnits)
}

func (p *solutionPlanUnitCollectionBaseImpl) Size() int {
	return len(p.solutionPlanUnits)
}

func (p *solutionPlanUnitCollectionBaseImpl) RandomDraw(n int) SolutionPlanUnits {
	return common.RandomElements(p.random, p.solutionPlanUnits, n)
}

func (p *solutionPlanUnitCollectionBaseImpl) add(solutionPlanUnit SolutionPlanUnit) {
	if _, ok := p.indices[solutionPlanUnit.ModelPlanUnit().Index()]; ok {
		return
	}
	p.indices[solutionPlanUnit.ModelPlanUnit().Index()] = len(p.solutionPlanUnits)
	p.solutionPlanUnits = append(p.solutionPlanUnits, solutionPlanUnit)
}

func (p *solutionPlanUnitCollectionBaseImpl) remove(solutionPlanUnit SolutionPlanUnit) {
	index, ok := p.indices[solutionPlanUnit.ModelPlanUnit().Index()]
	if !ok {
		return
	}
	lastIndex := len(p.solutionPlanUnits) - 1
	lastElement := p.solutionPlanUnits[lastIndex]
	p.solutionPlanUnits[index] = lastElement
	p.indices[lastElement.ModelPlanUnit().Index()] = index
	delete(p.indices, solutionPlanUnit.ModelPlanUnit().Index())
	p.solutionPlanUnits[lastIndex] = nil
	p.solutionPlanUnits = p.solutionPlanUnits[:lastIndex]
}

func (p *solutionPlanUnitCollectionBaseImpl) Iterator(quit <-chan struct{}) <-chan SolutionPlanUnit {
	ch := make(chan SolutionPlanUnit)
	go func() {
		defer close(ch)
		for _, solutionPlanUnit := range p.solutionPlanUnits {
			select {
			case <-quit:
				return
			case ch <- solutionPlanUnit:
			}
		}
	}()
	return ch
}

func (p *solutionPlanUnitCollectionBaseImpl) SolutionPlanUnit(
	modelPlanUnit ModelPlanUnit,
) SolutionPlanUnit {
	if index, ok := p.indices[modelPlanUnit.Index()]; ok {
		return p.solutionPlanUnits[index]
	}
	return nil
}
