# Stratux Build System - Quick Start

## 🎯 What You Need to Know

This repository can build **region-specific Stratux images** for Raspberry Pi:
- **US version**: UAT enabled (978 MHz for FAA weather/traffic)
- **EU version**: OGN enabled (868 MHz for glider tracking)

## 🚀 Fastest Path to Building

### If you have a Raspberry Pi available:
```bash
# On the Raspberry Pi:
./build-native.sh
# Choose your region, wait 5-15 minutes, done!
```

### If building on your laptop/desktop (x86_64):
```bash
# One-time setup:
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

# Then build:
./build-images.sh
# Wait 30-60 minutes, done!
```

## 📚 Full Documentation

- **[WHICH_BUILD_METHOD.md](WHICH_BUILD_METHOD.md)** ← Start here! Quick decision guide
- **[BUILD_INSTRUCTIONS.md](BUILD_INSTRUCTIONS.md)** - Complete build reference
- **[CLAUDE.md](CLAUDE.md)** - Architecture overview for AI assistants

## ⚡ Speed Comparison

| Method | Time | Hardware |
|--------|------|----------|
| Native (on Pi) | 5-15 min | Raspberry Pi 4/5 |
| Cross-compile | 30-60 min | x86_64 laptop/desktop |

## 🎬 Quick Examples

### Example 1: Build and Deploy on Same Pi
```bash
# Clone and build directly on the Pi
git clone <your-repo>
cd stratux
git submodule update --init --recursive
./build-native.sh
# Choose option, wait 10 minutes
sudo dpkg -i stratux-*.deb
sudo systemctl restart stratux
```

### Example 2: Build on Laptop, Deploy to Pi
```bash
# On your laptop:
git clone <your-repo>
cd stratux
git submodule update --init --recursive
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
./build-images.sh
# Wait 45 minutes
scp build_output/stratux-*.deb pi@192.168.1.x:~/

# On the Pi:
sudo dpkg -i stratux-*.deb
sudo systemctl restart stratux
```

## 🔧 Available Build Scripts

| Script | Purpose | Where to Run |
|--------|---------|--------------|
| `build-native.sh` | Fast native build | Raspberry Pi only |
| `build-images.sh` | Full build with images | x86_64 with Docker |
| `test-build.sh` | Quick test (one .deb) | x86_64 with Docker |

## 💡 Recommended Workflow

**For equipment deployment** (your use case):

1. **Development/Testing**: Use `./build-native.sh` on a test Raspberry Pi
   - Fastest feedback loop (5-15 min)
   - Test on actual hardware immediately

2. **Production Release**: Use `./build-images.sh` on your dev machine
   - Create both US and EU versions
   - Generate bootable images for new installations
   - Distribute .deb files for OTA updates

## 📦 What Gets Built

### .deb Packages (both methods)
- Contains all Stratux software
- Can be installed with `dpkg -i`
- Used for OTA updates
- ~50MB in size

### Full Images (cross-compile only)
- Bootable Raspberry Pi OS image
- Stratux pre-installed and configured
- Flash to SD card with Raspberry Pi Imager
- ~1.4GB compressed

## ❓ Common Questions

**Q: Which method should I use?**
A: See [WHICH_BUILD_METHOD.md](WHICH_BUILD_METHOD.md) for a decision guide

**Q: Can I build on Mac?**
A: Yes, with Docker installed. Use the cross-compile method.

**Q: Do I need a Raspberry Pi to build?**
A: No, you can cross-compile on x86_64. But native builds on Pi are much faster.

**Q: What's the version number?**
A: Built from current commit (`3339fd3`). No version tags yet in this fork.

**Q: Can I test changes quickly?**
A: Yes, use native builds on a test Pi - 5-15 minute turnaround time.

## 🐛 Troubleshooting

See [BUILD_INSTRUCTIONS.md](BUILD_INSTRUCTIONS.md) "Troubleshooting" section for:
- Docker network issues
- QEMU setup problems
- Out of space errors
- Slow build performance

## 🔗 Quick Links

- Main build guide: [BUILD_INSTRUCTIONS.md](BUILD_INSTRUCTIONS.md)
- Method comparison: [WHICH_BUILD_METHOD.md](WHICH_BUILD_METHOD.md)
- Architecture docs: [CLAUDE.md](CLAUDE.md)
- Stratux Wiki: https://github.com/stratux/stratux/wiki

## 🎉 Next Steps

1. Read [WHICH_BUILD_METHOD.md](WHICH_BUILD_METHOD.md) to choose your build method
2. Follow the appropriate instructions
3. Deploy to your equipment
4. Enjoy your Stratux!

---

**Note**: Cross-compilation works via QEMU emulation. It's slower but works anywhere Docker runs. Native builds on Pi are 5-10x faster and recommended for development.
