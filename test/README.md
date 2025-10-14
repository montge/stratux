# Stratux Test Utilities

This directory contains manual testing and debugging utilities for Stratux development. These are **NOT automated tests** - they are standalone programs for debugging, analysis, and hardware configuration.

For automated test coverage, see `/main/*_test.go` and [TEST_COVERAGE_ROADMAP.md](../TEST_COVERAGE_ROADMAP.md).

---

## Status: Phase 3.1 - Legacy Test Audit

**Objective:** Document all utilities, identify which are still useful, and plan migration to automated integration tests where applicable.

**Date:** 2025-10-14

---

## Utilities by Category

### 🎯 Integration Test Candidates (High Priority for Automation)

#### replay.go
**Purpose:** Replays recorded UAT and 1090ES traffic logs with timing simulation.

**Usage:**
```bash
go run replay.go <uat_replay_log> <es_replay_log> [speed_multiplier]
```

**Description:**
- Replays UAT messages to stdout
- Replays 1090ES messages via TCP socket on port 30003 (dump1090 format)
- Supports speed multipliers for fast-forward replay
- Skips gaps > 2 minutes automatically

**File Format:**
- CSV format: `timestamp_ns,message_data`
- Special `START` marker to reset timing

**Integration Test Potential:** ⭐⭐⭐⭐⭐ **EXCELLENT**
- **Action:** Convert to trace replay framework (similar to OGN/APRS tests)
- **Target:** Phase 3.3 - UAT integration tests, Phase 3.4 - 1090ES integration tests
- **Benefit:** Hardware-independent integration testing

**Recommendation:** **MIGRATE TO AUTOMATED TESTS** - This is the perfect candidate for automated integration testing using the trace replay methodology.

---

### 📊 Analysis and Debugging Tools (Keep for Development)

#### maxgap.go
**Purpose:** Analyzes METAR Quality of Service (QoS) by tracking update intervals.

**Usage:**
```bash
go run maxgap.go <replay_log>
```

**Description:**
- Tracks METAR updates in 5-minute windows
- Calculates QoS metric (average updates per airport)
- Generates `qos.png` graph showing FIS-B quality over time
- Based on AC 00-45G FIS-B transmission intervals

**Dependencies:**
- `gonum.org/v1/plot` (plotting library)

**Integration Test Potential:** ⭐⭐ LOW
- Analysis tool, not a test scenario

**Recommendation:** **KEEP** - Useful for FIS-B quality analysis and debugging.

---

#### packetrate.go
**Purpose:** Analyzes packet rate from replay logs and generates graphs.

**Usage:**
```bash
go run packetrate.go <replay_log>
```

**Description:**
- Tracks message rate in 1-minute windows
- Generates packet rate graphs
- Useful for diagnosing reception issues

**Dependencies:**
- `gonum.org/v1/plot`

**Integration Test Potential:** ⭐⭐ LOW
- Analysis tool

**Recommendation:** **KEEP** - Useful for performance analysis.

---

#### icao2reg.go
**Purpose:** Converts ICAO 24-bit addresses to aircraft tail numbers (US/Canada).

**Usage:**
```bash
go run icao2reg.go [ICAO_hex_code]
```

**Example:**
```bash
go run icao2reg.go A00001    # Outputs: N1
go run icao2reg.go AC82EC    # Example N-number
```

**Description:**
- Decodes US N-numbers (0xA00001-0xADF7C7)
- Decodes Canadian C-registrations (0xC00001-0xC0CDF8)
- Identifies military/non-civil registrations

**Integration Test Potential:** ⭐ NONE
- **NOTE:** This function is already 100% tested in `main/traffic_test.go`!
- See: `TestIcao2reg_USCivil`, `TestIcao2reg_Canada`, `TestIcao2reg_Australia`

**Recommendation:** **KEEP** - Useful CLI tool for quick ICAO lookups, but testing already complete.

---

#### nexrad_annunciator.go
**Purpose:** Monitors NEXRAD weather radar blocks and warns when weather is nearby.

**Usage:**
```bash
go run nexrad_annunciator.go <lat> <lon> < uat_log
```

**Description:**
- Parses UAT NEXRAD frames
- Calculates distance from ownship position to weather blocks
- Warns if weather within 10 nm (18.52 km)
- Uses block location calculations (48' x 4' grid)

