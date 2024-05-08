// © 2019-present nextmv.io inc

// Package schema contains the core schemas for nextroute.
package schema

// Output is the output of the  check.
type Output struct {
	// Error is the error raised during the check.
	Error *string `json:"error,omitempty"`
	// Remark is the remark of the check. It can be "ok", "timeout" or
	// anything else that should explain itself.
	Remark string `json:"remark"`
	// Verbosity is the verbosity of the check.
	Verbosity string `json:"verbosity"`
	// Duration is the input maximum duration in seconds of the check.
	DurationMaximum float64 `json:"duration_maximum"`
	// DurationUsed is the duration in seconds used for the check.
	DurationUsed float64 `json:"duration_used"`
	// Solution is the start solution of the check.
	Solution Solution `json:"solution"`
	// Summary is the summary of the check.
	Summary Summary `json:"summary"`
	// PlanUnits is the check of the individual plan units.
	PlanUnits []PlanUnit `json:"plan_units"`
	// Vehicles is the check of the vehicles.
	Vehicles []Vehicle `json:"vehicles"`
}

// Solution is the solution the check has been executed on.
type Solution struct {
	// StopsPlanned is the number of stops planned in the start solution.
	StopsPlanned int `json:"stops_planned"`
	// PlanUnitsPlanned is the number of units planned in the start
	// solution.
	PlanUnitsPlanned int `json:"plan_units_planned"`
	// PlanUnitsUnplanned is the number of units unplanned in the start
	// solution.
	PlanUnitsUnplanned int `json:"plan_units_unplanned"`
	// VehiclesUsed is the number of vehicles used in the start solution.
	VehiclesUsed int `json:"vehicles_used"`
	// VehiclesNotUsed is the number of vehicles not used in the start solution,
	// the empty vehicles.
	VehiclesNotUsed int `json:"vehicles_not_used"`
	// Objective is the objective of the start solution.
	Objective Objective `json:"objective"`
}

// Summary is the summary of the check.
type Summary struct {
	// PlanUnitsToBeChecked is the number of plan units to be checked.
	PlanUnitsToBeChecked int `json:"plan_units_to_be_checked"`
	// PlanUnitsChecked is the number of plan units checked. If this is less
	// than [PlanUnitsToBeChecked] the check timed out.
	PlanUnitsChecked int `json:"plan_units_checked"`
	// PlanUnitsMoveFoundExecutable is the number of plan units for which at
	// least one move has been found and the move is executable.
	PlanUnitsBestMoveFound int `json:"plan_units_best_move_found"`
	// PlanUnitsHaveNoMove is the number of plan units for which one of the two
	// statements is true: 1) the constraint estimation determined that there is
	// no move or 2) the constraint estimation determined that there is a move
	// but when we tried to execute it, it was not executable. This implies
	// there is no move that can be executed without violating a constraint.
	PlanUnitsHaveNoMove int `json:"plan_units_have_no_move"`
	// NumberOfPlanUnitsMakingObjectiveWorse is the number of plan units for
	// which the best move is executable but would increase the objective value
	// instead of decreasing it.
	NumberOfPlanUnitsMakingObjectiveWorse int `json:"number_of_plan_units_making_objective_worse"`
	// PlanUnitsBestMoveFailed is the number of plan units for which the best
	// move can not be planned. This should not happen if all the constraints
	// are implemented correct.
	PlanUnitsBestMoveFailed int `json:"plan_units_best_move_failed"`
	// MovesFailed is the number of moves that failed. A move can fail if the
	// estimate of a constraint is incorrect. A constraint is incorrect if
	// [ModelConstraint.EstimateIsViolated] returns false and one of the
	// violation checks returns true. Violation checks are implementations of
	// one or more of the interfaces [SolutionStopViolationCheck],
	// [SolutionVehicleViolationCheck] or [SolutionViolationCheck] on the same
	// constraint. Most constraints do not need and do not have violation
	// checks as the estimate is perfect. The number of moves failed can be more
	// than one per plan unit as we continue to try moves on different vehicles
	// until we find a move that is executable or all vehicles have been
	// visited.
	MovesFailed int `json:"moves_failed"`
}

