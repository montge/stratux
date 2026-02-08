#!/bin/bash
# Build a deployable package for the new Svelte web interface
# This creates a tarball that can be copied to a Stratux device for testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DATE=$(date +%Y%m%d)
VERSION=$(cd "$PROJECT_DIR/.." && ./scripts/getversion.sh 2>/dev/null || echo "dev")

# Package name with date to avoid conflicts
PACKAGE_NAME="stratux-web-svelte-${VERSION}-${BUILD_DATE}"

echo "=== Building Stratux Svelte Web Interface ==="
echo "Version: $VERSION"
echo "Date: $BUILD_DATE"
echo ""

# Build the Svelte app
cd "$PROJECT_DIR"
echo "Installing dependencies..."
npm install --silent

echo "Building production bundle..."
npm run build

# Create package directory
PACKAGE_DIR="/tmp/${PACKAGE_NAME}"
rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR/www-svelte"

# Copy built files
cp -r dist/* "$PACKAGE_DIR/www-svelte/"

# Create install script
cat > "$PACKAGE_DIR/install.sh" << 'INSTALL_SCRIPT'
#!/bin/bash
# Install Stratux Svelte Web Interface
# This installs alongside the existing interface, not replacing it

set -e

STRATUX_HOME="${STRATUX_HOME:-/opt/stratux}"
INSTALL_DIR="$STRATUX_HOME/www-svelte"

echo "=== Installing Stratux Svelte Web Interface ==="
echo "Install directory: $INSTALL_DIR"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo ./install.sh)"
    exit 1
fi

# Create install directory
mkdir -p "$INSTALL_DIR"

# Copy files
echo "Copying files..."
cp -r www-svelte/* "$INSTALL_DIR/"

# Set permissions
chown -R root:root "$INSTALL_DIR"
chmod -R 755 "$INSTALL_DIR"

echo ""
echo "=== Installation Complete ==="
echo ""
echo "The new Svelte interface is installed at: $INSTALL_DIR"
echo ""
echo "To test it, you have two options:"
echo ""
echo "1. RECOMMENDED: Run both interfaces side-by-side"
echo "   - Original: http://192.168.10.1/ (port 80)"
echo "   - Svelte:   http://192.168.10.1:8080/ (needs nginx config)"
echo ""
echo "2. TEMPORARY: Swap the interfaces"
echo "   sudo mv $STRATUX_HOME/www $STRATUX_HOME/www-legacy"
echo "   sudo mv $STRATUX_HOME/www-svelte $STRATUX_HOME/www"
echo "   sudo systemctl restart stratux"
echo ""
echo "   To revert:"
echo "   sudo mv $STRATUX_HOME/www $STRATUX_HOME/www-svelte"
echo "   sudo mv $STRATUX_HOME/www-legacy $STRATUX_HOME/www"
echo "   sudo systemctl restart stratux"
echo ""
INSTALL_SCRIPT

chmod +x "$PACKAGE_DIR/install.sh"

# Create README
cat > "$PACKAGE_DIR/README.md" << README
# Stratux Svelte Web Interface

**Version**: $VERSION
**Build Date**: $BUILD_DATE

This is a modern replacement for the AngularJS web interface, built with Svelte 5.

## Quick Install

\`\`\`bash
# Copy to your Stratux device
scp ${PACKAGE_NAME}.tar.gz pi@192.168.10.1:~/

# SSH to device and install
ssh pi@192.168.10.1
tar xzf ${PACKAGE_NAME}.tar.gz
cd ${PACKAGE_NAME}
sudo ./install.sh
\`\`\`

## Features

- **Smaller Bundle**: ~24KB gzipped (vs ~200KB+ for AngularJS)
- **Modern Framework**: Svelte 5 with reactive stores
- **Mobile-First**: Responsive design for tablet use in cockpit
- **Real-Time**: WebSocket updates for live status

## Current Status

- [x] Status page (full feature parity)
- [ ] Traffic page
- [ ] GPS/AHRS page
- [ ] Settings page
- [ ] Other pages

## Reverting

If you swapped the interfaces and want to revert:

\`\`\`bash
sudo mv /opt/stratux/www /opt/stratux/www-svelte
sudo mv /opt/stratux/www-legacy /opt/stratux/www
sudo systemctl restart stratux
\`\`\`
README

# Create tarball
cd /tmp
echo "Creating package: ${PACKAGE_NAME}.tar.gz"
tar czf "${PACKAGE_NAME}.tar.gz" "$PACKAGE_NAME"

# Move to project directory
mv "${PACKAGE_NAME}.tar.gz" "$PROJECT_DIR/"

# Cleanup
rm -rf "$PACKAGE_DIR"

echo ""
echo "=== Build Complete ==="
echo ""
echo "Package: $PROJECT_DIR/${PACKAGE_NAME}.tar.gz"
echo "Size: $(du -h "$PROJECT_DIR/${PACKAGE_NAME}.tar.gz" | cut -f1)"
echo ""
echo "To deploy:"
echo "  scp $PROJECT_DIR/${PACKAGE_NAME}.tar.gz pi@192.168.10.1:~/"
echo "  ssh pi@192.168.10.1 'tar xzf ${PACKAGE_NAME}.tar.gz && cd ${PACKAGE_NAME} && sudo ./install.sh'"
echo ""
