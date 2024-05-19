// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"

	"github.com/nextmv-io/nextroute/common"
)

// Move is a type alias for SolutionMoveStops. It is used to make the
// custom constraints and observers backward compatible.
type Move = SolutionMoveStops

// SolutionMoveStops is a move in a solution. A move is a change in the solution that
// can be executed. A move can be executed which may or may not result in a
// change in the solution. A move can be asked if it is executable, it is not
// executable if it is incomplete or if it is not executable because the move is
// not allowed. A move can be asked if it is an improvement, it is an
// improvement if it is executable and the move has a value less than zero.
// A move describes the change in the solution. The change in the solution
// is described by the stop positions. The stop positions describe for each
// stop in the associated units where in the existing solution the stop
// is supposed to be placed. The stop positions are ordered by the order
// of the stops in the unit. The first stop in the unit is the first
// stop in the stop positions. The last stop in the unit is the last
// stop in the stop positions.
type SolutionMoveStops interface {
	SolutionMove
	// Previous returns previous stop of the first to be planned
	// stop if it would be planned. Previous is the same stop as the
	// previous stop of the first stop position.
	Previous() SolutionStop

	// Next returns the next stop of the last to be planned
	// stop if it would be planned. Next is the same stop as the
	// next stop of the last stop position.
	Next() SolutionStop

	// PlanStopsUnit returns the [SolutionPlanStopsUnit] that is affected by the move.
	PlanStopsUnit() SolutionPlanStopsUnit

	// StopPositions returns the [StopPositions] that define the move and
	// how it will change the solution.
	StopPositions() StopPositions

	// StopPositionAt returns the stop position at the given index. The index
	// is the index is based on the positions in the move. Get the length of
	// the stop positions from [StopPositionsLength].
	StopPositionAt(index int) StopPosition

	// StopPositionsLength returns the length of the stop positions.
	StopPositionsLength() int

	// Vehicle returns the vehicle, if known, that is affected by the move. If
	// not known, nil is returned.
	Vehicle() SolutionVehicle

	// Solution returns the solution that is affected by the move.
	Solution() Solution
}

// StopPosition is the definition of the change in the solution for a
// specific stop. The change is defined by a Next and a Stop. The
// Next is a stop which is already part of the solution (it is planned)
// and the Stop is a stop which is not yet part of the solution (it is not
// planned). A stop position states that the stop should be moved from the
// unplanned set to the planned set by positioning it directly before the
// Next.
type StopPosition struct {
	solution          *solutionImpl
	previousStopIndex int
	stopIndex         int
	nextStopIndex     int
}

// StopPositions is a slice of stop positions.
type StopPositions []StopPosition

func newNotExecutableSolutionMoveStops(planUnit *solutionPlanStopsUnitImpl) *solutionMoveStopsImpl {
	return &solutionMoveStopsImpl{
		planUnit:  planUnit,
		valueSeen: 1,
		allowed:   false,
	}
}

type solutionMoveStopsImpl struct {
	planUnit      *solutionPlanStopsUnitImpl
	stopPositions []StopPosition
	valueSeen     int
	value         float64
	allowed       bool
}

// reset resets the move to its initial state.
// We assume the planUnit stays the same.
func (m *solutionMoveStopsImpl) reset() {
	m.stopPositions = m.stopPositions[:0]
	m.allowed = false
	m.value = 0.0
	m.valueSeen = 1
}

// replaceBy replaces the move by the new move.
// We assume the planunit and solution stays the same.
func (m *solutionMoveStopsImpl) replaceBy(newStop *solutionMoveStopsImpl, newValueSeen int) {
	m.reset()
	m.value = newStop.value
	m.valueSeen = newValueSeen
	m.allowed = newStop.allowed
	m.stopPositions = append(m.stopPositions, newStop.stopPositions...)
}

func (m *solutionMoveStopsImpl) String() string {
	return fmt.Sprintf("move{%v, vehicle=%v, %v, valueSeen=%v, value=%v, allowed=%v}",
		m.planUnit,
		m.Vehicle().Index(),
		m.stopPositions,
		m.valueSeen,
		m.value,
		m.allowed,
	)
}

