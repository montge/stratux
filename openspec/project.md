# Project Context

## Purpose
Stratux is an open-source ADS-B/UAT/OGN receiver for general aviation aircraft. It receives aviation traffic and weather data via SDR (Software Defined Radio) and broadcasts it to Electronic Flight Bag (EFB) applications using the GDL90 protocol over WiFi.

## Tech Stack
- **Go** - Main application (daemon, networking, protocol handling)
- **C** - SDR libraries (dump1090, dump978, rtl-ais)
- **JavaScript** - Web interface (vanilla JS, AngularJS for legacy UI)
- **HTML/CSS** - Web configuration portal
- **Target Platform**: Raspberry Pi (arm64) running Debian 12 (Bookworm)

## Project Conventions

### Code Style
- Go code follows standard Go conventions (`gofmt`)
- Use existing coding style in the file you're editing
- Comment significant functions and complex logic
- Keep code self-documenting where possible

### Architecture Patterns
- SDR hardware → C receivers → Go main application → GDL90 output
- Mutex-based thread safety for shared state
- Custom monotonic clock (`stratuxClock`) for consistent timing
- WebSocket for real-time updates to web interface

### Testing Strategy
- Unit tests in `*_test.go` files alongside source
- E2E tests for integration testing
- Coverage target: 80% (tracked via SonarCloud)
- Run tests with: `go test -v ./main/... ./common/...`

### Git Workflow
- Main branch: `master`
- Feature branches merged via PR
- CI runs on all pushes and PRs
- Semantic versioning (MAJOR.MINOR)

## Domain Context
- **ADS-B**: Automatic Dependent Surveillance-Broadcast (1090 MHz)
- **UAT**: Universal Access Transceiver (978 MHz, US only)
- **OGN**: Open Glider Network (868 MHz, EU)
- **GDL90**: Garmin Data Link protocol for EFB communication
- **FLARM**: Traffic awareness system popular in gliders

## Important Constraints
- Must run on Raspberry Pi with limited resources
- Real-time processing of radio signals
- WiFi hotspot mode for EFB connectivity
- Power consumption considerations for aircraft use

## External Dependencies
- RTL-SDR libraries for radio reception
- dump1090/dump978 for ADS-B/UAT decoding
- OGN receiver for European traffic
- GPS/GNSS receivers for position data
- AHRS/IMU sensors for attitude data
