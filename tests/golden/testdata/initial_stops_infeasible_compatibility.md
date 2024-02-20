# Initial stop infeasible example (initial_stops_infeasible_compatibility.json)

This example demonstrates the use of the `initial_stops` with optional (not
`fixed`) stops while some of the initial stops are infeasible.

Find some notes about the example below:

- The stop `Fushimi Inari Taisha` is infeasible with the vehicle it is initially
assigned to due to its `compatibility_attributes`. However, it can be assigned
to the other vehicle (and should be).
- The stop `Kiyomizu-dera` is incompatible with both vehicles and thus cannot be
assigned to any of them. It should not be planned at all, but we still expect a
solution including the other stops.
