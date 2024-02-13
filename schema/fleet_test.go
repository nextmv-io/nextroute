package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/nextmv-io/sdk/nextroute/schema"
)

// TestFleetConverter tests the fleet to nextroute input converter.
func TestFleetConverter(t *testing.T) {
	// Read fleet input.
	fleetInput := schema.FleetInput{}
	err := json.Unmarshal([]byte(fleetInputData), &fleetInput)
	if err != nil {
		t.Error(err)
	}

	// Convert fleet input to nextroute input.
	input, err := fleetInput.ToNextRoute()
	if err != nil {
		t.Error(err)
	}

	// Convert nextroute input to json for comparison because struct has
	// pointers.
	nextrouteInputGot, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		t.Error(err)
	}

	// Do the same for the expected nextroute input to get rid of formatting.
	inputWant := schema.Input{}
	err = json.Unmarshal([]byte(nextrouteInputWant), &inputWant)
	if err != nil {
		t.Error(err)
	}

	nextrouteInputWant, err := json.MarshalIndent(inputWant, "", "  ")
	if err != nil {
		t.Error(err)
	}

	// Compare the two jsons as string.
	if string(nextrouteInputGot) != string(nextrouteInputWant) {
		t.Errorf("got %s; want %s", string(nextrouteInputGot), string(nextrouteInputWant))
	}
}

const fleetInputData = `{
	"defaults": {
	  "vehicles": {
		"start": {
		  "lon": -96.659222,
		  "lat": 33.122746
		},
		"end": {
		  "lon": -96.659222,
		  "lat": 33.122746
		},
		"shift_start": "2021-10-17T09:00:00-06:00",
		"shift_end": "2021-10-17T11:00:00-06:00",
		"speed": 20,
		"capacity": 10,
		"max_stops": 15,
		"max_distance": 100000,
		"max_duration": 50000
	  },
	  "stops": {
		"stop_duration": 120,
		"unassigned_penalty": 200000,
		"earliness_penalty": 2,
		"lateness_penalty": 5,
		"quantity": -5,
		"max_wait": 300,
		"compatibility_attributes": ["A"],
		"hard_window": ["2021-10-17T09:00:00-06:00", "2021-10-17T10:00:00-06:00"]
	  }
	},
	"stop_groups": [["order-24-pickup-1", "order-24-dropoff"]],
	  "duration_groups": [
		{
		  "group": ["order-24-pickup-1", "order-24-dropoff"],
		  "duration": 10
		}
	  ],
	"vehicles": [
	  {
		"id": "vehicle-1",
		"capacity": 105,
		"backlog": ["order-24-pickup-1", "order-24-dropoff"],
		"compatibility_attributes": ["vehicle-1"],
		"speed": 14,
		"max_distance": 200000,
		"max_duration": 55000,
		"stop_duration_multiplier": 1,
		"start": {
		  "lon": -96.659222,
		  "lat": 33.122746
		},
		"end": {
		  "lon": -96.659222,
		  "lat": 33.122746
		},
		"initialization_cost": 2
	  },
	  {
		"id": "vehicle-2",
		"capacity": 95
	  }
	],
	"stops": [
	  {
		"id": "order-1-pickup-1",
		"compatibility_attributes": ["vehicle-1"],
		"position": {
		  "lon": -96.827094,
		  "lat": 33.004745
		},
		"precedes": "order-1-dropoff",
		"quantity": -27,
    "hard_window": [
        "2021-10-17T10:00:00-06:00",
        "2021-10-17T11:00:15-06:00"
      ]
	  },
	  {
		"id": "order-1-dropoff",
		"target_time": "2021-10-17T09:45:00-06:00",
		"position": {
		  "lon": -96.86074,
		  "lat": 33.005741
		},
		"quantity": 27,
		"hard_window": ["2021-10-17T09:00:00-06:00", "2021-10-17T10:00:00-06:00"]
	  },
	  {
		"id": "order-24-pickup-1",
		"position": {
		  "lon": -96.690515,
		  "lat": 32.995981
		},
		"precedes": "order-24-dropoff",
		"quantity": -9,
		"compatibility_attributes": ["vehicle-1"]
	  },
	  {
		"id": "order-24-dropoff",
		"target_time": "2021-10-17T10:55:00-06:00",
		"position": {
		  "lon": -96.783115,
		  "lat": 32.95055
		},
		"quantity": 9,
		"hard_window": ["2021-10-17T09:30:00-06:00", "2021-10-17T12:00:00-06:00"]
	  }
	]
  }`

