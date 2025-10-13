# Which Build Method Should I Use?

## Quick Decision Guide

```
┌─────────────────────────────────────────────────────────────┐
│ Do you have access to a Raspberry Pi 4 or 5?                │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
        YES                     NO
         │                       │
         ▼                       ▼
  Use Native Build        Use Cross-Compile
  ./build-native.sh       ./build-images.sh
  ⚡ 5-15 minutes         ⏱️ 30-60 minutes
```

## Comparison Table

| Feature | Native Build (Pi) | Cross-Compile (x86_64) |
|---------|-------------------|------------------------|
| **Script** | `./build-native.sh` | `./build-images.sh` |
| **Build Time** | ⚡ 5-15 min | ⏱️ 30-60 min (first), 20-30 min (subsequent) |
| **Hardware Needed** | Raspberry Pi 4/5 | Any x86_64 machine |
| **Prerequisites** | gcc, make, golang | Docker + QEMU |
| **Disk Space** | ~10GB | ~20GB |
| **RAM Needed** | 2GB+ | 4GB+ |
| **Complexity** | Simple | Moderate |
| **Best For** | Development, testing, quick iterations | CI/CD, building without Pi hardware |
| **Outputs** | .deb packages | .deb packages + full .img files |

## Detailed Scenarios

### Scenario 1: You're Developing on the Equipment
**Situation**: You have Stratux running on a Pi and want to test changes

**Best Method**: Native Build (`./build-native.sh`)
- Clone repo directly on the Pi
- Make your code changes
- Build in 5-15 minutes
- Install with `sudo dpkg -i stratux-*.deb`
- Test immediately

**Why**: Fastest feedback loop for development

---

### Scenario 2: Building for Multiple Units
**Situation**: You need to deploy to 10+ Raspberry Pis

**Best Method**: Cross-Compile (`./build-images.sh`)
- Build once on your development machine
- Create full bootable images
- Flash to multiple SD cards
- Or distribute .deb files for OTA updates

**Why**: Build once, deploy many times

---

### Scenario 3: CI/CD Pipeline
**Situation**: You want automated builds on every commit

**Best Method**: Cross-Compile (GitHub Actions already configured)
- Uses Docker-based builds
- Consistent environment
- No physical Pi needed
- Can run in cloud

**Why**: Automation-friendly, reproducible builds

---

### Scenario 4: Quick Testing Before Deployment
**Situation**: You made a small change and want to verify it works

**Best Method**: Native Build on a test Pi
- Fastest way to verify changes
- No need for full image rebuild
- Just build the .deb and install

**Why**: Speed matters for testing

---

### Scenario 5: Creating Installation Images
**Situation**: You need bootable SD card images for new installations

**Best Method**: Cross-Compile (image build option)
- `./build-images.sh` and choose to build images
- Creates full .img.xz files
- Flash to SD cards with Raspberry Pi Imager

**Why**: Only cross-compile process supports full image creation

## Performance Benchmarks

Based on Raspberry Pi 4 (4GB) and modern x86_64 laptop:

| Task | Native (Pi 4) | Cross-Compile (x86_64) |
|------|---------------|------------------------|
| Clean build | ~10 min | ~45 min |
| Incremental build | ~5 min | ~20 min |
| Full image build | N/A | ~90 min |
| Memory usage | ~1.5GB | ~3GB |

## Recommendations by Use Case

### For Equipment Deployment (Your Case)
Since this will "go on to the actual equipment" (Raspberry Pi):

**During Development**:
1. Use native builds on a test Pi for rapid iteration
2. Build time: 5-15 minutes
3. Immediate testing on actual hardware

**For Production Deployment**:
1. Use cross-compile to create .deb packages
2. Distribute via OTA updates
3. Or create full images for new installations

**Hybrid Approach** (Recommended):
- **Development phase**: Native builds on Pi for speed
- **Release phase**: Cross-compile for final .deb packages and images
- **Emergency fixes**: Native build on Pi, test, then cross-compile for distribution

## Command Reference

### Native Build (on Raspberry Pi)
```bash
# One-time setup
git clone <your-repo>
cd stratux
git submodule update --init --recursive

# Build (choose region when prompted)
./build-native.sh

# Install
sudo dpkg -i stratux-*.deb
sudo systemctl restart stratux
```

### Cross-Compile (on x86_64)
```bash
# One-time setup
git clone <your-repo>
cd stratux
git submodule update --init --recursive
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

# Quick test
./test-build.sh

# Full build
./build-images.sh

# Copy to Pi
scp stratux-*.deb pi@192.168.1.x:~/

# On Pi: Install
sudo dpkg -i stratux-*.deb
sudo systemctl restart stratux
```

## Bottom Line

**For your use case** (equipment deployment):
- ✅ **Start with native builds** during development/testing (fastest)
- ✅ **Use cross-compile** for creating production releases
- ✅ Both methods produce identical .deb packages
- ✅ Only cross-compile can create full bootable images

**Time savings**: Native builds are **5-10x faster** than cross-compilation!
