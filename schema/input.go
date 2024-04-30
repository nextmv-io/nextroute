// © 2019-present nextmv.io inc

// Package schema provides the input and output schema for
package schema

import (
	"time"

	"github.com/nextmv-io/sdk/measure"
)

// Input is the default input schema for nextroute.
type Input struct {
	// Options arbitrary options.
	Options any `json:"options,omitempty"`
	// CustomData arbitrary custom data.
	CustomData any `json:"custom_data,omitempty"`
	// Defaults default properties for vehicles and stops.
	Defaults *Defaults `json:"defaults,omitempty"`
	// StopGroups group of stops that must be part of the same route.
	StopGroups *[][]string `json:"stop_groups,omitempty"`
	// DurationMatrix matrix of durations in seconds between stops.
	DurationMatrix *[][]float64 `json:"duration_matrix,omitempty"`
	// DistanceMatrix matrix of distances in meters between stops.
	DistanceMatrix *[][]float64 `json:"distance_matrix,omitempty"`
	// DurationGroups duration in seconds added when approaching the group.
	DurationGroups *[]DurationGroup `json:"duration_groups,omitempty"`
	// Vehicles to route.
	Vehicles []Vehicle `json:"vehicles,omitempty"`
	// Stops that will be routed and assigned to the vehicles.
	Stops []Stop `json:"stops,omitempty"`
	// AlternateStops a set of alternate stops for vehicles.
	AlternateStops *[]AlternateStop `json:"alternate_stops,omitempty"`
}

// Defaults contains default values for vehicles and stops.
type Defaults struct {
	// Vehicles default values for vehicles.
	Vehicles *VehicleDefaults `json:"vehicles,omitempty"`
	// Stops default values for stops.
	Stops *StopDefaults `json:"stops,omitempty"`
}

// VehicleDefaults contains default values for vehicles.
type VehicleDefaults struct {
	// Capacity of the vehicle.
	Capacity any `json:"capacity,omitempty"`
	// StartLevel initial level of the vehicle.
	StartLevel any `json:"start_level,omitempty"`
	// StartLocation location where the vehicle starts..
	StartLocation *Location `json:"start_location,omitempty"`
	// EndLocation location where the vehicle ends..
	EndLocation *Location `json:"end_location,omitempty"`
	// Speed of the vehicle in meters per second.
	Speed *float64 `json:"speed,omitempty" minimumExclusive:"0"`
	// StartTime time when the vehicle starts its route.
	StartTime *time.Time `json:"start_time,omitempty"`
	// EndTime latest time at which the vehicle ends its route.
	EndTime *time.Time `json:"end_time,omitempty"`
	// MinStops minimum stops that a vehicle should visit.
	MinStops *int `json:"min_stops,omitempty" minimum:"0"`
	// MinStopsPenalty penalty for not visiting the minimum number of stops.
	MinStopsPenalty *float64 `json:"min_stops_penalty,omitempty" minimum:"0"`
	// MaxStops maximum number of stops that the vehicle can visit.
	MaxStops *int `json:"max_stops,omitempty" minimum:"0"`
	// MaxDistance maximum distance in meters that the vehicle can travel.
	MaxDistance *int `json:"max_distance,omitempty" minimum:"0"`
	// MaxDuration maximum duration in seconds that the vehicle can travel.
	MaxDuration *int `json:"max_duration,omitempty" minimum:"0"`
	// MaxWait maximum aggregated waiting time that the vehicle can wait across route stops.
	MaxWait *int `json:"max_wait,omitempty" minimum:"0"`
	// CompatibilityAttributes attributes that the vehicle is compatible with.
	CompatibilityAttributes *[]string `json:"compatibility_attributes,omitempty" uniqueItems:"true"`
	// ActivationPenalty penalty of using the vehicle.
	ActivationPenalty *int `json:"activation_penalty,omitempty" minimum:"0"`
	// AlternateStops a set of alternate stops for which only one should be serviced.
	AlternateStops *[]string `json:"alternate_stops,omitempty" uniqueItems:"true"`
}

