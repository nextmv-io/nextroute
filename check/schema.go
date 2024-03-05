// © 2019-present nextmv.io inc

// Package check contains the schema for the check configuration.
package check

import "strings"

// Verbosity is the verbosity of the check.
type Verbosity int

const (
	// Off does not run the check.
	Off Verbosity = iota
	// Low checks if there is at least one move per plan unit.
	Low
	// Medium checks the number of moves per plan unit and the
	// number of vehicles that have moves. It also reports the number of
	// constraints that are violated for each plan unit if it does not fit
	// on any vehicle.
	Medium
	// High is identical to medium.
	High
)

// ToVerbosity converts a string to a verbosity. The string can be
// anything that starts with "o", "l", "m", "h" or "v" case-insensitive.
// If the string does not start with one of these characters the
// verbosity is off.
func ToVerbosity(s string) Verbosity {
	ls := strings.ToLower(s)
	if strings.HasPrefix(ls, "o") {
		return Off
	}
	if strings.HasPrefix(ls, "l") {
		return Low
	}
	if strings.HasPrefix(ls, "m") {
		return Medium
	}
	if strings.HasPrefix(ls, "h") {
		return High
	}
	return Off
}

// String returns the string representation of the verbosity.
func (v Verbosity) String() string {
	switch v {
	case Off:
		return "off"
	case Low:
		return "low"
	case Medium:
		return "medium"
	case High:
		return "high"
	default:
		return "unknown"
	}
}
