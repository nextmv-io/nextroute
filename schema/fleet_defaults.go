// Package schema provides the input and output schema for nextroute.
package schema

func (fleetInput *FleetInput) applyVehicleDefaults() {
	// Specify some defaults for missing values
	capacity := 0
	vehicleCompatibilities := []string{}

	// Apply all vehicle defaults, if no explicit value is set
	// Apply general default, if neither explicit nor default value is given
	for v := range fleetInput.Vehicles {
		vehicleDefaults := fleetInput.Defaults != nil && fleetInput.Defaults.Vehicles != nil
		if fleetInput.Vehicles[v].Start == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.Start != nil {
				fleetInput.Vehicles[v].Start = fleetInput.Defaults.Vehicles.Start
			}
		}

		if fleetInput.Vehicles[v].End == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.End != nil {
				fleetInput.Vehicles[v].End = fleetInput.Defaults.Vehicles.End
			}
		}

		if fleetInput.Vehicles[v].Speed == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.Speed != nil {
				fleetInput.Vehicles[v].Speed = fleetInput.Defaults.Vehicles.Speed
			}
		}

		if fleetInput.Vehicles[v].Capacity == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.Capacity != nil {
				fleetInput.Vehicles[v].Capacity = fleetInput.Defaults.Vehicles.Capacity
			} else {
				fleetInput.Vehicles[v].Capacity = &capacity
			}
		}

		if fleetInput.Vehicles[v].ShiftStart == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.ShiftStart != nil {
				fleetInput.Vehicles[v].ShiftStart = fleetInput.Defaults.Vehicles.ShiftStart
			}
		}

		if fleetInput.Vehicles[v].ShiftEnd == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.ShiftEnd != nil {
				fleetInput.Vehicles[v].ShiftEnd = fleetInput.Defaults.Vehicles.ShiftEnd
			}
		}

		if fleetInput.Vehicles[v].CompatibilityAttributes == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.CompatibilityAttributes != nil {
				fleetInput.Vehicles[v].CompatibilityAttributes = fleetInput.Defaults.Vehicles.CompatibilityAttributes
			} else {
				fleetInput.Vehicles[v].CompatibilityAttributes = vehicleCompatibilities
			}
		}

		if fleetInput.Vehicles[v].MaxStops == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.MaxStops != nil {
				fleetInput.Vehicles[v].MaxStops = fleetInput.Defaults.Vehicles.MaxStops
			}
		}

		if fleetInput.Vehicles[v].MaxDistance == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.MaxDistance != nil {
				fleetInput.Vehicles[v].MaxDistance = fleetInput.Defaults.Vehicles.MaxDistance
			}
		}

		if fleetInput.Vehicles[v].MaxDuration == nil {
			if vehicleDefaults && fleetInput.Defaults.Vehicles.MaxDuration != nil {
				fleetInput.Vehicles[v].MaxDuration = fleetInput.Defaults.Vehicles.MaxDuration
			}
		}

		if fleetInput.Vehicles[v].StopDurationMultiplier == nil {
			multiplier := 1.0
			fleetInput.Vehicles[v].StopDurationMultiplier = &multiplier
		}
	}
}

// applyStopDefaults applies the given default values for all values of vehicles
// and stops not explicitly defined.
func (fleetInput *FleetInput) applyStopDefaults() {
	// Specify some defaults for missing values
	quantity := 0
	stopDuration := 0
	stopCompatibilities := []string{}

	// Apply all vehicle defaults, if no explicit value is set
	// Apply general default, if neither explicit nor default value is given
	for s := range fleetInput.Stops {
		stopDefaults := fleetInput.Defaults != nil && fleetInput.Defaults.Stops != nil

		if fleetInput.Stops[s].UnassignedPenalty == nil {
			if stopDefaults && fleetInput.Defaults.Stops.UnassignedPenalty != nil {
				fleetInput.Stops[s].UnassignedPenalty = fleetInput.Defaults.Stops.UnassignedPenalty
			}
		}

		if fleetInput.Stops[s].Quantity == nil {
			if stopDefaults && fleetInput.Defaults.Stops.Quantity != nil {
				fleetInput.Stops[s].Quantity = fleetInput.Defaults.Stops.Quantity
			} else {
				fleetInput.Stops[s].Quantity = &quantity
			}
		}

		if fleetInput.Stops[s].HardWindow == nil {
			if stopDefaults && fleetInput.Defaults.Stops.HardWindow != nil {
				fleetInput.Stops[s].HardWindow = fleetInput.Defaults.Stops.HardWindow
			}
		}

		if fleetInput.Stops[s].MaxWait == nil {
			if stopDefaults && fleetInput.Defaults.Stops.MaxWait != nil {
				fleetInput.Stops[s].MaxWait = fleetInput.Defaults.Stops.MaxWait
			}
		}

		if fleetInput.Stops[s].StopDuration == nil {
			if stopDefaults && fleetInput.Defaults.Stops.StopDuration != nil {
				fleetInput.Stops[s].StopDuration = fleetInput.Defaults.Stops.StopDuration
			} else {
				fleetInput.Stops[s].StopDuration = &stopDuration
			}
		}

		if fleetInput.Stops[s].TargetTime == nil {
			if stopDefaults && fleetInput.Defaults.Stops.TargetTime != nil {
				fleetInput.Stops[s].TargetTime = fleetInput.Defaults.Stops.TargetTime
			}
		}

		if fleetInput.Stops[s].EarlinessPenalty == nil {
			if stopDefaults && fleetInput.Defaults.Stops.EarlinessPenalty != nil {
				fleetInput.Stops[s].EarlinessPenalty = fleetInput.Defaults.Stops.EarlinessPenalty
			}
		}

		if fleetInput.Stops[s].LatenessPenalty == nil {
			if stopDefaults && fleetInput.Defaults.Stops.LatenessPenalty != nil {
				fleetInput.Stops[s].LatenessPenalty = fleetInput.Defaults.Stops.LatenessPenalty
			}
		}

		if fleetInput.Stops[s].CompatibilityAttributes == nil {
			if stopDefaults && fleetInput.Defaults.Stops.CompatibilityAttributes != nil {
				fleetInput.Stops[s].CompatibilityAttributes = fleetInput.Defaults.Stops.CompatibilityAttributes
			} else {
				fleetInput.Stops[s].CompatibilityAttributes = &stopCompatibilities
			}
		}
	}
}
