// Package schema provides the input and output schema for nextroute.
package schema

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

// FleetInput schema.
// DEPRECATION NOTICE: this part of the API is deprecated and is no longer
// maintained. It will be deleted soon. Please use [Input] instead.
type FleetInput struct {
	Options        *Options        `json:"options,omitempty"`
	Defaults       *FleetDefaults  `json:"defaults,omitempty"`
	Vehicles       []FleetVehicle  `json:"vehicles,omitempty"`
	Stops          []FleetStop     `json:"stops,omitempty"`
	StopGroups     [][]string      `json:"stop_groups,omitempty"`
	AlternateStops []FleetStop     `json:"alternate_stops,omitempty"`
	DurationGroups []DurationGroup `json:"duration_groups,omitempty"`
}

// FleetDefaults holds the fleet input default data.
// FleetInput schema.
// DEPRECATION NOTICE: this part of the API is deprecated and is no longer
// maintained. It will be deleted soon. Please use [Defaults] instead.
type FleetDefaults struct {
	Vehicles *FleetVehicleDefaults `json:"vehicles,omitempty"`
	Stops    *FleetStopDefaults    `json:"stops,omitempty"`
}

// FleetVehicleDefaults holds the fleet input vehicle default data.
// FleetInput schema.
// DEPRECATION NOTICE: this part of the API is deprecated and is no longer
// maintained. It will be deleted soon. Please use [VehicleDefaults] instead.
type FleetVehicleDefaults struct {
	Start                   *Location  `json:"start,omitempty"`
	End                     *Location  `json:"end,omitempty"`
	Speed                   *float64   `json:"speed,omitempty"`
	Capacity                any        `json:"capacity,omitempty"`
	ShiftStart              *time.Time `json:"shift_start,omitempty"`
	ShiftEnd                *time.Time `json:"shift_end,omitempty"`
	CompatibilityAttributes []string   `json:"compatibility_attributes,omitempty"`
	MaxStops                *int       `json:"max_stops,omitempty"`
	MaxDistance             *int       `json:"max_distance,omitempty"`
	MaxDuration             *int       `json:"max_duration,omitempty"`
}

// FleetStopDefaults holds the fleet input stop default data.
// DEPRECATION NOTICE: this part of the API is deprecated and is no longer
// maintained. It will be deleted soon. Please use [StopDefaults] instead.
type FleetStopDefaults struct {
	UnassignedPenalty       *int         `json:"unassigned_penalty,omitempty"`
	Quantity                any          `json:"quantity,omitempty"`
	HardWindow              *[]time.Time `json:"hard_window,omitempty"`
	MaxWait                 *int         `json:"max_wait,omitempty"`
	StopDuration            *int         `json:"stop_duration,omitempty"`
	TargetTime              *time.Time   `json:"target_time,omitempty"`
	EarlinessPenalty        *float64     `json:"earliness_penalty,omitempty"`
	LatenessPenalty         *float64     `json:"lateness_penalty,omitempty"`
	CompatibilityAttributes *[]string    `json:"compatibility_attributes,omitempty"`
}

// FleetVehicle holds the fleet input vehicle data.
// DEPRECATION NOTICE: this part of the API is deprecated and is no longer
// maintained. It will be deleted soon. Please use [Vehicle] instead.
type FleetVehicle struct {
	ID                      string    `json:"id,omitempty"`
	Start                   *Location `json:"start,omitempty"`
	End                     *Location `json:"end,omitempty"`
	Speed                   *float64  `json:"speed,omitempty"`
	Capacity                any       `json:"capacity,omitempty"`
	capacity                map[string]any
	ShiftStart              *time.Time `json:"shift_start,omitempty"`
	ShiftEnd                *time.Time `json:"shift_end,omitempty"`
	CompatibilityAttributes []string   `json:"compatibility_attributes,omitempty"`
	MaxStops                *int       `json:"max_stops,omitempty"`
	MaxDistance             *int       `json:"max_distance,omitempty"`
	MaxDuration             *int       `json:"max_duration,omitempty"`
	StopDurationMultiplier  *float64   `json:"stop_duration_multiplier,omitempty"`
	Backlog                 []string   `json:"backlog,omitempty"`
	AlternateStops          []string   `json:"alternate_stops,omitempty"`
	InitializationCost      int        `json:"initialization_cost,omitempty"`
}

