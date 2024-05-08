// © 2019-present nextmv.io inc

package check

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/check/schema"
	"github.com/nextmv-io/nextroute/common"
)

// ModelCheck is the check of a model returning a [Output].
func ModelCheck(
	model nextroute.Model,
	options Options,
) (schema.Output, error) {
	if model == nil {
		return schema.Output{}, fmt.Errorf("model is nil")
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		return schema.Output{}, err
	}

	err = removePlanUnits(solution)
	if err != nil {
		return schema.Output{}, err
	}

	return SolutionCheck(
		solution,
		options,
	)
}

// SolutionCheck is the check of a solution returning a [Output].
func SolutionCheck(
	solution nextroute.Solution,
	options Options,
) (schema.Output, error) {
	if solution == nil {
		return schema.Output{}, fmt.Errorf("solution is nil")
	}

	verbosity := ToVerbosity(options.Verbosity)
	if verbosity == Off || options.Duration.Seconds() == 0 {
		return schema.Output{}, nil
	}
	if int(verbosity) < int(Low) {
		return schema.Output{}, fmt.Errorf(
			"verbosity [%d] is not supported",
			verbosity,
		)
	}

	nextCheck := &checkImpl{
		solution:  solution,
		verbosity: verbosity,
		output: schema.Output{
			DurationMaximum: options.Duration.Seconds(),
			Verbosity:       verbosity.String(),
			Remark:          "completed",
			Summary: schema.Summary{
				PlanUnitsToBeChecked: len(solution.Model().PlanUnits()),
				PlanUnitsChecked:     0,
			},
			PlanUnits: make([]schema.PlanUnit, 0, len(solution.Model().PlanUnits())),
			Vehicles:  make([]schema.Vehicle, 0, len(solution.Model().Vehicles())),
			Solution: schema.Solution{
				StopsPlanned:       0,
				PlanUnitsUnplanned: 0,
				VehiclesUsed:       0,
				VehiclesNotUsed:    0,
			},
		},
	}

	nextCheck.Check()

	return nextCheck.output, nil
}

type checkImpl struct {
	solution  nextroute.Solution
	output    schema.Output
	verbosity Verbosity
}

func (m *checkImpl) checkStartSolution() {
	if m.solution == nil {
		return
	}

	for _, plannedPlanUnit := range m.solution.PlannedPlanUnits().SolutionPlanUnits() {
		for _, plannedPlanStopsUnit := range plannedPlanUnit.PlannedPlanStopsUnits() {
			m.output.Solution.StopsPlanned += len(plannedPlanStopsUnit.SolutionStops())
		}
	}

	m.output.Solution.PlanUnitsUnplanned = m.solution.UnPlannedPlanUnits().Size()
	m.output.Solution.PlanUnitsPlanned = m.solution.PlannedPlanUnits().Size()

	m.output.Solution.VehiclesUsed = len(common.Filter(
		m.solution.Vehicles(),
		func(v nextroute.SolutionVehicle) bool {
			return !v.IsEmpty()
		}),
	)

	m.output.Solution.VehiclesNotUsed = len(common.Filter(
		m.solution.Vehicles(),
		func(v nextroute.SolutionVehicle) bool {
			return v.IsEmpty()
		}),
	)

	m.output.Solution.Objective.Terms = make(
		[]schema.ObjectiveTerm,
		0,
		len(m.solution.Model().Objective().Terms()),
	)

	m.output.Solution.Objective.Value = m.solution.Score()

	for _, term := range m.solution.Model().Objective().Terms() {
		value := m.solution.ObjectiveValue(term.Objective())
		m.output.Solution.Objective.Terms = append(
			m.output.Solution.Objective.Terms,
			schema.ObjectiveTerm{
				Name:   fmt.Sprintf("%v", term.Objective()),
				Factor: term.Factor(),
				Base:   value / term.Factor(),
				Value:  value,
			},
		)
	}
}

