package check

import "strings"

// Verbosity is the verbosity of the check.
type Verbosity int

const (
	// Off does not run the check.
	Off Verbosity = iota
	// Low checks if there is at least one move per plan unit.
	Low
	// Medium checks the number of moves per plan unit and the
	// number of vehicles that have moves. It also reports the number of
	// constraints that are violated for each plan unit if it does not fit
	// on any vehicle.
	Medium
	// High is identical to medium.
	High
)

// ToVerbosity converts a string to a verbosity. The string can be
// anything that starts with "o", "l", "m", "h" or "v" case-insensitive.
// If the string does not start with one of these characters the
// verbosity is off.
func ToVerbosity(s string) Verbosity {
	ls := strings.ToLower(s)
	if strings.HasPrefix(ls, "o") {
		return Off
	}
	if strings.HasPrefix(ls, "l") {
		return Low
	}
	if strings.HasPrefix(ls, "m") {
		return Medium
	}
	if strings.HasPrefix(ls, "h") {
		return High
	}
	return Off
}

// String returns the string representation of the verbosity.
func (v Verbosity) String() string {
	switch v {
	case Off:
		return "off"
	case Low:
		return "low"
	case Medium:
		return "medium"
	case High:
		return "high"
	default:
		return "unknown"
	}
}

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
	// PlanUnitsHaveNoMove is the number of plan units for which no feasible
	// move has been found. This implies there is no move that can be executed
	// without violating a constraint.
	PlanUnitsHaveNoMove int `json:"plan_units_have_no_move"`
	// PlanUnitsMoveFoundTooExpensive is the number of plan units for which the
	// best move is executable but would increase the objective value instead
	// of decreasing it.
	PlanUnitsBestMoveIncreasesObjective int `json:"plan_units_best_move_increases_objective"`
	// PlanUnitsBestMoveFailed is the number of plan units for which the best
	// move can not be planned. This should not happen if all the constraints
	// are implemented correct.
	PlanUnitsBestMoveFailed int `json:"plan_units_best_move_failed"`
	// MovesFailed is the number of moves that failed. A move can fail if the
	// estimate of a constraint is incorrect. A constraint is incorrect if
	// [ModelConstraint.EstimateIsViolated] returns true and one of the
	// violation checks returns false. Violation checks are implementations of
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
	// HasBestMove is true if a move is found for the plan unit. A plan unit
	// has no move found if the plan unit is over-constrained or the move found
	// is too expensive.
	HasBestMove bool `json:"has_best_move"`
	// BestMoveIncreasesObjective is true if the best move for the plan unit
	// increases the objective.
	BestMoveIncreasesObjective bool `json:"best_move_increases_objective"`
	// BestMoveFailed is true if the plan unit best move failed to execute.
	BestMoveFailed bool `json:"best_move_failed"`
	// VehiclesHaveMoves is the number of vehicles that have moves for the plan
	// unit. Only calculated if the verbosity is medium or higher.
	VehiclesHaveMoves *int `json:"vehicles_have_moves,omitempty"`
	// VehicleWithMoves is the ID of the vehicles that have moves for the plan
	// unit. Only calculated if the verbosity is very high.
	VehiclesWithMoves *[]string `json:"vehicles_with_moves,omitempty"`
	// BestMoveObjective is the estimate of the objective of the best move if
	// the plan unit has a best move.
	BestMoveObjective *Objective `json:"best_move_objective"`

	// Constraints is the constraints that are violated for the plan unit.
	Constraints *map[string]int `json:"constraints,omitempty"`
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
