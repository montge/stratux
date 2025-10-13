# Build Status

## Latest Build (Third Attempt - With All Fixes)

**Run ID**: 18474465739
**Workflow**: Build US/EU Regions
**Regions**: Both US and EU
**Status**: In Progress
**Started**: 2025-10-13T18:09:28Z

### Fixes Applied

1. **First Issue**: Wrong make target
   - Problem: Used `make ddpkg` (Docker-based) on native ARM runners
   - Fix: Changed to `make dpkg` for direct native builds

2. **Second Issue**: Missing build dependencies
   - Problem: RTL-SDR libraries not installed on runner
   - Fix: Added step to install gcc, make, golang-go, ncurses-dev, libusb-1.0-0-dev, librtlsdr

3. **Third Issue**: Empty version string
   - Problem: `git describe --tags` fails when repository has no version tags
   - Fix: Modified `scripts/getversion.sh` to fall back to `dev-{hash}-{date}` format

### Commits
- fa5b699: "Fix build workflow: use 'make dpkg' instead of 'make ddpkg' on ARM runners"
- 4f5db5d: "Fix build workflow: install required dependencies (librtlsdr, golang, etc.)"
- 53e300a: "Fix getversion.sh: handle repositories without tags"

## Monitor Progress

```bash
# Watch live progress
gh run watch 18474465739

# Check status
gh run list --workflow=build-regions.yml --limit 1

# View in browser
open https://github.com/montge/stratux/actions/runs/18474465739
```

## Expected Timeline

- **Setup + Dependencies**: 1-2 minutes
- **Build time**: 13-18 minutes (both regions in parallel)
- **Total**: ~15-20 minutes

## What's Being Built

1. **US Version**
   - UAT Enabled: true
   - OGN Enabled: false
   - Developer Mode: false
   - RegionSelected: 1
   - Package: `stratux-US-dev-53e300a-20251013-arm64.deb`

2. **EU Version**
   - UAT Enabled: false
   - OGN Enabled: true
   - Developer Mode: true
   - RegionSelected: 2
   - Package: `stratux-EU-dev-53e300a-20251013-arm64.deb`

## When Build Completes

### Download Artifacts
```bash
# Download both packages
gh run download 18474465739

# This creates:
# - stratux-US-debian-package/stratux-US-*.deb
# - stratux-EU-debian-package/stratux-EU-*.deb
```

### Deploy to Crewdog (aarch64 Raspberry Pi 3B)
```bash
# Copy US version to crewdog
scp stratux-US-debian-package/stratux-US-*.deb pi@10.0.1.53:~/

# Install on crewdog
ssh pi@10.0.1.53
sudo dpkg -i stratux-US-*.deb
sudo systemctl restart stratux

# Verify
sudo systemctl status stratux
```

## Current Crewdog Info

- **Device**: Raspberry Pi 3 Model B Rev 1.2
- **Architecture**: aarch64 (ARM64) ✅
- **OS**: Debian GNU/Linux 12 (bookworm) ✅
- **IP**: 10.0.1.53

Perfect match for our builds!

## Previous Build Attempts

### Run 18474358447 (Second Attempt)
- Status: Failed
- Issue: Empty version string from getversion.sh
- Error: `'Version' field value '': version string is empty`

### Run 18474307860 (First Attempt)
- Status: Failed
- Issue: Missing RTL-SDR development libraries
- Error: `fatal error: rtl-sdr.h: No such file or directory`

### Run 18469782672 (Initial Attempt)
- Status: Failed
- Issue: Used `make ddpkg` which tries Docker-in-Docker
- Error: `Process completed with exit code 2`
