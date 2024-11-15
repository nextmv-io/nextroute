# Missing values in no mix constraint (no_mix_null.json)

This example demonstrates missing values when defining a _no mix_ constraint.
I.e., some of the stops either have no `mixing_items` defined or have it set to
`null`. These stops can be planned anywhere in a route, independent of the
current mix of items in the vehicle.

Find some notes about the example below:

- There is only one vehicle servicing all the stops.
- Stops `north` and `south` belong to mixing group `A`.
- Stops `east` and `west` belong to mixing group `B`.
- The transports from north to south and from east to west cannot overlap. This
  means that a route like `north -> east -> south -> west` is not feasible.
- All other stops can go anywhere in the route and should be planned in a way
  that minimizes the total travel time.
- All stops are located on a circle with their names indicating the cardinal
  direction they are located at. This is useful for checking the time-efficiency
  of the solution.
