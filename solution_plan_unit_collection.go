package nextroute

import (
	"math/rand"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

// NewSolutionPlanUnitCollection returns a new SolutionPlanUnitCollection.
func NewSolutionPlanUnitCollection(
	random *rand.Rand,
	planUnits nextroute.SolutionPlanUnits,
) nextroute.SolutionPlanUnitCollection {
	p := solutionPlanUnitCollectionImpl{
		solutionPlanUnitCollectionBaseImpl: solutionPlanUnitCollectionBaseImpl{
			random:            random,
			solutionPlanUnits: common.DefensiveCopy(planUnits),
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

func (s *solutionPlanUnitCollectionImpl) Add(solutionPlanUnit nextroute.SolutionPlanUnit) {
	s.add(solutionPlanUnit)
}

func (s *solutionPlanUnitCollectionImpl) Remove(solutionPlanUnit nextroute.SolutionPlanUnit) {
	s.remove(solutionPlanUnit)
}

func newSolutionPlanUnitCollectionBaseImpl(
	random *rand.Rand,
	initialCapacity int,
) solutionPlanUnitCollectionBaseImpl {
	return solutionPlanUnitCollectionBaseImpl{
		random:            random,
		solutionPlanUnits: make(nextroute.SolutionPlanUnits, 0, initialCapacity),
		indices:           make(map[int]int, initialCapacity),
	}
}

type solutionPlanUnitCollectionBaseImpl struct {
	random            *rand.Rand
	indices           map[int]int
	solutionPlanUnits nextroute.SolutionPlanUnits
}

func (p *solutionPlanUnitCollectionBaseImpl) SolutionPlanUnits() nextroute.SolutionPlanUnits {
	return common.DefensiveCopy(p.solutionPlanUnits)
}

func (p *solutionPlanUnitCollectionBaseImpl) RandomElement() nextroute.SolutionPlanUnit {
	return common.RandomElement(p.random, p.solutionPlanUnits)
}

func (p *solutionPlanUnitCollectionBaseImpl) Size() int {
	return len(p.solutionPlanUnits)
}

func (p *solutionPlanUnitCollectionBaseImpl) RandomDraw(n int) nextroute.SolutionPlanUnits {
	return common.RandomElements(p.random, p.solutionPlanUnits, n)
}

func (p *solutionPlanUnitCollectionBaseImpl) add(solutionPlanUnit nextroute.SolutionPlanUnit) {
	if _, ok := p.indices[solutionPlanUnit.ModelPlanUnit().Index()]; ok {
		return
	}
	p.indices[solutionPlanUnit.ModelPlanUnit().Index()] = len(p.solutionPlanUnits)
	p.solutionPlanUnits = append(p.solutionPlanUnits, solutionPlanUnit)
}

func (p *solutionPlanUnitCollectionBaseImpl) remove(solutionPlanUnit nextroute.SolutionPlanUnit) {
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

func (p *solutionPlanUnitCollectionBaseImpl) Iterator(quit <-chan struct{}) <-chan nextroute.SolutionPlanUnit {
	ch := make(chan nextroute.SolutionPlanUnit)
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
	modelPlanUnit nextroute.ModelPlanUnit,
) nextroute.SolutionPlanUnit {
	if index, ok := p.indices[modelPlanUnit.Index()]; ok {
		return p.solutionPlanUnits[index]
	}
	return nil
}