func (m *checkImpl) checkSolutionPlanUnits(
	ctx context.Context,
	solutionPlanUnits nextroute.SolutionPlanUnits,
) error {
	m.checkStartSolution()

	m.output.Summary.PlanUnitsToBeChecked = len(solutionPlanUnits)
	m.output.Summary.PlanUnitsChecked = 0

	if m.output.Summary.PlanUnitsToBeChecked == 0 {
		return nil
	}

	if int(m.verbosity) >= int(Medium) {
		for vIdx, solutionVehicle := range m.solution.Vehicles() {
			m.output.Vehicles = append(
				m.output.Vehicles,
				schema.Vehicle{
					ID:                 solutionVehicle.ModelVehicle().ID(),
					PlanUnitsHaveMoves: nil,
				},
			)
			planUnitsHaveMoves := 0
			m.output.Vehicles[vIdx].PlanUnitsHaveMoves = &planUnitsHaveMoves
		}
	}

	planUnitsObserver := newObserver()
	moveObserver := newObserver()

	if int(m.verbosity) >= int(Medium) {
		defer func() {
			m.solution.Model().RemoveSolutionObserver(planUnitsObserver)
			m.solution.Model().RemoveSolutionObserver(moveObserver)
		}()

		m.solution.Model().AddSolutionObserver(planUnitsObserver)
		m.solution.Model().AddSolutionObserver(moveObserver)
	}

SolutionPlanUnitLoop:
	for solutionPlanUnitIdx, solutionPlanUnit := range solutionPlanUnits {
		select {
		case <-ctx.Done():
			m.output.Remark = "timeout"
			break SolutionPlanUnitLoop
		default:
			planUnitsObserver.Reset()

			m.output.PlanUnits = append(
				m.output.PlanUnits,
				schema.PlanUnit{
					Stops:             toID(solutionPlanUnit.ModelPlanUnit()),
					VehiclesHaveMoves: nil,
					VehiclesWithMoves: nil,
				},
			)

			if int(m.verbosity) >= int(Medium) {
				vehiclesHaveMoves := 0
				m.output.PlanUnits[solutionPlanUnitIdx].VehiclesHaveMoves = &vehiclesHaveMoves
			}
			if int(m.verbosity) >= int(High) {
				m.output.PlanUnits[solutionPlanUnitIdx].VehiclesWithMoves = []*schema.VehiclesWithMovesDetail{}
			}

			moveMinimumValue := math.MaxFloat64
			moveIsImprovement := false
			movesFailed := false
		VehicleLoop:
			for solutionVehicleIdx, solutionVehicle := range m.solution.Vehicles() {
				bestMove := solutionVehicle.BestMove(ctx, solutionPlanUnit)

				if !bestMove.IsExecutable() {
					continue
				}

				if m.output.PlanUnits[solutionPlanUnitIdx].VehiclesHaveMoves != nil {
					*m.output.PlanUnits[solutionPlanUnitIdx].VehiclesHaveMoves++
				}

				value := bestMove.Value()
				vehicleDetails := &schema.VehiclesWithMovesDetail{
					VehicleID:              solutionVehicle.ModelVehicle().ID(),
					DeltaObjectiveEstimate: &value,
					FailedConstraints:      []string{},
				}

				if solutionMoveStops, ok := bestMove.(nextroute.SolutionMoveStops); ok {
					stopPositions := solutionMoveStops.StopPositions()
					vehicleDetails.Positions = make([]schema.Position, len(stopPositions))
					for stopIdx, stopPosition := range stopPositions {
						vehicleDetails.Positions[stopIdx] = schema.Position{
							Stop:     stopPosition.Stop().ModelStop().ID(),
							Previous: stopPosition.Previous().ModelStop().ID(),
							Next:     stopPosition.Next().ModelStop().ID(),
						}
					}
				}

				if m.output.PlanUnits[solutionPlanUnitIdx].VehiclesWithMoves != nil {
					m.output.PlanUnits[solutionPlanUnitIdx].VehiclesWithMoves = append(
						m.output.PlanUnits[solutionPlanUnitIdx].VehiclesWithMoves,
						vehicleDetails,
					)
				}

				if bestMove.IsImprovement() {
					actualScoreBeforeMove := m.solution.Score()
					moveObserver.Reset()

					planned, err := bestMove.Execute(ctx)
					if err != nil {
						return err
					}

					for _, constraint := range moveObserver.OnPlanFailedConstraints() {
						name := fmt.Sprintf("%v", constraint)
						vehicleDetails.FailedConstraints = append(
							vehicleDetails.FailedConstraints,
							name,
						)
					}

					if planned {
						moveIsImprovement = true
						vehicleDetails.WasPlannable = true
						deltaObjective := m.solution.Score() - actualScoreBeforeMove
						vehicleDetails.DeltaObjective = &deltaObjective

						m.output.PlanUnits[solutionPlanUnitIdx].HasPlannableBestMove = true

						if len(m.output.Vehicles) > solutionVehicleIdx {
							*m.output.Vehicles[solutionVehicleIdx].PlanUnitsHaveMoves++
						}

						if unplanned, err := solutionPlanUnit.UnPlan(); err != nil || !unplanned {
							return err
						}

						if int(m.verbosity) >= int(Medium) && bestMove.Value() < moveMinimumValue {
							moveMinimumValue = bestMove.Value()
							id := solutionVehicle.ModelVehicle().ID()
							nextCheckObjective := schema.Objective{
								Vehicle: &id,
								Value:   bestMove.Value(),
								Terms:   make([]schema.ObjectiveTerm, len(m.solution.Model().Objective().Terms())),
							}
							if solutionMoveStops, ok := bestMove.(nextroute.SolutionMoveStops); ok {
								for termIdx, term := range m.solution.Model().Objective().Terms() {
									base := term.Objective().EstimateDeltaValue(solutionMoveStops)
									nextCheckObjective.Terms[termIdx] = schema.ObjectiveTerm{
										Name:   fmt.Sprintf("%v", term.Objective()),
										Factor: term.Factor(),
										Base:   base,
										Value:  term.Factor() * base,
									}
								}
							} else {
								nextCheckObjective.Terms = make([]schema.ObjectiveTerm, 0)
							}

							nextCheckObjective.Value += bestMove.Value()
							m.output.PlanUnits[solutionPlanUnitIdx].BestMoveObjective = &nextCheckObjective
						}

						if m.verbosity == Low {
							break VehicleLoop
						}
					} else {
						// this hints at an inconsistency in the constraint
						// estimation compared to the constraint validation (if
						// available)
						m.output.Summary.MovesFailed++
						m.output.PlanUnits[solutionPlanUnitIdx].BestMoveFailed = true
						if !movesFailed {
							movesFailed = true
							m.output.Summary.PlanUnitsBestMoveFailed++
						}
					}
				}
			}

			if m.output.PlanUnits[solutionPlanUnitIdx].HasPlannableBestMove {
				m.output.Summary.PlanUnitsBestMoveFound++
				if !moveIsImprovement {
					m.output.PlanUnits[solutionPlanUnitIdx].PlanningMakesObjectiveWorse = true
					m.output.Summary.NumberOfPlanUnitsMakingObjectiveWorse++
				}
			} else {
				m.output.Summary.PlanUnitsHaveNoMove++
				constraints := make(map[string]int, len(m.solution.Model().Constraints()))
				m.output.PlanUnits[solutionPlanUnitIdx].Constraints = constraints
				for _, constraint := range planUnitsObserver.Constraints() {
					name := fmt.Sprintf("%v", constraint)
					if _, ok := constraints[name]; !ok {
						constraints[name] = 0
					}
					constraints[name]++
				}
			}
		}
		m.output.Summary.PlanUnitsChecked++
	}

	return nil
}

