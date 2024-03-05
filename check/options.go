// © 2019-present nextmv.io inc

package check

import "time"

// Options are the options for a check.
type Options struct {
	Duration  time.Duration `json:"duration" usage:"maximum duration of the check" default:"30s"`
	Verbosity string        `json:"verbosity"  usage:"{off, low, medium, high} verbosity of the check" default:"off"`
}