const nextrouteInputWant = `{
  "stop_groups": [
    [
      "order-24-pickup-1",
      "order-24-dropoff"
    ]
  ],
  "duration_groups": [
    {
      "group": [
        "order-24-pickup-1",
        "order-24-dropoff"
      ],
      "duration": 10
    }
  ],
  "vehicles": [
    {
      "capacity": {
        "default": 105
      },
      "start_level": {
        "default": 0
      },
      "compatibility_attributes": [
        "vehicle-1",
        "vehicle-1_order-24-pickup-1",
        "vehicle-1_order-24-dropoff"
      ],
      "max_distance": 200000,
      "stop_duration_multiplier": 1,
      "start_time": "2021-10-17T09:00:00-06:00",
      "end_time": "2021-10-17T11:00:00-06:00",
      "end_location": {
        "lon": -96.659222,
        "lat": 33.122746
      },
      "max_stops": 15,
      "speed": 14,
      "max_duration": 55000,
      "activation_penalty": 2,
      "start_location": {
        "lon": -96.659222,
        "lat": 33.122746
      },
      "initial_stops": [
        {
          "fixed": false,
          "id": "order-24-pickup-1"
        },
        {
          "fixed": false,
          "id": "order-24-dropoff"
        }
      ],
      "id": "vehicle-1"
    },
    {
      "capacity": {
        "default": 95
      },
      "start_level": {},
      "compatibility_attributes": [],
      "max_distance": 100000,
      "stop_duration_multiplier": 1,
      "start_time": "2021-10-17T09:00:00-06:00",
      "end_time": "2021-10-17T11:00:00-06:00",
      "end_location": {
        "lon": -96.659222,
        "lat": 33.122746
      },
      "max_stops": 15,
      "speed": 20,
      "max_duration": 50000,
      "activation_penalty": 0,
      "start_location": {
        "lon": -96.659222,
        "lat": 33.122746
      },
      "initial_stops": [],
      "id": "vehicle-2"
    }
  ],
  "stops": [
    {
      "precedes": "order-1-dropoff",
      "quantity": {
        "default": -27
      },
      "duration": 120,
      "max_wait": 300,
      "start_time_window": [
        "2021-10-17T10:00:00-06:00",
        "2021-10-17T11:00:00-06:00"
      ],
      "unplanned_penalty": 200000,
      "early_arrival_time_penalty": 2,
      "late_arrival_time_penalty": 5,
      "compatibility_attributes": [
        "vehicle-1"
      ],
      "id": "order-1-pickup-1",
      "location": {
        "lon": -96.827094,
        "lat": 33.004745
      }
    },
    {
      "quantity": {
        "default": 27
      },
      "duration": 120,
      "max_wait": 300,
      "start_time_window": [
        "2021-10-17T09:00:00-06:00",
        "2021-10-17T10:00:00-06:00"
      ],
      "unplanned_penalty": 200000,
      "early_arrival_time_penalty": 2,
      "late_arrival_time_penalty": 5,
      "compatibility_attributes": [
        "A"
      ],
      "target_arrival_time": "2021-10-17T09:45:00-06:00",
      "id": "order-1-dropoff",
      "location": {
        "lon": -96.86074,
        "lat": 33.005741
      }
    },
    {
      "precedes": "order-24-dropoff",
      "quantity": {
        "default": -9
      },
      "duration": 120,
      "max_wait": 300,
      "start_time_window": [
        "2021-10-17T09:00:00-06:00",
        "2021-10-17T10:00:00-06:00"
      ],
      "unplanned_penalty": 200000,
      "early_arrival_time_penalty": 2,
      "late_arrival_time_penalty": 5,
      "compatibility_attributes": [
        "vehicle-1_order-24-pickup-1"
      ],
      "id": "order-24-pickup-1",
      "location": {
        "lon": -96.690515,
        "lat": 32.995981
      }
    },
    {
      "quantity": {
        "default": 9
      },
      "duration": 120,
      "max_wait": 300,
      "start_time_window": [
        "2021-10-17T09:31:00-06:00",
        "2021-10-17T12:00:00-06:00"
      ],
      "unplanned_penalty": 200000,
      "early_arrival_time_penalty": 2,
      "late_arrival_time_penalty": 5,
      "compatibility_attributes": [
        "A_order-24-dropoff"
      ],
      "target_arrival_time": "2021-10-17T10:55:00-06:00",
      "id": "order-24-dropoff",
      "location": {
        "lon": -96.783115,
        "lat": 32.95055
      }
    }
  ]
}`
