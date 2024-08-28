package nextroute

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var version string

// Version returns the version of the nextroute module.
func Version() string {
	return strings.TrimSpace(version)
}
