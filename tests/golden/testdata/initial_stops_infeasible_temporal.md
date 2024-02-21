# Initial stop infeasible example (initial_stops_infeasible_temporal.json)

This example demonstrates the use of the `initial_stops` with optional (not
`fixed`) stops while one of the initial stops is infeasible due to temporal
constraints.

Find some notes about the example below:

- The stop `stop6` is infeasible in the sequence it is initially assigned as,
since its window closes very early. However, it can simply be approached first
to still get planned (and should be).
- The stop `stop3` is challenging to handle as it is part of a plan unit with
`stop1`. Hence, it will get assigned before `stop2` when its `max_wait` would
cause a temporal violation. However, this is resolved when `stop2` eventually
gets planned between `stop1` and `stop3`. Its `stop_duration` is sufficient to
avoid long waiting time.

TODO: add further stops that need to be removed for the route to become feasible
