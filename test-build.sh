#!/bin/bash
# Quick test build script - just builds the .deb packages without full images
# This is much faster for testing (5-10 minutes vs 30+ minutes per image)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Testing .deb package build ==="
echo "This will build a single .deb package to verify the build system works"
echo ""

cd "${SCRIPT_DIR}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "ERROR: Docker is not installed"
    exit 1
fi

echo "[1/3] Cleaning previous builds..."
make clean || true

echo ""
echo "[2/3] Building .deb package using Docker..."
echo "This will take 5-10 minutes..."
echo ""

make ddpkg

echo ""
echo "[3/3] Checking result..."
DEB_FILE=$(ls -1t stratux-*.deb 2>/dev/null | head -1)

if [ -z "$DEB_FILE" ]; then
    echo "ERROR: No .deb file was created"
    exit 1
fi

echo ""
echo "SUCCESS! Build completed."
echo "Generated file: $DEB_FILE"
echo ""
ls -lh "$DEB_FILE"
echo ""
echo "You can now run ./build-images.sh to build both US and EU versions"
