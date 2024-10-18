#!/bin/bash

# This script builds the binary to be shipped with the Python wheel and
# puts it in the expected location. We build the binary in static mode
# to avoid any runtime dependencies.
# For now, we are building on the target platform, thus, we don't need to
# cross-compile.

set -euo pipefail

# Change to project root
cd "$(dirname "$0")/.."

# Build the binary
mkdir -p ./src/nextroute/bin
CGO_ENABLED=0 go build -o ./src/nextroute/bin/nextroute.exe ./cmd
