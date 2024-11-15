# © 2019-present nextmv.io inc

"""
The Nextroute Python interface.

Nextroute is a flexible engine for solving Vehicle Routing Problems (VRPs). The
core of Nextroute is written in Go and this package provides a Python interface
to it.
"""

from .__about__ import __version__
from .options import Options as Options
from .options import Verbosity as Verbosity
from .solve import solve as solve
from .version import nextroute_version as nextroute_version

VERSION = __version__
"""The version of the Nextroute Python package."""
