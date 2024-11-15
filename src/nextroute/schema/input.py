# © 2019-present nextmv.io inc

"""
Defines the input class.
"""

from datetime import datetime
from typing import Any, List, Optional, Union

from nextroute.base_model import BaseModel
from nextroute.schema.stop import AlternateStop, Stop, StopDefaults
from nextroute.schema.vehicle import Vehicle, VehicleDefaults


class Defaults(BaseModel):
    """Default values for vehicles and stops."""

    stops: Optional[StopDefaults] = None
    """Default values for stops."""
    vehicles: Optional[VehicleDefaults] = None
    """Default values for vehicles."""


class DurationGroup(BaseModel):
    """Represents a group of stops that get additional duration whenever a stop
    of the group is approached for the first time."""

    duration: int
    """Duration to add when visiting the group."""
    group: List[str]
    """Stop IDs contained in the group."""


class MatrixTimeFrame(BaseModel):
    """Represents a time-dependent duration matrix or scaling factor."""

    start_time: datetime
    """Start time of the time frame."""
    end_time: datetime
    """End time of the time frame."""
    matrix: Optional[List[List[float]]] = None
    """Duration matrix for the time frame."""
    scaling_factor: Optional[float] = None
    """Scaling factor for the time frame."""


class TimeDependentMatrix(BaseModel):
    """Represents time-dependent duration matrices."""

    vehicle_ids: Optional[List[str]] = None
    """Vehicle IDs for which the duration matrix is defined."""
    default_matrix: List[List[float]]
    """Default duration matrix."""
    matrix_time_frames: Optional[List[MatrixTimeFrame]] = None
    """Time-dependent duration matrices."""


class Input(BaseModel):
    """Input schema for Nextroute."""

    stops: List[Stop]
    """Stops that must be visited by the vehicles."""
    vehicles: List[Vehicle]
    """Vehicles that service the stops."""

    alternate_stops: Optional[List[AlternateStop]] = None
    """A set of alternate stops for the vehicles."""
    custom_data: Optional[Any] = None
    """Arbitrary data associated with the input."""
    defaults: Optional[Defaults] = None
    """Default values for vehicles and stops."""
    distance_matrix: Optional[List[List[float]]] = None
    """Matrix of travel distances in meters between stops."""
    duration_groups: Optional[List[DurationGroup]] = None
    """Duration in seconds added when approaching the group."""
    duration_matrix: Optional[
        Union[
            List[List[float]],
            TimeDependentMatrix,
            List[TimeDependentMatrix],
        ]
    ] = None
    """Matrix of travel durations in seconds between stops as a single matrix or duration matrices."""
    options: Optional[Any] = None
    """Arbitrary options."""
    stop_groups: Optional[List[List[str]]] = None
    """Groups of stops that must be part of the same route."""
