# Multi window example (multi_window.json)

This example demonstrates the use of multiple service windows defined for stops.

Find some notes about the example below:

- The vehicles spend virtually all time waiting, since stops are so close and
stop duration is short.
- **Vehicles**:
  - `v1`: Vehicle starts when the earliest window opens.
- **Stops**:
  - Defaults: stops have a default multi `start_time_window` that allows
  servicing between 12:00 & 12:05 and again between 12:30 & 12:35.
  - `Kyoto Imperial Palace` & `Gionmachi`: Share the same time windows, since
  they don't have a specific window defined.
  - `Kinkaku-ji`: Has a specific time window that allows servicing between
  12:20 & 12:25.
