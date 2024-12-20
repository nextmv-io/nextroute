import os
import re
import subprocess
import unittest
from typing import Any, List, Tuple

import nextroute

NEXTROUTE_CMD = os.path.join(os.path.dirname(__file__), "..", "..", "cmd")


def _compile_nextroute():
    """Compiles the nextroute binary."""
    subprocess.run(["go", "build", "-o", "nextroute.exe"], cwd=NEXTROUTE_CMD, check=True)


def _extract_option(line_1: str, line_2: str) -> Tuple[str, type, Any, str]:
    """Extracts an option from the help lines of the nextroute binary."""
    components = line_1.split(" ")
    name = components[0].replace(".", "_").upper()[1:]
    t = bool
    default = False
    default_match = re.compile(r"\(default (.+?)\)").search(line_2)
    default_extracted = None
    if default_match:
        default_extracted = default_match.group(1)
    if len(components) > 1:
        if components[1] == "duration":
            t = float
            default = float(default_extracted[:-1]) if default_extracted else 0.0
        elif components[1] == "string":
            t = str
            default = default_extracted[1:-1] if default_extracted else ""
        elif components[1] == "int":
            t = int
            default = int(default_extracted) if default_extracted else 0
        elif components[1] == "float":
            t = float
            default = float(default_extracted) if default_extracted else 0.0
        elif components[1] == "value":
            t = Any
            default = None  # No default extraction for value as this can be anything.
        else:
            raise ValueError(f"Unsupported type: {components[1]}")

    descr_end = line_2.find("(env") - 1
    if descr_end < 0:
        raise ValueError(f"Could not find the end of the description in line: {line_2}")
    descr = line_2[:descr_end].strip()

    return name, t, default, descr


def _extract_options() -> List[Tuple[str, type, Any, str]]:
    """Extracts the options from the nextroute binary."""
    _compile_nextroute()
    executable = os.path.join(NEXTROUTE_CMD, "nextroute.exe")
    output = subprocess.check_output([executable, "--help"], stderr=subprocess.STDOUT, text=True)

    options = []
    lines = output.splitlines()
    for line_idx, line in enumerate(lines):
        if not line.startswith("  -"):
            continue

        line = line.strip()

        if len(lines) <= line_idx + 1:
            raise ValueError(f'Could not find the description for option: {line.split(" ")[0]}')
        next_line = lines[line_idx + 1]

        options.append(_extract_option(line, next_line))
    return options


class TestOptions(unittest.TestCase):
    def test_options_default_values(self):
        IGNORED_OPTIONS = {
            "RUNNER_INPUT_PATH",
            "RUNNER_OUTPUT_PATH",
            "RUNNER_OUTPUT_SOLUTIONS",
            "RUNNER_PROFILE_CPU",
            "RUNNER_PROFILE_MEMORY",
        }
        opt = nextroute.Options()
        options_dict = opt.to_dict()
        # Check that all (relevant) options are present and have the correct default
        # values. We get the options from the nextroute binary help output to ensure
        # that we are in sync with the Go implementation.
        bin_options = _extract_options()
        for name, t, default, _ in bin_options:
            if name in IGNORED_OPTIONS:
                continue
            self.assertIn(name, options_dict, f"Option {name} is missing.")
            if t is not Any:
                self.assertEqual(
                    options_dict[name],
                    default,
                    f"Option {name} has the wrong default value. {options_dict[name]} != {default} (got != expected)",
                )

    def test_options_to_args(self):
        # Default options should not produce any arguments.
        opt = nextroute.Options()
        args = opt.to_args()
        self.assertListEqual(args, [])

        # Only options that are not default should produce arguments.
        opt2 = nextroute.Options(
            CHECK_DURATION=4,
            CHECK_VERBOSITY=nextroute.Verbosity.MEDIUM,
            SOLVE_DURATION=4,
            SOLVE_ITERATIONS=-1,  # Default value should be skipped.
            MODEL_CONSTRAINTS_DISABLE_ATTRIBUTES=True,
            MODEL_VALIDATE_ENABLE_MATRIX=False,  # This option should be skipped because it is bool False.
        )
        args2 = opt2.to_args()
        self.assertListEqual(
            args2,
            [
                "-check.duration",
                "4.0s",
                "-check.verbosity",
                "medium",
                "-model.constraints.disable.attributes",  # Bool flags do not have values.
                "-solve.duration",
                "4.0s",
            ],
        )