// FleetStop holds the fleet input stop data.
// DEPRECATION NOTICE: this part of the API is deprecated and is no longer
// maintained. It will be deleted soon. Please use [Stop] instead.
type FleetStop struct {
	ID                      string   `json:"id,omitempty"`
	Position                Location `json:"position,omitempty"`
	UnassignedPenalty       *int     `json:"unassigned_penalty,omitempty"`
	Quantity                any      `json:"quantity,omitempty"`
	quantity                map[string]any
	Precedes                any `json:"precedes,omitempty"`
	precedes                []string
	Succeeds                any `json:"succeeds,omitempty"`
	succeeds                []string
	HardWindow              *[]time.Time `json:"hard_window,omitempty"`
	MaxWait                 *int         `json:"max_wait,omitempty"`
	StopDuration            *int         `json:"stop_duration,omitempty"`
	TargetTime              *time.Time   `json:"target_time,omitempty"`
	EarlinessPenalty        *float64     `json:"earliness_penalty,omitempty"`
	LatenessPenalty         *float64     `json:"lateness_penalty,omitempty"`
	CompatibilityAttributes *[]string    `json:"compatibility_attributes,omitempty"`
}

// Options adds solver options to the input.
// DEPRECATION NOTICE: this part of the API is deprecated and is no longer
// maintained. It will be deleted soon. Please use [solve.Options] instead.
type Options struct {
	Solver *SolverOptions `json:"solver,omitempty"`
}

// SolverOptions represent the solver runtime duration in legacy fleet.
// DEPRECATION NOTICE: this part of the API is deprecated and is no longer
// maintained. It will be deleted soon. Please use [solve.Options] instead.
type SolverOptions struct {
	Limits *Limits `json:"limits,omitempty"`
}

// Limits represent the solver runtime limitation in fleet.
// DEPRECATION NOTICE: this part of the API is deprecated and is no longer
// maintained. It will be deleted soon. Please use [solve.Options] instead.
type Limits struct {
	Duration string `json:"duration,omitempty"`
}

// dynamicDefaultKey is used when only scalar quantities / capacities are given.
const dynamicDefaultKey = "default"

// ToNextRoute converters a legacy cloud fleet input into nextroute input format.
func (fleetInput FleetInput) ToNextRoute() (Input, error) {
	input := Input{}

	fleetInput.applyStopDefaults()
	fleetInput.applyVehicleDefaults()

	// Get stopMap and update precedes/succeeds any fields.
	stopMap, err := fleetInput.stopMapUpdate()
	if err != nil {
		return input, err
	}

	// Create vehicles with special logic for backlog legacy needs.
	backlogStops := make(map[string]struct{})
	vehicles := make([]Vehicle, len(fleetInput.Vehicles))
	for i, v := range fleetInput.Vehicles {
		v := v
		newAttributes := make([]string, len(v.CompatibilityAttributes))
		copy(newAttributes, v.CompatibilityAttributes)
		if len(v.CompatibilityAttributes) > 0 {
			for _, b := range v.Backlog {
				for _, ca := range v.CompatibilityAttributes {
					newAttributes = append(newAttributes, fmt.Sprintf("%s_%s", ca, b))
				}
			}
		} else {
			newAttributes = append(newAttributes, v.Backlog...)
		}
		newBacklog := make([]InitialStop, 0)
		falseBool := false
		isInBacklog := make(map[string]struct{})
		startLevel := make(map[string]any)
		for _, b := range v.Backlog {
			backlogStops[b] = struct{}{}
			backlogStop := stopMap[b]
			if _, ok := isInBacklog[b]; !ok {
				newBacklog = append(newBacklog, InitialStop{
					Fixed: &falseBool,
					ID:    b,
				})
				isInBacklog[b] = struct{}{}
			}

			for _, p := range backlogStop.precedes {
				if _, ok := isInBacklog[p]; !ok {
					newBacklog = append(newBacklog, InitialStop{
						Fixed: &falseBool,
						ID:    p,
					})
					backlogStops[p] = struct{}{}
					isInBacklog[p] = struct{}{}
				}
			}

			for _, s := range backlogStop.succeeds {
				if _, ok := isInBacklog[s]; !ok {
					newBacklog = append(newBacklog, InitialStop{
						Fixed: &falseBool,
						ID:    s,
					})
					backlogStops[s] = struct{}{}
					isInBacklog[s] = struct{}{}
				}
			}
		}

		for _, b := range newBacklog {
			backlogStop := stopMap[b.ID]
			for k, q := range backlogStop.quantity {
				level, ok := convertToInt(q)
				if !ok {
					return input, fmt.Errorf("could not convert quantity for stop %s to int", backlogStop.ID)
				}
				if _, ok := startLevel[k]; !ok {
					startLevel[k] = 0
				}
				currentLevel, ok := convertToInt(startLevel[k])
				if !ok {
					return input, fmt.Errorf("could not convert quantity for stop %s to int", backlogStop.ID)
				}
				currentLevel += level
				startLevel[k] = currentLevel
			}
		}
		vehicles[i] = Vehicle{
			Capacity:                v.capacity,
			CompatibilityAttributes: &newAttributes,
			MaxDistance:             v.MaxDistance,
			StopDurationMultiplier:  v.StopDurationMultiplier,
			StartTime:               v.ShiftStart,
			EndTime:                 v.ShiftEnd,
			StartLocation:           v.Start,
			EndLocation:             v.End,
			MaxStops:                v.MaxStops,
			Speed:                   v.Speed,
			MaxDuration:             v.MaxDuration,
			InitialStops:            &newBacklog,
			ActivationPenalty:       &v.InitializationCost,
			ID:                      v.ID,
			StartLevel:              startLevel,
			CustomData:              nil,
			MinStops:                nil,
			MinStopsPenalty:         nil,
			MaxWait:                 nil,
		}
	}

	// Create stops with legacy backlog feature.
	stops := createStops(fleetInput, backlogStops)

	// Put new input format together and return it.
	if fleetInput.StopGroups != nil {
		input.StopGroups = &fleetInput.StopGroups
	}
	if fleetInput.DurationGroups != nil {
		input.DurationGroups = &fleetInput.DurationGroups
	}
	input.Vehicles = vehicles
	input.Stops = stops

	return input, nil
}

