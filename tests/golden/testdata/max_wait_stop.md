# Max wait time example (max_wait.json)

This example demonstrates the use of the `max_wait` parameter to set a maximum
wait time for a single stop.

Find some notes about the example below:

- The vehicles spend virtually all time waiting, since stops are so close and
stop duration is short.
- **Vehicles**:
  - `v1`: Starts its shift at 12:00.
- **Stops**:
  - Defaults: stops have default `max_wait` of 30 minutes.
  - `Kyoto Imperial Palace`: Can't be serviced. The window opens _after_ the
  stop's maximum wait time and we can't spend sufficient time at other stops
  beforehand.
  - `Gionmachi`: Can be serviced. The window opens _before_ the stop's maximum
  wait time.
  - `Kinkaku-ji`: Can't be serviced. The _inherited_ maximum wait time is 30
  minutes, but the window opens _after_ that. Agaim, we can't spend sufficient
  time at other stops beforehand.
  - `Arashiyama Bamboo Forest`: Can be serviced. The stop has no window, so the
  maximum wait time is not relevant.