func (m *solutionMoveStopsImpl) Solution() Solution {
	return m.planUnit.solution()
}

func (m *solutionMoveStopsImpl) Vehicle() SolutionVehicle {
	if len(m.stopPositions) == 0 {
		return SolutionVehicle{}
	}
	return m.stopPositions[len(m.stopPositions)-1].Next().Vehicle()
}

func (m *solutionMoveStopsImpl) vehicle() SolutionVehicle {
	return m.stopPositions[len(m.stopPositions)-1].Next().vehicle()
}

func (m *solutionMoveStopsImpl) Next() SolutionStop {
	if next, ok := m.next(); ok {
		return next
	}
	return SolutionStop{}
}

func (m *solutionMoveStopsImpl) next() (SolutionStop, bool) {
	if len(m.stopPositions) == 0 {
		return SolutionStop{}, false
	}
	return m.stopPositions[len(m.stopPositions)-1].Next(), true
}

func (m *solutionMoveStopsImpl) Previous() SolutionStop {
	previous, ok := m.previous()
	if !ok {
		return SolutionStop{}
	}
	return previous
}

func (m *solutionMoveStopsImpl) previous() (SolutionStop, bool) {
	if len(m.stopPositions) == 0 {
		return SolutionStop{}, false
	}
	return m.stopPositions[0].Previous(), true
}

func (m *solutionMoveStopsImpl) Execute(_ context.Context) (bool, error) {
	if !m.IsExecutable() {
		return false, nil
	}

	m.planUnit.solution().model.OnPlan(m)

	if _, isElementOfPlanUnitsUnit := m.planUnit.ModelPlanUnit().PlanUnitsUnit(); !isElementOfPlanUnitsUnit {
		m.planUnit.solution().unPlannedPlanUnits.remove(m.planUnit)
		m.planUnit.solution().plannedPlanUnits.add(m.planUnit)
	}

	startPropagate, err := m.attach()
	if err != nil {
		return false, err
	}

	constraint, _, err := m.planUnit.solution().isFeasible(startPropagate, true)
	if err != nil {
		return false, err
	}

	if constraint != nil {
		m.planUnit.solution().model.OnPlanFailed(m, constraint)
		if _, isElementOfPlanUnitsUnit := m.planUnit.ModelPlanUnit().PlanUnitsUnit(); !isElementOfPlanUnitsUnit {
			m.planUnit.solution().unPlannedPlanUnits.add(m.planUnit)
			m.planUnit.solution().plannedPlanUnits.remove(m.planUnit)
		}

		for _, position := range m.stopPositions {
			position.Stop().detach()
		}

		constraint, _, err := m.planUnit.solution().isFeasible(startPropagate, true)
		if err != nil {
			return false, err
		}

		if constraint != nil {
			return false, fmt.Errorf(
				"undoing failed solutionMoveStopsImpl %v failed: %v",
				m, constraint,
			)
		}

		return false, nil
	}
	m.planUnit.solution().model.OnPlanSucceeded(m)

	return true, nil
}

func (m *solutionMoveStopsImpl) attach() (int, error) {
	startPropagate := -1
	for i := len(m.stopPositions) - 1; i >= 0; i-- {
		stopPosition := m.stopPositions[i]
		m.planUnit.solutionStops[i] = stopPosition.Stop()
		beforeStop := stopPosition.Next()
		if stopPosition.Stop().IsPlanned() {
			return -1, fmt.Errorf(
				"stop %v is already planned",
				stopPosition.Stop(),
			)
		}
		if beforeStop.IsFirst() {
			return -1, fmt.Errorf(
				"nextStop %v is first",
				beforeStop,
			)
		}
		startPropagate = stopPosition.Stop().attach(
			beforeStop.PreviousIndex(),
		)
	}
	return startPropagate, nil
}

func (m *solutionMoveStopsImpl) PlanUnit() SolutionPlanUnit {
	return m.planUnit
}

func (m *solutionMoveStopsImpl) PlanStopsUnit() SolutionPlanStopsUnit {
	if m.planUnit == nil {
		return nil
	}
	return m.planUnit
}

