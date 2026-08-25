# © 2019-present nextmv.io inc

"""
This module contains definitions for the schema in the Nextroute check.
"""


from nextroute.base_model import BaseModel


class ObjectiveTerm(BaseModel):
    """Check of the individual terms of the objective for a move."""

    base: float | None = None
    """Base of the objective term."""
    factor: float | None = None
    """Factor of the objective term."""
    name: str | None = None
    """Name of the objective term."""
    value: float | None = None
    """Value of the objective term, which is equivalent to `self.base *
    self.factor`."""


class Objective(BaseModel):
    """Estimate of an objective of a move."""

    terms: list[ObjectiveTerm] | None = None
    """Check of the individual terms of the objective."""
    value: float | None = None
    """Value of the objective."""
    vehicle: str | None = None
    """ID of the vehicle for which it reports the objective."""


class Solution(BaseModel):
    """Solution that the check has been executed on."""

    objective: Objective | None = None
    """Objective of the start solution."""
    plan_units_planned: int | None = None
    """Number of plan units planned in the start solution."""
    plan_units_unplanned: int | None = None
    """Number of plan units unplanned in the start solution."""
    stops_planned: int | None = None
    """Number of stops planned in the start solution."""
    vehicles_not_used: int | None = None
    """Number of vehicles not used in the start solution."""
    vehicles_used: int | None = None
    """Number of vehicles used in the start solution."""


class Summary(BaseModel):
    """Summary of the check."""

    moves_failed: int | None = None
    """number of moves that failed. A move can fail if the estimate of a
    constraint is incorrect. A constraint is incorrect if `ModelConstraint.
    EstimateIsViolated` returns true and one of the violation checks returns
    false. Violation checks are implementations of one or more of the
    interfaces [SolutionStopViolationCheck], [SolutionVehicleViolationCheck] or
    [SolutionViolationCheck] on the same constraint. Most constraints do not
    need and do not have violation checks as the estimate is perfect. The
    number of moves failed can be more than one per plan unit as we continue to
    try moves on different vehicles until we find a move that is executable or
    all vehicles have been visited."""
    plan_units_best_move_failed: int | None = None
    """Number of plan units for which the best move can not be planned. This
    should not happen if all the constraints are implemented correct."""
    plan_units_best_move_found: int | None = None
    """Number of plan units for which at least one move has been found and the
    move is executable."""
    plan_units_best_move_increases_objective: int | None = None
    """Number of plan units for which the best move is executable but would
    increase the objective value instead of decreasing it."""
    plan_units_checked: int | None = None
    """Number of plan units that have been checked. If this is less than
    `self.plan_units_to_be_checked` the check timed out."""
    plan_units_have_no_move: int | None = None
    """Number of plan units for which no feasible move has been found. This
    implies there is no move that can be executed without violating a
    constraint."""
    plan_units_to_be_checked: int | None = None
    """Number of plan units to be checked."""


class PlanUnit(BaseModel):
    """Check of a plan unit."""

    best_move_failed: bool | None = None
    """True if the plan unit's best move failed to execute."""
    best_move_increases_objective: bool | None = None
    """True if the best move for the plan unit increases the objective."""
    best_move_objective: Objective | None = None
    """Estimate of the objective of the best move if the plan unit has a best
    move."""
    constraints: dict[str, int] | None = None
    """Constraints that are violated for the plan unit."""
    has_best_move: bool | None = None
    """True if a move is found for the plan unit. A plan unit has no move found
    if the plan unit is over-constrained or the move found is too expensive."""
    stops: list[str] | None = None
    """IDs of the sops in the plan unit."""
    vehicles_have_moves: int | None = None
    """Number of vehicles that have moves for the plan unit. Only calculated if
    the verbosity is very high."""
    vehicles_with_moves: list[str] | None = None
    """IDs of the vehicles that have moves for the plan unit. Only calculated
    if the verbosity is very high."""


class Vehicle(BaseModel):
    """Check of a vehicle."""

    id: str
    """ID of the vehicle."""

    plan_units_have_moves: int | None = None
    """Number of plan units that have moves for the vehicle. Only calculated if
    the depth is medium."""


class Output(BaseModel):
    """Output of a feasibility check."""

    duration_maximum: float | None = None
    """Maximum duration of the check, in seconds."""
    duration_used: float | None = None
    """Duration used by the check, in seconds."""
    error: str | None = None
    """Error raised during the check."""
    plan_units: list[PlanUnit] | None = None
    """Check of the individual plan units."""
    remark: str | None = None
    """Remark of the check. It can be "ok", "timeout" or anything else that
    should explain itself."""
    solution: Solution | None = None
    """Start soltuion of the check."""
    summary: Summary | None = None
    """Summary of the check."""
    vehicles: list[Vehicle] | None = None
    """Check of the vehicles."""
    verbosity: str | None = None
    """Verbosity level of the check."""
