#!/bin/bash

# Ensure the script is executed from the root of the project
cd "$(dirname "$0")/.." || exit 1

CONFIG_FILE="workflow-configuration.yml"

# Extract the .build key and iterate over each object.
BUILDS=$(yq e '.build' $CONFIG_FILE)
for i in $(seq 0 $(($(echo "$BUILDS" | yq e 'length' -) - 1)))
do
    # Extract GOOS and GOARCH values
    BUILD_GOOS=$(echo "$BUILDS" | yq e ".[$i].GOOS" -)
    BUILD_GOARCH=$(echo "$BUILDS" | yq e ".[$i].GOARCH" -)
    
    # Step 5: Construct and run the build command
    echo "🐰 Compiling Nextroute binary for OS: $BUILD_GOOS; ARCH: $BUILD_GOARCH"
    if [ "$BUILD_GOOS" = "windows" ]; then
        OUTPUT_SUFFIX=".exe"
    else
        OUTPUT_SUFFIX=""
    fi
    GOOS=$BUILD_GOOS \
        GOARCH=$BUILD_GOARCH \
        go build -o nextroute/bin/nextroute-${BUILD_GOOS}-${BUILD_GOARCH}${OUTPUT_SUFFIX} \
        cmd/main.go
done
