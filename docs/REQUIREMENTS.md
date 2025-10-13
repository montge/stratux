# Stratux System Requirements Specification (SRS)

**Document ID**: SRS-STRATUX-001
**Version**: 1.0 DRAFT
**Date**: 2025-10-13
**Classification**: SAL-3 (Software Assurance Level 3 per DO-278A)
**Status**: Draft for Review

---

## 1. Introduction

### 1.1 Purpose

This document specifies the functional and non-functional requirements for the Stratux ADS-B/UAT/OGN receiver system.

### 1.2 Scope

Stratux is a portable aviation receiver that:
- Receives ADS-B, UAT, OGN/FLARM, and AIS radio signals
- Processes and fuses traffic and weather information
- Provides GPS position and AHRS attitude data
- Outputs data to Electronic Flight Bag (EFB) applications via multiple protocols

### 1.3 Intended Use

**PRIMARY USE**: Supplemental situational awareness for pilots
**NOT INTENDED FOR**:
- Primary navigation
- Terrain avoidance (TAWS)
- Collision avoidance (TCAS)
- Primary attitude reference
- Meeting FAA ADS-B Out mandate

### 1.4 Definitions

- **ADS-B**: Automatic Dependent Surveillance-Broadcast
- **UAT**: Universal Access Transceiver (978 MHz)
- **1090ES**: 1090 MHz Extended Squitter
- **OGN**: Open Glider Network (868 MHz)
- **GDL90**: Garmin Data Link protocol (output format)
- **AHRS**: Attitude and Heading Reference System
- **EFB**: Electronic Flight Bag
- **Ownship**: The aircraft carrying the Stratux receiver

---

## 2. System Overview

### 2.1 System Context

