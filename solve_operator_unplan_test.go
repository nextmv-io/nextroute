// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

// TestUnPlanOperatorStopGroupConsistency is a regression test for the
// stop-group unplan bug. A stop-group becomes a PlanAll units-unit whose
// members must all be planned on the same vehicle. The Vehicle and Island
// unplan operators used to unplan individual member stops-units. Unplanning a
// single member marks the whole units-unit as unplanned (see
// solutionPlanStopsUnitImpl.UnPlan) while leaving the sibling members on the
// route. That leaves the solution in an inconsistent state: the group is
// "unplanned" yet some of its stops are still routed. The output builder
// (factory.toSolutionOutputStops) then lists every group stop as unplanned,
// so a stop appears BOTH on a vehicle route AND in the unplanned list.
//
// The invariant we assert: no stop that is currently routed may belong to a
// plan unit that sits in the unplanned collection.
func TestUnPlanOperatorStopGroupConsistency(t *testing.T) {
	// Try many seeds. The operators skip stops probabilistically, so a single
	// seed may happen to unplan the whole group; across many seeds a partial
	// unplan (the bug) is essentially certain with the old behaviour.
	const seeds = 60

	type opCase struct {
		name string
		run  func(op *solveOperatorUnPlanImpl, s Solution) error
	}
	cases := []opCase{
		{
			name: "Vehicle",
			run: func(op *solveOperatorUnPlanImpl, s Solution) error {
				_, err := op.unplanSomeStopsOfOneVehicle(s, 0.5)
				return err
			},
		},
		{
			name: "Island",
			run: func(op *solveOperatorUnPlanImpl, s Solution) error {
				_, err := op.unplanOneIsland(s, 3)
				return err
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			violations := 0
			for seed := int64(0); seed < seeds; seed++ {
				solution := buildPlannedGroupSolution(t, seed, 6, 6)

				op := &solveOperatorUnPlanImpl{}
				if err := c.run(op, solution); err != nil {
					t.Fatalf("seed %d: operator failed: %v", seed, err)
				}

				if viol := groupInvariantViolations(solution); len(viol) > 0 {
					violations++
					if violations <= 3 {
						t.Errorf(
							"seed %d: %d stop(s) are routed while their plan unit is unplanned: %v",
							seed, len(viol), viol,
						)
					}
				}
			}
			if violations > 0 {
				t.Errorf("%s operator: %d/%d seeds produced an inconsistent solution",
					c.name, violations, seeds)
			}
		})
	}
}

// groupInvariantViolations returns the ids of stops that are routed while the
// plan unit they belong to is in the unplanned collection. Mirrors the output
// builder: a PlanAll units-unit contributes all of its member stops.
func groupInvariantViolations(solution Solution) []string {
	routed := map[int]bool{}
	for _, v := range solution.Vehicles() {
		for _, s := range v.SolutionStops() {
			if !s.IsFirst() && !s.IsLast() && s.IsPlanned() {
				routed[s.ModelStop().Index()] = true
			}
		}
	}

	var out []string
	for _, u := range solution.UnPlannedPlanUnits().SolutionPlanUnits() {
		for _, ms := range planUnitModelStops(u) {
			if routed[ms.Index()] {
				out = append(out, ms.ID())
			}
		}
	}
	return out
}

// planUnitModelStops expands a solution plan unit to its member model stops the
// same way the output builder does (PlanAll units-units expand to all members).
func planUnitModelStops(u SolutionPlanUnit) ModelStops {
	switch v := u.(type) {
	case SolutionPlanStopsUnit:
		return v.ModelPlanStopsUnit().Stops()
	case SolutionPlanUnitsUnit:
		if !v.ModelPlanUnitsUnit().PlanAll() {
			return nil
		}
		var out ModelStops
		for _, child := range v.SolutionPlanUnits() {
			out = append(out, planUnitModelStops(child)...)
		}
		return out
	}
	return nil
}

// buildPlannedGroupSolution builds a single-vehicle model with one PlanAll
// same-vehicle stop-group of groupSize stops plus fillerCount ungrouped stops,
// then plans everything onto the vehicle. The solution's RNG is seeded so the
// subsequent operator call is deterministic.
func buildPlannedGroupSolution(t *testing.T, seed int64, groupSize, fillerCount int) Solution {
	t.Helper()

	model, err := NewModel()
	if err != nil {
		t.Fatal(err)
	}
	model.SetRandom(rand.New(rand.NewSource(seed)))

	serviceDuration := NewStopDurationExpression("serviceDuration", 0.0)
	vehicleType, err := model.NewVehicleType(
		NewTimeIndependentDurationExpression(
			NewTravelDurationExpression(
				NewHaversineExpression(),
				common.NewSpeed(10, common.MetersPerSecond),
			),
		),
		NewDurationExpression("travelDuration", serviceDuration, common.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Spread stops out on a small grid so they have distinct valid locations.
	lon, lat := 0.0, 0.0
	newStop := func(id string) ModelStop {
		lon += 0.01
		lat += 0.005
		loc, err := common.NewLocation(lon, lat)
		if err != nil {
			t.Fatal(err)
		}
		s, err := model.NewStop(loc)
		if err != nil {
			t.Fatal(err)
		}
		s.SetID(id)
		return s
	}

	// The stop group: groupSize single-stop units combined into a PlanAll
	// same-vehicle units-unit.
	groupUnits := make([]ModelPlanUnit, 0, groupSize)
	for i := 0; i < groupSize; i++ {
		u, err := model.NewPlanSingleStop(newStop(fmt.Sprintf("g%02d", i)))
		if err != nil {
			t.Fatal(err)
		}
		groupUnits = append(groupUnits, u)
	}
	if _, err := model.NewPlanAllPlanUnits(true, groupUnits...); err != nil {
		t.Fatal(err)
	}

	// Ungrouped filler stops.
	for i := 0; i < fillerCount; i++ {
		if _, err := model.NewPlanSingleStop(newStop(fmt.Sprintf("f%02d", i))); err != nil {
			t.Fatal(err)
		}
	}

	depot := newStop("depot")
	if _, err := model.NewVehicle(vehicleType, model.Epoch(), depot, depot); err != nil {
		t.Fatal(err)
	}

	solution, err := NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	// Plan everything we can onto the (single) vehicle.
	ctx := context.Background()
	for {
		unplanned := solution.UnPlannedPlanUnits().SolutionPlanUnits()
		if len(unplanned) == 0 {
			break
		}
		progress := false
		for _, u := range unplanned {
			move := solution.BestMove(ctx, u)
			if !move.IsExecutable() {
				continue
			}
			planned, err := move.Execute(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if planned {
				progress = true
			}
		}
		if !progress {
			break
		}
	}

	if solution.UnPlannedPlanUnits().Size() != 0 {
		t.Fatalf("seed %d: expected all units planned before unplan operator, %d remain",
			seed, solution.UnPlannedPlanUnits().Size())
	}

	return solution
}
