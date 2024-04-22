// © 2019-present nextmv.io inc

package schema

import (
	"time"

	"github.com/nextmv-io/nextroute/check/schema"
)

// SolutionOutput represents a solution as JSON.
type SolutionOutput struct {
	// Unplanned is the list of stops that were not planned in the solution.
	Unplanned []StopOutput `json:"unplanned"`
	// Vehicles is the list of vehicles that were used in the solution.
	Vehicles []VehicleOutput `json:"vehicles"`
	// Objective is the objective of the solution.
	Objective ObjectiveOutput `json:"objective"`
	// Check is the check of the solution.
	Check *schema.Output `json:"check,omitempty"`
}

// StopOutput is the basic struct for a stop.
type StopOutput struct {
	// ID is the ID of the stop.
	ID string `json:"id"`
	// Location is the location of the stop.
	Location Location `json:"location"`
	// CustomData is the custom data of the stop.
	CustomData any `json:"custom_data,omitempty"`
}

// VehicleOutput holds the solution of the ModelVehicle Routing Problem.
type VehicleOutput struct {
	// ID is the ID of the vehicle.
	ID string `json:"id"`
	// Route is the list of stops in the route of the vehicle.
	Route []PlannedStopOutput `json:"route"`
	// RouteTravelDuration is the total travel duration of the vehicle.
	RouteTravelDuration int `json:"route_travel_duration"`
	// RouteTravelDistance is the total travel distance of the vehicle.
	RouteTravelDistance int `json:"route_travel_distance,omitempty"`
	// RouteStopsDuration is the total stops duration of the vehicle.
	RouteStopsDuration int `json:"route_stops_duration,omitempty"`
	// RouteWaitingDuration is the total waiting duration of the vehicle.
	RouteWaitingDuration int `json:"route_waiting_duration,omitempty"`
	// RouteDuration is the total duration of the vehicle.
	RouteDuration int `json:"route_duration"`
	// CustomData is the custom data of the vehicle.
	CustomData any `json:"custom_data,omitempty"`
	// AlternateStops is the list of alternate stops selected.
	AlternateStops *[]string `json:"alternate_stops,omitempty"`
}

// PlannedStopOutput adds information to the input stop.
type PlannedStopOutput struct {
	// Stop is the input stop.
	Stop StopOutput `json:"stop"`
	// TravelDuration is the travel duration of the stop.
	TravelDuration int `json:"travel_duration"`
	// CumulativeTravelDuration is the total travel duration to get to this location.
	CumulativeTravelDuration int `json:"cumulative_travel_duration"`
	// TravelDistance is the travel distance of the stop.
	TravelDistance int `json:"travel_distance,omitempty"`
	// CumulativeTravelDistance is the total travel distance to get to this location.
	CumulativeTravelDistance int `json:"cumulative_travel_distance,omitempty"`
	// TargetArrivalTime is the target arrival time of the stop.
	TargetArrivalTime *time.Time `json:"target_arrival_time,omitempty"`
	// ArrivalTime is the arrival time of the stop.
	ArrivalTime *time.Time `json:"arrival_time,omitempty"`
	// WaitingDuration is the waiting duration of the stop in seconds.
	WaitingDuration int `json:"waiting_duration,omitempty"`
	// StartTime is the start time of the stop.
	StartTime *time.Time `json:"start_time,omitempty"`
	// Duration is the duration of the stop in seconds.
	Duration int `json:"duration,omitempty"`
	// EndTime is the end time of the stop.
	EndTime *time.Time `json:"end_time,omitempty"`
	// EarlyArrivalDuration is the early arrival duration of the stop in seconds.
	EarlyArrivalDuration int `json:"early_arrival_duration,omitempty"`
	// LateArrivalDuration is the late arrival duration of the stop in seconds.
	LateArrivalDuration int `json:"late_arrival_duration,omitempty"`
	// MixItems is the mix items of the stop.
	MixItems any `json:"mix_items,omitempty"`
	// CustomData is the custom data of the stop.
	CustomData any `json:"custom_data,omitempty"`
}

// ObjectiveOutput represents an objective as JSON.
type ObjectiveOutput struct {
	// Name is the name of the objective.
	Name string `json:"name"`
	// Objectives is the list of objectives.
	Objectives []ObjectiveOutput `json:"objectives,omitempty"`
	// Factor is the factor of the objective.
	Factor float64 `json:"factor,omitempty"`
	// Base is the value of the objective before the factor is applied.
	Base float64 `json:"base,omitempty"`
	// Value is the value of the objective after the factor is applied.
	Value float64 `json:"value"`
	// CustomData is the custom data of the objective.
	CustomData any `json:"custom_data,omitempty"`
}

// CustomResultStatistics is an example of custom result statistics that can be
// added to the output and used in experiments.
type CustomResultStatistics struct {
	// ActivatedVehicles is the number of vehicles that were used in the
	// solution.
	ActivatedVehicles int `json:"activated_vehicles"`
	// UnplannedStops is the number of stops that were not planned in the
	// solution.
	UnplannedStops int `json:"unplanned_stops"`
	// MaxTravelDuration is the maximum travel duration of a vehicle in the
	// solution.
	MaxTravelDuration int `json:"max_travel_duration"`
	// MaxDuration is the maximum duration of a vehicle (including waiting
	// times) in the solution.
	MaxDuration int `json:"max_duration"`
	// MinTravelDuration is the minimum travel duration of a vehicle in the
	// solution, excluding vehicles that were not used.
	MinTravelDuration int `json:"min_travel_duration"`
	// MinDuration is the minimum duration of a vehicle (including waiting
	// times) in the solution, excluding vehicles that were not used.
	MinDuration int `json:"min_duration"`
	// MaxStopsInRoute is the maximum number of stops in a vehicle's route in
	// the solution. The start and end stops of the vehicle are not considered.
	MaxStopsInVehicle int `json:"max_stops_in_vehicle"`
	// MinStopsInRoute is the minimum number of stops in a vehicle's route in
	// the solution. The start and end stops of the vehicle are not considered.
	MinStopsInVehicle int `json:"min_stops_in_vehicle"`
}
