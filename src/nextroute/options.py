# © 2019-present nextmv.io inc

"""
Options for working with the Nextroute engine.
"""

import json
from enum import Enum
from typing import Any, Dict, List

from pydantic import Field

from nextroute.base_model import BaseModel

# Arguments that require a duration suffix.
_DURATIONS_ARGS = [
    "-check.duration",
    "-solve.duration",
]

# Arguments that require a string enum.
_STR_ENUM_ARGS = [
    "CHECK_VERBOSITY",
]


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
    """Options for using Nextroute."""

    CHECK_DURATION: float = 30
    """Maximum duration of the check, in seconds."""
    CHECK_VERBOSITY: Verbosity = Verbosity.OFF
    """Verbosity of the check engine."""
    FORMAT_DISABLE_PROGRESSION: bool = False
    """Whether to disable the progression series."""
    MODEL_CONSTRAINTS_DISABLE_ATTRIBUTES: bool = False
    """Ignore the compatibility attributes constraint."""
    MODEL_CONSTRAINTS_DISABLE_CAPACITIES: List[str] = Field(default_factory=list)
    """Ignore the capacity constraint for the given resource names."""
    MODEL_CONSTRAINTS_DISABLE_CAPACITY: bool = False
    """Ignore the capacity constraint for all resources."""
    MODEL_CONSTRAINTS_DISABLE_DISTANCELIMIT: bool = False
    """Ignore the distance limit constraint."""
    MODEL_CONSTRAINTS_DISABLE_GROUPS: bool = False
    """Ignore the groups constraint."""
    MODEL_CONSTRAINTS_DISABLE_MAXIMUMDURATION: bool = False
    """Ignore the maximum duration constraint."""
    MODEL_CONSTRAINTS_DISABLE_MAXIMUMSTOPS: bool = False
    """Ignore the maximum stops constraint."""
    MODEL_CONSTRAINTS_DISABLE_MAXIMUMWAITSTOP: bool = False
    """Ignore the maximum stop wait constraint."""
    MODEL_CONSTRAINTS_DISABLE_MAXIMUMWAITVEHICLE: bool = False
    """Ignore the maximum vehicle wait constraint."""
    MODEL_CONSTRAINTS_DISABLE_MIXINGITEMS: bool = False
    """Ignore the do not mix items constraint."""
    MODEL_CONSTRAINTS_DISABLE_PRECEDENCE: bool = False
    """Ignore the precedence (pickups & deliveries) constraint."""
    MODEL_CONSTRAINTS_DISABLE_STARTTIMEWINDOWS: bool = False
    """Ignore the start time windows constraint."""
    MODEL_CONSTRAINTS_DISABLE_VEHICLEENDTIME: bool = False
    """Ignore the vehicle end time constraint."""
    MODEL_CONSTRAINTS_DISABLE_VEHICLESTARTTIME: bool = False
    """Ignore the vehicle start time constraint."""
    MODEL_CONSTRAINTS_ENABLE_CLUSTER: bool = False
    """Enable the cluster constraint."""
    MODEL_OBJECTIVES_CAPACITIES: str = ""
    """
    Capacity objective, provide triple for each resource
    `name:default;factor:1.0;offset;0.0`.
    """
    MODEL_OBJECTIVES_CLUSTER: float = 0.0
    """Factor to weigh the cluster objective."""
    MODEL_OBJECTIVES_EARLYARRIVALPENALTY: float = 1.0
    """Factor to weigh the early arrival objective."""
    MODEL_OBJECTIVES_LATEARRIVALPENALTY: float = 1.0
    """Factor to weigh the late arrival objective."""
    MODEL_OBJECTIVES_MINSTOPS: float = 1.0
    """Factor to weigh the min stops objective."""
    MODEL_OBJECTIVES_TRAVELDURATION: float = 0.0
    """Factor to weigh the travel duration objective."""
    MODEL_OBJECTIVES_UNPLANNEDPENALTY: float = 1.0
    """Factor to weigh the unplanned objective."""
    MODEL_OBJECTIVES_VEHICLEACTIVATIONPENALTY: float = 1.0
    """Factor to weigh the vehicle activation objective."""
    MODEL_OBJECTIVES_VEHICLESDURATION: float = 1.0
    """Factor to weigh the vehicles duration objective."""
    MODEL_PROPERTIES_DISABLE_DURATIONGROUPS: bool = False
    """Ignore the durations groups of stops."""
    MODEL_PROPERTIES_DISABLE_DURATIONS: bool = False
    """Ignore the durations of stops."""
    MODEL_PROPERTIES_DISABLE_INITIALSOLUTION: bool = False
    """Ignore the initial solution."""
    MODEL_PROPERTIES_DISABLE_STOPDURATIONMULTIPLIERS: bool = False
    """Ignore the stop duration multipliers defined on vehicles."""
    MODEL_VALIDATE_DISABLE_RESOURCES: bool = False
    """Disable the resources validation."""
    MODEL_VALIDATE_DISABLE_STARTTIME: bool = False
    """Disable the start time validation."""
    MODEL_VALIDATE_ENABLE_MATRIX: bool = False
    """Enable matrix validation."""
    MODEL_VALIDATE_ENABLE_MATRIXASYMMETRYTOLERANCE: int = 20
    """Percentage of acceptable matrix asymmetry, requires matrix validation enabled."""
    SOLVE_DURATION: float = 5
    """Maximum duration, in seconds, of the solver."""
    SOLVE_ITERATIONS: int = -1
    """
    Maximum number of iterations, -1 assumes no limit; iterations are counted
    after start solutions are generated.
    """
    SOLVE_PARALLELRUNS: int = -1
    """
    Maximum number of parallel runs, -1 results in using all available
    resources.
    """
    SOLVE_RUNDETERMINISTICALLY: bool = False
    """Run the parallel solver deterministically."""
    SOLVE_STARTSOLUTIONS: int = -1
    """
    Number of solutions to generate on top of those passed in; one solution
    generated with sweep algorithm, the rest generated randomly.
    """

    def to_args(self) -> List[str]:
        """
        Convert the options to command-line arguments.

        Returns
        ----------
        List[str]
            The flattened options as a list of strings.
        """

        opt_dict = self.to_dict()

        default_options = Options()
        default_options_dict = default_options.to_dict()

        args = []
        for key, value in opt_dict.items():
            # We only care about custom options, so we skip the default ones.
            default_value = default_options_dict.get(key)
            if value == default_value:
                continue

            key = f"-{key.replace('_', '.').lower()}"

            str_value = json.dumps(value)
            if key in _DURATIONS_ARGS:
                str_value = str_value + "s"  # Transforms into seconds.

            if str_value.startswith('"') and str_value.endswith('"'):
                str_value = str_value[1:-1]

            # Nextroute’s Go implementation does not support boolean flags with
            # values. If the value is a boolean, then we only append the key if
            # the value is True.
            should_append_value = True
            if isinstance(value, bool):
                if not value:
                    continue

                should_append_value = False

            args.append(key)
            if should_append_value:
                args.append(str_value)

        return args

    @classmethod
    def extract_from_dict(cls, data: Dict[str, Any]) -> "Options":
        """
        Extracts options from a dictionary. This dictionary may contain more
        keys that are not part of the Nextroute options.

        Parameters
        ----------
        data : Dict[str, Any]
            The dictionary to extract options from.

        Returns
        ----------
        Options
            The Nextroute options.
        """

        options = cls()
        for key, value in data.items():
            key = key.upper()
            if not hasattr(options, key):
                continue

            # Enums need to be handled manually.
            if key == "CHECK_VERBOSITY":
                value = Verbosity(value)

            setattr(options, key, value)

        return options
