# © 2019-present nextmv.io inc

"""
Defines the location class.
"""

from nextroute.base_model import BaseModel


class Location(BaseModel):
    """Location represents a geographical location."""

    lat: float
    """Latitude of the location."""
    lon: float
    """Longitude of the location."""
