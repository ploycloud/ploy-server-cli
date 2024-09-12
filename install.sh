#!/bin/bash

# Ploy CLI Installer
#
# This script installs the latest version of the Ploy CLI on your system.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/cloudoploy/ploy-cli/main/install.sh | bash

set -e

# Determine OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then
    ARCH="amd64"
elif [[ "$ARCH" == "arch64" || "$ARCH" == "arm64" ]]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Get latest release from GitHub API
REPO="cloudoploy/ploy-cli"
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest)
TAG_NAME=$(echo "$LATEST_RELEASE" | grep -oP '"tag_name": "\K(.*)(?=")')
echo "Latest release: $TAG_NAME"
DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep -oP '"browser_download_url": "\K(.*'${OS}'-'${ARCH}'.tar.gz)(?=")')

if [ -z "$DOWNLOAD_URL" ]; then
    echo "No suitable binary found for OS: $OS and architecture: $ARCH"
    exit 1
fi

# Download the latest release
TEMP_DIR=$(mktemp -d)
DOWNLOAD_FILE="$TEMP_DIR/ploy-$OS-$ARCH.tar.gz"

echo "Downloading $DOWNLOAD_URL..."
curl -L -o "$DOWNLOAD_FILE" "$DOWNLOAD_URL"

# Extract and install
echo "Extracting $DOWNLOAD_FILE..."
tar -xzf "$DOWNLOAD_FILE" -C "$TEMP_DIR"

echo "Installing to /usr/local/bin/fly..."
sudo mv "$TEMP_DIR/fly-$OS-$ARCH" /usr/local/bin/ploy
sudo chmod +x /usr/local/bin/ploy

# Clean up
rm -rf "$TEMP_DIR"

# Verify installation
if command -v ploy &> /dev/null; then
    echo "Ploy CLI has been successfully installed!"
    ploy --version
else
    echo "Installation failed. Please check your permissions and try again."
    exit 1
fi

echo "You can now use the 'ploy' command to manage your cloud deployments."