**Integration Test Potential:** ⭐⭐⭐ MEDIUM
- Could test block location calculations
- Weather proximity alerting logic

**Recommendation:** **KEEP** - Useful for weather testing, consider automated tests for block_location() (already at 100% coverage).

---

### 🔍 Data Extraction Tools (Keep for Debugging)

#### es_dump_csv.go
**Purpose:** Exports 1090ES messages from SQLite database to CSV format.

**Usage:**
```bash
go run es_dump_csv.go <sqlite_file> > output.csv
```

**Description:**
- Reads `es_messages` table from SQLite
- Extracts all dump1090Data fields to CSV
- 23 columns: ICAO, DF, CA, TypeCode, position, altitude, etc.

**Dependencies:**
- `github.com/mattn/go-sqlite3`

**Integration Test Potential:** ⭐ NONE
- Data export utility

**Recommendation:** **KEEP** - Useful for analyzing logged 1090ES data.

---

#### getairmet.go
**Purpose:** Extracts AIRMET/NOTAM messages from UAT and groups by geographic location.

**Usage:**
```bash
# From stdin:
cat uat_log.txt | go run getairmet.go -stdin

# Single message:
go run getairmet.go
```

**Description:**
- Parses UAT AIRMET/NOTAM frames
- Groups geographic points by GeoHash (precision 5)
- Outputs JSON with report number, location, start/end times
- Useful for debugging weather message decoding

**Dependencies:**
- `github.com/gansidui/geohash`

**Integration Test Potential:** ⭐⭐⭐ MEDIUM
- Could test AIRMET decoding logic

**Recommendation:** **KEEP** - Useful for AIRMET/NOTAM debugging.

---

#### extract_latlng.go
**Purpose:** Extracts latitude/longitude from raw UAT uplink frames.

**Usage:**
```bash
cat uat_log.txt | go run extract_latlng.go
```

**Description:**
- Decodes first 6 bytes of UAT uplink frames
- Extracts 24-bit lat/lon coordinates
- Outputs decimal degrees
- Simple coordinate parsing verification tool

**Integration Test Potential:** ⭐ NONE
- Basic debugging tool

**Recommendation:** **KEEP** - Quick coordinate extraction for debugging.

---

#### extract_metar.go
**Purpose:** Extracts METAR text reports from UAT messages.

**Usage:**
```bash
cat uat_log.txt | go run extract_metar.go
```

**Description:**
- Decodes UAT uplink frames
- Extracts DLAC-encoded text (METAR format)
- Uses DLAC alphabet for text decoding
- Outputs raw METAR strings

**Integration Test Potential:** ⭐⭐ LOW
- Text extraction debugging

**Recommendation:** **KEEP** - Useful for METAR debugging.

---

#### uatsummary.go
**Purpose:** Summarizes UAT message statistics from replay logs.

**Usage:**
```bash
go run uatsummary.go <uat_log>
```

**Description:**
- Parses UAT messages
- Counts message types (traffic, weather, etc.)
- Provides quick statistics

**Integration Test Potential:** ⭐ NONE
- Summary tool

**Recommendation:** **KEEP** - Quick UAT log analysis.

---

### 🔧 Hardware Configuration (Keep for Setup)

#### mtk_config.sh
**Purpose:** Configures MTK3339 GPS receivers for Stratux.

**Usage:**
```bash
sudo ./mtk_config.sh
```

**Description:**
- Resets MTK3339 to 9600 baud
- Enables WAAS (Wide Area Augmentation System)
- Sets 5 Hz position reporting
- Enables required NMEA sentences
- Increases to 38400 baud for high-rate output

**Configuration:**
- Serial port: `/dev/ttyAMA0`
- Final baud: 38400
- Update rate: 5 Hz

**Integration Test Potential:** ⭐ NONE
- Hardware setup script

**Recommendation:** **KEEP** - Essential for MTK3339 GPS setup.

---

#### sirf_config.sh
**Purpose:** Configures BU-353-S4 (SiRF chipset) GPS receivers for Stratux.

**Usage:**
```bash
sudo ./sirf_config.sh
```

**Description:**
- Resets SiRF receiver to 4800 baud
- Enables WAAS
- Sets 5 Hz position reporting
- Enables required NMEA sentences
- Increases to 38400 baud

**Configuration:**
- Serial port: `/dev/ttyUSB0`
- Final baud: 38400
- Update rate: 5 Hz

