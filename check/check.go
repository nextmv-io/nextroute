package check

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
	"github.com/nextmv-io/sdk/nextroute/check"
)

// ModelCheck is the check of a model returning a [Output].
func ModelCheck(
	model nextroute.Model,
	options check.Options,
) (check.Output, error) {
	if model == nil {
		return check.Output{}, fmt.Errorf("model is nil")
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		return check.Output{}, err
	}

	err = removePlanUnits(solution)
	if err != nil {
		return check.Output{}, err
	}

	return SolutionCheck(
		solution,
		options,
	)
}

// SolutionCheck is the check of a solution returning a [Output].
func SolutionCheck(
	solution nextroute.Solution,
	options check.Options,
) (check.Output, error) {
	if solution == nil {
		return check.Output{}, fmt.Errorf("solution is nil")
	}

	verbosity := check.ToVerbosity(options.Verbosity)
	if verbosity == check.Off || options.Duration.Seconds() == 0 {
		return check.Output{}, nil
	}
	if int(verbosity) < int(check.Low) {
		return check.Output{}, fmt.Errorf(
			"verbosity [%d] is not supported",
			verbosity,
		)
	}

	nextCheck := &checkImpl{
		solution:  solution,
		verbosity: verbosity,
		output: check.Output{
			DurationMaximum: options.Duration.Seconds(),
			Verbosity:       verbosity.String(),
			Remark:          "completed",
			Summary: check.Summary{
				PlanUnitsToBeChecked: len(solution.Model().PlanUnits()),
				PlanUnitsChecked:     0,
			},
			PlanUnits: make([]check.PlanUnit, 0, len(solution.Model().PlanUnits())),
			Vehicles:  make([]check.Vehicle, 0, len(solution.Model().Vehicles())),
			Solution: check.Solution{
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
	verbosity check.Verbosity
	output    check.Output
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
		[]check.ObjectiveTerm,
		0,
		len(m.solution.Model().Objective().Terms()),
	)

	m.output.Solution.Objective.Value = m.solution.Score()

	for _, term := range m.solution.Model().Objective().Terms() {
		value := m.solution.ObjectiveValue(term.Objective())
		m.output.Solution.Objective.Terms = append(
			m.output.Solution.Objective.Terms,
			check.ObjectiveTerm{
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

	if int(m.verbosity) >= int(check.Medium) {
		for vIdx, solutionVehicle := range m.solution.Vehicles() {
			m.output.Vehicles = append(
				m.output.Vehicles,
				check.Vehicle{
					ID:                 solutionVehicle.ModelVehicle().ID(),
					PlanUnitsHaveMoves: nil,
				},
			)
			planUnitsHaveMoves := 0
			m.output.Vehicles[vIdx].PlanUnitsHaveMoves = &planUnitsHaveMoves
		}
	}

	observer := newObserver()

	if int(m.verbosity) >= int(check.Medium) {
		defer func() {
			m.solution.Model().RemoveSolutionObserver(observer)
		}()

		m.solution.Model().AddSolutionObserver(observer)
	}

SolutionPlanUnitLoop:
	for solutionPlanUnitIdx, solutionPlanUnit := range solutionPlanUnits {
		select {
		case <-ctx.Done():
			m.output.Remark = "timeout"
			break SolutionPlanUnitLoop
		default:
			observer.Reset()

			m.output.PlanUnits = append(
				m.output.PlanUnits,
				check.PlanUnit{
					Stops:             toID(solutionPlanUnit.ModelPlanUnit()),
					VehiclesHaveMoves: nil,
					VehiclesWithMoves: nil,
				},
			)

			if int(m.verbosity) >= int(check.Medium) {
				vehiclesHaveMoves := 0
				m.output.PlanUnits[solutionPlanUnitIdx].VehiclesHaveMoves = &vehiclesHaveMoves
			}
			if int(m.verbosity) >= int(check.High) {
				m.output.PlanUnits[solutionPlanUnitIdx].VehiclesWithMoves = &[]string{}
			}

			moveMinimumValue := math.MaxFloat64
			moveIsImprovement := false
			movesFailed := false
		VehicleLoop:
			for solutionVehicleIdx, solutionVehicle := range m.solution.Vehicles() {
				move := solutionVehicle.BestMove(ctx, solutionPlanUnit)
				if !move.IsExecutable() {
					continue
				}

				if m.output.PlanUnits[solutionPlanUnitIdx].VehiclesHaveMoves != nil {
					*m.output.PlanUnits[solutionPlanUnitIdx].VehiclesHaveMoves++
				}

				if m.output.PlanUnits[solutionPlanUnitIdx].VehiclesWithMoves != nil {
					*m.output.PlanUnits[solutionPlanUnitIdx].VehiclesWithMoves = append(
						*m.output.PlanUnits[solutionPlanUnitIdx].VehiclesWithMoves,
						solutionVehicle.ModelVehicle().ID(),
					)
				}

				if move.IsImprovement() {
					executed, err := move.Execute(ctx)
					if err != nil {
						return err
					}
					if executed {
						moveIsImprovement = true

						m.output.PlanUnits[solutionPlanUnitIdx].HasBestMove = true

						if len(m.output.Vehicles) > solutionVehicleIdx {
							*m.output.Vehicles[solutionVehicleIdx].PlanUnitsHaveMoves++
						}

						if unplanned, err := solutionPlanUnit.UnPlan(); err != nil || !unplanned {
							return err
						}

						if int(m.verbosity) >= int(check.Medium) && move.Value() < moveMinimumValue {
							moveMinimumValue = move.Value()
							id := solutionVehicle.ModelVehicle().ID()
							nextCheckObjective := check.Objective{
								Vehicle: &id,
								Value:   move.Value(),
								Terms:   make([]check.ObjectiveTerm, len(m.solution.Model().Objective().Terms())),
							}
							if solutionMoveStops, ok := move.(nextroute.SolutionMoveStops); ok {
								for termIdx, term := range m.solution.Model().Objective().Terms() {
									base := term.Objective().EstimateDeltaValue(solutionMoveStops)
									nextCheckObjective.Terms[termIdx] = check.ObjectiveTerm{
										Name:   fmt.Sprintf("%v", term.Objective()),
										Factor: term.Factor(),
										Base:   base,
										Value:  term.Factor() * base,
									}
								}
							} else {
								nextCheckObjective.Terms = make([]check.ObjectiveTerm, 0)
							}

							nextCheckObjective.Value += move.Value()
							m.output.PlanUnits[solutionPlanUnitIdx].BestMoveObjective = &nextCheckObjective
						}

						if m.verbosity == check.Low {
							break VehicleLoop
						}
					} else {
						movesFailed = true
						m.output.Summary.MovesFailed++
					}
				}
			}

			if m.output.PlanUnits[solutionPlanUnitIdx].HasBestMove {
				m.output.Summary.PlanUnitsBestMoveFound++
				if !moveIsImprovement {
					m.output.PlanUnits[solutionPlanUnitIdx].BestMoveIncreasesObjective = true
					m.output.Summary.PlanUnitsBestMoveIncreasesObjective++
				}
			} else {
				if movesFailed {
					m.output.PlanUnits[solutionPlanUnitIdx].BestMoveFailed = true
				}
				m.output.Summary.PlanUnitsHaveNoMove++
				constraints := make(map[string]int, len(m.solution.Model().Constraints()))
				m.output.PlanUnits[solutionPlanUnitIdx].Constraints = &constraints
				for _, constraint := range observer.Constraints() {
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
