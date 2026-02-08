#!/bin/bash
# Build Stratux .deb package using cross-compiled ARM64 binaries

set -e

STRATUX_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$STRATUX_DIR/build_arm64"
DEBPKG_BASE="/tmp/dpkg-stratux/stratux"
DEBPKG_HOME="$DEBPKG_BASE/opt/stratux"
VERSIONSTR=$($STRATUX_DIR/scripts/getversion.sh)
ARCH="arm64"

echo "=== Building Stratux .deb package v$VERSIONSTR for $ARCH ==="

# Verify binaries exist
for bin in stratuxrun dump1090 rtl_ais fancontrol libdump978.so; do
    if [ ! -f "$BUILD_DIR/$bin" ]; then
        echo "ERROR: Missing $BUILD_DIR/$bin - run cross-compile-arm64.sh first"
        exit 1
    fi
done

echo "=== Building web interface ==="
make -C "$STRATUX_DIR/web" STRATUX_HOME="$DEBPKG_HOME"

echo "=== Fetching OGN database ==="
cd "$STRATUX_DIR/ogn" && ./fetch_ddb.sh || true

echo "=== Preparing dpkg directory structure ==="
rm -rf "$DEBPKG_BASE"
mkdir -p "$DEBPKG_BASE/DEBIAN"
mkdir -p "$DEBPKG_HOME/bin"
mkdir -p "$DEBPKG_HOME/www"
mkdir -p "$DEBPKG_HOME/ogn"
mkdir -p "$DEBPKG_HOME/softrf"
mkdir -p "$DEBPKG_HOME/cfg"
mkdir -p "$DEBPKG_HOME/lib"
mkdir -p "$DEBPKG_HOME/mapdata"
chmod a+rwx "$DEBPKG_HOME/mapdata"

echo "=== Copying binaries ==="
cp -f "$BUILD_DIR/stratuxrun" "$DEBPKG_HOME/bin/"
cp -f "$BUILD_DIR/fancontrol" "$DEBPKG_HOME/bin/"
cp -f "$BUILD_DIR/dump1090" "$DEBPKG_HOME/bin/"
cp -f "$BUILD_DIR/rtl_ais" "$DEBPKG_HOME/bin/"
cp -f "$STRATUX_DIR/ogn/ogn-rx-eu_aarch64" "$DEBPKG_HOME/bin/ogn-rx-eu"
chmod +x "$DEBPKG_HOME/bin/"*

echo "=== Copying libraries ==="
cp -f "$BUILD_DIR/libdump978.so" "$DEBPKG_HOME/lib/"

echo "=== Copying web interface ==="
cp -ru "$STRATUX_DIR/web/"* "$DEBPKG_HOME/www/" 2>/dev/null || true
rm -rf "$DEBPKG_HOME/www/Makefile" "$DEBPKG_HOME/www/plates/Makefile"

echo "=== Copying map data ==="
cp -ru "$STRATUX_DIR/mapdata/"* "$DEBPKG_HOME/mapdata/"

echo "=== Copying OGN files ==="
cp -f "$STRATUX_DIR/ogn/ddb.json" "$DEBPKG_HOME/ogn/" 2>/dev/null || true
cp -f "$STRATUX_DIR/ogn/"*ogn-tracker-bin-*.zip "$DEBPKG_HOME/ogn/" 2>/dev/null || true
cp -f "$STRATUX_DIR/ogn/install-ogntracker-firmware-pi.sh" "$DEBPKG_HOME/ogn/" 2>/dev/null || true
cp -f "$STRATUX_DIR/ogn/fetch_ddb.sh" "$DEBPKG_HOME/ogn/" 2>/dev/null || true
cp -f "$STRATUX_DIR/softrf/"*.zip "$DEBPKG_HOME/softrf/" 2>/dev/null || true
cp -f "$STRATUX_DIR/softrf/"*.sh "$DEBPKG_HOME/softrf/" 2>/dev/null || true

