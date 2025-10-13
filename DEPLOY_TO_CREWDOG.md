# Deploying to Crewdog (10.0.1.53)

## Overview

This guide covers deploying Stratux builds to your device "crewdog" at 10.0.1.53

## Option A: Build Directly on Crewdog (RECOMMENDED - Fastest)

### Step 1: Copy Repository to Crewdog
```bash
# From your development machine:
ssh pi@10.0.1.53 "mkdir -p ~/stratux-build"
rsync -avz --exclude '.git' --exclude 'build_output' \
  /home/e/Development/stratux/ pi@10.0.1.53:~/stratux-build/
```

### Step 2: SSH into Crewdog and Build
```bash
# SSH to the device
ssh pi@10.0.1.53

# Navigate to the build directory
cd ~/stratux-build

# Initialize submodules (first time only)
git submodule update --init --recursive

# Build! (5-15 minutes)
./build-native.sh
# Choose your region when prompted
```

### Step 3: Install
```bash
# Still on crewdog:
sudo dpkg -i stratux-*.deb
sudo systemctl restart stratux
```

### Step 4: Verify
```bash
# Check service status
sudo systemctl status stratux

# View logs
sudo journalctl -u stratux -f
```

---

## Option B: Cross-Compile and Deploy

### Step 1: Build on Your Machine
```bash
# On your development machine:
cd /home/e/Development/stratux

# One-time QEMU setup (if not done):
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

# Build
./build-images.sh
# This creates build_output/stratux-*.deb files
```

### Step 2: Copy to Crewdog
```bash
# From your development machine:
scp build_output/stratux-*.deb pi@10.0.1.53:~/
```

### Step 3: Install on Crewdog
```bash
# SSH to crewdog
ssh pi@10.0.1.53

# Install the package
sudo dpkg -i stratux-*.deb

# Restart service
sudo systemctl restart stratux
```

---

## Option C: Development Iteration Workflow (Best for Testing)

When making code changes and want to test quickly:

### Setup (One Time)
```bash
# Mount crewdog's filesystem via SSHFS for easy editing
mkdir -p ~/mnt/crewdog
sshfs pi@10.0.1.53:/home/pi/stratux-build ~/mnt/crewdog

# Or use VSCode Remote-SSH:
# 1. Install "Remote - SSH" extension
# 2. Connect to pi@10.0.1.53
# 3. Open folder ~/stratux-build
```

### Iteration Loop
```bash
# Edit code locally or via VSCode Remote-SSH

# Build on crewdog (fastest - 5-15 min)
ssh pi@10.0.1.53 "cd ~/stratux-build && ./build-native.sh"

# Install
ssh pi@10.0.1.53 "cd ~/stratux-build && sudo dpkg -i stratux-*.deb && sudo systemctl restart stratux"

# Check logs
ssh pi@10.0.1.53 "sudo journalctl -u stratux -f"
```

---

## Quick Commands Reference

### Status Checks
```bash
# Check if Stratux is running
ssh pi@10.0.1.53 "sudo systemctl status stratux"

# View live logs
ssh pi@10.0.1.53 "sudo journalctl -u stratux -f"

# Check version
ssh pi@10.0.1.53 "dpkg -l | grep stratux"

# Check CPU/memory
ssh pi@10.0.1.53 "top -n 1 | head -20"
```

### Service Management
```bash
# Stop Stratux
ssh pi@10.0.1.53 "sudo systemctl stop stratux"

# Start Stratux
ssh pi@10.0.1.53 "sudo systemctl start stratux"

# Restart Stratux
ssh pi@10.0.1.53 "sudo systemctl restart stratux"

# Disable auto-start
ssh pi@10.0.1.53 "sudo systemctl disable stratux"

# Enable auto-start
ssh pi@10.0.1.53 "sudo systemctl enable stratux"
```

### Backup and Restore
```bash
# Backup current configuration
ssh pi@10.0.1.53 "sudo cp /boot/firmware/stratux.conf ~/stratux.conf.backup"

# Backup entire /opt/stratux
ssh pi@10.0.1.53 "sudo tar czf ~/stratux-backup.tar.gz /opt/stratux"

# Copy backup to your machine
scp pi@10.0.1.53:~/stratux-backup.tar.gz ~/backups/

# Restore configuration
ssh pi@10.0.1.53 "sudo cp ~/stratux.conf.backup /boot/firmware/stratux.conf"
```

### Troubleshooting
```bash
# Check for errors in logs
ssh pi@10.0.1.53 "sudo journalctl -u stratux --since '10 minutes ago' | grep -i error"

# Check disk space
ssh pi@10.0.1.53 "df -h"

# Check temperature (important!)
ssh pi@10.0.1.53 "vcgencmd measure_temp"

# List USB devices (SDR dongles)
ssh pi@10.0.1.53 "lsusb"

# Check network interfaces
ssh pi@10.0.1.53 "ip addr show"
```

---

## Recommended Workflow for Crewdog

Given you have direct access to the device:

### For Development/Testing:
1. ✅ **Use Option A** (build directly on crewdog)
   - Fastest turnaround: 5-15 minutes
   - No cross-compilation overhead
   - Immediate testing

### For Production Deployment:
1. Build on your dev machine once
2. Create both US and EU versions
3. Deploy to crewdog and other units as needed

### For Quick Iterations:
1. ✅ **Use Option C** (development iteration workflow)
   - Edit code with VSCode Remote-SSH
   - Build on crewdog: `./build-native.sh`
   - Install: `sudo dpkg -i stratux-*.deb`
   - Test immediately

---

## First Time Setup Checklist

- [ ] Ensure crewdog is accessible: `ping 10.0.1.53`
- [ ] Test SSH access: `ssh pi@10.0.1.53`
- [ ] Copy repository to crewdog
- [ ] Initialize submodules on crewdog
- [ ] Run first build: `./build-native.sh`
- [ ] Backup existing configuration
- [ ] Install new build
- [ ] Verify Stratux is running
- [ ] Access web interface: http://10.0.1.53 (or http://stratux.local)

---

## Safety Notes

⚠️ **Before deploying to crewdog:**
1. Backup existing configuration
2. Test on a separate Pi if possible
3. Have the original SD card image as backup
4. Don't deploy during flight operations!

📝 **After deployment:**
1. Verify web interface is accessible
2. Check that GPS is acquiring
3. Verify SDR dongles are detected
4. Monitor temperature
5. Check for any error messages in logs

---

## Performance Expectations

On Raspberry Pi 4 (typical for Stratux):
- Build time: 10-15 minutes
- Installation: <1 minute
- Service restart: 5-10 seconds
- Total deployment: ~15 minutes

---

## Next Steps

1. Choose your deployment method (A, B, or C above)
2. Backup crewdog's current configuration
3. Deploy your first build
4. Test thoroughly
5. Iterate as needed

For questions about the build system itself, see:
- [BUILD_README.md](BUILD_README.md) - Quick start
- [WHICH_BUILD_METHOD.md](WHICH_BUILD_METHOD.md) - Method comparison
- [BUILD_INSTRUCTIONS.md](BUILD_INSTRUCTIONS.md) - Full reference
