# © 2019-present nextmv.io inc

"""
Defines the stop class.
"""

from datetime import datetime
from typing import Any

from nextroute.base_model import BaseModel
from nextroute.schema.location import Location


class StopDefaults(BaseModel):
    """Default values for a stop."""

    compatibility_attributes: list[str] | None = None
    """Attributes that the stop is compatible with."""
    duration: int | None = None
    """Duration of the stop in seconds."""
    early_arrival_time_penalty: float | None = None
    """Penalty per second for arriving at the stop before the target arrival time."""
    late_arrival_time_penalty: float | None = None
    """Penalty per second for arriving at the stop after the target arrival time."""
    max_wait: int | None = None
    """Maximum waiting duration in seconds at the stop."""
    quantity: Any | None = None
    """Quantity of the stop."""
    start_time_window: Any | None = None
    """Time window in which the stop can start service."""
    target_arrival_time: datetime | None = None
    """Target arrival time at the stop."""
    unplanned_penalty: int | None = None
    """Penalty for not planning a stop."""


class Stop(StopDefaults):
    """Stop is a location that must be visited by a vehicle in a Vehicle
    Routing Problem (VRP.)"""

    id: str
    """Unique identifier for the stop."""
    location: Location
    """Location of the stop."""

    custom_data: Any | None = None
    """Arbitrary data associated with the stop."""
    mixing_items: Any | None = None
    """Defines the items that are inserted or removed from the vehicle when visiting the stop."""
    precedes: Any | None = None
    """Stops that must be visited after this one on the same route."""
    succeeds: Any | None = None
    """Stops that must be visited before this one on the same route."""


class AlternateStop(StopDefaults):
    """An alternate stop can be serviced instead of another stop."""

    id: str
    """Unique identifier for the stop."""
    location: Location
    """Location of the stop."""

    custom_data: Any | None = None
    """Arbitrary data associated with the stop."""
