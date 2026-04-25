#!/usr/bin/env bash

# shortk install script
# Usage: curl -sSL https://raw.githubusercontent.com/user/shortk/main/install.sh | bash

set -e

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Hypothetical download URL
# URL="https://github.com/user/shortk/releases/latest/download/shortk-${OS}-${ARCH}"

# For now, since we are in the repo, let's assume we build it
if command -v go >/dev/null 2>&1; then
    echo "Building shortk from source..."
    go build -o shortk main.go
else
    echo "Go not found. In a real scenario, I would download the pre-built binary here."
    # curl -L $URL -o shortk
fi

# Install to ~/bin if it exists and is in PATH, otherwise /usr/local/bin
INSTALL_DIR="$HOME/.local/bin"
if [[ ! ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
    INSTALL_DIR="/usr/local/bin"
fi

echo "Installing to $INSTALL_DIR..."
if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR"
fi

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
