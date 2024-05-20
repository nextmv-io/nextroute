// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
	"sync"

	"github.com/nextmv-io/nextroute/common"
)

// SolutionPlanStopsUnit is a set of stops that are planned to be visited by
// a vehicle.
type SolutionPlanStopsUnit interface {
	SolutionPlanUnit
	// ModelPlanStopsUnit returns the [ModelPlanStopsUnit] this unit is
	// based upon.
	ModelPlanStopsUnit() ModelPlanStopsUnit

	// SolutionStop returns the solution stop for the given model stop.
	// Will panic if the stop is not part of the unit.
	SolutionStop(stop ModelStop) SolutionStop
	// SolutionStops returns the solution stops in this unit.
	SolutionStops() SolutionStops
	// StopPositions returns the stop positions of the invoking plan unit.
	// The stop positions are the positions of the stops in the solution.
	// If the unit is unplanned, the stop positions will be empty.
	StopPositions() StopPositions
}

// SolutionPlanStopsUnits is a slice of [SolutionPlanStopsUnit].
type SolutionPlanStopsUnits []SolutionPlanStopsUnit

type solutionPlanStopsUnitImpl struct {
	modelPlanStopsUnit ModelPlanStopsUnit
	solutionStops      []SolutionStop
}

func (p *solutionPlanStopsUnitImpl) String() string {
	return fmt.Sprintf("solutionPlanStopsUnit{%v, planned=%v}",
		p.modelPlanStopsUnit,
		p.IsPlanned(),
	)
}

func (p *solutionPlanStopsUnitImpl) SolutionStop(stop ModelStop) SolutionStop {
	return p.solutionStop(stop)
}

func (p *solutionPlanStopsUnitImpl) solutionStop(stop ModelStop) SolutionStop {
	for _, solutionStop := range p.solutionStops {
		if solutionStop.ModelStop().Index() == stop.Index() {
			return solutionStop
		}
	}
	panic(
		fmt.Errorf("solution stop for model stop %s [%v] not found in unit %v",
			stop.ID(),
			stop.Index(),
			p.modelPlanStopsUnit.Index(),
		),
	)
}

func (p *solutionPlanStopsUnitImpl) PlannedPlanStopsUnits() SolutionPlanStopsUnits {
	if p.IsPlanned() {
		return SolutionPlanStopsUnits{p}
	}
	return SolutionPlanStopsUnits{}
}

func (p *solutionPlanStopsUnitImpl) ModelPlanUnit() ModelPlanUnit {
	return p.modelPlanStopsUnit
}

func (p *solutionPlanStopsUnitImpl) ModelPlanStopsUnit() ModelPlanStopsUnit {
	return p.modelPlanStopsUnit
}

func (p *solutionPlanStopsUnitImpl) Index() int {
	return p.modelPlanStopsUnit.Index()
}

func (p *solutionPlanStopsUnitImpl) Solution() Solution {
	return p.solutionStops[0].Solution()
}

func (p *solutionPlanStopsUnitImpl) solution() *solutionImpl {
	return p.solutionStops[0].solution
}

func (p *solutionPlanStopsUnitImpl) Stops() ModelStops {
	return p.modelPlanStopsUnit.Stops()
}

func (p *solutionPlanStopsUnitImpl) SolutionStops() SolutionStops {
	solutionStops := make(SolutionStops, len(p.solutionStops))
	copy(solutionStops, p.solutionStops)
	return solutionStops
}

func (p *solutionPlanStopsUnitImpl) solutionStopsImpl() []SolutionStop {
	return p.solutionStops
}

func (p *solutionPlanStopsUnitImpl) IsPlanned() bool {
	if len(p.solutionStops) == 0 {
		return false
	}
	for _, solutionStop := range p.solutionStops {
		if !solutionStop.IsPlanned() {
			return false
		}
	}
	return true
}

func (p *solutionPlanStopsUnitImpl) IsFixed() bool {
	for _, solutionStop := range p.solutionStops {
		if solutionStop.ModelStop().IsFixed() {
			return true
		}
	}
	return false
}

func (p *solutionPlanStopsUnitImpl) UnPlan() (bool, error) {
	if !p.IsPlanned() || p.IsFixed() {
		return false, nil
	}

	solution := p.Solution().(*solutionImpl)

	solution.Model().OnUnPlan(p)

	if planUnitsUnit, isMemberOf := p.modelPlanStopsUnit.PlanUnitsUnit(); isMemberOf {
		solutionPlanUnitsUnit := solution.SolutionPlanUnit(planUnitsUnit)
		solution.plannedPlanUnits.remove(solutionPlanUnitsUnit)
		solution.unPlannedPlanUnits.add(solutionPlanUnitsUnit)
	} else {
		solution.plannedPlanUnits.remove(p)
		solution.unPlannedPlanUnits.add(p)
	}

	success, err := p.unplan()
	if err != nil {
		success = false
	}

	if success {
		solution.Model().OnUnPlanSucceeded(p)
	} else {
		if planUnitsUnit, isMemberOf := p.modelPlanStopsUnit.PlanUnitsUnit(); isMemberOf {
			solutionPlanUnitsUnit := solution.SolutionPlanUnit(planUnitsUnit)
			solution.unPlannedPlanUnits.remove(solutionPlanUnitsUnit)
			solution.plannedPlanUnits.add(solutionPlanUnitsUnit)
		} else {
			solution.unPlannedPlanUnits.remove(p)
			solution.plannedPlanUnits.add(p)
		}
		solution.Model().OnUnPlanFailed(p)
	}
	return success, err
}

func (p *solutionPlanStopsUnitImpl) StopPositions() StopPositions {
	if p.IsPlanned() {
		return common.Map(p.solutionStops, func(solutionStop SolutionStop) StopPosition {
			return newStopPosition(
				solutionStop.Previous(),
				solutionStop,
				solutionStop.Next(),
			)
		})
	}
	return StopPositions{}
}

var unplanSolutionMove = sync.Pool{
	New: func() any {
		return &solutionMoveStopsImpl{
			stopPositions: make([]StopPosition, 0, 64),
		}
	},
}

func (p *solutionPlanStopsUnitImpl) unplan() (bool, error) {
	solution := p.solutionStops[0].Solution().(*solutionImpl)

	// TODO: solutionStop.detach() modifies the solution so we have to
	// create the move here, even though we don't need it only if
	// isFeasible() returns a constraint.
	move := unplanSolutionMove.Get().(*solutionMoveStopsImpl)
	defer func() {
		move.stopPositions = move.stopPositions[:0]
		unplanSolutionMove.Put(move)
	}()
	move.planUnit = p
	move.value = 0.0
	move.valueSeen = 0
	move.allowed = true
	for _, solutionStop := range p.solutionStops {
		move.stopPositions = append(move.stopPositions, newStopPosition(
			solutionStop.Previous(),
			solutionStop,
			solutionStop.Next(),
		))
	}

	idx := p.solutionStops[0].PreviousIndex()
	for _, solutionStop := range p.solutionStops {
		solutionStop.detach()
	}

	constraint, _, err := solution.isFeasible(idx, true)
	if err != nil {
		return false, err
	}
	if constraint != nil {
		planned, err := move.Execute(context.Background())
		if err != nil {
			return false, err
		}
		if !planned {
			return false,
				fmt.Errorf(
					"failed undoing failed unplan",
				)
		}
		return false, nil
	}
	return true, nil
}
