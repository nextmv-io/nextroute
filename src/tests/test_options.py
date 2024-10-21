import unittest

import nextroute


class TestOptions(unittest.TestCase):
    def test_options_default_values(self):
        opt = nextroute.Options()
        options_dict = opt.to_dict()
        self.assertDictEqual(
            options_dict,
            {
                "CHECK_DURATION": 30.0,
                "CHECK_VERBOSITY": "off",
                "FORMAT_DISABLE_PROGRESSION": False,
                "MODEL_CONSTRAINTS_DISABLE_ATTRIBUTES": False,
                "MODEL_CONSTRAINTS_DISABLE_CAPACITIES": [],
                "MODEL_CONSTRAINTS_DISABLE_CAPACITY": False,
                "MODEL_CONSTRAINTS_DISABLE_DISTANCELIMIT": False,
                "MODEL_CONSTRAINTS_DISABLE_GROUPS": False,
                "MODEL_CONSTRAINTS_DISABLE_MAXIMUMDURATION": False,
                "MODEL_CONSTRAINTS_DISABLE_MAXIMUMSTOPS": False,
                "MODEL_CONSTRAINTS_DISABLE_MAXIMUMWAITSTOP": False,
                "MODEL_CONSTRAINTS_DISABLE_MAXIMUMWAITVEHICLE": False,
                "MODEL_CONSTRAINTS_DISABLE_MIXINGITEMS": False,
                "MODEL_CONSTRAINTS_DISABLE_PRECEDENCE": False,
                "MODEL_CONSTRAINTS_DISABLE_STARTTIMEWINDOWS": False,
                "MODEL_CONSTRAINTS_DISABLE_VEHICLEENDTIME": False,
                "MODEL_CONSTRAINTS_DISABLE_VEHICLESTARTTIME": False,
                "MODEL_CONSTRAINTS_ENABLE_CLUSTER": False,
                "MODEL_OBJECTIVES_CAPACITIES": "",
                "MODEL_OBJECTIVES_CLUSTER": 0.0,
                "MODEL_OBJECTIVES_EARLYARRIVALPENALTY": 1.0,
                "MODEL_OBJECTIVES_LATEARRIVALPENALTY": 1.0,
                "MODEL_OBJECTIVES_MINSTOPS": 1.0,
                "MODEL_OBJECTIVES_TRAVELDURATION": 0.0,
                "MODEL_OBJECTIVES_UNPLANNEDPENALTY": 1.0,
                "MODEL_OBJECTIVES_VEHICLEACTIVATIONPENALTY": 1.0,
                "MODEL_OBJECTIVES_VEHICLESDURATION": 1.0,
                "MODEL_PROPERTIES_DISABLE_DURATIONGROUPS": False,
                "MODEL_PROPERTIES_DISABLE_DURATIONS": False,
                "MODEL_PROPERTIES_DISABLE_INITIALSOLUTION": False,
                "MODEL_PROPERTIES_DISABLE_STOPDURATIONMULTIPLIERS": False,
                "MODEL_VALIDATE_DISABLE_RESOURCES": False,
                "MODEL_VALIDATE_DISABLE_STARTTIME": False,
                "MODEL_VALIDATE_ENABLE_MATRIX": False,
                "MODEL_VALIDATE_ENABLE_MATRIXASYMMETRYTOLERANCE": 20,
                "SOLVE_DURATION": 5.0,
                "SOLVE_ITERATIONS": -1,
                "SOLVE_PARALLELRUNS": -1,
                "SOLVE_RUNDETERMINISTICALLY": False,
                "SOLVE_STARTSOLUTIONS": -1,
            },
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