echo "=== Copying scripts ==="
cp "$STRATUX_DIR/debian/stratux-pre-start.sh" "$DEBPKG_HOME/bin/"
chmod 744 "$DEBPKG_HOME/bin/stratux-pre-start.sh"
cp -f "$STRATUX_DIR/debian/stratux-wifi.sh" "$DEBPKG_HOME/bin/"
cp -f "$STRATUX_DIR/debian/sdr-tool.sh" "$DEBPKG_HOME/bin/"
chmod 755 "$DEBPKG_HOME/bin/"*

echo "=== Copying config templates ==="
cp -f "$STRATUX_DIR/debian/stratux-dnsmasq.conf.template" "$DEBPKG_HOME/cfg/"
cp -f "$STRATUX_DIR/debian/interfaces.template" "$DEBPKG_HOME/cfg/"
cp -f "$STRATUX_DIR/debian/wpa_supplicant.conf.template" "$DEBPKG_HOME/cfg/"
cp -f "$STRATUX_DIR/debian/wpa_supplicant_ap.conf.template" "$DEBPKG_HOME/cfg/"

echo "=== Creating DEBIAN control files ==="
cp -f "$STRATUX_DIR/debian/control.dpkg" "$DEBPKG_BASE/DEBIAN/control"
cp -f "$STRATUX_DIR/debian/conffiles.dpkg" "$DEBPKG_BASE/DEBIAN/conffiles"
cp -f "$STRATUX_DIR/debian/preinst.dpkg" "$DEBPKG_BASE/DEBIAN/preinst"
cp -f "$STRATUX_DIR/debian/postinst.dpkg" "$DEBPKG_BASE/DEBIAN/postinst"
cp -f "$STRATUX_DIR/debian/prerm.dpkg" "$DEBPKG_BASE/DEBIAN/prerm"

echo "=== Creating udev rules ==="
mkdir -p "$DEBPKG_BASE/etc/udev/rules.d/"
cp -f "$STRATUX_DIR/debian/10-stratux.rules" "$DEBPKG_BASE/etc/udev/rules.d/"
cp -f "$STRATUX_DIR/debian/99-uavionix.rules" "$DEBPKG_BASE/etc/udev/rules.d/"
cp -f "$STRATUX_DIR/debian/99-pong.rules" "$DEBPKG_BASE/etc/udev/rules.d/"

echo "=== Creating systemd services ==="
mkdir -p "$DEBPKG_BASE/lib/systemd/system/"
cp "$STRATUX_DIR/debian/stratux.service" "$DEBPKG_BASE/lib/systemd/system/"
chmod 644 "$DEBPKG_BASE/lib/systemd/system/stratux.service"
cp "$STRATUX_DIR/debian/stratux_fancontrol.service" "$DEBPKG_BASE/lib/systemd/system/"
chmod 644 "$DEBPKG_BASE/lib/systemd/system/stratux_fancontrol.service"

echo "=== Setting version and architecture ==="
sed -i "s/VERSION/$VERSIONSTR/g" "$DEBPKG_BASE/DEBIAN/control"
sed -i "s/ARCH/$ARCH/g" "$DEBPKG_BASE/DEBIAN/control"

echo "=== Setting permissions ==="
chmod 755 "$DEBPKG_BASE/DEBIAN/control"
chmod 755 "$DEBPKG_BASE/DEBIAN/preinst"
chmod 755 "$DEBPKG_BASE/DEBIAN/postinst"
chmod 755 "$DEBPKG_BASE/DEBIAN/prerm"

echo "=== Creating default US config ==="
echo '{"UAT_Enabled": true,"OGN_Enabled": false,"DeveloperMode": false,"RegionSelected": 1}' > "$DEBPKG_HOME/cfg/stratux.conf.default"

echo "=== Building .deb package ==="
dpkg-deb -b "$DEBPKG_BASE"

echo "=== Moving .deb to build directory ==="
mv -f "$DEBPKG_BASE/../stratux.deb" "$BUILD_DIR/stratux-$VERSIONSTR-$ARCH.deb"

echo "=== Done! ==="
ls -lh "$BUILD_DIR/stratux-$VERSIONSTR-$ARCH.deb"