func (m *solutionMoveStopsImpl) Value() float64 {
	return m.value
}

func (m *solutionMoveStopsImpl) ValueSeen() int {
	return m.valueSeen
}

func (m solutionMoveStopsImpl) IncrementValueSeen(inc int) SolutionMove {
	m.valueSeen += inc
	return &m
}

func (m *solutionMoveStopsImpl) StopPositions() StopPositions {
	stopPositions := make(StopPositions, len(m.stopPositions))
	copy(stopPositions, m.stopPositions)
	return stopPositions
}

func (m *solutionMoveStopsImpl) StopPositionAt(index int) StopPosition {
	return m.stopPositions[index]
}

func (m *solutionMoveStopsImpl) StopPositionsLength() int {
	return len(m.stopPositions)
}

func (m *solutionMoveStopsImpl) IsExecutable() bool {
	return m.stopPositions != nil &&
		!m.planUnit.IsPlanned() &&
		m.allowed &&
		!m.planUnit.IsFixed()
}

func (m *solutionMoveStopsImpl) IsImprovement() bool {
	return m.IsExecutable() && m.value < 0
}

func (m *solutionMoveStopsImpl) TakeBest(that SolutionMove) SolutionMove {
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
	if m.planUnit.solution().random.Intn(m.ValueSeen()+that.ValueSeen()) == 0 {
		m.valueSeen++
		return m
	}
	return that.IncrementValueSeen(m.ValueSeen())
}

// deltaStopTravelDurationValue computes the sume of deltaStopDurationValue()
// and deltaTravelDurationValue() in one pass.
// This is more efficient than calling the two functions separately.
// But only call it if the travel time is not dependent on time.
func (m *solutionMoveStopsImpl) deltaStopTravelDurationValue(
	vehicleType ModelVehicleType,
) float64 {
	if len(m.stopPositions) == 0 || m.stopPositions[0].Stop().IsPlanned() {
		return 0
	}
	deltaStopDurationValue := 0.0
	travelDuration := 0.0
	vehicleTravelDuration := vehicleType.TravelDurationExpression()
	vehicleDuration := vehicleType.DurationExpression()
	for _, stopPosition := range m.stopPositions {
		modelStop := stopPosition.Stop().ModelStop()
		nextStop := stopPosition.Next().ModelStop()
		previousStop := stopPosition.Previous().ModelStop()
		if stopPosition.Next().IsPlanned() {
			deltaStopDurationValue -= stopPosition.Next().DurationValue()
			travelDuration -= stopPosition.Next().TravelDurationValue()
			travelDuration += vehicleTravelDuration.Value(
				vehicleType,
				modelStop,
				nextStop,
			)
		}
		deltaStopDurationValue += vehicleDuration.Value(
			vehicleType,
			previousStop,
			modelStop,
		)
		deltaStopDurationValue += vehicleDuration.Value(
			vehicleType,
			modelStop,
			nextStop,
		)
		travelDuration += vehicleTravelDuration.Value(
			vehicleType,
			previousStop,
			modelStop,
		)
	}
	return deltaStopDurationValue + travelDuration
}

