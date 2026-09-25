#!/bin/sh
# Kidon Installer Script
# Usage: curl -sfL https://raw.githubusercontent.com/uddeshya-world/-kidon-security/main/install.sh | sh

set -e

REPO="uddeshya-world/-kidon-security"
BINARY="kidon"
INSTALL_DIR="/usr/local/bin"

# Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
esac

echo "⚔️  KIDON INSTALLER"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "OS: $OS | Arch: $ARCH"
echo ""

# Download latest release
LATEST=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
    LATEST="v0.3.0"
fi

echo "📦 Downloading Kidon $LATEST..."
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/kidon-${OS}-${ARCH}"

if command -v curl > /dev/null; then
    curl -sL "$DOWNLOAD_URL" -o "$BINARY" || {
        echo "Note: Pre-built binary not found. Building from source..."
        echo "Run: go build -o kidon cmd/kidon/main.go"
        exit 1
    }
elif command -v wget > /dev/null; then
    wget -q "$DOWNLOAD_URL" -O "$BINARY" || exit 1
fi

chmod +x "$BINARY"

echo ""
echo "✅ Kidon installed successfully!"
echo ""
echo "Run: ./$BINARY dashboard"
echo "Or move to PATH: sudo mv $BINARY $INSTALL_DIR/"
