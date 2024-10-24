#!/bin/bash

set -e

# Set binary name
BINARY_NAME="ploy"

# Determine the build number (use the GitHub run number if available, otherwise set to 0)
BUILD_NUMBER=${GITHUB_RUN_NUMBER:-0}

# Set the ldflags
LDFLAGS="-X 'github.com/ploycloud/ploy-server-cli/cmd.BuildNumber=${BUILD_NUMBER}'"

# Create or recreate the build folder
BUILD_DIR="build"
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Function to build for a specific OS and architecture
build() {
    local os=$1
    local arch=$2
    local output="${BUILD_DIR}/${BINARY_NAME}-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi

    echo "Building for ${os}/${arch}..."
    if ! GOOS=$os GOARCH=$arch go build -ldflags "${LDFLAGS}" -o "${output}" .; then
        echo "Failed to build for ${os}/${arch}"
        exit 1
    fi

    if [ "$os" = "windows" ]; then
        zip "${output}.zip" "${output}"
        rm "${output}"
    else
        tar -czf "${output}.tar.gz" "${output}"
        rm "${output}"
    fi
}

# Build for various platforms
build linux amd64
build linux arm64
build darwin amd64
build darwin arm64
#build windows amd64
#build windows arm64

# Generate checksums
cd "$BUILD_DIR"
if command -v sha256sum > /dev/null; then
    sha256sum ${BINARY_NAME}-*.tar.gz > checksums.txt
elif command -v shasum > /dev/null; then
    shasum -a 256 ${BINARY_NAME}-*.tar.gz > checksums.txt
else
    echo "Neither sha256sum nor shasum command found. Skipping checksum generation."
fi
cd ..

echo "Build completed successfully. Artifacts are in the '$BUILD_DIR' directory."
