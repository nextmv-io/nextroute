# © 2019-present nextmv.io inc

"""
Defines the vehicle class.
"""

from datetime import datetime
from typing import Any

from nextroute.base_model import BaseModel
from nextroute.schema.location import Location


class InitialStop(BaseModel):
    """Represents a stop that is already planned on a vehicle."""

    id: str
    """Unique identifier of the stop."""

    fixed: bool | None = None
    """Whether the stop is fixed on the route."""


class VehicleDefaults(BaseModel):
    """Default values for vehicles."""

    activation_penalty: int | None = None
    """Penalty of using the vehicle."""
    alternate_stops: list[str] | None = None
    """A set of alternate stops for which only one should be serviced."""
    capacity: Any | None = None
    """Capacity of the vehicle."""
    compatibility_attributes: list[str] | None = None
    """Attributes that the vehicle is compatible with."""
    end_location: Location | None = None
    """Location where the vehicle ends."""
    end_time: datetime | None = None
    """Latest time at which the vehicle ends its route."""
    max_distance: int | None = None
    """Maximum distance in meters that the vehicle can travel."""
    max_duration: int | None = None
    """Maximum duration in seconds that the vehicle can travel."""
    max_stops: int | None = None
    """Maximum number of stops that the vehicle can visit."""
    max_wait: int | None = None
    """Maximum aggregated waiting time that the vehicle can wait across route stops."""
    min_stops: int | None = None
    """Minimum stops that a vehicle should visit."""
    min_stops_penalty: float | None = None
    """Penalty for not visiting the minimum number of stops."""
    speed: float | None = None
    """Speed of the vehicle in meters per second."""
    start_level: Any | None = None
    """Initial level of the vehicle."""
    start_location: Location | None = None
    """Location where the vehicle starts."""
    start_time: datetime | None = None
    """Time when the vehicle starts its route."""


class Vehicle(VehicleDefaults):
    """A vehicle services stops in a Vehicle Routing Problem (VRP)."""

    id: str
    """Unique identifier of the vehicle."""

    custom_data: Any | None = None
    """Arbitrary custom data."""
    initial_stops: list[InitialStop] | None = None
    """Initial stops planned on the vehicle."""
    stop_duration_multiplier: float | None = None
    """Multiplier for the duration of stops."""
