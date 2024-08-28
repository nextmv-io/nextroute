// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
)

func newSolutionMoveUnits(
	planUnit *solutionPlanUnitsUnitImpl,
	moves SolutionMoves,
) solutionMoveUnitsImpl {
	if len(moves) != len(planUnit.solutionPlanUnits) {
		panic(
			fmt.Sprintf("moves and SolutionPlanUnits must have the same length: %v != %v",
				len(moves),
				len(planUnit.solutionPlanUnits),
			),
		)
	}

	value := 0.0
	for _, move := range moves {
		value += move.Value()
	}

	return solutionMoveUnitsImpl{
		solution:  planUnit.Solution().(*solutionImpl),
		planUnit:  planUnit,
		moves:     moves,
		value:     value,
		valueSeen: 1,
		allowed:   true,
	}
}

func newNotExecutableSolutionMoveUnits(planUnit *solutionPlanUnitsUnitImpl) *solutionMoveUnitsImpl {
	return &solutionMoveUnitsImpl{
		solution:  planUnit.Solution().(*solutionImpl),
		planUnit:  planUnit,
		valueSeen: 1,
	}
}

type solutionMoveUnitsImpl struct {
	planUnit  *solutionPlanUnitsUnitImpl
	solution  *solutionImpl
	moves     SolutionMoves
	valueSeen int
	value     float64
	allowed   bool
}

func (m solutionMoveUnitsImpl) String() string {
	return fmt.Sprintf("move{%v, valueSeen=%v, value=%v, allowed=%v}",
		m.planUnit,
		m.valueSeen,
		m.value,
		m.allowed,
	)
}

func (m solutionMoveUnitsImpl) Execute(ctx context.Context) (bool, error) {
	if !m.IsExecutable() {
		return false, nil
	}

	m.solution.model.OnPlan(m)

	m.solution.unPlannedPlanUnits.remove(m.planUnit)
	m.solution.plannedPlanUnits.add(m.planUnit)

	for idx, move := range m.moves {
		if planned, err := move.Execute(ctx); err != nil || !planned {
			m.solution.unPlannedPlanUnits.add(m.planUnit)
			m.solution.plannedPlanUnits.remove(m.planUnit)

			for i := idx - 1; i >= 0; i-- {
				executedMove := m.moves[i]
				unPlanned, err := executedMove.PlanUnit().UnPlan()
				if err != nil || !unPlanned {
					return false, fmt.Errorf(
						"failed to unplan %v: %w",
						executedMove.PlanUnit(),
						err,
					)
				}
			}
			return false, err
		}
	}

	return true, nil
}

func (m solutionMoveUnitsImpl) PlanUnit() SolutionPlanUnit {
	return m.planUnit
}

func (m solutionMoveUnitsImpl) PlanUnitsUnit() SolutionPlanUnitsUnit {
	if m.planUnit == nil {
		return nil
	}
	return m.planUnit
}

func (m solutionMoveUnitsImpl) Value() float64 {
	return m.value
}

func (m solutionMoveUnitsImpl) ValueSeen() int {
	return m.valueSeen
}

func (m solutionMoveUnitsImpl) IncrementValueSeen(inc int) SolutionMove {
	m.valueSeen += inc
	return m
}

func (m solutionMoveUnitsImpl) IsExecutable() bool {
	return m.moves != nil &&
		!m.planUnit.IsPlanned() &&
		m.allowed &&
		!m.planUnit.IsFixed()
}

func (m solutionMoveUnitsImpl) IsImprovement() bool {
	return m.IsExecutable() && m.value < 0
}

func (m solutionMoveUnitsImpl) TakeBest(that SolutionMove) SolutionMove {
	if !that.IsExecutable() {
		return m
	}
	if !m.IsExecutable() {
		return that
	}
	if m.value > that.Value() {
		return that
	}
	if m.value < that.Value() {
		return m
	}
	if m.solution.random.Intn(m.ValueSeen()+that.ValueSeen()) == 0 {
		m.valueSeen++
		return m
	}
	return that.IncrementValueSeen(m.ValueSeen())
}
