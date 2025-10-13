#!/bin/bash
# Native build script for Raspberry Pi
# Use this when building directly on a Raspberry Pi (much faster than cross-compile)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ARCH=$(uname -m)

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're on ARM architecture
if [ "$ARCH" != "aarch64" ] && [ "$ARCH" != "armv7l" ]; then
    log_error "This script is for native builds on Raspberry Pi (ARM architecture)"
    log_info "You're on $ARCH - use ./build-images.sh for cross-compilation instead"
    exit 1
fi

log_info "=== Native Raspberry Pi Build ==="
log_info "Architecture: $ARCH"
echo ""

cd "${SCRIPT_DIR}"

# Function to build a specific region
build_region() {
    local region=$1
    local makefile="${SCRIPT_DIR}/Makefile"

    log_info "Building ${region} configuration..."

    # Backup original Makefile
    cp "$makefile" "${makefile}.bak"

    # Modify for region
    if [ "$region" = "US" ]; then
        sed -i "s/echo '{\"UAT_Enabled\": .*}'/echo '{\"UAT_Enabled\": true,\"OGN_Enabled\": false,\"DeveloperMode\": false,\"RegionSelected\": 1}'/" "$makefile"
    elif [ "$region" = "EU" ]; then
        sed -i "s/echo '{\"UAT_Enabled\": .*}'/echo '{\"UAT_Enabled\": false,\"OGN_Enabled\": true,\"DeveloperMode\": true,\"RegionSelected\": 2}'/" "$makefile"
    fi

    # Clean and build
    log_info "Cleaning previous build..."
    make clean

    log_info "Building .deb package..."
    make dpkg

    # Find and rename the deb
    DEB_FILE=$(ls -1t stratux-*.deb 2>/dev/null | head -1)
    if [ -z "$DEB_FILE" ]; then
        log_error "Failed to create .deb file"
        mv "${makefile}.bak" "$makefile"
        return 1
    fi

    # Rename with region
    VERSION=$(echo "$DEB_FILE" | sed 's/stratux-\(.*\)-.*.deb/\1/')
    ARCH_NAME=$(echo "$DEB_FILE" | sed 's/stratux-.*-\(.*\).deb/\1/')
    NEW_NAME="stratux-${region}-${VERSION}-${ARCH_NAME}.deb"

    mv "$DEB_FILE" "$NEW_NAME"
    log_info "Created: $NEW_NAME"

    # Restore Makefile
    mv "${makefile}.bak" "$makefile"

    echo "$NEW_NAME"
}

# Check what user wants to build
echo "What do you want to build?"
echo "  1) US version only"
echo "  2) EU version only"
echo "  3) Both US and EU"
echo ""
read -p "Enter choice [1-3]: " choice

case $choice in
    1)
        log_info "Building US version..."
        build_region "US"
        ;;
    2)
        log_info "Building EU version..."
        build_region "EU"
        ;;
    3)
        log_info "Building both versions..."
        US_DEB=$(build_region "US")
        echo ""
        EU_DEB=$(build_region "EU")
        echo ""
        log_info "Build complete!"
        log_info "US: $US_DEB"
        log_info "EU: $EU_DEB"
        ;;
    *)
        log_error "Invalid choice"
        exit 1
        ;;
esac

echo ""
log_info "=== Build Complete ==="
log_info "To install: sudo dpkg -i stratux-*.deb"
log_info "To test: sudo systemctl restart stratux"
