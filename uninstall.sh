#!/usr/bin/env bash

# shortk uninstall script

set -e

echo "Uninstalling shortk..."

# 1. Remove the binary
INSTALL_DIR="$HOME/.local/bin"
if [ -f "$INSTALL_DIR/shortk" ]; then
    rm "$INSTALL_DIR/shortk"
    echo "Removed binary from $INSTALL_DIR/shortk"
fi

if [ -f "/usr/local/bin/shortk" ]; then
    sudo rm "/usr/local/bin/shortk"
    echo "Removed binary from /usr/local/bin/shortk"
fi

# 2. Remove configuration and wrappers
CONFIG_DIR="$HOME/.config/shortk"
if [ -d "$CONFIG_DIR" ]; then
    rm -rf "$CONFIG_DIR"
    echo "Removed configuration directory: $CONFIG_DIR"
fi

# 3. Clean up shell profiles
START_MARKER="# <<< shortk initialize <<<"
END_MARKER="# >>> shortk initialize >>>"

clean_profile() {
    local profile="$1"
    if [ -f "$profile" ] && grep -q "$START_MARKER" "$profile"; then
        # Use sed to remove the block between markers (including the markers)
        # We use a temporary file to ensure safe replacement
        sed "/$START_MARKER/,/$END_MARKER/d" "$profile" > "${profile}.tmp"
        mv "${profile}.tmp" "$profile"
        echo "Cleaned up $profile"
    fi
}

clean_profile "$HOME/.zshrc"
clean_profile "$HOME/.bashrc"
clean_profile "$HOME/.bash_profile"

echo ""
echo "shortk has been successfully uninstalled and all global configurations removed."
echo "Note: Local .shortk files in your project directories are kept intact."
echo "Please restart your terminal to fully apply the changes."
