import unittest

import nextroute
import nextroute.check
import nextroute.factory
from nextroute import options


class TestOptions(unittest.TestCase):
    def test_options_default_values(self):
        opt = nextroute.Options()
        options_dict = opt.to_dict()
        self.assertDictEqual(
            options_dict,
            {
                "check": {"duration": 30.0, "verbosity": "off"},
                "format": {"disable": {"progression": False}},
                "model": {
                    "constraints": {
                        "disable": {
                            "attributes": False,
                            "capacity": False,
                            "capacities": [],
                            "distance_limit": False,
                            "groups": False,
                            "maximum_duration": False,
                            "maximum_stops": False,
                            "maximum_wait_stop": False,
                            "maximum_wait_vehicle": False,
                            "mixing_items": False,
                            "precedence": False,
                            "vehicle_start_time": False,
                            "vehicle_end_time": False,
                            "start_time_windows": False,
                        },
                        "enable": {"cluster": False},
                    },
                    "objectives": {
                        "capacities": "",
                        "min_stops": 1.0,
                        "early_arrival_penalty": 1.0,
                        "late_arrival_penalty": 1.0,
                        "vehicle_activation_penalty": 1.0,
                        "travel_duration": 0.0,
                        "vehicles_duration": 1.0,
                        "unplanned_penalty": 1.0,
                        "cluster": 0.0,
                    },
                    "properties": {
                        "disable": {
                            "durations": False,
                            "stop_duration_multipliers": False,
                            "duration_groups": False,
                            "initial_solution": False,
                        }
                    },
                    "validate": {
                        "disable": {"start_time": False, "resources": False},
                        "enable": {"matrix": False, "matrix_asymmetry_tolerance": 20},
                    },
                },
                "solve": {
                    "iterations": -1,
                    "duration": 5.0,
                    "parallel_runs": -1,
                    "start_solutions": -1,
                    "run_deterministically": False,
                },
            },
        )

    def test_flatten(self):
        nested = {
            "foo": {
                "bar": False,
                "baz": 1,
                "roh": "doh",
            },
            "bar": {
                "baz": {
                    "bah": "roh",
                }
            },
            "baz": False,
            "bah": 1,
        }
        flat = options._flatten(nested)
        self.assertDictEqual(
            flat,
            {
                "-foo.bar": False,
                "-foo.baz": 1,
                "-foo.roh": "doh",
                "-bar.baz.bah": "roh",
                "-baz": False,
                "-bah": 1,
            },
        )

    def test_options_to_args(self):
        # Default options should not produce any arguments.
        opt = nextroute.Options()
        args = opt.to_args()
        self.assertListEqual(args, [])

        # Only options that are not default should produce arguments.
        opt2 = nextroute.Options(
            check=nextroute.check.Options(
                duration=4,
                verbosity=nextroute.check.Verbosity.MEDIUM,
            ),
            solve=nextroute.ParallelSolveOptions(
                duration=4,
                iterations=-1,  # Default value should be skipped.
            ),
            model=nextroute.factory.Options(
                constraints=nextroute.factory.Constraints(
                    disable=nextroute.factory.DisableConstraints(
                        attributes=True,
                    ),
                ),
                validate=nextroute.factory.Validate(
                    enable=nextroute.factory.EnableValidate(
                        matrix=False,  # This option should be skipped because it is bool False.
                    ),
                ),
            ),
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
