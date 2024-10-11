# © 2019-present nextmv.io inc

"""
Options for the Nextroute check engine.
"""

from enum import Enum

from nextroute.base_model import BaseModel


class Verbosity(str, Enum):
    """Format of an `Input`."""

    OFF = "off"
    """The check engine is not run."""
    LOW = "low"
    """Low verbosity for the check engine."""
    MEDIUM = "medium"
    """Medium verbosity for the check engine."""
    HIGH = "high"
    """High verbosity for the check engine."""


class Options(BaseModel):
    """Options for the Nextroute check engine."""

    duration: float = 30
    """Maximum duration of the check, in seconds."""
    verbosity: Verbosity = Verbosity.OFF
    """Verbosity of the check engine."""
