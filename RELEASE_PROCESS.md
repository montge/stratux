# Stratux Release Process

This document explains how to create releases for Stratux, including nightly builds and official releases.

## Overview

Stratux uses GitHub Actions for automated builds with three workflows:

1. **Nightly Builds** (`.github/workflows/nightly.yml`) - Automated .deb packages every night
2. **Region Builds** (`.github/workflows/build-regions.yml`) - On-demand .deb package builds
3. **Full Release Images** (`.github/workflows/release-images.yml`) - Complete SD card images for official releases

## 1. Nightly Builds (Automated)

### What It Does
- Runs automatically at 2 AM UTC (6 PM PST / 7 PM PDT) every day
- Only builds if there are commits in the last 24 hours
- Creates both US and EU .deb packages
- Publishes as a prerelease on GitHub

### Output
- `stratux-US-0.0.YYYYMMDD-<hash>-arm64.deb`
- `stratux-EU-0.0.YYYYMMDD-<hash>-arm64.deb`

### Manual Trigger
```bash
# Trigger manually (useful for testing)
gh workflow run nightly.yml
```

### Use Cases
- Daily testing builds
- Latest features for beta testers
- Continuous integration verification

---

## 2. On-Demand .deb Builds

### What It Does
- Manually triggered build of US and EU packages
- Faster than full image builds (~2 minutes vs 45-60 minutes)
- For testing changes quickly

### Trigger
```bash
# Build both regions
gh workflow run build-regions.yml -f region=both

# Or just one region
gh workflow run build-regions.yml -f region=US
gh workflow run build-regions.yml -f region=EU
```

### Output
- `stratux-US-0.0.YYYYMMDD-<hash>-arm64.deb`
- `stratux-EU-0.0.YYYYMMDD-<hash>-arm64.deb`

### Download
```bash
# List recent runs
gh run list --workflow=build-regions.yml --limit 5

# Download artifacts
gh run download <run-id>
```

---

## 3. Official Releases (Full SD Card Images)

### What It Does
- Triggered by pushing a version tag (e.g., `v1.6.0`)
- Builds complete bootable SD card images for both US and EU
- Includes both .img files and .deb packages
- Creates a draft release on GitHub
- Takes 45-60 minutes (images are large)

### Prerequisites

Before creating a release:

1. **Update version documentation**
   - Update `README.md` with new version info
   - Update `CHANGELOG.md` with release notes

2. **Test thoroughly**
   - Install nightly build on hardware
   - Verify all features work
   - Test both US and EU configurations if applicable

3. **Merge all changes to master**

### Create a Release

#### Step 1: Create and Push Tag

```bash
# Make sure you're on master with latest changes
git checkout master
git pull

# Create version tag (use semantic versioning)
git tag -a v1.6.0 -m "Release version 1.6.0"

# Push the tag (this triggers the workflow)
git push origin v1.6.0
```

#### Step 2: Monitor the Build

```bash
# Watch the build progress
gh run list --workflow=release-images.yml

# Watch specific run
gh run watch <run-id>
```

The build will take 45-60 minutes and produce:
- `stratux-lite-v1.6.0-US.img` (and .zip)
- `stratux-lite-v1.6.0-EU.img` (and .zip)
- `stratux-US-1.6.0-arm64.deb`
- `stratux-EU-1.6.0-arm64.deb`

#### Step 3: Finalize the Release

1. Go to https://github.com/yourusername/stratux/releases
2. Find the draft release
3. Review the release notes
4. Edit if needed (add detailed changelog, known issues, etc.)
5. Click "Publish release"

### Manual Release Build (Alternative)

If you need to trigger the release build without a tag:

```bash
gh workflow run release-images.yml -f release_name=v1.6.0
```

---

## Version Numbering

Stratux uses semantic versioning with the format: `MAJOR.MINOR.PATCH`

### Examples:
- `v1.6.0` - Major/minor release
- `v1.6.1` - Patch release (bug fixes)
- `v2.0.0` - Major version (breaking changes)

### For Development Builds:
- Nightly: `0.0.YYYYMMDD-<commit-hash>`
- On-demand: `0.0.YYYYMMDD-<commit-hash>`

---

## Build Artifacts

### .deb Packages (Debian packages)
- **Size**: ~80-90 MB
- **Use**: Upgrade existing Stratux installations
- **Installation**: `sudo dpkg -i stratux-*.deb && sudo systemctl restart stratux`

### .img Files (SD Card Images)
- **Size**: ~2-4 GB (compressed to ~500MB-1GB in .zip)
- **Use**: New installations or complete system reinstall
- **Installation**: Write to SD card with Raspberry Pi Imager or balenaEtcher

---

## Regions

### US Region
- **UAT Enabled**: true (978 MHz)
- **OGN Enabled**: false
- **Developer Mode**: false
- **Region**: 1
- **Use**: United States, FAA ADS-B reception

### EU Region
- **UAT Enabled**: false
- **OGN Enabled**: true (868 MHz)
- **Developer Mode**: true
- **Region**: 2
- **Use**: Europe, FLARM/OGN reception

---

## Troubleshooting

### Build Failed

Check the logs:
```bash
gh run view <run-id> --log-failed
```

Common issues:
- **Missing dependencies**: Check the "Install build dependencies" step
- **Version script error**: Check `scripts/getversion.sh`
- **Region config error**: Check Makefile modification step

### Release Not Appearing

- Check if tag was pushed: `git ls-remote --tags origin`
- Verify workflow triggered: `gh run list --workflow=release-images.yml`
- Check if release is in draft state

### Artifacts Too Large

GitHub has a 2GB limit per artifact. If images exceed this:
- Images are automatically zipped
- Download may take several minutes

---

## Quick Reference

```bash
# Nightly build (automatic, or manual trigger)
gh workflow run nightly.yml

# Quick .deb build for testing
gh workflow run build-regions.yml -f region=both

# Official release
git tag -a v1.6.0 -m "Release 1.6.0"
git push origin v1.6.0

# Check build status
gh run list --limit 5

# Download artifacts
gh run download <run-id>
```

---

## Testing Releases

### Before Publishing

1. **Download pre-release artifacts**
2. **Test .deb installation**:
   ```bash
   scp stratux-US-*.deb pi@testpi:~/
   ssh pi@testpi "sudo dpkg -i stratux-US-*.deb && sudo systemctl restart stratux"
   ```
3. **Test image**:
   - Write to test SD card
   - Boot on Raspberry Pi
   - Verify web interface accessible
   - Check GPS, ADS-B, traffic reception
4. **Verify both US and EU if changes affect region-specific features**

---

## Notes

- All builds run on native ARM64 GitHub runners (fast, no emulation)
- Nightly builds automatically clean up old nightlies (keep last 7 days)
- Draft releases must be manually published
- Tags cannot be moved once pushed (delete and recreate if needed)

---

## CI/CD Pipeline Summary

```
Development
    ↓
Commit to master
    ↓
├─→ Nightly Build (2 AM UTC)
│   └─→ Prerelease with .deb files
│
├─→ On-demand build (manual)
│   └─→ Artifacts only (no release)
│
└─→ Push version tag
    └─→ Full release build
        └─→ Draft release with images + .deb files
            └─→ Manual publish
```