func (fleetInput *FleetInput) stopMapUpdate() (map[string]FleetStop, error) {
	stopMap := make(map[string]FleetStop)
	for idx, stop := range fleetInput.Stops {
		if stop.Precedes == nil { //nolint:gocritic
			fleetInput.Stops[idx].precedes = []string{}
		} else if reflect.ValueOf(stop.Precedes).Kind() == reflect.Slice {
			slice := stop.Precedes.([]any)
			precedes := make([]string, len(slice))
			for i, p := range slice {
				val, ok := p.(string)
				if !ok {
					return nil, fmt.Errorf("could not parse Precedes field for stop %s", stop.ID)
				}
				precedes[i] = val
			}
			fleetInput.Stops[idx].precedes = precedes
		} else {
			precedes, ok := stop.Precedes.(string)
			if !ok {
				return nil, fmt.Errorf("could not parse Precedes field for stop %s", stop.ID)
			}
			fleetInput.Stops[idx].precedes = []string{precedes}
		}

		if stop.Succeeds == nil { //nolint:gocritic
			fleetInput.Stops[idx].succeeds = []string{}
		} else if reflect.ValueOf(stop.Succeeds).Kind() == reflect.Slice {
			slice := stop.Succeeds.([]any)
			succeeds := make([]string, len(slice))
			for i, p := range slice {
				val, ok := p.(string)
				if !ok {
					return nil, fmt.Errorf("could not parse Succeeds field for stop %s", stop.ID)
				}
				succeeds[i] = val
			}
			fleetInput.Stops[idx].succeeds = succeeds
		} else {
			succeeds, ok := stop.Succeeds.(string)
			if !ok {
				return nil, fmt.Errorf("could not parse Succeeds field for stop %s", stop.ID)
			}
			fleetInput.Stops[idx].succeeds = []string{succeeds}
		}

		// Handle multi capacity fields
		if reflect.ValueOf(stop.Quantity).Kind() == reflect.Map {
			// We detected a map and need to handle it
			quantities, ok := stop.Quantity.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("could not parse Quanitity field for stop %s", stop.ID)
			}
			// Init map
			fleetInput.Stops[idx].quantity = make(map[string]any)
			// Set all values
			for kind, quant := range quantities {
				rounded, ok := convertToInt(quant)
				if !ok {
					return nil, fmt.Errorf("could not parse Quanitity field for stop %s", stop.ID)
				}
				fleetInput.Stops[idx].quantity[kind] = rounded
			}
		} else {
			// Since it is not a map, we expect a scalar value
			quantity, ok := convertToInt(stop.Quantity)
			if !ok {
				return nil, fmt.Errorf("could not parse Quanitity field for stop %s", stop.ID)
			}
			fleetInput.Stops[idx].quantity = map[string]any{
				dynamicDefaultKey: quantity,
			}
		}
		stopMap[stop.ID] = fleetInput.Stops[idx]
	}
	for v, vehicle := range fleetInput.Vehicles {
		vehicle, err := handleCapacity(vehicle)
		if err != nil {
			return stopMap, err
		}
		fleetInput.Vehicles[v] = vehicle
	}
	return stopMap, nil
}

