# © 2019-present nextmv.io inc

"""
This module contains definitions for the schema in the Nextroute check.
"""

from typing import Dict, List, Optional

from nextroute.base_model import BaseModel


class ObjectiveTerm(BaseModel):
    """Check of the individual terms of the objective for a move."""

    base: Optional[float] = None
    """Base of the objective term."""
    factor: Optional[float] = None
    """Factor of the objective term."""
    name: Optional[str] = None
    """Name of the objective term."""
    value: Optional[float] = None
    """Value of the objective term, which is equivalent to `self.base *
    self.factor`."""


class Objective(BaseModel):
    """Estimate of an objective of a move."""

    terms: Optional[List[ObjectiveTerm]] = None
    """Check of the individual terms of the objective."""
    value: Optional[float] = None
    """Value of the objective."""
    vehicle: Optional[str] = None
    """ID of the vehicle for which it reports the objective."""


class Solution(BaseModel):
    """Solution that the check has been executed on."""

    objective: Optional[Objective] = None
    """Objective of the start solution."""
    plan_units_planned: Optional[int] = None
    """Number of plan units planned in the start solution."""
    plan_units_unplanned: Optional[int] = None
    """Number of plan units unplanned in the start solution."""
    stops_planned: Optional[int] = None
    """Number of stops planned in the start solution."""
    vehicles_not_used: Optional[int] = None
    """Number of vehicles not used in the start solution."""
    vehicles_used: Optional[int] = None
    """Number of vehicles used in the start solution."""


class Summary(BaseModel):
    """Summary of the check."""

    moves_failed: Optional[int] = None
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
    plan_units_best_move_failed: Optional[int] = None
    """Number of plan units for which the best move can not be planned. This
    should not happen if all the constraints are implemented correct."""
    plan_units_best_move_found: Optional[int] = None
    """Number of plan units for which at least one move has been found and the
    move is executable."""
    plan_units_best_move_increases_objective: Optional[int] = None
    """Number of plan units for which the best move is executable but would
    increase the objective value instead of decreasing it."""
    plan_units_checked: Optional[int] = None
    """Number of plan units that have been checked. If this is less than
    `self.plan_units_to_be_checked` the check timed out."""
    plan_units_have_no_move: Optional[int] = None
    """Number of plan units for which no feasible move has been found. This
    implies there is no move that can be executed without violating a
    constraint."""
    plan_units_to_be_checked: Optional[int] = None
    """Number of plan units to be checked."""


class PlanUnit(BaseModel):
    """Check of a plan unit."""

    best_move_failed: Optional[bool] = None
    """True if the plan unit's best move failed to execute."""
    best_move_increases_objective: Optional[bool] = None
    """True if the best move for the plan unit increases the objective."""
    best_move_objective: Optional[Objective] = None
    """Estimate of the objective of the best move if the plan unit has a best
    move."""
    constraints: Optional[Dict[str, int]] = None
    """Constraints that are violated for the plan unit."""
    has_best_move: Optional[bool] = None
    """True if a move is found for the plan unit. A plan unit has no move found
    if the plan unit is over-constrained or the move found is too expensive."""
    stops: Optional[List[str]] = None
    """IDs of the sops in the plan unit."""
    vehicles_have_moves: Optional[int] = None
    """Number of vehicles that have moves for the plan unit. Only calculated if
    the verbosity is very high."""
    vehicles_with_moves: Optional[List[str]] = None
    """IDs of the vehicles that have moves for the plan unit. Only calculated
    if the verbosity is very high."""


class Vehicle(BaseModel):
    """Check of a vehicle."""

    id: str
    """ID of the vehicle."""

    plan_units_have_moves: Optional[int] = None
    """Number of plan units that have moves for the vehicle. Only calculated if
    the depth is medium."""


class Output(BaseModel):
    """Output of a feasibility check."""

    duration_maximum: Optional[float] = None
    """Maximum duration of the check, in seconds."""
    duration_used: Optional[float] = None
    """Duration used by the check, in seconds."""
    error: Optional[str] = None
    """Error raised during the check."""
    plan_units: Optional[List[PlanUnit]] = None
    """Check of the individual plan units."""
    remark: Optional[str] = None
    """Remark of the check. It can be "ok", "timeout" or anything else that
    should explain itself."""
    solution: Optional[Solution] = None
    """Start soltuion of the check."""
    summary: Optional[Summary] = None
    """Summary of the check."""
    vehicles: Optional[List[Vehicle]] = None
    """Check of the vehicles."""
    verbosity: Optional[str] = None
    """Verbosity level of the check."""
