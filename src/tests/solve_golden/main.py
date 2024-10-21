# This script is copied to the `src` root so that the `nextroute` import is
# resolved. It is fed an input via stdin and is meant to write the output to
# stdout.
from typing import Any, Dict

import nextmv

import nextroute


def main() -> None:
    """Entry point for the program."""

    parameters = [
        nextmv.Parameter("input", str, "", "Path to input file. Default is stdin.", False),
        nextmv.Parameter("output", str, "", "Path to output file. Default is stdout.", False),
    ]

    nextroute_options = nextroute.Options()
    for name, default_value in nextroute_options.to_dict().items():
        parameters.append(nextmv.Parameter(name.lower(), type(default_value), default_value, name, False))

    options = nextmv.Options(*parameters)

    input = nextmv.load_local(options=options, path=options.input)

    nextmv.log("Solving vehicle routing problem:")
    nextmv.log(f"  - stops: {len(input.data.get('stops', []))}")
    nextmv.log(f"  - vehicles: {len(input.data.get('vehicles', []))}")

    output = solve(input, options)
    nextmv.write_local(output, path=options.output)


def solve(input: nextmv.Input, options: nextmv.Options) -> Dict[str, Any]:
    """Solves the given problem and returns the solution."""

    nextroute_input = nextroute.schema.Input.from_dict(input.data)
    nextroute_options = nextroute.Options.extract_from_dict(options.to_dict())
    nextroute_output = nextroute.solve(nextroute_input, nextroute_options)

    return nextroute_output.to_dict()


if __name__ == "__main__":
    main()