func (m *solutionMoveStopsImpl) deltaTravelDurationValue() float64 {
	if len(m.stopPositions) == 0 || m.stopPositions[0].Stop().IsPlanned() {
		return 0
	}

	vehicle := m.vehicle()

	vehicleType := vehicle.ModelVehicle().VehicleType()

	isDependentOnTime := vehicleType.TravelDurationExpression().IsDependentOnTime()

	if isDependentOnTime {
		if len(m.stopPositions) == 1 {
			solutionStop := m.stopPositions[0].Stop()
			previousStop, _ := m.previous()
			departure := previousStop.EndValue()
			fromDuration, _, _, _ := vehicleType.TemporalValues(
				departure,
				previousStop.ModelStop(),
				solutionStop.ModelStop(),
			)
			nextStop, _ := m.next()
			toDuration, _, _, _ := vehicleType.TemporalValues(
				departure,
				solutionStop.ModelStop(),
				nextStop.ModelStop(),
			)
			return fromDuration + toDuration - nextStop.TravelDurationValue()
		}

		newTravelDuration := 0.0

		generator := newSolutionStopGenerator(
			*m,
			false,
			isDependentOnTime,
		)

		previousStop, _ := generator.next()
		departure := previousStop.EndValue()

		for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
			travelDuration, _, _, end := vehicleType.TemporalValues(
				departure,
				previousStop.ModelStop(),
				solutionStop.ModelStop(),
			)

			newTravelDuration += travelDuration

			previousStop = solutionStop
			departure = end
		}

		currentTravelDuration := previousStop.CumulativeTravelDurationValue() -
			m.Previous().CumulativeTravelDurationValue()
		return newTravelDuration - currentTravelDuration
	}

	travelDuration := 0.0

	for _, stopPosition := range m.stopPositions {
		modelStop := stopPosition.Stop().ModelStop()
		if stopPosition.Next().IsPlanned() {
			travelDuration -= stopPosition.Next().TravelDurationValue()
			travelDuration += vehicleType.TravelDurationExpression().Value(
				vehicleType,
				modelStop,
				stopPosition.Next().ModelStop(),
			)
		}
		travelDuration += vehicleType.TravelDurationExpression().Value(
			vehicleType,
			stopPosition.Previous().ModelStop(),
			modelStop,
		)
	}

	return travelDuration
}

// NewMoveStops creates a new move and checks if the move is allowed and if so
// estimates the delta score.
func NewMoveStops(
	planUnit SolutionPlanStopsUnit,
	stopPositions StopPositions,
) (SolutionMoveStops, error) {
	return newMoveStops(planUnit, stopPositions, true)
}

