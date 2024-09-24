# © 2019-present nextmv.io inc

import json
import os
import unittest

from nextroute.schema import Input, Stop, Vehicle


class TestInput(unittest.TestCase):
    filepath = os.path.join(os.path.dirname(__file__), "input.json")

    def test_from_json(self):
        with open(self.filepath) as f:
            json_data = json.load(f)

        nextroute_input = Input.from_dict(json_data)
        parsed = nextroute_input.to_dict()

        for s, stop in enumerate(parsed["stops"]):
            original_stop = json_data["stops"][s]
            self.assertEqual(
                stop,
                original_stop,
                f"stop: parsed({stop}) and original ({original_stop}) should be equal",
            )

        for v, vehicle in enumerate(parsed["vehicles"]):
            original_vehicle = json_data["vehicles"][v]
            self.assertEqual(
                vehicle,
                original_vehicle,
                f"vehicle: parsed ({vehicle}) and original ({original_vehicle}) should be equal",
            )

        self.assertEqual(
            parsed["defaults"],
            json_data["defaults"],
            f"defaults: parsed ({parsed['defaults']}) and original ({json_data['defaults']}) should be equal",
        )

    def test_from_dict(self):
        with open(self.filepath) as f:
            json_data = json.load(f)

        nextroute_input = Input.from_dict(json_data)
        stops = nextroute_input.stops
        for stop in stops:
            self.assertTrue(
                isinstance(stop, Stop),
                f"Stop ({stop}) should be of type Stop.",
            )

        vehicles = nextroute_input.vehicles
        for vehicle in vehicles:
            self.assertTrue(
                isinstance(vehicle, Vehicle),
                f"Vehicle ({vehicle}) should be of type Vehicle.",
            )
