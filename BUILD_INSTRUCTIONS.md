# Stratux Build Instructions

## About the Version

**This repository is building from the current master branch**, which appears to be a **development version** based on the commit history. There are **no version tags** in this repository fork yet.

Looking at the git history:
- Latest commit: `3339fd3` - "Correctly process multiple NMEA messages on 1 line"
- Recent merge: Region selection UI changes (commit `80f6baf`)
- This is a fork from the main Stratux repository at `git@github.com:stratux/stratux.git`

The official Stratux project has moved from separate US/EU images to a **unified image** where users select their region on first boot. However, the build scripts I've created for you will generate **separate US and EU images** with pre-configured settings.

## Build Methods: Cross-Compile vs Native

### Cross-Compilation (x86_64 → ARM64)
**When**: Building on your development machine (laptop/desktop)
**Speed**: Slower (30-60 min for first build) due to ARM emulation
**Pros**: Build on your main machine, no need for Raspberry Pi hardware
**Cons**: Requires Docker + QEMU, slower than native

### Native Build (ARM64 → ARM64)
**When**: Building directly on a Raspberry Pi
**Speed**: Fast (5-15 min) - native compilation
**Pros**: Much faster, simpler setup, no emulation overhead
**Cons**: Requires access to a Raspberry Pi for building

## Prerequisites

### For Cross-Compilation (x86_64 systems)
- **Docker** installed and running
- **Git** with submodules support
- **QEMU** for ARM64 emulation
- At least **20GB** free disk space for builds
- At least **4GB** RAM (8GB recommended)

### For Native Build (Raspberry Pi)
- **Raspberry Pi 4 or 5** (recommended for speed)
- **Raspberry Pi OS** (Debian Bookworm or later)
- **Git** with submodules support
- **Build tools**: gcc, make, golang
- At least **10GB** free disk space
- At least **2GB** RAM (4GB recommended)

### One-Time Setup

#### 1. Install QEMU for ARM64 Emulation (x86_64 systems only)
```bash
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
```

This enables your x86_64 system to build ARM64 Docker containers for Raspberry Pi.

#### 2. Initialize Submodules
This has already been done, but for future reference:
```bash
git submodule update --init --recursive
```

## Build Options

### ⚡ FASTEST: Native Build on Raspberry Pi

If you have access to a Raspberry Pi, this is **10x faster** than cross-compilation:

```bash
./build-native.sh
```

**Time**: 5-15 minutes per build
**Output**: Region-specific `.deb` files
**Note**: Automatically detects you're on ARM and uses native compilation (no Docker/QEMU needed)

The script will ask which region(s) to build:
1. US only
2. EU only
3. Both US and EU

---

### 🔄 Cross-Compilation Options (x86_64 Development Machine)

### Option 1: Quick Test Build (Recommended First)
Build a single .deb package to verify your setup works:

```bash
./test-build.sh
```

**Time**: 5-10 minutes
**Output**: Single `.deb` file in current directory
**Use**: Verify build system works before committing to full image builds

### Option 2: Build Both US and EU Images (Full Build)
Build complete Raspberry Pi images for both regions:

```bash
./build-images.sh
```

**Time**:
- .deb packages: ~10-15 minutes each (total 20-30 min)
- Full images: ~30-45 minutes each (total 60-90 min)

**Output**: Located in `build_output/` directory
- `stratux-US-[VERSION]-arm64.deb`
- `stratux-EU-[VERSION]-arm64.deb`
- `stratux-[DATE]-US.img.xz` (if you chose to build images)
- `stratux-[DATE]-EU.img.xz` (if you chose to build images)

The script will:
1. Build US .deb package (UAT enabled, OGN disabled)
2. Build EU .deb package (OGN enabled, UAT disabled, Developer mode on)
3. Ask if you want to build full Raspberry Pi images (optional)

### Option 3: Manual Build Process

#### Build .deb package only:
```bash
make clean
make ddpkg
```

#### Build full Raspberry Pi image:
```bash
cd image_build
./build.sh
```

