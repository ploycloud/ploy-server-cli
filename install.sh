#!/bin/bash

# Ploy CLI Installer
#
# This script installs the latest version of the Ploy CLI on your system.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/cloudoploy/ploy-cli/main/install.sh | bash

set -e

# Fetch the latest release version
echo "Fetching the latest Ploy CLI version..."
PLOY_VERSION=$(curl -s https://api.github.com/repos/cloudoploy/ploy-cli/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/v//')

if [ -z "$PLOY_VERSION" ]; then
    echo "Failed to fetch the latest version. Please check your internet connection and try again."
    exit 1
fi

echo "Latest Ploy CLI version: $PLOY_VERSION"

# Determine system architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Determine operating system
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case $OS in
    linux|darwin)
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Download URL
DOWNLOAD_URL="https://github.com/cloudoploy/ploy-cli/releases/download/v${PLOY_VERSION}/ploy-${OS}-${ARCH}"

# Installation directory
INSTALL_DIR="/usr/local/bin"

# Temporary directory for download
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download Ploy CLI
echo "Downloading Ploy CLI..."
curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/ploy"

# Make the binary executable
chmod +x "$TMP_DIR/ploy"

# Move the binary to the installation directory
echo "Installing Ploy CLI to $INSTALL_DIR..."
sudo mv "$TMP_DIR/ploy" "$INSTALL_DIR/ploy"

# Verify installation
if command -v ploy &> /dev/null; then
    echo "Ploy CLI has been successfully installed!"
    ploy -version
else
    echo "Installation failed. Please check your permissions and try again."
    exit 1
fi

echo "You can now use the 'ploy' command to manage your cloud deployments."
