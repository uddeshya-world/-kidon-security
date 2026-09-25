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

# A missing GitHub release is an HTTP error. Do not substitute a tag.
LATEST=""
if API=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest"); then
    LATEST=$(printf '%s\n' "$API" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
fi
if [ -z "$LATEST" ]; then
    echo "No release binaries are published yet. Build from source: go build -o kidon ./cmd/kidon"
    exit 1
fi

echo "📦 Downloading Kidon $LATEST..."
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/kidon-${OS}-${ARCH}"

downloaded=0
if command -v curl > /dev/null; then
    if curl -fsSL "$DOWNLOAD_URL" -o "$BINARY"; then
        downloaded=1
    fi
elif command -v wget > /dev/null; then
    if wget -O "$BINARY" "$DOWNLOAD_URL"; then
        downloaded=1
    fi
else
    echo "curl or wget is required."
    exit 1
fi

if [ "$downloaded" -ne 1 ] || ! [ -s "$BINARY" ]; then
    rm -f "$BINARY"
    echo "Download failed: $DOWNLOAD_URL"
    exit 1
fi

chmod +x "$BINARY"

echo ""
echo "✅ Kidon installed successfully!"
echo ""
echo "Run: ./$BINARY dashboard"
echo "Or move to PATH: sudo mv $BINARY $INSTALL_DIR/"
