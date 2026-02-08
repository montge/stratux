#!/bin/bash -e
# Script to build Stratux images with pi-gen using pre-built cross-compiled .deb

arch=`uname -m`
echo "arch ${arch}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STRATUX_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$STRATUX_DIR/build_arm64"

# Find the .deb package
DEB_FILE=$(ls -1t "$BUILD_DIR"/*.deb 2>/dev/null | head -1)
if [ -z "$DEB_FILE" ]; then
    echo "ERROR: No .deb package found in $BUILD_DIR"
    echo "Run scripts/build-deb-cross.sh first"
    exit 1
fi
echo "Using pre-built .deb: $DEB_FILE"

# disable full image (stage 5), normal image (stage 4) and desktop image (stage 3)
touch pi-gen/stage3/SKIP pi-gen/stage4/SKIP pi-gen/stage5/SKIP

# and disable building the images based on the SKIPPED stages
touch pi-gen/stage3/SKIP_IMAGES pi-gen/stage4/SKIP_IMAGES pi-gen/stage5/SKIP_IMAGES

# copy the config file into place
cp config pi-gen/

# copy the stage2 10-stratux files into place
rsync -av --delete stage2/10-stratux/ pi-gen/stage2/10-stratux/

# Create a minimal stratux directory with just the .deb package
# (instead of cloning the entire repo and running make ddpkg)
rm -rf pi-gen/stratux
mkdir -p pi-gen/stratux
cp "$DEB_FILE" pi-gen/stratux/

echo "Copied $(basename "$DEB_FILE") to pi-gen/stratux/"

# build via docker
#
# pass PRESERVE_CONTAINER=1 to keep the container in the case of error
# to enable debugging
(cd pi-gen && ./build-docker.sh)
