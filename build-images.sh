#!/bin/bash
# Build script for creating both US and EU Stratux images
# This script builds separate .deb packages and Raspberry Pi images for US and EU regions

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="${SCRIPT_DIR}/build_output"
IMAGE_BUILD_DIR="${SCRIPT_DIR}/image_build"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to modify Makefile for region-specific builds
modify_makefile_for_region() {
    local region=$1
    local makefile="${SCRIPT_DIR}/Makefile"

    log_info "Modifying Makefile for ${region} build..."

    if [ "$region" = "US" ]; then
        # US settings: UAT enabled, OGN disabled
        sed -i "s/echo '{\"UAT_Enabled\": .*}'/echo '{\"UAT_Enabled\": true,\"OGN_Enabled\": false,\"DeveloperMode\": false,\"RegionSelected\": 1}'/" "$makefile"
    elif [ "$region" = "EU" ]; then
        # EU settings: UAT disabled, OGN enabled, Developer mode on
        sed -i "s/echo '{\"UAT_Enabled\": .*}'/echo '{\"UAT_Enabled\": false,\"OGN_Enabled\": true,\"DeveloperMode\": true,\"RegionSelected\": 2}'/" "$makefile"
    fi
}

# Function to restore original Makefile
restore_makefile() {
    log_info "Restoring original Makefile..."
    cd "${SCRIPT_DIR}"
    git checkout Makefile
}

# Function to build a .deb package for a specific region
build_deb_package() {
    local region=$1
    log_info "Building ${region} .deb package..."

    cd "${SCRIPT_DIR}"

    # Modify Makefile for this region
    modify_makefile_for_region "$region"

    # Clean previous builds
    make clean

    # Build the .deb package using Docker
    log_info "Running 'make ddpkg' (this may take a while)..."
    make ddpkg

    # Find the generated .deb file
    DEB_FILE=$(ls -1t stratux-*.deb 2>/dev/null | head -1)

    if [ -z "$DEB_FILE" ]; then
        log_error "Failed to find generated .deb file for ${region}"
        restore_makefile
        return 1
    fi

    # Create output directory
    mkdir -p "${BUILD_DIR}"

    # Rename and move the .deb file
    VERSION=$(echo "$DEB_FILE" | sed 's/stratux-\(.*\)-.*.deb/\1/')
    ARCH=$(echo "$DEB_FILE" | sed 's/stratux-.*-\(.*\).deb/\1/')
    NEW_NAME="stratux-${region}-${VERSION}-${ARCH}.deb"

    cp "$DEB_FILE" "${BUILD_DIR}/${NEW_NAME}"
    log_info "Created: ${BUILD_DIR}/${NEW_NAME}"

    # Restore Makefile
    restore_makefile

    echo "$NEW_NAME"
}

# Function to build a Raspberry Pi image for a specific region
build_pi_image() {
    local region=$1
    local deb_file=$2

    log_info "Building ${region} Raspberry Pi image..."

    cd "${IMAGE_BUILD_DIR}"

    # The build.sh script expects the .deb to be in the parent directory
    # Copy our region-specific .deb there temporarily
    cp "${BUILD_DIR}/${deb_file}" "${SCRIPT_DIR}/"

    # Temporarily rename it to what the build script expects
    TEMP_DEB_NAME=$(echo "$deb_file" | sed "s/stratux-${region}-/stratux-/")
    mv "${SCRIPT_DIR}/${deb_file}" "${SCRIPT_DIR}/${TEMP_DEB_NAME}"

    # Run the image build
    log_info "Running image build script (this will take 30+ minutes)..."
    ./build.sh

    # Find the generated image
    IMG_FILE=$(find pi-gen/deploy -name "*.img" -o -name "*.img.xz" 2>/dev/null | head -1)

    if [ -z "$IMG_FILE" ]; then
        log_error "Failed to find generated image file for ${region}"
        # Clean up temp deb
        rm -f "${SCRIPT_DIR}/${TEMP_DEB_NAME}"
        return 1
    fi

    # Move image to output directory with region-specific name
    IMG_BASENAME=$(basename "$IMG_FILE")
    NEW_IMG_NAME="${IMG_BASENAME%.img*}-${region}.img$(echo "$IMG_BASENAME" | sed 's/.*\.img//')"

    mkdir -p "${BUILD_DIR}"
    mv "$IMG_FILE" "${BUILD_DIR}/${NEW_IMG_NAME}"
    log_info "Created: ${BUILD_DIR}/${NEW_IMG_NAME}"

    # Clean up
    rm -f "${SCRIPT_DIR}/${TEMP_DEB_NAME}"

    cd "${SCRIPT_DIR}"
}

# Main script logic
main() {
    log_info "=== Stratux US/EU Image Build Script ==="
    log_info "This script will build both US and EU versions"
    echo ""

    # Check if we're on the right architecture
    ARCH=$(uname -m)
    log_info "Detected architecture: ${ARCH}"

    if [ "$ARCH" != "x86_64" ] && [ "$ARCH" != "aarch64" ]; then
        log_warn "Building on ${ARCH} - ensure Docker is properly configured"
    fi

    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    # Check if submodules are initialized
    if [ ! -f "${IMAGE_BUILD_DIR}/pi-gen/build-docker.sh" ]; then
        log_error "Submodules not initialized. Run: git submodule update --init --recursive"
        exit 1
    fi

    # Create output directory
    mkdir -p "${BUILD_DIR}"

    # Build US version
    log_info ""
    log_info "========================================="
    log_info "Building US version..."
    log_info "========================================="
    US_DEB=$(build_deb_package "US")

    # Build EU version
    log_info ""
    log_info "========================================="
    log_info "Building EU version..."
    log_info "========================================="
    EU_DEB=$(build_deb_package "EU")

    log_info ""
    log_info "========================================="
    log_info "Debian packages built successfully!"
    log_info "========================================="
    log_info "US: ${BUILD_DIR}/${US_DEB}"
    log_info "EU: ${BUILD_DIR}/${EU_DEB}"
    echo ""

    # Ask if user wants to build full images
    read -p "Do you want to build full Raspberry Pi images? This will take 30+ minutes per image. (y/N): " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Build US image
        log_info ""
        log_info "========================================="
        log_info "Building US Raspberry Pi image..."
        log_info "========================================="
        build_pi_image "US" "$US_DEB"

        # Build EU image
        log_info ""
        log_info "========================================="
        log_info "Building EU Raspberry Pi image..."
        log_info "========================================="
        build_pi_image "EU" "$EU_DEB"

        log_info ""
        log_info "========================================="
        log_info "All builds completed successfully!"
        log_info "========================================="
    else
        log_info "Skipping Raspberry Pi image builds."
    fi

    log_info ""
    log_info "Build artifacts are in: ${BUILD_DIR}"
    ls -lh "${BUILD_DIR}"
}

# Run main function
main "$@"
