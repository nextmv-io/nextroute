# © 2019-present nextmv.io inc

"""
Defines the input class.
"""

from datetime import datetime
from typing import Any

from nextroute.base_model import BaseModel
from nextroute.schema.stop import AlternateStop, Stop, StopDefaults
from nextroute.schema.vehicle import Vehicle, VehicleDefaults


class Defaults(BaseModel):
    """Default values for vehicles and stops."""

    stops: StopDefaults | None = None
    """Default values for stops."""
    vehicles: VehicleDefaults | None = None
    """Default values for vehicles."""


class DurationGroup(BaseModel):
    """Represents a group of stops that get additional duration whenever a stop
    of the group is approached for the first time."""

    duration: int
    """Duration to add when visiting the group."""
    group: list[str]
    """Stop IDs contained in the group."""


class MatrixTimeFrame(BaseModel):
    """Represents a time-dependent duration matrix or scaling factor."""

    start_time: datetime
    """Start time of the time frame."""
    end_time: datetime
    """End time of the time frame."""
    matrix: list[list[float]] | None = None
    """Duration matrix for the time frame."""
    scaling_factor: float | None = None
    """Scaling factor for the time frame."""


class TimeDependentMatrix(BaseModel):
    """Represents time-dependent duration matrices."""

    vehicle_ids: list[str] | None = None
    """Vehicle IDs for which the duration matrix is defined."""
    default_matrix: list[list[float]]
    """Default duration matrix."""
    matrix_time_frames: list[MatrixTimeFrame] | None = None
    """Time-dependent duration matrices."""


class Input(BaseModel):
    """Input schema for Nextroute."""

    stops: list[Stop]
    """Stops that must be visited by the vehicles."""
    vehicles: list[Vehicle]
    """Vehicles that service the stops."""

    alternate_stops: list[AlternateStop] | None = None
    """A set of alternate stops for the vehicles."""
    custom_data: Any | None = None
    """Arbitrary data associated with the input."""
    defaults: Defaults | None = None
    """Default values for vehicles and stops."""
    distance_matrix: list[list[float]] | None = None
    """Matrix of travel distances in meters between stops."""
    duration_groups: list[DurationGroup] | None = None
    """Duration in seconds added when approaching the group."""
    duration_matrix: list[list[float]] | TimeDependentMatrix | list[TimeDependentMatrix] | None = None
    """Matrix of travel durations in seconds between stops as a single matrix or duration matrices."""
    options: Any | None = None
    """Arbitrary options."""
    stop_groups: list[list[str]] | None = None
    """Groups of stops that must be part of the same route."""
