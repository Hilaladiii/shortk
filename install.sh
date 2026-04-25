#!/usr/bin/env bash

# shortk install script
# Usage: curl -fsSL https://raw.githubusercontent.com/Hilaladiii/shortk/main/install.sh | bash

set -e

# --- CONFIGURATION ---
REPO="Hilaladiii/shortk"
# ---------------------

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $OS in
    darwin) OS="darwin" ;;
    linux)  OS="linux" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

BINARY_NAME="shortk-${OS}-${ARCH}"

echo "Checking for latest release of shortk..."

# Try to get the latest release version from GitHub API
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -n "$LATEST_RELEASE" ]; then
    echo "Found version $LATEST_RELEASE. Downloading $BINARY_NAME..."
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_RELEASE}/${BINARY_NAME}"
    
    if curl -L --fail "$DOWNLOAD_URL" -o shortk; then
        echo "Successfully downloaded binary."
    else
        echo "Failed to download pre-built binary. Falling back to source build..."
        LATEST_RELEASE="" # Trigger fallback
    fi
fi

# Fallback: Build from source if download failed or no release found
if [ -z "$LATEST_RELEASE" ]; then
    if command -v go >/dev/null 2>&1; then
        if [ ! -f "main.go" ]; then
            echo "Cloning source code for build..."
            TMP_DIR=$(mktemp -d)
            git clone --depth 1 "https://github.com/${REPO}.git" "$TMP_DIR"
            cd "$TMP_DIR"
        fi
        echo "Building shortk from source..."
        go build -o shortk main.go
    else
        echo "Error: Could not download pre-built binary and Go is not installed."
        echo "Please ensure you have set the correct REPO in this script or install Go."
        exit 1
    fi
fi

# Determine install directory
INSTALL_DIR="$HOME/.local/bin"
if [[ ! ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
    INSTALL_DIR="/usr/local/bin"
fi

echo "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"

if [ -w "$INSTALL_DIR" ]; then
    mv shortk "$INSTALL_DIR/shortk"
    chmod +x "$INSTALL_DIR/shortk"
else
    sudo mv shortk "$INSTALL_DIR/shortk"
    sudo chmod +x "$INSTALL_DIR/shortk"
fi

# Initialize
"$INSTALL_DIR/shortk" init

echo "shortk installed successfully!"
