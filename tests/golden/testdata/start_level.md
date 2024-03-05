# Start level example (start_level.json)

This example demonstrates the use of the `start_level` parameter of the capacity
feature.

Find some notes about the example below:

- All stops have a `-1` quantity (all pickups), hence, the vehicles can service
  as many as they have capacity but no more.
- The first vehicle `v1` defines a start level of 0 and can therefore service
  as many stops as it has capacity.
- The second vehicle `v2` defines a start level equal to its capacity and can
  therefore service no stops at all.
