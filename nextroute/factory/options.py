# © 2019-present nextmv.io inc

"""
Options for the Nextroute factory.
"""

from typing import List

from pydantic import Field

from nextroute.base_model import BaseModel


class DisableConstraints(BaseModel):
    """Options for disabling specific constraints."""

    attributes: bool = False
    """Ignore the compatibility attributes constraint."""
    capacity: bool = False
    """Ignore the capacity constraint for all resources."""
    capacities: List[str] = Field(default_factory=list)
    """Ignore the capacity constraint for the given resource names."""
    distance_limit: bool = False
    """Ignore the distance limit constraint."""
    groups: bool = False
    """Ignore the groups constraint."""
    maximum_duration: bool = False
    """Ignore the maximum duration constraint."""
    maximum_stops: bool = False
    """Ignore the maximum stops constraint."""
    maximum_wait_stop: bool = False
    """Ignore the maximum stop wait constraint."""
    maximum_wait_vehicle: bool = False
    """Ignore the maximum vehicle wait constraint."""
    mixing_items: bool = False
    """Ignore the do not mix items constraint."""
    precedence: bool = False
    """Ignore the precedence (pickups & deliveries) constraint."""
    vehicle_start_time: bool = False
    """Ignore the vehicle start time constraint."""
    vehicle_end_time: bool = False
    """Ignore the vehicle end time constraint."""
    start_time_windows: bool = False
    """Ignore the start time windows constraint."""


class EnableConstraints(BaseModel):
    """Options for enabling specific constraints."""

    cluster: bool = False
    """Enable the cluster constraint."""


class Constraints(BaseModel):
    """Options for configuring constraints."""

    disable: DisableConstraints = Field(default_factory=DisableConstraints)
    """Options for disabling specific constraints."""
    enable: EnableConstraints = Field(default_factory=EnableConstraints)
    """Options for enabling specific constraints."""


class Objectives(BaseModel):
    """Options for configuring objectives."""

    capacities: str = ""
    """
    Capacity objective, provide triple for each resource
    `name:default;factor:1.0;offset;0.0`.
    """
    min_stops: float = 1.0
    """Factor to weigh the min stops objective."""
    early_arrival_penalty: float = 1.0
    """Factor to weigh the early arrival objective."""
    late_arrival_penalty: float = 1.0
    """Factor to weigh the late arrival objective."""
    vehicle_activation_penalty: float = 1.0
    """Factor to weigh the vehicle activation objective."""
    travel_duration: float = 0.0
    """Factor to weigh the travel duration objective."""
    vehicles_duration: float = 1.0
    """Factor to weigh the vehicles duration objective."""
    unplanned_penalty: float = 1.0
    """Factor to weigh the unplanned objective."""
    cluster: float = 0.0
    """Factor to weigh the cluster objective."""


class DisableProperties(BaseModel):
    """Options for disabling specific properties."""

    durations: bool = False
    """Ignore the durations of stops."""
    stop_duration_multipliers: bool = False
    """Ignore the stop duration multipliers defined on vehicles."""
    duration_groups: bool = False
    """Ignore the durations groups of stops."""
    initial_solution: bool = False
    """Ignore the initial solution."""


class Properties(BaseModel):
    """Options for configuring properties."""

    disable: DisableProperties = Field(default_factory=DisableProperties)
    """Options for disabling specific properties."""


class DisableValidate(BaseModel):
    """Options for disabling specific validations."""

    start_time: bool = False
    """Disable the start time validation."""
    resources: bool = False
    """Disable the resources validation."""


class EnableValidate(BaseModel):
    """Options for enabling specific validations."""

    matrix: bool = False
    """Enable matrix validation."""
    matrix_asymmetry_tolerance: int = 20
    """Percentage of acceptable matrix asymmetry, requires matrix validation enabled."""


class Validate(BaseModel):
    """Options for configuring validations."""

    disable: DisableValidate = Field(default_factory=DisableValidate)
    """Options for disabling specific validations"""
    enable: EnableValidate = Field(default_factory=EnableValidate)
    """Options for enabling specific validations"""


class Options(BaseModel):
    """Options that configure how a Nextroute model is built."""

    constraints: Constraints = Field(default_factory=Constraints)
    """Options for configuring constraints."""
    objectives: Objectives = Field(default_factory=Objectives)
    """Options for configuring objectives."""
    properties: Properties = Field(default_factory=Properties)
    """Options for configuring properties."""
    validate_options: Validate = Field(default_factory=Validate, alias="validate")
    """Options for configuring validations."""
