# © 2019-present nextmv.io inc

"""
Defines the stop class.
"""

from datetime import datetime
from typing import Any, List, Optional

from nextroute.base_model import BaseModel
from nextroute.schema.location import Location


class StopDefaults(BaseModel):
    """Default values for a stop."""

    compatibility_attributes: Optional[List[str]] = None
    """Attributes that the stop is compatible with."""
    duration: Optional[int] = None
    """Duration of the stop in seconds."""
    early_arrival_time_penalty: Optional[float] = None
    """Penalty per second for arriving at the stop before the target arrival time."""
    late_arrival_time_penalty: Optional[float] = None
    """Penalty per second for arriving at the stop after the target arrival time."""
    max_wait: Optional[int] = None
    """Maximum waiting duration in seconds at the stop."""
    quantity: Optional[Any] = None
    """Quantity of the stop."""
    start_time_window: Optional[Any] = None
    """Time window in which the stop can start service."""
    target_arrival_time: Optional[datetime] = None
    """Target arrival time at the stop."""
    unplanned_penalty: Optional[int] = None
    """Penalty for not planning a stop."""


class Stop(StopDefaults):
    """Stop is a location that must be visited by a vehicle in a Vehicle
    Routing Problem (VRP.)"""

    id: str
    """Unique identifier for the stop."""
    location: Location
    """Location of the stop."""

    custom_data: Optional[Any] = None
    """Arbitrary data associated with the stop."""
    mixing_items: Optional[Any] = None
    """Defines the items that are inserted or removed from the vehicle when visiting the stop."""
    precedes: Optional[Any] = None
    """Stops that must be visited after this one on the same route."""
    succeeds: Optional[Any] = None
    """Stops that must be visited before this one on the same route."""


class AlternateStop(StopDefaults):
    """An alternate stop can be serviced instead of another stop."""

    id: str
    """Unique identifier for the stop."""
    location: Location
    """Location of the stop."""

    custom_data: Optional[Any] = None
    """Arbitrary data associated with the stop."""
