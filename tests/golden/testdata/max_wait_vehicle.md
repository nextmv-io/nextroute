# Max wait time example (max_wait.json)

This example demonstrates the use of the `max_wait` parameter to set maximum
accumulative wait times for vehicles.

Find some notes about the example below:

- The vehicles spend virtually all time waiting, since stops are so close and
stop duration is short.
- **Vehicles**:
  - `v1`: Has a maximum wait time of 30 minutes (inherited from defaults).
  - `v2`: Has a maximum wait time of 20 minutes.
- **Stops**:
  - `Fushimi Inari Taisha`: Cannot be assigned, as the window opens _after_ the
  maximum accumulated time of all vehicles.
  - `Kiyomizu-dera`: Can only be serviced by the first vehicle, since the window
  opens after the maximum accumulated time of the second vehicle.
  - `Nijō Castle`: Can be serviced by both vehicles, since the window opens
  before the maximum accumulated time of both vehicles.
  - `Kyoto Imperial Palace`: Can only be serviced after waiting until the window
  opens at other stops beforehand (as it also has a per stop `max_wait`).
  - `Gionmachi`: Can be serviced by both vehicles, since its window is within
  accumulated vehicle and individual stop maxima.
  - `Kinkaku-ji`: Can be serviced, since the inherited individual stop maximum
  is not exceeded.
  - `Arashiyama Bamboo Forest`: Can be serviced, since it does not have a window
  that can cause waiting.