**Integration Test Potential:** ⭐ NONE
- Hardware setup script

**Recommendation:** **KEEP** - Essential for SiRF GPS setup.

---

#### extract_gps.sh
**Purpose:** Extracts GPS PUBX sentences from replay logs.

**Usage:**
```bash
./extract_gps.sh <replay_log>
```

**Description:**
- Filters out START/PAUSE/UNPAUSE markers
- Extracts PUBX,00 sentences (position data)
- Outputs selected fields: timestamp, lat, lon, alt, speed, track
- Quick GPS data extraction

**Integration Test Potential:** ⭐ NONE
- Text processing script

**Recommendation:** **KEEP** - Quick GPS data extraction.

---

### 🚧 Hardware-Dependent (Requires Physical Devices)

#### uat_read.go
**Purpose:** Reads UAT data directly from RTL-SDR hardware.

**Usage:**
```bash
go run uat_read.go
```

**Description:**
- Directly interfaces with RTL-SDR dongle
- Uses godump978 for UAT decoding
- Real-time UAT reception testing
- Requires physical SDR hardware

**Dependencies:**
- `github.com/jpoirier/gortlsdr` (RTL-SDR library)
- `github.com/stratux/stratux/godump978` (UAT decoder)
- **Build tag:** `//go:build ignore` (doesn't build by default)

**Integration Test Potential:** ⭐⭐⭐⭐ HIGH (for Phase 4)
- Perfect candidate for Phase 4.3 - SDR Integration Tests
- Requires physical hardware and recorded IQ files

**Recommendation:** **KEEP** - Essential for Phase 4 SDR testing, but requires hardware.

---

#### bmp180_read.go
**Purpose:** Reads barometric pressure sensor (BMP180) via I2C.

**Usage:**
```bash
go run bmp180_read.go
```

**Description:**
- Reads BMP180 temperature and pressure
- Calculates altitude from pressure
- Converts to feet
- Tests I2C communication

**Dependencies:**
- `github.com/kidoman/embd` (embedded device I/O library)
- Physical BMP180 sensor on I2C bus

**Integration Test Potential:** ⭐⭐⭐⭐ HIGH (for Phase 4)
- Phase 4.2 - Hardware Sensor Integration Tests
- Barometric pressure sensor validation

**Recommendation:** **KEEP** - Essential for Phase 4 sensor testing, but requires hardware.

---

#### real_read_bmp388.go
**Purpose:** Reads barometric pressure sensor (BMP388) via I2C.

**Usage:**
```bash
go run real_read_bmp388.go
```

**Description:**
- Reads BMP388 temperature and pressure
- Outputs CSV format: `pressure,temperature`
- 1 ms update rate (1000 Hz)
- Tests newer BMP388 sensor

**Dependencies:**
- `github.com/stratux/stratux/sensors/bmp388`
- `github.com/kidoman/embd`
- Physical BMP388 sensor on I2C bus

**Integration Test Potential:** ⭐⭐⭐⭐ HIGH (for Phase 4)
- Phase 4.2 - Hardware Sensor Integration Tests

**Recommendation:** **KEEP** - Essential for Phase 4 sensor testing, but requires hardware.

---

### 🗂️ Third-Party Libraries

#### metar-to-text/
**Purpose:** C library for parsing METAR weather reports to human-readable text.

**Description:**
- mdsplib METAR decoder library
- Not used in current Stratux codebase
- Appears to be legacy code

**Integration Test Potential:** ⭐ NONE
- External library

**Recommendation:** **EVALUATE** - Check if this is actually used. If not, consider removing to reduce repository size.

---

## Test Data Migration Plan

### Phase 3.2: Convert to Automated Integration Tests

Based on this audit, the following utilities should be converted to automated integration tests:

#### 🎯 High Priority (Phase 3.3-3.6)

1. **replay.go → Integration Test Framework**
   - **Target:** Phase 3.3 (UAT), Phase 3.4 (1090ES)
   - **Method:** Convert replay logs to `.trace.gz` format
   - **Location:** `/main/testdata/uat/`, `/main/testdata/1090es/`
   - **Expected Coverage:** +5-10% main package
   - **Effort:** Medium (2-3 weeks)

2. **getairmet.go → AIRMET/NOTAM Integration Tests**
   - **Target:** Phase 3.3 (UAT Integration Tests)
   - **Method:** Create trace files with AIRMET messages
   - **Location:** `/main/testdata/weather/`
   - **Expected Coverage:** +2-3% for decodeAirmet()
   - **Effort:** Low (1 week)

3. **nexrad_annunciator.go → NEXRAD Integration Tests**
   - **Target:** Phase 3.3 (UAT Integration Tests)
   - **Method:** Test NEXRAD block location and proximity calculations
   - **Note:** block_location() already at 100% coverage
   - **Expected Coverage:** +1-2% for decodeNexradFrame()
   - **Effort:** Low (1 week)

#### 🔄 Medium Priority (Phase 4)

4. **uat_read.go → SDR Integration Tests**
   - **Target:** Phase 4.3 (SDR Integration Tests)
   - **Method:** Use recorded IQ files for repeatable tests
   - **Requires:** Physical Pi or CGO_ENABLED=1 build environment
   - **Expected Coverage:** SDR reception validation (not code coverage)
   - **Effort:** High (2-3 weeks)

5. **bmp180_read.go, real_read_bmp388.go → Sensor Integration Tests**
   - **Target:** Phase 4.2 (Hardware Sensor Integration Tests)
   - **Method:** On-device test suite
   - **Requires:** Physical Raspberry Pi with sensors
   - **Expected Coverage:** Sensor communication validation
   - **Effort:** Medium (1-2 weeks)

---

## Utilities to Keep

These utilities remain valuable for development and debugging:

### Keep (Essential for Development)
- ✅ **replay.go** - After migration, keep as manual testing tool
- ✅ **maxgap.go** - FIS-B quality analysis
- ✅ **packetrate.go** - Performance analysis
- ✅ **icao2reg.go** - Quick ICAO lookup (tests already complete)
- ✅ **nexrad_annunciator.go** - Weather proximity testing
- ✅ **es_dump_csv.go** - Data export
- ✅ **getairmet.go** - AIRMET debugging
- ✅ **extract_latlng.go** - Coordinate debugging
- ✅ **extract_metar.go** - METAR debugging
- ✅ **uatsummary.go** - Quick stats
- ✅ **mtk_config.sh** - GPS setup
- ✅ **sirf_config.sh** - GPS setup
- ✅ **extract_gps.sh** - GPS data extraction
- ✅ **uat_read.go** - Real-time UAT testing
- ✅ **bmp180_read.go** - BMP180 sensor testing
- ✅ **real_read_bmp388.go** - BMP388 sensor testing

### Evaluate for Removal
- ⚠️ **metar-to-text/** - Check if used, consider removing if legacy

---

## Next Steps (Phase 3.2)

1. ✅ Complete `/test/` audit (this document)
2. 🚧 Audit `/test-data/` directory
3. 🚧 Create trace file generators for replay logs
4. 🚧 Implement UAT integration tests (Phase 3.3)
5. 🚧 Implement 1090ES integration tests (Phase 3.4)
6. 🚧 Implement GPS integration tests (Phase 3.5)
7. 🚧 Implement end-to-end integration tests (Phase 3.6)

---

## Building and Running Utilities

### Requirements
```bash
# Install dependencies
go get gonum.org/v1/plot
go get github.com/mattn/go-sqlite3
go get github.com/gansidui/geohash
go get github.com/kidoman/embd
```

### Hardware-Dependent Builds
```bash
# For RTL-SDR utilities (requires CGO and librtlsdr):
CGO_ENABLED=1 go build uat_read.go

# For sensor utilities (requires I2C on Raspberry Pi):
GOARCH=arm64 GOOS=linux CGO_ENABLED=1 go build bmp180_read.go
```

### Standard Builds
```bash
# Most utilities build without CGO:
go build replay.go
go build maxgap.go
go build icao2reg.go
# etc.
```

---

## References

- **Test Coverage Roadmap:** [TEST_COVERAGE_ROADMAP.md](../TEST_COVERAGE_ROADMAP.md)
- **Coverage Summary:** [coverage_summary.md](../coverage_summary.md)
- **Integration Test Examples:** `/main/integration_*_test.go`
- **Trace File Format:** See `/main/testdata/ogn/generate_trace.go`

---

**Last Updated:** 2025-10-14
**Phase:** 3.1 - Legacy Test Audit Complete
**Next Phase:** 3.2 - Test Data Migration Planning