// newMoveStops creates a new move.
func newMoveStops(
	planUnit SolutionPlanStopsUnit,
	stopPositions StopPositions,
	checkConstraintsAndEstimateDeltaScore bool,
) (SolutionMoveStops, error) {
	if planUnit == nil {
		return nil, fmt.Errorf("planUnit is nil")
	}
	if len(stopPositions) != len(planUnit.SolutionStops()) {
		return nil, fmt.Errorf(
			"stopPositions length %v must be equal to planUnit.SolutionStops length %v",
			len(stopPositions),
			len(planUnit.SolutionStops()),
		)
	}
	if !stopPositions[0].Previous().IsPlanned() {
		return nil, fmt.Errorf("previous stop %s of first stop position must be planned",
			stopPositions[0].Previous().ModelStop().ID(),
		)
	}
	if !stopPositions[len(stopPositions)-1].Next().IsPlanned() {
		return nil, fmt.Errorf("next stop %s of last stop positions must be planned",
			stopPositions[len(stopPositions)-1].Next().ModelStop().ID(),
		)
	}

	stops := common.Map(stopPositions, func(i StopPosition) ModelStop {
		return i.Stop().ModelStop()
	})

	allowed, err := planUnit.ModelPlanStopsUnit().DirectedAcyclicGraph().IsAllowed(stops)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, fmt.Errorf(
			"move is not allowed, " +
				"the stops are in a sequence violating the Directed Acyclic Graph of the plan unit",
		)
	}

	vehicle := stopPositions[0].Previous().vehicle()

	lastPlannedPreviousStop := stopPositions[0].Previous()

	position := stopPositions[0].Previous().Position()

	for index, sp := range stopPositions {
		stopPosition := sp
		if stopPosition.Stop().PlanStopsUnit() != planUnit {
			return nil,
				fmt.Errorf(
					"stop %v of stop position %v is not part of the provided plan unit",
					stopPosition.Stop().ModelStop().ID(),
					index,
				)
		}
		if stopPosition.Stop().IsPlanned() {
			return nil,
				fmt.Errorf(
					"stop %v of stop position %v is already planned",
					stopPosition.Stop().ModelStop().ID(),
					index,
				)
		}

		if stopPosition.Previous().IsPlanned() {
			if stopPosition.Previous().Position() < position {
				return nil,
					fmt.Errorf("previous stop %s of stop position %v is planned at position %v,"+
						" which is before or at the last planned previous stop %s at position %v",
						stopPosition.Previous().ModelStop().ID(),
						index,
						stopPosition.Previous().Position(),
						lastPlannedPreviousStop.ModelStop().ID(),
						lastPlannedPreviousStop.Position(),
					)
			}
			position = stopPosition.Previous().Position()

			lastPlannedPreviousStop = stopPosition.Previous()
		}
		if stopPosition.Next().IsPlanned() {
			if stopPosition.Next().Position() < position {
				return nil,
					fmt.Errorf("next stop %s of stop position %v is planned at position %v,"+
						" which is before or at the last planned previous stop %s at position %v",
						stopPosition.Next().ModelStop().ID(),
						index,
						stopPosition.Next().Position(),
						lastPlannedPreviousStop.ModelStop().ID(),
						lastPlannedPreviousStop.Position(),
					)
			}
			position = stopPosition.Next().Position()
		}
		if stopPosition.Next().IsPlanned() && !stopPosition.Previous().IsPlanned() {
			if lastPlannedPreviousStop.Position() != stopPosition.Next().Position()-1 {
				return nil,
					fmt.Errorf("stop positions are not allowed, planned previous stop %v is not adjacent"+
						" to the planned next stop %v of the next stop position",
						lastPlannedPreviousStop.ModelStop().ID(),
						stopPosition.Next().ModelStop().ID(),
					)
			}
		}

		if stopPosition.Next().IsPlanned() && stopPosition.Previous().IsPlanned() {
			if stopPosition.Next().Position() != stopPosition.Previous().Position()+1 {
				return nil,
					fmt.Errorf("stop positions are not allowed, planned previous stop %v is not adjacent"+
						" to the planned next stop %v of stop position %v",
						stopPosition.Previous().ModelStop().ID(),
						stopPosition.Next().ModelStop().ID(),
						index,
					)
			}
		}

		if !stopPosition.Previous().IsPlanned() {
			if stopPositions[index-1].Stop() != stopPosition.Previous() {
				return nil,
					fmt.Errorf("the previous stop %s of stop position %v"+
						" must be the stop %s of the previous stop position %v if it is unplanned",
						stopPosition.Previous().ModelStop().ID(),
						index,
						stopPositions[index-1].Stop().ModelStop().ID(),
						index-1,
					)
			}
		}

		if !stopPosition.Next().IsPlanned() {
			if stopPositions[index+1].Stop() != stopPosition.Next() {
				return nil,
					fmt.Errorf("the next stop %s of stop position %v"+
						" must be the stop %s of the next stop position %v if it is unplanned",
						stopPosition.Next().ModelStop().ID(),
						index,
						stopPositions[index+1].Stop().ModelStop().ID(),
						index+1,
					)
			}
		}

		if stopPosition.Previous().IsPlanned() && stopPosition.Previous().vehicle().index != vehicle.index {
			return nil,
				fmt.Errorf(
					"planned previous stop %v of stop position %v vehicle mismatch: %v != %v",
					stopPosition.Previous().ModelStop().ID(),
					index,
					stopPosition.Previous().vehicle().ModelVehicle().ID(),
					vehicle.ModelVehicle().ID(),
				)
		}
		if stopPosition.Next().IsPlanned() && stopPosition.Next().vehicle().index != vehicle.index {
			return nil,
				fmt.Errorf(
					"planned next stop %v of stop position %v vehicle mismatch: %v != %v",
					stopPosition.Next().ModelStop().ID(),
					index,
					stopPosition.Next().vehicle().ModelVehicle().ID(),
					vehicle.ModelVehicle().ID(),
				)
		}
	}

	stopPositionsImpl := make([]StopPosition, len(stopPositions))
	copy(stopPositionsImpl, stopPositions)
	move := &solutionMoveStopsImpl{
		planUnit:      planUnit.(*solutionPlanStopsUnitImpl),
		stopPositions: stopPositionsImpl,
		value:         0.0,
		valueSeen:     0,
		allowed:       true,
	}
	if checkConstraintsAndEstimateDeltaScore {
		value, allowed, _ := planUnit.(*solutionPlanStopsUnitImpl).solution().checkConstraintsAndEstimateDeltaScore(move)
		move.value = value
		move.allowed = allowed
		move.valueSeen = 1
	}
	return move, nil
}