func handleCapacity(vehicle FleetVehicle) (FleetVehicle, error) {
	if reflect.ValueOf(vehicle.Capacity).Kind() == reflect.Map {
		capacities, ok := vehicle.Capacity.(map[string]any)
		if !ok {
			return vehicle, fmt.Errorf("could not parse Capacity field for vehicle %s", vehicle.ID)
		}

		vehicle.capacity = make(map[string]any)

		for kind, cap := range capacities {
			rounded, ok := convertToInt(cap)
			if !ok {
				return vehicle, fmt.Errorf("could not parse Capacity field for vehicle %s", vehicle.ID)
			}
			vehicle.capacity[kind] = rounded
		}
	} else {
		capacity, ok := convertToInt(vehicle.Capacity)
		if !ok {
			return vehicle, fmt.Errorf("could not parse Capacity field for vehicle %s", vehicle.ID)
		}
		vehicle.capacity = map[string]any{
			dynamicDefaultKey: capacity,
		}
	}
	return vehicle, nil
}

func createStops(fleetInput FleetInput, backlogStops map[string]struct{}) []Stop {
	stops := make([]Stop, len(fleetInput.Stops))
	for i, s := range fleetInput.Stops {
		compats := make([]string, 0)
		if s.CompatibilityAttributes != nil && len(*s.CompatibilityAttributes) > 0 {
			if _, ok := backlogStops[s.ID]; ok {
				for _, ca := range *s.CompatibilityAttributes {
					compats = append(compats, fmt.Sprintf("%s_%s", ca, s.ID))
				}
			} else {
				compats = *s.CompatibilityAttributes
			}
		} else if _, ok := backlogStops[s.ID]; ok {
			compats = append(compats, s.ID)
		}
		stops[i] = Stop{
			Precedes:                s.Precedes,
			Quantity:                s.quantity,
			Succeeds:                s.Succeeds,
			Duration:                s.StopDuration,
			MaxWait:                 s.MaxWait,
			UnplannedPenalty:        s.UnassignedPenalty,
			EarlyArrivalTimePenalty: s.EarlinessPenalty,
			LateArrivalTimePenalty:  s.LatenessPenalty,
			CompatibilityAttributes: &compats,
			TargetArrivalTime:       s.TargetTime,
			ID:                      s.ID,
			Location:                s.Position,
			CustomData:              nil,
		}
		if s.HardWindow != nil {
			timeWindow := *s.HardWindow
			for i := range timeWindow {
				timeWindow[i] = roundToMinute(timeWindow[i])
			}
			stops[i].StartTimeWindow = timeWindow
		}
	}
	return stops
}

func roundToMinute(t time.Time) time.Time {
	if t.Minute() > 29 {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()+1, 0, 0, t.Location())
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
}

// intTolerance specified the absolute offset tolerable for dynamic fields that
// are naturally float, but where integer values are expected.
const intTolerance = 0.00001

// convertToInt tries to convert the given unknown typed value to int. Returns
// true, iff the conversion was successful.
func convertToInt(unknown interface{}) (int, bool) {
	// Convert unknown to float
	floatType := reflect.TypeOf(float64(0))
	v := reflect.ValueOf(unknown)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, false
	}
	fv := v.Convert(floatType)
	floatValue := fv.Float()
	// Round to next int (not ok if offset out of tolerance)
	rounded := math.Round(floatValue)
	if math.Abs(rounded-floatValue) > intTolerance {
		return 0, false
	}
	return int(rounded), true
}
