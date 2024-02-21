# Initial stop infeasible example (initial_stops_infeasible_max_duration.json)

This example demonstrates the use of the `initial_stops` with optional (not
`fixed`) stops when facing a constraint on the last stop. I.e., the max duration
of the vehicle imposes a last start time on the end stop of the vehicle.

Find some notes about the example below:

- The only vehicle starts with all stops pre-assigned as `initial_stops`.
- Every stop takes 10 minutes to be served and _some_ (fast) driving occurs
  between the stops. This means that only 3 stops can be served within the max
  duration of 35 minutes. All others should get removed in reverse order, but
  should be assigned again since the route is shortest as 5-4-3.