// PlanUnit is the check of a plan unit.
type PlanUnit struct {
	// ID is the ID of the plan unit. The ID of the plan unit is the slice of
	// ID's of the stops in the plan unit.
	Stops []string `json:"stops"`
	// HasPlannableBestMove is true if a move is found and we were able to
	// execute it successfully for the plan unit. A plan unit has no move found
	// if the plan unit is over-constrained or the move found is too expensive
	// or the constraints estimated it should be executable but when we tried to
	// execute it, it wasn't executable - this is an error in the constraint
	// estimation.
	HasPlannableBestMove bool `json:"has_plannable_best_move"`
	// PlanningMakesObjectiveWorse is true if the best move for the plan unit
	// increases the objective.
	PlanningMakesObjectiveWorse bool `json:"planning_makes_objective_worse"`
	// BestMoveFailed is true if the plan unit best move failed to execute.
	BestMoveFailed bool `json:"best_move_failed"`
	// VehiclesHaveMoves is the number of vehicles that have moves for the plan
	// unit. Only calculated if the verbosity is medium or higher.
	VehiclesHaveMoves *int `json:"vehicles_have_moves,omitempty"`
	// VehicleWithMoves is the ID of the vehicles that have moves for the plan
	// unit. Only calculated if the verbosity is very high.
	VehiclesWithMoves []*VehiclesWithMovesDetail `json:"vehicles_with_moves,omitempty"`
	// BestMoveObjective is the estimate of the objective of the best move if
	// the plan unit has a best move.
	BestMoveObjective *Objective `json:"best_move_objective,omitempty"`
	// Constraints is the constraints that are violated for the plan unit.
	Constraints map[string]int `json:"constraints,omitempty"`
}

// VehiclesWithMovesDetail shows details of the vehicles that have moves.
type VehiclesWithMovesDetail struct {
	// Vehicle is the ID of the vehicle.
	VehicleID string `json:"vehicle_id"`
	// DeltaObjectiveEstimate is the estimate of the delta of the objective of
	// that will be incurred by the move.
	DeltaObjectiveEstimate *float64 `json:"delta_objective_estimate,omitempty"`
	// DeltaObjective is the actual delta of the objective of that will be
	// incurred by the move.
	DeltaObjective *float64 `json:"delta_objective,omitempty"`
	// FailedConstraints are the constraints that are violated for the move.
	FailedConstraints []string `json:"failed_constraints,omitempty"`
	// WasPlannable is true if the move was plannable, false otherwise.
	WasPlannable bool `json:"was_plannable"`
	// Positions defines where the stop should be inserted.
	Positions []Position `json:"positions"`
}

// Position is the equivalent of a StopPosition. It expresses after which stop
// and before which other stop a stop should be inserted.
type Position struct {
	// Previous is the ID of the stop after which the stop should be inserted.
	Previous string `json:"previous"`
	// Stop is the ID of the stop to be inserted.
	Stop string `json:"stop"`
	// Next is the ID of the stop before which the stop should be inserted.
	Next string `json:"next"`
}

// ObjectiveTerm is the check of the individual terms of
// the objective for a move.
type ObjectiveTerm struct {
	// Name is the name of the objective term.
	Name string `json:"name"`
	// Factor is the factor of the objective term.
	Factor float64 `json:"factor"`
	// Base is the base value of the objective term.
	Base float64 `json:"base"`
	// Value is the value of the objective term which is the Factor times Base.
	Value float64 `json:"value"`
}

// Objective is the estimate of an objective of a move.
type Objective struct {
	// Vehicle is the ID of the vehicle for which it reports the objective.
	Vehicle *string `json:"vehicle,omitempty"`
	// Value is the value of the objective.
	Value float64 `json:"value"`
	// Terms is the check of the individual terms of the objective.
	Terms []ObjectiveTerm `json:"terms"`
}

// Vehicle is the check of a vehicle.
type Vehicle struct {
	// ID is the ID of the vehicle.
	ID string `json:"id"`
	// PlanUnitsHaveMoves is the number of plan units that have moves for the
	// vehicle. Only calculated if the depth is medium.
	PlanUnitsHaveMoves *int `json:"plan_units_have_moves,omitempty"`
}