// StopDefaults contains default values for stops.
type StopDefaults struct {
	// UnplannedPenalty penalty for not planning a stop.
	UnplannedPenalty *int `json:"unplanned_penalty,omitempty" minimum:"0"`
	// Quantity of the stop.
	Quantity any `json:"quantity,omitempty"`
	// StartTimeWindow time window in which the stop can start service.
	StartTimeWindow any `json:"start_time_window,omitempty"`
	// MaxWait maximum waiting duration in seconds at the stop.
	MaxWait *int `json:"max_wait,omitempty" minimum:"0"`
	// Duration in seconds that the stop takes.
	Duration *int `json:"duration,omitempty" minimum:"0"`
	// TargetArrivalTime at the stop.
	TargetArrivalTime *time.Time `json:"target_arrival_time,omitempty"`
	// EarlyArrivalTimePenalty penalty per second for arriving at the stop before the target arrival time.
	EarlyArrivalTimePenalty *float64 `json:"early_arrival_time_penalty,omitempty" minimum:"0"`
	// LateArrivalTimePenalty penalty per second for arriving at the stop after the target arrival time.
	LateArrivalTimePenalty *float64 `json:"late_arrival_time_penalty,omitempty" minimum:"0"`
	// CompatibilityAttributes attributes that the stop is compatible with.
	CompatibilityAttributes *[]string `json:"compatibility_attributes,omitempty" uniqueItems:"true"`
}

// Vehicle represents a vehicle.
type Vehicle struct {
	// Capacity of the vehicle.
	Capacity any `json:"capacity,omitempty"`
	// StartLevel initial level of the vehicle.
	StartLevel any `json:"start_level,omitempty"`
	// CustomData arbitrary custom data.
	CustomData any `json:"custom_data,omitempty"`
	// CompatibilityAttributes attributes that the vehicle is compatible with.
	CompatibilityAttributes *[]string `json:"compatibility_attributes,omitempty" uniqueItems:"true"`
	// MaxDistance maximum distance in meters that the vehicle can travel.
	MaxDistance *int `json:"max_distance,omitempty" minimum:"0"`
	// StopDurationMultiplier multiplier for the duration of stops.
	StopDurationMultiplier *float64 `json:"stop_duration_multiplier,omitempty"`
	// StartTime time when the vehicle starts its route.
	StartTime *time.Time `json:"start_time,omitempty"`
	// EndTime latest time at which the vehicle ends its route.
	EndTime *time.Time `json:"end_time,omitempty"`
	// EndLocation location where the vehicle ends.
	EndLocation *Location `json:"end_location,omitempty"`
	// MinStops minimum stops that a vehicle should visit.
	MinStops *int `json:"min_stops,omitempty" minimum:"0"`
	// MinStopsPenalty penalty for not visiting the minimum number of stops.
	MinStopsPenalty *float64 `json:"min_stops_penalty,omitempty" minimum:"0"`
	// MaxStops maximum number of stops that the vehicle can visit.
	MaxStops *int `json:"max_stops,omitempty"`
	// Speed of the vehicle in meters per second.
	Speed *float64 `json:"speed,omitempty" minimumExclusive:"0"`
	// MaxDuration maximum duration in seconds that the vehicle can travel.
	MaxDuration *int `json:"max_duration,omitempty" minimum:"0"`
	// MaxWait maximum aggregated waiting time that the vehicle can wait across route stops.
	MaxWait *int `json:"max_wait,omitempty" minimum:"0"`
	// ActivationPenalty penalty of using the vehicle.
	ActivationPenalty *int `json:"activation_penalty,omitempty" minimum:"0"`
	// StartLocation location where the vehicle starts.
	StartLocation *Location `json:"start_location,omitempty"`
	// AlternateStops a set of alternate stops for which only one should be serviced.
	AlternateStops *[]string `json:"alternate_stops,omitempty" uniqueItems:"true"`
	// InitialStops initial stops planned on the vehicle.
	InitialStops *[]InitialStop `json:"initial_stops,omitempty" uniqueItems:"true"`
	// ID of the vehicle.
	ID string `json:"id,omitempty"`
}

// InitialStop represents an initial stop.
type InitialStop struct {
	// Fixed whether the stop is fixed or not.
	Fixed *bool `json:"fixed,omitempty"`
	// ID unique identifier for the stop.
	ID string `json:"id"`
}