```
┌──────────────────────────────────────────────────────────────────────┐
│                        EXTERNAL INTERFACES                           │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  RF Inputs          GPS Input           Sensors          Config      │
│  ┌─────────┐      ┌──────────┐        ┌────────┐      ┌─────────┐  │
│  │ 1090MHz │      │UART/USB  │        │I2C IMU │      │Web UI   │  │
│  │ 978MHz  │──┐   │NMEA/TCP  │──┐     │I2C Baro│──┐   │Serial   │  │
│  │ 868MHz  │  │   └──────────┘  │     └────────┘  │   │File     │  │
│  │ 162MHz  │  │                  │                 │   └─────────┘  │
│  └─────────┘  │                  │                 │         │       │
│               ▼                  ▼                 ▼         ▼       │
│         ┌──────────────────────────────────────────────────────┐   │
│         │                 STRATUX CORE                         │   │
│         │  ┌────────────┐  ┌──────────┐  ┌────────────────┐  │   │
│         │  │Radio       │  │GPS       │  │AHRS            │  │   │
│         │  │Processing  │  │Processing│  │Processing      │  │   │
│         │  └──────┬─────┘  └────┬─────┘  └───────┬────────┘  │   │
│         │         │             │                 │           │   │
│         │         ▼             ▼                 ▼           │   │
│         │  ┌────────────────────────────────────────────┐    │   │
│         │  │        Traffic & Weather Fusion             │    │   │
│         │  └───────────────────┬────────────────────────┘    │   │
│         │                      │                              │   │
│         │         ┌────────────┴────────────┐                │   │
│         │         ▼                          ▼                │   │
│         │  ┌─────────────┐          ┌──────────────┐         │   │
│         │  │GDL90 Output │          │NMEA Output   │         │   │
│         │  └──────┬──────┘          └──────┬───────┘         │   │
│         └─────────┼────────────────────────┼─────────────────┘   │
│                   │                         │                      │
│                   ▼                         ▼                      │
│         ┌──────────────────────────────────────────┐              │
│         │    Network Distribution                  │              │
│         │  UDP │ TCP │ Serial │ Bluetooth │ Web   │              │
│         └──┬───┴──┬──┴────┬───┴──────┬────┴───┬───┘              │
│            │      │       │          │        │                   │
│            ▼      ▼       ▼          ▼        ▼                   │
│         ┌──────────────────────────────────────────┐              │
│         │         EFB APPLICATIONS                 │              │
│         │  ForeFlight │ WingX │ Garmin Pilot │... │              │
│         └──────────────────────────────────────────┘              │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 3. Functional Requirements

### 3.1 Radio Reception (FR-100 series)

#### FR-101: 1090MHz ADS-B Reception
**Priority**: HIGH | **Verification**: Test | **Traceability**: DO-260B

The system SHALL receive and decode 1090 MHz ADS-B Extended Squitter messages.

**Acceptance Criteria**:
- Messages decoded per DO-260B
- CRC validation performed
- Invalid messages discarded
- Position, velocity, identification extracted

#### FR-102: 978MHz UAT Reception
**Priority**: HIGH | **Verification**: Test | **Traceability**: DO-282B

The system SHALL receive and decode 978 MHz UAT messages containing traffic and weather.

**Acceptance Criteria**:
- Messages decoded per DO-282B
- FIS-B weather products extracted
- TIS-B traffic extracted
- Reed-Solomon error correction applied

#### FR-103: 868MHz OGN Reception
**Priority**: MEDIUM | **Verification**: Test

The system SHALL receive and decode 868 MHz OGN/FLARM messages.

**Acceptance Criteria**:
- FLARM protocol messages decoded
- OGN APRS messages decoded
- FANET messages decoded
- PAW (PilotAware) messages decoded

#### FR-104: 162MHz AIS Reception
**Priority**: LOW | **Verification**: Test

The system SHALL optionally receive and decode 162 MHz marine AIS messages.

**Acceptance Criteria**:
- AIVDM/AIVDO sentences decoded
- Vessel position and identification extracted
- Maritime traffic distinguished from aircraft

#### FR-105: SDR Device Management
**Priority**: HIGH | **Verification**: Test

The system SHALL automatically detect and configure RTL-SDR devices.

**Acceptance Criteria**:
- Auto-detection by USB serial number pattern
- Frequency assignment (1090/978/868/162)
- Gain control (automatic and manual)
- Automatic reconnection after device failure

---

### 3.2 GPS Subsystem (FR-200 series)

#### FR-201: GPS Position Acquisition
**Priority**: CRITICAL | **Verification**: Test

The system SHALL acquire and maintain GPS position fix.

**Acceptance Criteria**:
- 3D fix acquired within 60 seconds (hot start)
- Position accuracy ≤ 10 meters (95% with WAAS)
- Update rate ≥ 1 Hz
- Time-to-first-fix ≤ 35 seconds (warm start)

#### FR-202: GPS Device Support
**Priority**: HIGH | **Verification**: Test

The system SHALL support multiple GPS receiver types.

**Acceptance Criteria**:
- u-blox (6/7/8/9/10 series) support
- Prolific USB (SIRF) support
- Generic NMEA support
- Network NMEA (TCP port 30011) support
- OGN Tracker as GPS source

#### FR-203: GPS Configuration
**Priority**: MEDIUM | **Verification**: Test

The system SHALL configure GPS receivers for aviation use.

**Acceptance Criteria**:
- u-blox: Airborne <4g dynamic model
- u-blox: SBAS/WAAS enabled
- u-blox: Rate set to 10 Hz (if supported)
- u-blox: Power management disabled

#### FR-204: GPS Data Validation
**Priority**: CRITICAL | **Verification**: Test

The system SHALL validate GPS data before use.

**Acceptance Criteria**:
- Fix quality checked (2D/3D/DGPS)
- HDOP threshold (reject if > 4.0)
- Satellite count minimum (≥ 4)
- Position sanity check (< 600 kt ground speed)
- Time validity check

#### FR-205: GPS Status Reporting
**Priority**: HIGH | **Verification**: Test

The system SHALL report GPS status to users.

**Acceptance Criteria**:
- Fix type indicated (None/2D/3D/DGPS)
- Satellite count reported
- Horizontal accuracy (meters) reported
- Position timestamp provided
- GPS device type identified

---

### 3.3 AHRS Subsystem (FR-300 series)

#### FR-301: AHRS Sensor Support
**Priority**: MEDIUM | **Verification**: Test

The system SHALL optionally support AHRS sensors.

**Acceptance Criteria**:
- MPU-9250/9255 IMU support
- MPU-6500 IMU support
- ICM-20948 IMU support
- BMP-280 barometer support
- BMP-388 barometer support
- I2C auto-detection and configuration

#### FR-302: AHRS Data Output
**Priority**: MEDIUM | **Verification**: Test

The system SHALL provide attitude data when AHRS sensors available.

**Acceptance Criteria**:
- Roll angle (±180°)
- Pitch angle (±90°)
- Heading (0-360°, magnetic)
- Turn rate (°/second)
- Slip/skid indicator
- G-load (with min/max tracking)

#### FR-303: Barometric Altitude
**Priority**: HIGH | **Verification**: Test

The system SHALL provide barometric pressure altitude.

**Acceptance Criteria**:
- Altitude relative to 29.92" Hg
- Vertical speed (ft/min)
- Temperature compensation
- Update rate ≥ 10 Hz

#### FR-304: AHRS Calibration
**Priority**: MEDIUM | **Verification**: Test

The system SHALL support AHRS sensor calibration.

**Acceptance Criteria**:
- Magnetometer calibration procedure
- Accelerometer calibration
- Calibration values stored persistently
- Calibration status indicated

#### FR-305: AHRS Safety Disclaimer
**Priority**: CRITICAL | **Verification**: Inspection

The system SHALL mark AHRS data as advisory only.

**Acceptance Criteria**:
- GDL90 AHRS message includes "AHRS not for primary" indicator
- Web UI displays warning "Not for primary flight reference"
- Documentation includes safety disclaimer

---

### 3.4 Traffic Fusion (FR-400 series)

#### FR-401: Multi-Source Traffic Fusion
**Priority**: HIGH | **Verification**: Test

The system SHALL fuse traffic from multiple sources (1090ES, UAT, OGN, AIS).

**Acceptance Criteria**:
- Same target from multiple sources consolidated
- Most recent/accurate data used
- Source indicated (ADS-B/UAT/OGN/AIS)
- Duplicate suppression by ICAO address

#### FR-402: Traffic Position Extrapolation
**Priority**: HIGH | **Verification**: Test

The system SHALL extrapolate traffic positions between updates.

**Acceptance Criteria**:
- Dead reckoning based on last velocity
- Maximum extrapolation time: 10 seconds
- Extrapolation indicated in output
- Stale traffic aged out (>60 seconds)

#### FR-403: Ownship Detection and Filtering
**Priority**: CRITICAL | **Verification**: Test

The system SHALL detect and filter ownship from traffic display.

**Acceptance Criteria**:
- Position-based ownship detection (< 0.01 nm)
- ICAO address-based suppression (if configured)
- Ownship altitude correlation
- False ownship detection prevention

#### FR-404: Relative Position Calculation
**Priority**: HIGH | **Verification**: Test

The system SHALL calculate traffic relative position to ownship.

**Acceptance Criteria**:
- Distance (nm) calculated
- Bearing (°) from ownship calculated
- Relative altitude (ft) calculated
- Vertical rate relative to ownship

#### FR-405: Signal-Based Range Estimation
**Priority**: MEDIUM | **Verification**: Test

For Mode-S traffic without position (FR-405a), the system SHALL estimate range using signal strength.

**Acceptance Criteria**:
- RSSI used for range estimation
- Range estimate accuracy ± 50%
- Estimate marked as "approximate"

#### FR-406: ICAO to Registration Conversion
**Priority**: LOW | **Verification**: Test

The system SHALL convert ICAO addresses to registration numbers where possible.

**Acceptance Criteria**:
- US (N-numbers) conversion
- Canada (C-numbers) conversion
- Australia (VH-numbers) conversion
- Database lookup for other countries

#### FR-407: Traffic Alerting
**Priority**: HIGH | **Verification**: Test

The system SHALL alert on proximate traffic.

**Acceptance Criteria**:
- Alert flag set for traffic < 2 nm horizontal
- Alert flag set for traffic < ±500 ft vertical (if both altitudes valid)
- Alert cleared when traffic exits alert volume
- Alert communicated in GDL90 messages

---

### 3.5 Weather Processing (FR-500 series)

#### FR-501: FIS-B Weather Reception
**Priority**: HIGH | **Verification**: Test | **Traceability**: AC 00-63

The system SHALL receive and decode FIS-B weather products from UAT.

**Acceptance Criteria**:
- NEXRAD (regional and CONUS)
- METAR
- TAF
- Winds Aloft
- PIREPs
- SIGMETs
- AIRMETs
- NOTAMs
- Lightning
- Icing
- Turbulence
- Cloud tops

#### FR-502: Weather Data Management
**Priority**: MEDIUM | **Verification**: Test

The system SHALL manage weather product lifecycle.

**Acceptance Criteria**:
- Products retained for 15 minutes
- Stale products purged automatically
- Product timestamp indicated
- Product type indicated

#### FR-503: Weather Geographic Filtering
**Priority**: LOW | **Verification**: Test

The system SHALL optionally filter weather by geographic region.

**Acceptance Criteria**:
- NEXRAD blocks filtered by region
- Only relevant weather transmitted to clients
- Filter configurable by user

---

### 3.6 GDL90 Output (FR-600 series)

#### FR-601: GDL90 Protocol Compliance
**Priority**: CRITICAL | **Verification**: Test | **Traceability**: GDL90 ICD

The system SHALL output data in GDL90 format per FAA specification.

**Acceptance Criteria**:
- Message framing correct (0x7E delimiters)
- CRC calculation per spec
- Byte stuffing (0x7D escape) implemented
- Message IDs per specification

#### FR-602: GDL90 Heartbeat
**Priority**: CRITICAL | **Verification**: Test

The system SHALL transmit GDL90 heartbeat every 1.0 ±0.1 seconds.

**Acceptance Criteria**:
- Message ID 0x00
- GPS status included
- UAT status included
- Timestamp included

#### FR-603: GDL90 Ownship Report
**Priority**: HIGH | **Verification**: Test

The system SHALL transmit ownship position when GPS valid.

**Acceptance Criteria**:
- Message ID 0x0A (geometric altitude)
- Message ID 0x0B (basic ownship)
- Position accuracy (NACp) indicated
- Velocity included if available

#### FR-604: GDL90 Traffic Report
**Priority**: HIGH | **Verification**: Test

The system SHALL transmit traffic reports for all tracked targets.

**Acceptance Criteria**:
- Message ID 0x14
- Position, altitude, velocity included
- Track direction included
- Callsign/tail number included
- Alert flag set for proximate traffic

#### FR-605: GDL90 Weather Products
**Priority**: HIGH | **Verification**: Test

The system SHALL transmit FIS-B weather in GDL90 format.

**Acceptance Criteria**:
- Message ID 0x63 (FIS-B/APDU)
- Product segmentation for large messages
- Product type indicated
- Timestamp included

#### FR-606: ForeFlight Extensions
**Priority**: MEDIUM | **Verification**: Test

The system SHALL support ForeFlight-specific extensions.

**Acceptance Criteria**:
- Message ID 0x65 (ForeFlight ID)
- AHRS attitude message
- Device capabilities indicated

---

### 3.7 NMEA Output (FR-700 series)

#### FR-701: NMEA GPS Sentences
**Priority**: MEDIUM | **Verification**: Test | **Traceability**: NMEA 0183

The system SHALL output GPS data in NMEA format.

**Acceptance Criteria**:
- $GPRMC (position, velocity, time)
- $GPGGA (position, fix quality, satellites)
- $GPGSA (satellite status, HDOP)
- $GPGSV (satellites in view)
- $PGRMZ (barometric altitude)

#### FR-702: FLARM NMEA Sentences
**Priority**: MEDIUM | **Verification**: Test

The system SHALL output traffic in FLARM NMEA format.

**Acceptance Criteria**:
- $PFLAU (traffic summary, alert status)
- $PFLAA (individual traffic, relative position)
- Traffic prioritized by threat level
- Maximum 10 PFLAA sentences per second

---

### 3.8 Network Distribution (FR-800 series)

#### FR-801: UDP Broadcast
**Priority**: HIGH | **Verification**: Test

The system SHALL broadcast GDL90 data via UDP.

**Acceptance Criteria**:
- Default port 4000
- Additional ports configurable
- Broadcast to 255.255.255.255
- Multicast support (optional)

#### FR-802: TCP Streaming
**Priority**: MEDIUM | **Verification**: Test

The system SHALL support TCP streaming of NMEA data.

**Acceptance Criteria**:
- TCP server on port 2000
- Multiple simultaneous clients
- Automatic client cleanup on disconnect

#### FR-803: Serial Output
**Priority**: MEDIUM | **Verification**: Test

The system SHALL support serial output.

**Acceptance Criteria**:
- Configurable baud rate (9600-921600)
- GDL90 or NMEA format selectable
- Multiple serial devices (/dev/serialout0-9)
- DTR/RTS control

#### FR-804: Bluetooth Output
**Priority**: LOW | **Verification**: Test

The system SHALL optionally support Bluetooth LE output.

**Acceptance Criteria**:
- GATT service for data transfer
- Compatible with SoftRF/Stratux GATT profile
- NMEA or GDL90 over BLE

#### FR-805: Client Sleep Detection
**Priority**: MEDIUM | **Verification**: Test

The system SHALL detect sleeping/inactive clients.

**Acceptance Criteria**:
- ICMP ping to DHCP clients
- Message rate throttling for sleepers
- Wake on high-priority messages (traffic alerts)

---

### 3.9 Web Management Interface (FR-900 series)

#### FR-901: Status Display
**Priority**: HIGH | **Verification**: Inspection

The system SHALL provide web-based status display.

**Acceptance Criteria**:
- GPS status and position
- Traffic count (total, with position, ADS-B vs UAT)
- Weather message count
- Tower locations
- System uptime
- Error indications

#### FR-902: Configuration Management
**Priority**: HIGH | **Verification**: Test

The system SHALL allow web-based configuration.

**Acceptance Criteria**:
- GPS device selection
- AHRS enable/disable
- Region selection (US/EU)
- Network settings (SSID, password)
- Ownship settings (ICAO, tail)

#### FR-903: Live Traffic Display
**Priority**: MEDIUM | **Verification**: Inspection

The system SHALL display live traffic graphically.

**Acceptance Criteria**:
- 2D radar-style display
- Traffic symbols with altitude/callsign
- Ownship at center
- Range rings
- Heading indicator

#### FR-904: Weather Display
**Priority**: MEDIUM | **Verification**: Inspection

The system SHALL display weather products.

**Acceptance Criteria**:
- METAR text display
- TAF text display
- NEXRAD imagery
- Weather age indication

#### FR-905: Diagnostic Logging
**Priority**: HIGH | **Verification**: Test

The system SHALL provide diagnostic logging via web UI.

**Acceptance Criteria**:
- System log viewer
- Log level filtering
- Log download capability
- Log rotation

---

### 3.10 Data Logging (FR-1000 series)

#### FR-1001: SQLite Database Logging
**Priority**: MEDIUM | **Verification**: Test

The system SHALL log traffic and weather to SQLite database.

**Acceptance Criteria**:
- Traffic messages logged
- Weather messages logged
- GPS track logged
- Database rotation at 100 MB
- Database cleanup at 95% disk usage

#### FR-1002: Replay Capability
**Priority**: MEDIUM | **Verification**: Test

The system SHALL support replay of logged data.

**Acceptance Criteria**:
- `--replay <file>` command line option
- Compressed log format
- Timestamped messages
- Playback rate control

#### FR-1003: AHRS CSV Logging
**Priority**: LOW | **Verification**: Test

The system SHALL log AHRS sensor data to CSV.

**Acceptance Criteria**:
- Timestamp, roll, pitch, heading
- Acceleration (X/Y/Z)
- Angular rate (X/Y/Z)
- Temperature, pressure

---

### 3.11 System Management (FR-1100 series)

#### FR-1101: OTA Updates
**Priority**: HIGH | **Verification**: Test

The system SHALL support over-the-air software updates.

**Acceptance Criteria**:
- Update via web UI
- Update via command line
- Version verification
- Automatic rollback on failure
- Update progress indication

#### FR-1102: Configuration Persistence
**Priority**: HIGH | **Verification**: Test

The system SHALL persist configuration across reboots.

**Acceptance Criteria**:
- Settings stored in `/boot/firmware/stratux.conf`
- JSON format
- Automatic backup on write
- Factory reset capability

#### FR-1103: Error Handling and Recovery
**Priority**: CRITICAL | **Verification**: Test

The system SHALL handle errors gracefully.

**Acceptance Criteria**:
- Error tracking system (globalStatus.Errors)
- LED indication of errors
- Automatic reconnection to failed devices
- Watchdog for hang detection
- Graceful degradation

#### FR-1104: System Health Monitoring
**Priority**: HIGH | **Verification**: Test

The system SHALL monitor system health.

**Acceptance Criteria**:
- CPU temperature monitoring
- Disk usage monitoring
- Memory usage monitoring
- Network connectivity
- Service status (stratux, fancontrol)

---

## 4. Non-Functional Requirements

### 4.1 Performance (NFR-100 series)

#### NFR-101: Boot Time
**Priority**: MEDIUM | **Verification**: Test

The system SHALL boot and be operational within 120 seconds.

#### NFR-102: Message Latency
**Priority**: HIGH | **Verification**: Test

The system SHALL output messages with ≤ 100ms latency from reception.

#### NFR-103: Client Capacity
**Priority**: MEDIUM | **Verification**: Test

The system SHALL support ≥ 5 simultaneous clients without degradation.

#### NFR-104: Message Throughput
**Priority**: HIGH | **Verification**: Test

The system SHALL handle ≥ 500 ADS-B messages/second without loss.

#### NFR-105: CPU Usage
**Priority**: LOW | **Verification**: Test

The system SHALL use ≤ 80% CPU under normal conditions.

---

### 4.2 Reliability (NFR-200 series)

#### NFR-201: Continuous Operation
**Priority**: HIGH | **Verification**: Test

The system SHALL operate continuously for ≥ 8 hours without failure.

#### NFR-202: Error Recovery
**Priority**: HIGH | **Verification**: Test

The system SHALL recover from transient errors without user intervention.

#### NFR-203: Data Integrity
**Priority**: CRITICAL | **Verification**: Test

The system SHALL validate all received data before use.

#### NFR-204: Fail-Safe Behavior
**Priority**: CRITICAL | **Verification**: Test

The system SHALL fail to safe state (no data) rather than incorrect data.

---

### 4.3 Maintainability (NFR-300 series)

#### NFR-301: Modularity
**Priority**: MEDIUM | **Verification**: Inspection

The system code SHALL be modular with clear interfaces.

#### NFR-302: Documentation
**Priority**: HIGH | **Verification**: Inspection

All modules SHALL have documentation describing purpose and interfaces.

#### NFR-303: Logging
**Priority**: HIGH | **Verification**: Inspection

The system SHALL log significant events with timestamps and severity.

---

### 4.4 Portability (NFR-400 series)

#### NFR-401: Hardware Platform
**Priority**: HIGH | **Verification**: Test

The system SHALL run on Raspberry Pi 3B, 3B+, 4, 5, Zero 2 W.

#### NFR-402: Operating System
**Priority**: HIGH | **Verification**: Test

The system SHALL run on Debian Bookworm (ARM64).

#### NFR-403: Storage
**Priority**: HIGH | **Verification**: Test

The system SHALL fit on ≥ 4 GB microSD card (recommend 8 GB).

---

### 4.5 Security (NFR-500 series)

#### NFR-501: Authentication
**Priority**: MEDIUM | **Verification**: Test

The web management interface SHALL support authentication (optional).

#### NFR-502: Input Validation
**Priority**: CRITICAL | **Verification**: Test

All user inputs SHALL be validated before use.

#### NFR-503: Secure Updates
**Priority**: HIGH | **Verification**: Test

OTA updates SHALL be verified before installation.

#### NFR-504: Credential Protection
**Priority**: HIGH | **Verification**: Inspection

WiFi passwords SHALL be stored encrypted or protected.

---

## 5. Requirements Traceability

### 5.1 High-Level Requirements to Detailed Requirements

| High-Level | Detailed Requirements | Priority |
|------------|----------------------|----------|
| Receive ADS-B | FR-101, FR-105 | HIGH |
| Receive UAT | FR-102, FR-105, FR-501 | HIGH |
| Receive OGN | FR-103, FR-105 | MEDIUM |
| GPS Position | FR-201-205 | CRITICAL |
| AHRS Attitude | FR-301-305 | MEDIUM |
| Traffic Fusion | FR-401-407 | HIGH |
| Weather Processing | FR-501-503 | HIGH |
| GDL90 Output | FR-601-606 | CRITICAL |
| NMEA Output | FR-701-702 | MEDIUM |
| Network Distribution | FR-801-805 | HIGH |
| Web Interface | FR-901-905 | HIGH |
| Data Logging | FR-1001-1003 | MEDIUM |
| System Management | FR-1101-1104 | HIGH |

### 5.2 Requirements to Test Cases

See `TEST_PLAN.md` for test case traceability (to be created).

---

## 6. Requirements Verification Matrix

| Requirement ID | Verification Method | Test Procedure | Status |
|----------------|--------------------|--------------|---------|
| FR-101 | Test | TP-101 | ❌ Not Started |
| FR-102 | Test | TP-102 | ❌ Not Started |
| ... | ... | ... | ... |

*(Full matrix to be completed)*

---

## 7. Open Issues and TBDs

1. **TBD-001**: Specific performance targets for message latency need validation against real-world usage
2. **TBD-002**: Security requirements (authentication, HTTPS) need stakeholder input
3. **TBD-003**: Test coverage targets need to be established (recommend 80% for SAL-3)
4. **TBD-004**: Formal hazard analysis needs to be performed per DO-278A
5. **TBD-005**: Requirements priority needs review by subject matter experts

---

## 8. Document Control

### 8.1 Change History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1 | 2025-10-13 | System Analysis | Initial draft from reverse engineering |
| 1.0 DRAFT | 2025-10-13 | Requirements Engineering | First complete draft for review |

### 8.2 Review and Approval

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Author | TBD | | |
| Technical Review | TBD | | |
| Safety Review | TBD | | |
| Approval | TBD | | |

---

## 9. References

1. RTCA DO-278A: Software Integrity Assurance for CNS/ATM Systems
2. RTCA DO-260B: Minimum Operational Performance Standards for 1090 MHz Extended Squitter ADS-B
3. RTCA DO-282B: Minimum Operational Performance Standards for UAT ADS-B
4. FAA GDL90 Public Interface Control Document
5. NMEA 0183 Standard for Interfacing Marine Electronic Devices
6. Stratux Code Repository: https://github.com/cyoung/stratux

---

**END OF DOCUMENT**

**Next Steps**:
1. Review and approve this requirements document
2. Create detailed test plan (TEST_PLAN.md)
3. Establish requirements traceability matrix
4. Begin unit test development
5. Implement coverage measurement

**Total Requirements**: 84 functional, 17 non-functional = **101 requirements**
**Estimated Test Cases**: ~250-300 (including unit, integration, system tests)
