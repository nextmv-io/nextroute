# © 2019-present nextmv.io inc

import json
import math
import os
import unittest

from nextroute import check as nextrouteCheck
from nextroute.schema import (
    Location,
    ObjectiveOutput,
    Output,
    PlannedStopOutput,
    ResultStatistics,
    RunStatistics,
    SeriesData,
    Solution,
    Statistics,
    StopOutput,
    VehicleOutput,
    Version,
)


class TestOutput(unittest.TestCase):
    filepath = os.path.join(os.path.dirname(__file__), "output.json")

    def test_from_json(self):
        with open(self.filepath) as f:
            json_data = json.load(f)

        nextroute_output = Output.from_dict(json_data)
        parsed = nextroute_output.to_dict()
        solution = parsed["solutions"][0]

        for s, stop in enumerate(solution["unplanned"]):
            original_stop = json_data["solutions"][0]["unplanned"][s]
            self.assertEqual(
                stop,
                original_stop,
                f"stop: parsed({stop}) and original ({original_stop}) should be equal",
            )

        for v, vehicle in enumerate(solution["vehicles"]):
            original_vehicle = json_data["solutions"][0]["vehicles"][v]
            self.assertEqual(
                vehicle,
                original_vehicle,
                f"vehicle: parsed ({vehicle}) and original ({original_vehicle}) should be equal",
            )

        self.assertEqual(
            solution["objective"],
            json_data["solutions"][0]["objective"],
            f"objective: parsed ({solution['objective']}) and "
            f"original ({json_data['solutions'][0]['objective']}) should be equal",
        )

        statistics = parsed["statistics"]
        self.assertEqual(
            statistics["run"],
            json_data["statistics"]["run"],
            f"run: parsed ({statistics['run']}) and original ({json_data['statistics']['run']}) should be equal",
        )
        self.assertEqual(
            statistics["result"],
            json_data["statistics"]["result"],
            f"result: parsed ({statistics['result']}) and "
            f"original ({json_data['statistics']['result']}) should be equal",
        )

    def test_from_dict(self):
        with open(self.filepath) as f:
            json_data = json.load(f)

        nextroute_output = Output.from_dict(json_data)

        version = nextroute_output.version
        self.assertTrue(isinstance(version, Version), "Version should be of type Version.")

        solutions = nextroute_output.solutions
        for solution in solutions:
            self.assertTrue(
                isinstance(solution, Solution),
                f"Solution ({solution}) should be of type Solution.",
            )

            unplanned = solution.unplanned
            for stop in unplanned:
                self.assertTrue(
                    isinstance(stop, StopOutput),
                    f"Stop ({stop}) should be of type StopOutput.",
                )
                self.assertTrue(
                    stop.id is not None,
                    f"Stop ({stop}) should have an id.",
                )
                self.assertNotEqual(
                    stop.id,
                    "",
                    f"Stop ({stop}) should have a valid id.",
                )
                self.assertTrue(
                    isinstance(stop.location, Location),
                    f"Stop ({stop}) should have a location.",
                )
                self.assertGreaterEqual(
                    stop.location.lat,
                    -90,
                    f"Stop ({stop}) should have a valid latitude.",
                )
                self.assertLessEqual(
                    stop.location.lat,
                    90,
                    f"Stop ({stop}) should have a valid latitude.",
                )
                self.assertGreaterEqual(
                    stop.location.lon,
                    -180,
                    f"Stop ({stop}) should have a valid longitude.",
                )
                self.assertLessEqual(
                    stop.location.lon,
                    180,
                    f"Stop ({stop}) should have a valid longitude.",
                )

            vehicles = solution.vehicles
            for vehicle in vehicles:
                self.assertTrue(
                    isinstance(vehicle, VehicleOutput),
                    f"Vehicle ({vehicle}) should be of type VehicleOutput.",
                )
                self.assertTrue(
                    vehicle.id is not None,
                    f"Vehicle ({vehicle}) should have an id.",
                )
                self.assertNotEqual(
                    vehicle.id,
                    "",
                    f"Vehicle ({vehicle}) should have a valid id.",
                )
                self.assertGreaterEqual(
                    vehicle.route_duration,
                    0,
                    f"Vehicle ({vehicle}) should have a valid route duration.",
                )
                self.assertGreaterEqual(
                    vehicle.route_stops_duration,
                    0,
                    f"Vehicle ({vehicle}) should have a valid route stops duration.",
                )
                self.assertGreaterEqual(
                    vehicle.route_travel_distance,
                    0,
                    f"Vehicle ({vehicle}) should have a valid route travel distance.",
                )
                self.assertGreaterEqual(
                    vehicle.route_travel_duration,
                    0,
                    f"Vehicle ({vehicle}) should have a valid route travel duration.",
                )

                for stop in vehicle.route:
                    self.assertTrue(
                        isinstance(stop, PlannedStopOutput),
                        f"Stop ({stop}) should be of type PlannedStopOutput.",
                    )
                    self.assertTrue(
                        isinstance(stop.stop, StopOutput),
                        f"Stop ({stop}) should have a stop.",
                    )
                    self.assertGreaterEqual(
                        stop.travel_duration,
                        0,
                        f"Stop ({stop}) should have a valid travel duration.",
                    )
                    self.assertGreaterEqual(
                        stop.cumulative_travel_duration,
                        0,
                        f"Stop ({stop}) should have a valid cumulative travel duration.",
                    )

            objective = solution.objective
            self.assertTrue(
                isinstance(objective, ObjectiveOutput),
                f"Objective ({objective}) should be of type ObjectiveOutput.",
            )

        statistics = nextroute_output.statistics
        self.assertTrue(
            isinstance(statistics, Statistics),
            f"Statistics ({statistics}) should be of type Statistics.",
        )

        run_statistics = statistics.run
        self.assertTrue(
            isinstance(run_statistics, RunStatistics),
            f"Run statistics ({run_statistics}) should be of type RunStatistics.",
        )
        self.assertGreaterEqual(
            run_statistics.duration,
            0,
            f"Run statistics ({run_statistics}) should have a valid duration.",
        )
        self.assertGreaterEqual(
            run_statistics.iterations,
            0,
            f"Run statistics ({run_statistics}) should have a valid number of iterations.",
        )

        result_statistics = statistics.result
        self.assertTrue(
            isinstance(result_statistics, ResultStatistics),
            f"Result statistics ({result_statistics}) should be of type ResultStatistics.",
        )
        self.assertGreaterEqual(
            result_statistics.duration,
            0,
            f"Result statistics ({result_statistics}) should have a valid duration.",
        )
        self.assertGreaterEqual(
            result_statistics.value,
            0,
            f"Result statistics ({result_statistics}) should have a valid value.",
        )

        series_data = statistics.series_data
        self.assertTrue(
            isinstance(series_data, SeriesData),
            f"Series data ({series_data}) should be of type SeriesData.",
        )

    def test_with_check(self):
        with open(os.path.join(os.path.dirname(__file__), "output_with_check.json")) as f:
            json_data = json.load(f)

        nextroute_output = Output.from_dict(json_data)
        check = nextroute_output.solutions[0].check
        self.assertTrue(
            isinstance(check, nextrouteCheck.Output),
            f"Check ({check}) should be of type nextrouteCheck.Output.",
        )
        self.assertTrue(
            isinstance(check.solution, nextrouteCheck.Solution),
            f"Solution ({check.solution}) should be of type nextrouteCheck.checkSolution.",
        )
        self.assertTrue(
            isinstance(check.summary, nextrouteCheck.Summary),
            f"Summary ({check.summary}) should be of type nextrouteCheck.Summary.",
        )
        for plan_unit in check.plan_units:
            self.assertTrue(
                isinstance(plan_unit, nextrouteCheck.PlanUnit),
                f"Plan unit ({plan_unit}) should be of type nextrouteCheck.PlanUnit.",
            )

        for vehicle in check.vehicles:
            self.assertTrue(
                isinstance(vehicle, nextrouteCheck.Vehicle),
                f"Vehicle ({vehicle}) should be of type nextrouteCheck.Vehicle",
            )

    def test_result_statistics_decoding(self):
        test_cases = [
            {
                "name": "value is float",
                "json_stats": '{"duration": 0.1, "value": 1.23}',
            },
            {
                "name": "value is nan",
                "json_stats": '{"duration": 0.1, "value": "nan"}',
            },
            {
                "name": "value is infinity",
                "json_stats": '{"duration": 0.1, "value": "inf"}',
            },
            {
                "name": "value is infinity 2",
                "json_stats": '{"duration": 0.1, "value": "+inf"}',
            },
            {
                "name": "value is -infinity",
                "json_stats": '{"duration": 0.1, "value": "-inf"}',
            },
        ]

        for test in test_cases:
            dict_stats = json.loads(test["json_stats"])
            stats = ResultStatistics.from_dict(dict_stats)
            self.assertTrue(isinstance(stats, ResultStatistics))
            self.assertTrue(isinstance(stats.value, float))

    def test_result_statistics_encoding(self):
        test_cases = [
            {
                "name": "value is float",
                "stats": ResultStatistics(duration=0.1, value=1.23),
            },
            {
                "name": "value is nan",
                "stats": ResultStatistics(duration=0.1, value=math.nan),
            },
            {
                "name": "value is infinity",
                "stats": ResultStatistics(duration=0.1, value=math.inf),
            },
            {
                "name": "value is -infinity",
                "stats": ResultStatistics(duration=0.1, value=-1 * math.inf),
            },
        ]

        for test in test_cases:
            stats = test["stats"]
            dict_stats = stats.to_dict()
            self.assertTrue(isinstance(dict_stats, dict))
            self.assertTrue(isinstance(dict_stats["value"], float))