// AlternateStop represents an alternate stop.
type AlternateStop struct {
	// Quantity of the stop.
	Quantity any `json:"quantity,omitempty"`
	// Duration in seconds that the stop takes.
	Duration *int `json:"duration,omitempty" minimum:"0"`
	// CustomData arbitrary custom data.
	CustomData any `json:"custom_data,omitempty"`
	// MaxWait maximum waiting duration in seconds at the stop.
	MaxWait *int `json:"max_wait,omitempty" minimum:"0"`
	// StartTimeWindow time window in which the stop can start service.
	StartTimeWindow any `json:"start_time_window,omitempty"`
	// UnplannedPenalty penalty for not planning a stop.
	UnplannedPenalty *int `json:"unplanned_penalty,omitempty" minimum:"0"`
	// EarlyArrivalTimePenalty penalty per second for arriving at the stop before the target arrival time.
	EarlyArrivalTimePenalty *float64 `json:"early_arrival_time_penalty,omitempty" minimum:"0"`
	// LateArrivalTimePenalty penalty per second for arriving at the stop after the target arrival time.
	LateArrivalTimePenalty *float64 `json:"late_arrival_time_penalty,omitempty" minimum:"0"`
	// TargetArrivalTime at the stop.
	TargetArrivalTime *time.Time `json:"target_arrival_time,omitempty"`
	// ID unique identifier for the stop.
	ID string `json:"id,omitempty"`
	// Location where the stop is.
	Location Location `json:"location,omitempty"`
}

// Stop represents a stop.
type Stop struct {
	// Precedes stops that must be visited after this one on the same route.
	Precedes any `json:"precedes,omitempty"`
	// Quantity of the stop.
	Quantity any `json:"quantity,omitempty"`
	// Succeeds stops that must be visited before this one on the same route.
	Succeeds any `json:"succeeds,omitempty"`
	// CustomData arbitrary custom data.
	CustomData any `json:"custom_data,omitempty"`
	// Duration in seconds that the stop takes.
	Duration *int `json:"duration,omitempty" minimum:"0"`
	// MaxWait maximum waiting duration in seconds at the stop.
	MaxWait *int `json:"max_wait,omitempty" minimum:"0"`
	// StartTimeWindow time window in which the stop can start service.
	StartTimeWindow any `json:"start_time_window,omitempty"`
	// UnplannedPenalty penalty for not planning a stop.
	UnplannedPenalty *int `json:"unplanned_penalty,omitempty" minimum:"0"`
	// EarlyArrivalTimePenalty penalty per second for arriving at the stop before the target arrival time.
	EarlyArrivalTimePenalty *float64 `json:"early_arrival_time_penalty,omitempty" minimum:"0"`
	// LateArrivalTimePenalty penalty per second for arriving at the stop after the target arrival time.
	LateArrivalTimePenalty *float64 `json:"late_arrival_time_penalty,omitempty" minimum:"0"`
	// CompatibilityAttributes attributes that the stop is compatible with.
	CompatibilityAttributes *[]string `json:"compatibility_attributes,omitempty" uniqueItems:"true"`
	// TargetArrivalTime at the stop.
	TargetArrivalTime *time.Time `json:"target_arrival_time,omitempty"`
	// ID unique identifier for the stop.
	ID string `json:"id,omitempty"`
	// Location where the stop is.
	Location Location `json:"location,omitempty"`
	// MixingItems defines the items that are inserted or removed from the vehicle when visiting the stop.
	MixingItems any `json:"mixing_items,omitempty"`
}

// MixItem is an item that is used to specify the type of mix.
type MixItem struct {
	// Name is the name of the mix item.
	Name string `json:"name"`
	// Quantity is the number units of the mix items are inserted or removed from a
	// vehicle.
	Quantity int `json:"quantity"`
}

// Location represents a geographical location.
type Location struct {
	// Lon longitude of the location.
	Lon float64 `json:"lon" minimum:"-180" maximum:"180"`
	// Lat latitude of the location.
	Lat float64 `json:"lat" minimum:"-90" maximum:"90"`
}

// ToPoint converts a schema.Location to a measure.Point.
func (l Location) ToPoint() measure.Point {
	return measure.Point{l.Lon, l.Lat}
}

// DurationGroup represents a group of stops that get additional duration
// whenever a stop of the group is approached for the first time.
type DurationGroup struct {
	// Group stop IDs contained in the group.
	Group []string `json:"group,omitempty" uniqueItems:"true"`
	// Duration to add when visiting the group.
	Duration int `json:"duration,omitempty" minimum:"0"`
}
