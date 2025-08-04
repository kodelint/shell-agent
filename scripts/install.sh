#!/bin/bash

set -e

# Shell Agent installation script

BINARY_NAME="shell-agent"
INSTALL_DIR="/usr/local/bin"
REPO_URL="https://github.com/yourusername/shell-agent"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Download URL
DOWNLOAD_URL="${REPO_URL}/releases/latest/download/${BINARY_NAME}-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
    BINARY_NAME="${BINARY_NAME}.exe"
fi

echo "Installing Shell Agent..."
echo "OS: $OS"
echo "Architecture: $ARCH"
echo "Download URL: $DOWNLOAD_URL"

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download binary
echo "Downloading Shell Agent..."
if command -v curl >/dev/null 2>&1; then
    curl -L -o "$BINARY_NAME" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -O "$BINARY_NAME" "$DOWNLOAD_URL"
else
    echo "Error: curl or wget is required to download Shell Agent"
    exit 1
fi

# Make executable
chmod +x "$BINARY_NAME"

# Install binary
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
else
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
fi

# Cleanup
cd - >/dev/null
rm -rf "$TMP_DIR"

echo "Shell Agent installed successfully!"
echo "Run 'shell-agent --help' to get started."
echo ""
echo "To download an AI model, run:"
echo "  shell-agent download"
echo ""
echo "To start interactive mode, run:"
echo "  shell-agent"
