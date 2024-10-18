# © 2019-present nextmv.io inc

"""
Methods for solving a Vehicle Routing Problem with Nextroute.
"""

import json
import os
import subprocess
from typing import Any, Dict, Union

from nextroute.options import Options
from nextroute.schema.input import Input
from nextroute.schema.output import Output

SUPPORTED_OS = ["linux", "windows", "darwin"]
"""The operating systems supported by the Nextroute engine."""
SUPPORTED_ARCHITECTURES = ["x86_64", "arm64", "aarch64"]
"""The architectures supported by the Nextroute engine."""

_ARCHITECTURE_TRANSLATION = {
    "x86_64": "amd64",
    "arm64": "arm64",
    "aarch64": "arm64",
}


def solve(
    input: Union[Input, Dict[str, Any]],
    options: Union[Options, Dict[str, Any]],
) -> Output:
    """
    Solve a Vehicle Routing Problem (VRP) using the Nextroute engine. The input
    and options are passed to the engine, and the output is returned. The input
    and options can be provided as dictionaries or as objects, although the
    recommended way is to use the classes, as they provide validation.

    Examples
    --------

    * Using default options to load an input from a file.
        ```python
        import json

        import nextroute

        with open("input.json") as f:
            data = json.load(f)

        input = nextroute.schema.Input.from_dict(data)
        options = nextroute.Options()
        output = nextroute.solve(input, options)
        print(output)
        ```

    * Using custom options to load an input from a file.
        ```python
        import json

        import nextroute

        with open("input.json") as f:
            data = json.load(f)

        input = nextroute.schema.Input.from_dict(data)
        options = nextroute.Options(
            solve=nextroute.ParallelSolveOptions(duration=2),
        )
        output = nextroute.solve(input, options)
        print(output)
        ```

    * Using custom dict options to load an input from a file.
        ```python
        import json

        import nextroute

        with open("input.json") as f:
            data = json.load(f)

        input = nextroute.schema.Input.from_dict(data)
        options = {
            "solve": {
                "duration": 2,
            },
        }
        output = nextroute.solve(input, options)
        print(output)
        ```


    Parameters
    ----------
    input : Union[schema.Input, Dict[str, Any]]
        The input to the Nextroute engine. If a dictionary is provided, it will
        be converted to an Input object to validate it.
    options : Union[Options, Dict[str, Any]]
        The options for the Nextroute engine. If a dictionary is provided, it
        will be converted to an Options object.

    Returns
    -------
    schema.Output
        The output of the Nextroute engine. You can call the `to_dict` method
        on this object to get a dictionary representation of the output.
    """

    if isinstance(input, dict):
        input = Input.from_dict(input)

    input_stream = json.dumps(input.to_dict())

    if isinstance(options, dict):
        options = Options.from_dict(options)

    os_name = os.uname().sysname.lower()
    if os_name not in SUPPORTED_OS:
        raise Exception(f'unsupported operating system: "{os_name}", supported os are: {", ".join(SUPPORTED_OS)}')

    architecture = os.uname().machine.lower()
    if architecture not in SUPPORTED_ARCHITECTURES:
        raise Exception(
            f'unsupported architecture: "{architecture}", supported arch are: {", ".join(SUPPORTED_ARCHITECTURES)}'
        )

    binary_name = f"nextroute-{os_name}-{_ARCHITECTURE_TRANSLATION[architecture]}"
    if os_name == "windows":
        binary_name += ".exe"

    executable = os.path.join(os.path.dirname(__file__), "bin", binary_name)
    if not os.path.exists(executable):
        raise Exception(f"missing Nextroute binary: {executable}")

    option_args = options.to_args()
    args = [executable] + option_args

    try:
        result = subprocess.run(
            args,
            env=os.environ,
            check=True,
            text=True,
            capture_output=True,
            input=input_stream,
        )

    except subprocess.CalledProcessError as e:
        raise Exception(f"error running Nextroute binary: {e.stderr}") from e

    raw_output = result.stdout
    output = Output.from_dict(json.loads(raw_output))

    return output
