# © 2019-present nextmv.io inc

"""
Defines the vehicle class.
"""

from datetime import datetime
from typing import Any, List, Optional

from nextroute.base_model import BaseModel
from nextroute.schema.location import Location


class InitialStop(BaseModel):
    """Represents a stop that is already planned on a vehicle."""

    id: str
    """Unique identifier of the stop."""

    fixed: Optional[bool] = None
    """Whether the stop is fixed on the route."""


class VehicleDefaults(BaseModel):
    """Default values for vehicles."""

    activation_penalty: Optional[int] = None
    """Penalty of using the vehicle."""
    alternate_stops: Optional[List[str]] = None
    """A set of alternate stops for which only one should be serviced."""
    capacity: Optional[Any] = None
    """Capacity of the vehicle."""
    compatibility_attributes: Optional[List[str]] = None
    """Attributes that the vehicle is compatible with."""
    end_location: Optional[Location] = None
    """Location where the vehicle ends."""
    end_time: Optional[datetime] = None
    """Latest time at which the vehicle ends its route."""
    max_distance: Optional[int] = None
    """Maximum distance in meters that the vehicle can travel."""
    max_duration: Optional[int] = None
    """Maximum duration in seconds that the vehicle can travel."""
    max_stops: Optional[int] = None
    """Maximum number of stops that the vehicle can visit."""
    max_wait: Optional[int] = None
    """Maximum aggregated waiting time that the vehicle can wait across route stops."""
    min_stops: Optional[int] = None
    """Minimum stops that a vehicle should visit."""
    min_stops_penalty: Optional[float] = None
    """Penalty for not visiting the minimum number of stops."""
    speed: Optional[float] = None
    """Speed of the vehicle in meters per second."""
    start_level: Optional[Any] = None
    """Initial level of the vehicle."""
    start_location: Optional[Location] = None
    """Location where the vehicle starts."""
    start_time: Optional[datetime] = None
    """Time when the vehicle starts its route."""


class Vehicle(VehicleDefaults):
    """A vehicle services stops in a Vehicle Routing Problem (VRP)."""

    id: str
    """Unique identifier of the vehicle."""

    custom_data: Optional[Any] = None
    """Arbitrary custom data."""
    initial_stops: Optional[List[InitialStop]] = None
    """Initial stops planned on the vehicle."""
    stop_duration_multiplier: Optional[float] = None
    """Multiplier for the duration of stops."""
