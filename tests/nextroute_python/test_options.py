import unittest

import nextroute
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
        opt = nextroute.Options()
        args = opt.to_args()
        self.assertListEqual(
            args,
            [
                "-check.duration",
                "30.0s",
                "-check.verbosity",
                '"off"',
                "-format.disable.progression",
                "false",
                "-model.constraints.disable.attributes",
                "false",
                "-model.constraints.disable.capacity",
                "false",
                "-model.constraints.disable.capacities",
                "[]",
                "-model.constraints.disable.distancelimit",
                "false",
                "-model.constraints.disable.groups",
                "false",
                "-model.constraints.disable.maximumduration",
                "false",
                "-model.constraints.disable.maximumstops",
                "false",
                "-model.constraints.disable.maximumwaitstop",
                "false",
                "-model.constraints.disable.maximumwaitvehicle",
                "false",
                "-model.constraints.disable.mixingitems",
                "false",
                "-model.constraints.disable.precedence",
                "false",
                "-model.constraints.disable.vehiclestarttime",
                "false",
                "-model.constraints.disable.vehicleendtime",
                "false",
                "-model.constraints.disable.starttimewindows",
                "false",
                "-model.constraints.enable.cluster",
                "false",
                "-model.objectives.capacities",
                '""',
                "-model.objectives.minstops",
                "1.0",
                "-model.objectives.earlyarrivalpenalty",
                "1.0",
                "-model.objectives.latearrivalpenalty",
                "1.0",
                "-model.objectives.vehicleactivationpenalty",
                "1.0",
                "-model.objectives.travelduration",
                "0.0",
                "-model.objectives.vehiclesduration",
                "1.0",
                "-model.objectives.unplannedpenalty",
                "1.0",
                "-model.objectives.cluster",
                "0.0",
                "-model.properties.disable.durations",
                "false",
                "-model.properties.disable.stopdurationmultipliers",
                "false",
                "-model.properties.disable.durationgroups",
                "false",
                "-model.properties.disable.initialsolution",
                "false",
                "-model.validate.disable.starttime",
                "false",
                "-model.validate.disable.resources",
                "false",
                "-model.validate.enable.matrix",
                "false",
                "-model.validate.enable.matrixasymmetrytolerance",
                "20",
                "-solve.iterations",
                "-1",
                "-solve.duration",
                "5.0s",
                "-solve.parallelruns",
                "-1",
                "-solve.startsolutions",
                "-1",
                "-solve.rundeterministically",
                "false",
            ],
        )
