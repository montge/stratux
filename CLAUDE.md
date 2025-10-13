# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Stratux is an ADS-B/UAT/OGN receiver for general aviation aircraft, built to run on Raspberry Pi. It receives aviation traffic and weather data via SDR (Software Defined Radio) and broadcasts it to Electronic Flight Bag (EFB) applications using the GDL90 protocol over WiFi.

**Target Platform**: Raspberry Pi (arm64) running Debian 12 (Bookworm)
**Languages**: Go (main application), C (SDR libraries), JavaScript (web interface)
**Version**: Semantic versioning (MAJOR.MINOR), e.g., v3.6

## Build Commands

### Local Development Build
```bash
# Initialize submodules first (required)
git submodule update --init --recursive

# Standard build (for local development/testing)
make all

# Install to /opt/stratux/
sudo make optinstall

# Build and install Debian package locally
make dpkg
sudo make install
```

### Docker-Based Build (Recommended for Release)
```bash
# Build everything in Docker container
make dall

# Build Debian package in Docker (matches CI environment)
make ddpkg
```

### Testing
```bash
# Run tests
make -C test
```

### Web Interface
```bash
# Build web interface
make -C web
```

### Cleanup
```bash
make clean
```

## Architecture Overview

### Core Components

**main/** - Main application package containing the Stratux daemon
- `gen_gdl90.go` - GDL90 protocol implementation, heartbeat, ownship reports
- `traffic.go` - Traffic target tracking and GDL90 traffic message generation
- `network.go` - Network output handling (GDL90, NMEA, etc.)
- `managementinterface.go` - Web interface HTTP/WebSocket handlers
- `sdr.go` - SDR (Software Defined Radio) management for UAT/1090ES receivers
- `gps.go` - GPS/GNSS receiver handling and NMEA parsing
- `sensors.go` - AHRS/IMU and barometric pressure sensor integration
- `ogn.go` - Open Glider Network receiver integration (EU)
- `ais.go` - AIS (marine traffic) receiver integration

**common/** - Shared utilities and helper functions
- `helpers.go` - General utility functions
- `equations.go` - Aviation-related calculations
- `cputemp.go` - CPU temperature monitoring

**dump1090/** - 1090MHz ADS-B receiver (C-based, submodule)

**dump978/** - 978MHz UAT receiver (C library with Go bindings)

**rtl-ais/** - AIS receiver for marine traffic (C-based, submodule)

**web/** - Web-based configuration portal
- HTML/CSS/JavaScript single-page application
- Real-time status updates via WebSockets
- Configuration management UI

### Data Flow

```
SDR Hardware → dump1090/dump978/ogn-rx → main/sdr.go → parseInput()
                                                ↓
                                         Traffic Tracking
                                                ↓
                                    GDL90 Message Generation
                                                ↓
                                      Network Distribution
                                                ↓
                              Connected EFB Apps (UDP unicast)
```

### Message Classes

Stratux tracks different types of traffic sources:
- `MSGCLASS_UAT` (0) - UAT messages (978 MHz, US)
- `MSGCLASS_ES` (1) - 1090ES messages (Mode-S Extended Squitter)
- `MSGCLASS_OGN` (2) - Open Glider Network (868 MHz, EU)
- `MSGCLASS_AIS` (3) - Marine AIS messages

### Configuration

**Runtime config**: `/boot/firmware/stratux.conf` (JSON)
**Default config template**: `/opt/stratux/cfg/stratux.conf.default`

Settings are managed via the web interface and stored in the runtime config file.

## OTA (Over-The-Air) Update Process

Stratux supports OTA updates via two mechanisms:

1. **Debian Package (.deb)** - Preferred method for application updates
   - Uploaded via web interface → `/overlay/robase/root/`
   - On boot: `stratux-pre-start.sh` installs via `dpkg -i`
   - Handles all Stratux executables and configuration

2. **Update Script** - For system-level updates outside the Stratux application
   - Legacy method, still available for special cases
   - Similar process but executes a script instead of dpkg

## Development Environment

### Remote Development (Recommended)
Work directly on a Raspberry Pi via SSH/VSCode Remote-SSH:
1. Enable persistent logging on web interface (makes filesystem writable)
2. Remove CPU frequency limits in `/boot/firmware/config.txt` for faster compilation
3. Install Go and build tools on the Pi
4. Use VSCode with Remote-SSH extension

### Local x86 Development (Advanced)
Build and run on x86/x64 Linux for faster iteration:
- Hardware sensors (AHRS, GPS) won't be available
- Some SDR functionality may not work
- Use `make optinstall` to avoid installing systemd services
- Config location remapped to `~/.stratux.conf` when not running as root

## Key Code Patterns

### GPS Types and Protocols
GPS device types are encoded in a combined nibble format (see `main/gen_gdl90.go:100-120`):
- Lower nibble: GPS hardware type (1-15)
- Upper nibble: Protocol type (e.g., `GPS_PROTOCOL_NMEA = 0x10`)
- **Important**: This enumeration has a JavaScript duplicate in `web/plates/js/status.js` that must be kept in sync manually

### Mutex Usage
The codebase uses mutexes extensively for thread safety:
- `mySituation` has separate mutexes for GPS, attitude, baro, and satellite data
- `ADSBTowerMutex` protects the `ADSBTowers` map
- `msgLogMutex` protects the message log

### Stratux Clock
Uses a custom monotonic clock (`stratuxClock`) instead of `time.Now()` in most places to ensure consistent time tracking even if system time changes.

## Debian Package Structure

**What goes in the .deb package:**
- All Stratux executables (`stratuxrun`, `fancontrol`, `dump1090`, `rtl_ais`, `ogn-rx-eu`)
- Libraries (`libdump978.so`)
- Scripts (`stratux-pre-start.sh`, `stratux-wifi.sh`, `sdr-tool.sh`)
- Config templates
- Systemd service files
- udev rules

**What goes in the system image (not .deb):**
- Base OS and libraries
- System configuration referencing Stratux scripts
- Custom bluez/librtlsdr builds (until Debian 13)

## Web Interface Integration

The management interface (`main/managementinterface.go`) exposes REST and WebSocket endpoints:

**HTTP Endpoints:**
- `GET /getStatus` - Device status and statistics
- `GET /getSettings` - Current settings
- `POST /setSettings` - Update settings
- `GET /getSituation` - GPS/AHRS data
- `GET /getTowers` - ADS-B tower list with signal stats
- `POST /calibrateAHRS`, `/cageAHRS`, `/resetGMeter` - Sensor operations
- `POST /restart`, `/reboot`, `/shutdown` - System control

**WebSocket Endpoints:**
- `ws://192.168.10.1/traffic` - Real-time traffic updates
- Weather and status updates (see `main/managementinterface.go`)

See `docs/app-vendor-integration.md` for detailed API documentation.

## Region-Specific Features

Stratux supports US and EU configurations:
- **US**: UAT enabled (978 MHz), OGN disabled
- **EU**: UAT disabled, OGN enabled (868 MHz), Developer mode enabled
- Region selection is managed in settings (`RegionSelected`: 0=none, 1=US, 2=EU)

## CI/CD

GitHub Actions workflows:
- **ci.yml** - Builds .deb package on every push using Docker + Debian Bookworm
- **release.yml** - Creates releases with .deb package and full Raspberry Pi image

Both use `make ddpkg` (Docker-based dpkg build) to ensure consistent build environment.

## Important Files

- `scripts/getversion.sh` - Version string extraction
- `scripts/getarch.sh` - Architecture detection
- `debian/stratux-pre-start.sh` - Boot-time initialization and OTA update handler
- `debian/stratux-wifi.sh` - WiFi configuration management
- `debian/sdr-tool.sh` - SDR device management utilities

## Testing and Debugging

### Replay Mode
Stratux can replay recorded data for testing:
```bash
./stratuxrun -replay -uatlog <logfile> -speed 2
```

### Trace Replay
Record and replay all system inputs:
```bash
# Recording is enabled via web interface (TraceLog setting)
# Replay:
./stratuxrun -trace <tracefile> -traceSpeed 1.0 -traceFilter "dump1090,godump978"
```

### Development Mode
Enable in settings for additional logging and features useful for debugging.

## Coding Conventions

- Use existing coding style in the file you're editing
- Go code follows standard Go conventions
- Comment significant functions and complex logic
- Keep code self-documenting where possible
