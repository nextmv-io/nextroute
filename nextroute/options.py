# © 2019-present nextmv.io inc

"""
Options for working with the Nextroute engine.
"""

import json
from typing import Any, Dict, List

from pydantic import Field

from nextroute.base_model import BaseModel
from nextroute.check.options import Options as CheckOptions
from nextroute.factory.options import Options as FactoryOptions

_DURATIONS_ARGS = [
    "-check.duration",
    "-solve.duration",
]


class ParallelSolveOptions(BaseModel):
    """Options for the parallel solver."""

    iterations: int = -1
    """
    Maximum number of iterations, -1 assumes no limit; iterations are counted
    after start solutions are generated.
    """
    duration: float = 5
    """Maximum duration, in seconds, of the solver."""
    parallel_runs: int = -1
    """
    Maximum number of parallel runs, -1 results in using all available
    resources.
    """
    start_solutions: int = -1
    """
    Number of solutions to generate on top of those passed in; one solution
    generated with sweep algorithm, the rest generated randomly.
    """
    run_deterministically: bool = False
    """Run the parallel solver deterministically."""


class DisableFormatOptions(BaseModel):
    """Options for disabling/enabling the progression series."""

    progression: bool = False
    """Whether to disable the progression series."""


class FormatOptions(BaseModel):
    """Options for formatting the output of the solver."""

    disable: DisableFormatOptions = Field(default_factory=DisableFormatOptions)
    """Options for disabling/enabling the progression series."""


class Options(BaseModel):
    """Options for using Nextroute."""

    check: CheckOptions = Field(default_factory=CheckOptions)
    """Options for enabling the check engine."""
    format: FormatOptions = Field(default_factory=FormatOptions)
    """Options for the output format."""
    model: FactoryOptions = Field(default_factory=FactoryOptions)
    """Options for the ready-to-go model."""
    solve: ParallelSolveOptions = Field(default_factory=ParallelSolveOptions)
    """Options for the parallel solver."""

    def to_args(self) -> List[str]:
        """
        Convert the options to command-line arguments.

        Returns
        ----------
        List[str]
            The flattened options as a list of strings.
        """

        opt_dict = self.to_dict()
        flattened = _flatten(opt_dict)

        default_options = Options()
        default_options_dict = default_options.to_dict()
        default_flattened = _flatten(default_options_dict)

        args = []
        for key, value in flattened.items():
            # We only care about custom options, so we skip the default ones.
            default_value = default_flattened.get(key)
            if value == default_value:
                continue

            key = key.replace("_", "")

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


def _flatten(nested: Dict[str, Any]) -> Dict[str, Any]:
    """Flatten a nested dict."""

    flattened = {}
    for child_key, child_value in nested.items():
        root_key = f"-{child_key}"
        __set_children(flattened, root_key, child_value)

    return flattened


def __set_children(flattened: Dict[str, Any], parent_key: str, parent_value: Any):
    """Helper function for `__flatten`. it is invoked recursively on a child
    value. If the child is not a dict, then the value is simply set on the
    flattened dict. If the child is a dict, then the function is invoked
    recursively on the child’s values, unitl a non-dict values is hit."""

    new_key = parent_key

    if parent_value is None:
        flattened[new_key] = parent_value
        return

    if isinstance(parent_value, dict):
        for child_key, child_value in parent_value.items():
            new_key = f"{parent_key}.{child_key}"
            __set_children(flattened, new_key, child_value)
        return

    flattened[new_key] = parent_value
