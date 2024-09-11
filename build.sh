#!/bin/bash

set -e

# Get the version from the tag or use "dev" if not set
VERSION=${GITHUB_REF_NAME:-dev}

# Remove 'v' prefix if present
VERSION=${VERSION#v}

# Set binary name
BINARY_NAME="ploy"

# Determine the build number (use the GitHub run number if available, otherwise set to 0)
BUILD_NUMBER=${GITHUB_RUN_NUMBER:-0}

# Set the ldflags
LDFLAGS="-X 'github.com/cloudoploy/ploy-cli/cmd.Version=${VERSION}' -X 'github.com/cloudoploy/ploy-cli/cmd.BuildNumber=${BUILD_NUMBER}'"

# Create or recreate the build folder
BUILD_DIR="build"
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Function to build for a specific OS and architecture
build() {
    os=$1
    arch=$2
    output="${BUILD_DIR}/${BINARY_NAME}-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi

    echo "Building for ${os}/${arch}..."
    GOOS=$os GOARCH=$arch go build -ldflags "${LDFLAGS}" -o "${output}" .

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
#build darwin amd64
#build darwin arm64
#build windows amd64

# Generate checksums
cd "$BUILD_DIR"
if command -v sha256sum > /dev/null; then
    sha256sum ${BINARY_NAME}-*.tar.gz ${BINARY_NAME}-*.zip > checksums.txt
elif command -v shasum > /dev/null; then
    shasum -a 256 ${BINARY_NAME}-*.tar.gz ${BINARY_NAME}-*.zip > checksums.txt
else
    echo "Neither sha256sum nor sass command found. Skipping checksum generation."
fi
cd ..

echo "Build completed successfully. Artifacts are in the '$BUILD_DIR' directory."
