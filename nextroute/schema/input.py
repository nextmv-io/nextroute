# © 2019-present nextmv.io inc

"""
Defines the input class.
"""

from typing import Any, List, Optional

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
    duratrion_groups: Optional[List[DurationGroup]] = None
    """Duration in seconds added when approaching the group."""
    duration_matrix: Optional[List[List[float]]] = None
    """Matrix of travel durations in seconds between stops."""
    options: Optional[Any] = None
    """Arbitrary options."""
    stop_groups: Optional[List[List[str]]] = None
    """Groups of stops that must be part of the same route."""