func (m *checkImpl) Check() {
	start := time.Now()

	localCtx, cancelFn := context.WithDeadline(
		context.Background(),
		start.Add(time.Duration(m.output.DurationMaximum*float64(time.Second))),
	)

	defer func() { cancelFn() }()

	err := m.checkSolutionPlanUnits(
		localCtx,
		m.solution.UnPlannedPlanUnits().SolutionPlanUnits(),
	)

	if err != nil {
		errorStr := err.Error()
		m.output.Error = &errorStr
	}

	m.output.DurationUsed = time.Since(start).Seconds()
}

// toID returns the IDs of the stops of the plan unit.
func toID(modelPlanUnit nextroute.ModelPlanUnit) []string {
	if modelPlanStopsUnit, ok := modelPlanUnit.(nextroute.ModelPlanStopsUnit); ok {
		return common.MapSlice(
			modelPlanStopsUnit.Stops(),
			func(stop nextroute.ModelStop) []string {
				return []string{stop.ID()}
			})
	}
	if modelPlanUnitsUnit, ok := modelPlanUnit.(nextroute.ModelPlanUnitsUnit); ok {
		return common.MapSlice(
			modelPlanUnitsUnit.PlanUnits(),
			toID,
		)
	}
	return []string{"unknown plan unit"}
}

func removePlanUnits(solution nextroute.Solution) error {
	if solution == nil {
		return fmt.Errorf("solution is nil")
	}
	for _, solutionPlanUnit := range solution.PlannedPlanUnits().SolutionPlanUnits() {
		if solutionPlanUnit.IsFixed() {
			continue
		}
		unplanned, err := solutionPlanUnit.UnPlan()
		if err != nil {
			return err
		}
		if !unplanned {
			return fmt.Errorf(
				"planned plan unit [%s] is not fixed and cannot be unplanned after initial solution creation",
				toID(solutionPlanUnit.ModelPlanUnit()),
			)
		}
	}
	return nil
}