## Region Differences

### US Configuration
- **UAT Enabled**: Yes (978 MHz for US weather/traffic)
- **OGN Enabled**: No
- **Developer Mode**: No
- **RegionSelected**: 1

### EU Configuration
- **UAT Enabled**: No
- **OGN Enabled**: Yes (868 MHz for glider tracking)
- **Developer Mode**: Yes (additional logging/features)
- **RegionSelected**: 2

## What Gets Built

### Debian Package (.deb)
Contains:
- All Stratux executables (`stratuxrun`, `fancontrol`, `dump1090`, `rtl_ais`, `ogn-rx-eu`)
- Libraries (`libdump978.so`)
- Scripts (`stratux-pre-start.sh`, `stratux-wifi.sh`, `sdr-tool.sh`)
- Config templates
- Systemd service files
- udev rules
- Region-specific default configuration

### Raspberry Pi Image (.img.xz)
Complete bootable image containing:
- Debian 12 (Bookworm) base system
- Stratux .deb package (pre-installed)
- Custom bluez 5.79 (for Bluetooth LE)
- Custom librtlsdr 2.0.2
- Network configuration (WiFi AP mode)
- Optimized for SD card longevity

## Troubleshooting

### "exec format error"
You need to setup QEMU for ARM emulation:
```bash
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
```

### "all predefined address pools have been fully subnetted"
Docker network issue. Run:
```bash
docker system prune -a -f --volumes
# Or restart Docker daemon
```

### "fatal: No names found, cannot describe anything"
This is normal - the repository has no version tags. The version will be generated from the git commit hash.

### Build fails with "out of space"
Ensure you have at least 20GB free. Clean up Docker:
```bash
docker system prune -a -f --volumes
```

### Very slow build on x86_64
This is normal - ARM emulation via QEMU is 10-20x slower than native builds.

**Solutions** (in order of speed):
1. **Use `./build-native.sh` on a Raspberry Pi** (fastest - 5-15 min)
2. Use a cloud ARM instance (AWS Graviton, Oracle ARM free tier, etc.)
3. Be patient with cross-compilation (first build ~60 min, subsequent ~30 min)

## Testing Your Build

### Test .deb Package
```bash
# On a Raspberry Pi running Debian Bookworm:
sudo dpkg -i stratux-[VERSION]-arm64.deb
sudo systemctl start stratux
```

### Test Full Image
1. Flash the `.img.xz` file to an SD card using:
   - Raspberry Pi Imager
   - balenaEtcher
   - `dd` command

2. Boot Raspberry Pi with the SD card
3. Connect to WiFi network "Stratux" (password: none by default)
4. Navigate to http://192.168.10.1

## Build Artifacts Location

All build outputs are saved to `build_output/` directory:
```
build_output/
├── stratux-US-[VERSION]-arm64.deb
├── stratux-EU-[VERSION]-arm64.deb
├── stratux-[DATE]-US.img.xz (if built)
└── stratux-[DATE]-EU.img.xz (if built)
```

## Recommended Workflow

### If You Have a Raspberry Pi Available:
1. Copy this repository to your Raspberry Pi
2. Run `./build-native.sh` (fastest method - 5-15 min)
3. Install directly: `sudo dpkg -i stratux-*.deb`

### If Building on x86_64 (Cross-Compile):
1. Setup QEMU: `docker run --rm --privileged multiarch/qemu-user-static --reset -p yes`
2. Test build: `./test-build.sh` (verify environment)
3. Full build: `./build-images.sh` (builds both regions)
4. Copy `.deb` to Raspberry Pi and install

### For Production Deployment:
1. Use `.deb` packages for OTA updates on existing systems
2. Use `.img` files for fresh SD card installations
3. Native builds on Pi are recommended for development/testing iterations

## Notes

- First build will be slow (30-60 min) due to Docker image building
- Subsequent builds are faster due to Docker layer caching
- The build process creates a clean environment each time
- No root/sudo required for the build process itself
- The scripts automatically restore the Makefile after region-specific modifications
