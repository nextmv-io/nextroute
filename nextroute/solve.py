# © 2019-present nextmv.io inc

"""
Methods for solving a Vehicle Routing Problem with Nextroute.
"""

import json
import os
import subprocess
from typing import Union

from nextroute import check, options, schema


def solve(input: schema.Input, options: options.Options) -> Union[schema.Output, check.Output]:
    """
    Solve a Vehicle Routing Problem (VRP) using the Nextroute engine.
    """

    input_stream = json.dumps(input.to_dict())
    options_args = options.to_args()

    args = ["./main"] + options_args

    try:
        result = subprocess.run(
            args,
            env=os.environ,
            check=True,
            text=True,
            capture_output=True,
            stdin=input_stream,
        )

    except subprocess.CalledProcessError as e:
        raise Exception(f"error running Nextroute binary: {e.stderr}") from e

    raw_output = result.stdout

    return raw_output
