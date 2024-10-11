# © 2019-present nextmv.io inc

"""
The Nextroute Python interface.

Nextroute is a flexible engine for solving Vehicle Routing Problems (VRPs). The
core of Nextroute is written in Go and this package provides a Python interface
to it.
"""

from .__about__ import __version__
from .options import DisableFormatOptions as DisableFormatOptions
from .options import FormatOptions as FormatOptions
from .options import Options as Options
from .options import ParallelSolveOptions as ParallelSolveOptions
from .solve import solve as solve

VERSION = __version__
"""The version of the Nextroute Python package."""
