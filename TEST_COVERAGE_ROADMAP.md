# Stratux Test Coverage Roadmap

## Overview
This roadmap outlines the systematic approach to achieving comprehensive test coverage for the Stratux project, progressing from unit tests of pure functions to full system integration testing on physical hardware.

**Current Status:** Phase 2.4 Complete
**Current Coverage:** 18.9%
**Target Coverage:** 80% (long-term goal)

---

## Phase 1: Test Infrastructure ✅ COMPLETE

**Goal:** Establish test infrastructure and achieve coverage for packages without hardware dependencies.

### Completed Work:
- ✅ Common package: 90.2% coverage (45+ test functions, 715 lines)
- ✅ UATparse package: 29.7% coverage (17 test functions, 670 lines)
- ✅ Test framework setup with `go test` and coverage reporting
- ✅ CI/CD integration with GitHub Actions
- ✅ Coverage reporting and artifact generation

### Functions at 100% Coverage (21 total):
**Common Package (16):**
- LinReg, LinRegWeighted, Mean, Stdev
- ArrayMin, ArrayMax, Radians, Degrees
- Distance, RoundToInt16, CalcAltitude
- IMin, IMax, IsRunningAsRoot, IsCPUTempValid

**UATparse Package (4):**
- formatDLACData, airmetParseDate, airmetLatLng, block_location

**Main Package (1):**
- getProductNameFromId

### Test Files Created:
- `common/helpers_test.go` (715 lines)
- `uatparse/uatparse_test.go` (670 lines)
- `main/gen_gdl90_test.go` (110 lines added)

---

## Phase 2: Protocol Parser Integration Tests

**Goal:** Add integration tests for protocol parsers using trace file replay methodology.

### Phase 2.1: UAT Parser Tests ✅ COMPLETE
- ✅ UAT message parsing and decoding
- ✅ Uplink message handling
- ✅ Text report extraction

### Phase 2.2: GPS/NMEA Parser Tests ✅ COMPLETE
**Test Files:** 414 lines in `main/gps_test.go`

**Functions at 100% Coverage (5):**
- chksumUBX() - UBX protocol Fletcher-like checksum
- makeUBXCFG() - UBX message construction with sync chars
- makeNMEACmd() - NMEA command construction with XOR checksum
- validateNMEAChecksum() - NMEA sentence validation
- calculateNACp() - Navigation Accuracy Category calculation

**Test Coverage:**
- 7 test cases for UBX checksum (empty, single/multiple bytes, edge cases)
- 3 test cases for UBX message construction (payload sizes, length encoding)
- 3 test cases for NMEA command construction (framing, checksum)
- 9 test cases for NMEA validation (valid sentences, invalid checksums, missing delimiters)
- 32 test cases for NACp calculation (all levels, precise boundaries)

### Phase 2.3: FLARM/OGN NMEA Utilities ✅ COMPLETE
**Test Files:** 672 lines in `main/flarm-nmea_test.go`

**Functions Tested (6):**
- appendNmeaChecksum() - NMEA checksum calculation (XOR-based)
- computeAlarmLevel() - FLARM collision alarm levels
- gdl90EmitterCatToNMEA() - GDL90 to NMEA aircraft type conversion
- nmeaAircraftTypeToGdl90() - Reverse conversion
- atof32() - String to float32 with error handling
- getIdTail() - OGN ID and tail parsing

**Test Coverage:**
- 6 test cases for NMEA checksum (various sentence types, edge cases)
- 21 test cases for alarm level calculation (boundaries, vertical separation)
- 33 test cases for aircraft type conversion (all emitter categories, bidirectional)
- 13 test cases for atof32 (integers, decimals, scientific notation, errors)
- 17 test cases for OGN ID parsing (ID/tail formats, truncation, hex decoding)

### Phase 2.4: OGN/APRS Parser Integration Tests ✅ COMPLETE
**Test Files:** 645 lines in `main/integration_ogn_aprs_test.go`

**Coverage Achieved:** 18.9% main package (+2.2% from 16.7%)

**Functions Tested:**
- parseOgnMessage(): 100% coverage
- importOgnTrafficMessage(): 82.8% coverage
- importOgnStatusMessage(): 80% coverage
- parseAprsMessage(): 72.9% coverage

**Trace Files Created:**
- `main/testdata/ogn/basic_ogn.trace.gz` (11 messages: 2 status + 9 traffic)
- `main/testdata/aprs/basic_aprs.trace.gz` (12 messages, various protocols)
- `main/testdata/ogn/generate_trace.go` (87 lines)
- `main/testdata/aprs/generate_trace.go` (91 lines)

**Test Functions (22):**
- TestOGNBasicParsing - Message counter, status exclusion
- TestOGNStatusMessage - Background noise, gain, TX status parsing
- TestOGNTrafficParsing - Position, altitude, track, speed, climb rate
- TestOGNAddressTypes - ICAO, FLARM, Skylines, PAW address handling
- TestOGNAircraftTypes - Glider, powered, helicopter, etc.
- TestOGNRegistrationUpdate - Registration-only messages
- TestOGNSignalStrength - SNR and DOP parsing
- TestOGNInvalidMessages - Malformed JSON, missing fields
- TestOGNAltitudeConversion - MSL vs HAE altitude
- TestOGNEmitterCategory - Hex acft_cat parsing
- TestOGNHardwareType - Stratux hardware identification
- TestOGNZeroValues - Zero altitude, speed, climb handling
- TestOGNNegativeValues - Negative climb rates
- TestAPRSBasicParsing - Message counter, ground station filtering
- TestAPRSMessageParsing - Coordinate parsing (degrees-minutes to decimal)
- TestAPRSProtocolTypes - FLR, OGN, ICA, SKY, PAW, FAN protocols
- TestAPRSInvalidMessages - Malformed messages, missing fields
- TestAPRSOptionalFields - Messages without track/speed/altitude
- TestAPRSCoordinateFormats - Various precision levels
- TestAPRSMinimalMessages - Bare minimum required fields
- TestOGNAPRSTrafficSource - Verify MSGCLASS_OGN (2) assignment
- TestOGNAPRSStateMixing - State isolation between tests

**Methodology:**
- Trace file replay: Gzipped CSV format (timestamp, protocol, message)
- Hardware-independent testing
- State isolation with resetOGNAPRSState()
- Fake GPS position for distance checks

### Phase 2.5: Datalog and Output Format Tests ✅ COMPLETE
**Test Files:**
- `main/datalog_test.go` (429 lines)
- `main/xplane_test.go` (459 lines)

**Datalog Functions (8):**
- boolMarshal() - Boolean to SQL INTEGER conversion
- intMarshal() - All signed integer types to SQL
- uintMarshal() - All unsigned integer types to SQL
- floatMarshal() - Float to SQL with precision
- stringMarshal() - String passthrough
- notsupportedMarshal() - Unsupported types handling
- structCanBeMarshalled() - Reflection method detection
- structMarshal() - Struct to SQL via String()

**X-Plane Functions (4):**
- convertKnotsToXPlaneSpeed() - Knots to m/s conversion
- createXPlaneGpsMsg() - GPS position messages
- createXPlaneAttitudeMsg() - Attitude messages
- createXPlaneTrafficMsg() - Traffic messages with callsign cleaning

**Test Coverage:**
- 9 test cases for integer types (all signed integer types, max values)
- 7 test cases for unsigned integer types (all unsigned types, max values)
- 6 test cases for float types (positive/negative, scientific notation)
- 21 test cases for X-Plane message formatting
- 6 test cases for callsign cleaning (special characters, case preservation)

### Phase 2.6: Traffic Processing Tests ✅ COMPLETE
**Test Files:** `main/traffic_test.go` (extended with targeted tests)

**Additional Tests (7):**
- TestEstimateDistance_LearningPositiveError - Learning algorithm when underestimating
- TestComputeTrafficPriority_NoBaroAlt - GPS altitude fallback
- TestIsOwnshipTrafficInfo_OGNTrackerWithValidGPS - OGN tracker with GPS validation
- TestIsOwnshipTrafficInfo_NoAltitudeVerification - Altitude verification disabled
- TestMakeTrafficReportMsg_GNSSAltitude - GNSS altitude conversion
- TestMakeTrafficReportMsg_OutOfBoundsAltitude - Altitude boundary encoding
- TestMakeTrafficReportMsg_OnGroundFlag - On-ground flag encoding

### Phase 2.7: Aircraft Type Mapping ✅ COMPLETE
**Test Files:** `main/tracker_test.go` (376 lines)

**Functions (1):**
- mapAircraftType() - Bidirectional type mapping for OGN trackers

**Test Coverage:**
- 10 test cases for bidirectional mapping
- 2 test cases for empty table handling
- 4 test cases for single entry scenarios
- 2 test cases for duplicate handling
- 4 test cases for negative values
- Property-based symmetry testing
- 4 test cases for zero value distinction

### Phase 2.8: MessageQueue Data Structure ✅ COMPLETE
**Test Files:** `main/messagequeue_test.go` (533 lines)

**Tests Created (14):**
- TestNewMessageQueue - Constructor and initialization
- TestMessageQueuePutAndPeek - Non-destructive reads
- TestMessageQueuePutAndPop - Destructive reads
- TestMessageQueuePriorityOrdering - Lowest priority first
- TestMessageQueueEmptyQueue - Empty state handling
- TestMessageQueueSamePriorityFIFO - FIFO within same priority
- TestMessageQueuePruning - Automatic size limit enforcement
- TestMessageQueueGetQueueDump - Queue inspection
- TestMessageQueueClose - Graceful shutdown
- TestMessageQueueExpiredEntries - Time-based expiration
- TestMessageQueueFindInsertPosition - Binary search insertion
- TestMessageQueueMixedPriorities - Complex scenarios
- TestMessageQueueGetQueueDumpWithPrune - Forced pruning
- TestMessageQueueDataAvailableChannel - Channel notifications

**Status:** Tests ready but blocked by C library dependencies (requires CGO_ENABLED=1)

---

## Phase 3: Legacy Test Migration 🚧 IN PROGRESS

**Goal:** Audit, convert, and replace manual testing utilities with automated integration tests.

### Phase 3.1: Audit `/test/` Directory
**Objective:** Identify which utilities are still useful vs obsolete.

**Tasks:**
- [ ] Catalog all programs in `/test/` directory
- [ ] Document purpose and usage of each utility
- [ ] Identify which provide unique functionality for debugging
- [ ] Mark obsolete utilities for removal
- [ ] Document recommended utilities in CLAUDE.md

**Test Directory Contents (Known):**
```
/test/
├── Various debugging utilities
└── Manual testing programs
```

**Deliverables:**
- `/test/README.md` - Documentation of utilities
- List of utilities to keep vs remove
- Integration test candidates

### Phase 3.2: Audit `/test-data/` Directory
**Objective:** Convert manual test scenarios to automated trace replay tests.

**Tasks:**
- [ ] Catalog all files in `/test-data/` directory
- [ ] Identify what protocol/scenario each file represents
- [ ] Extract representative test cases
- [ ] Convert to gzipped trace files in `/main/testdata/`
- [ ] Create integration tests using trace replay methodology
- [ ] Remove obsolete test-data files after conversion

**Test Data Migration Plan:**
```
/test-data/              →  /main/testdata/
├── raw_logs/            →  ├── protocol_name/
│   ├── uat_*.log       →  │   ├── scenario_name.trace.gz
│   ├── 1090es_*.log    →  │   ├── generate_trace.go
│   └── gps_*.log       →  │   └── README.md
└── ...                  →  └── ...
```

**Integration Test Template:**
1. Generate trace file from old test data
2. Create replay function (like `replayOGNTraceDirect()`)
3. Add state reset function
4. Create test functions for scenarios
5. Verify expected behavior

**Deliverables:**
- Automated integration tests for each scenario
- Documentation in testdata subdirectories
- Removal of obsolete test-data files
- Updated CI to run new integration tests

### Phase 3.3: UAT Integration Tests (from test-data)
**Priority:** HIGH (UAT is core functionality)

**Expected Coverage:**
- [ ] UAT uplink message scenarios
- [ ] UAT weather decoding (NEXRAD, METAR, TAF)
- [ ] UAT traffic messages
- [ ] Multi-message sequences
- [ ] Error/malformed message handling

**Target Functions:**
- uatparse.DecodeUplink()
- parseUplinkBlock()
- decodeNexradFrame()
- decodeAirmet()

**Estimated Impact:** +5-10% main package coverage

### Phase 3.4: 1090ES Integration Tests (from test-data)
**Priority:** HIGH (1090ES is core functionality)

**Expected Coverage:**
- [ ] Mode-S message decoding
- [ ] ADS-B position messages
- [ ] Velocity messages
- [ ] Identification messages
- [ ] Multi-target scenarios

**Target Functions:**
- parseInput() for 1090ES
- decodeModeS()
- esListen()

**Estimated Impact:** +3-5% main package coverage

### Phase 3.5: GPS Integration Tests (from test-data)
**Priority:** MEDIUM

**Expected Coverage:**
- [ ] NMEA sentence sequences (GPRMC, GPGGA, GPGSA, etc.)
- [ ] UBX protocol messages
- [ ] GPS fix quality transitions
- [ ] Satellite tracking
- [ ] Multi-constellation GNSS

**Target Functions:**
- processNMEALine()
- parseGPRMC(), parseGPGGA(), parseGPGSA()
- processUBXMessage()

**Estimated Impact:** +2-4% main package coverage

### Phase 3.6: Complete Integration Scenarios
**Priority:** MEDIUM

**Expected Coverage:**
- [ ] End-to-end data flow: SDR → Parser → GDL90 → Network
- [ ] Multi-source traffic fusion (UAT + 1090ES + OGN)
- [ ] Ownship detection with various GPS types
- [ ] Traffic extrapolation over time
- [ ] Network output formatting (GDL90, NMEA, FLARM)

**Methodology:**
- Use trace files with multiple protocols
- Simulate time progression
- Verify output messages
- Check state consistency

**Estimated Impact:** +5-8% main package coverage

---

## Phase 4: Physical Device Testing Suite 🔮 FUTURE

**Goal:** Create a comprehensive testing suite that runs on actual Raspberry Pi hardware to validate sensor integration and real-world behavior.

### Phase 4.1: Development Mode Test Framework
**Objective:** Build test harness that runs on physical devices.

**Requirements:**
- [ ] Test mode flag in stratux.conf
- [ ] Development mode in web interface
- [ ] Test result reporting endpoint
- [ ] Automated test execution on device
- [ ] Remote test triggering via API

**Architecture:**
```
Development Mode Test Suite
├── Hardware Sensor Tests
├── SDR Integration Tests
├── GPS/AHRS Validation
├── Network Output Verification
└── End-to-End System Tests
```

**Deliverables:**
- `test/device/` directory with test suite
- Web UI integration for test results
- API endpoints for remote testing
- Documentation in CLAUDE.md

### Phase 4.2: Hardware Sensor Integration Tests
**Priority:** HIGH (validates physical sensors)

**Test Coverage:**
- [ ] **GPS Modules**
  - UART communication
  - Fix acquisition timing
  - Position accuracy validation
  - Satellite tracking
  - NMEA/UBX protocol compliance

- [ ] **AHRS/IMU Sensors**
  - I2C/SPI communication
  - Calibration procedures
  - Attitude accuracy (pitch, roll, heading)
  - Acceleration/gyro readings
  - Sensor fusion algorithms

- [ ] **Barometric Pressure Sensors**
  - I2C communication
  - Pressure readings vs known altitude
  - Temperature compensation
  - Altitude calculation accuracy

- [ ] **CPU Temperature**
  - Temperature reading accuracy
  - Thermal management validation

**Test Methodology:**
- Known positions (GPS coordinates at test location)
- Static attitude references (level surface, known angles)
- Reference barometer for altitude validation
- Multi-device comparison

**Deliverables:**
- Hardware test suite in `test/device/hardware/`
- Reference data for test location
- Pass/fail criteria for each sensor
- Diagnostic output for failures

### Phase 4.3: SDR Integration Tests
**Priority:** HIGH (core ADS-B/UAT/OGN reception)

**Test Coverage:**
- [ ] **dump1090 Integration**
  - SDR device detection
  - 1090 MHz reception
  - Message decoding accuracy
  - Performance under load
  - Signal strength calibration

- [ ] **dump978 Integration**
  - 978 MHz UAT reception
  - Uplink message decoding
  - Weather product extraction
  - Traffic message parsing

- [ ] **ogn-rx-eu Integration**
  - 868 MHz OGN reception
  - FLARM message decoding
  - APRS parsing
  - Tracker position accuracy

- [ ] **rtl-ais Integration**
  - Marine AIS reception
  - Message decoding
  - Position calculation

**Test Methodology:**
- Recorded IQ files for repeatable tests
- Signal generator for controlled tests
- Real-world reception during test flights
- Multi-SDR configurations

**Test Scenarios:**
```
test/device/sdr/
├── recorded_iq/
│   ├── 1090es_heavy_traffic.iq
│   ├── uat_weather.iq
│   ├── ogn_gliders.iq
│   └── ais_marina.iq
├── signal_generator/
│   └── test_patterns.txt
└── test_sdr_integration.go
```

**Deliverables:**
- SDR test suite with recorded signals
- Performance benchmarks
- Signal quality metrics
- Decoding accuracy validation

### Phase 4.4: Network Output Verification Tests
**Priority:** MEDIUM

**Test Coverage:**
- [ ] **GDL90 Output**
  - Message format compliance
  - Heartbeat timing
  - Traffic report accuracy
  - Ownship report correctness
  - FIS-B weather formatting

- [ ] **NMEA Output**
  - FLARM sentence generation
  - Checksum accuracy
  - Update rate compliance
  - Compatibility with EFBs

- [ ] **X-Plane Output**
  - GPS data format
  - Traffic data format
  - Attitude data format
  - ForeFlight compatibility

- [ ] **JSON WebSocket**
  - Real-time traffic updates
  - Status updates
  - Configuration changes
  - Error reporting

**Test Methodology:**
- Packet capture and analysis
- EFB simulator integration
- Protocol conformance checking
- Performance under load

**Deliverables:**
- Network output test suite
- EFB compatibility matrix
- Protocol validation tools
- Performance benchmarks

### Phase 4.5: End-to-End System Tests
**Priority:** MEDIUM (validates complete system)

**Test Scenarios:**
- [ ] **Ground Testing**
  - Static receiver test (no GPS)
  - Mobile receiver test (GPS + motion)
  - Multi-target tracking
  - Weather reception and display

- [ ] **Flight Testing**
  - In-flight reception quality
  - Ownship detection accuracy
  - Traffic alerting correctness
  - AHRS attitude accuracy
  - Battery life and thermal performance

- [ ] **Stress Testing**
  - Heavy traffic density (>100 targets)
  - Sustained operation (24+ hours)
  - Thermal stress (high ambient temp)
  - Low power scenarios (battery near empty)

**Test Methodology:**
- Automated test flights with reference data
- Controlled environment testing
- Real-world flight validation
- Comparison with certified ADS-B receivers

**Flight Test Data Collection:**
```
test/device/flight_tests/
├── reference_data/
│   ├── gps_track.gpx
│   ├── known_traffic.json
│   └── certified_adsb_log.txt
├── stratux_logs/
│   ├── stratux.log
│   ├── traffic_log.json
│   └── gps_log.nmea
└── analysis/
    ├── compare_tracks.py
    ├── validate_traffic.py
    └── generate_report.py
```

**Deliverables:**
- End-to-end test suite
- Flight test procedures
- Data analysis tools
- Certification-ready test reports

### Phase 4.6: Automated Regression Testing
**Priority:** LOW (infrastructure for ongoing validation)

**Objectives:**
- Continuous testing on development devices
- Automated nightly tests
- Performance regression detection
- Hardware compatibility testing

**Infrastructure:**
- [ ] Test Raspberry Pi devices in lab
- [ ] Automated test execution
- [ ] Result reporting dashboard
- [ ] Alert system for failures

**Test Matrix:**
```
Hardware Variants:
- Raspberry Pi 3B+
- Raspberry Pi 4 (4GB)
- Raspberry Pi 5
- Different SDR dongles (RTL-SDR v3, v4)
- Different GPS modules (u-blox, GlobalSat)
- Different AHRS modules (MPU6050, MPU9250)

Software Variants:
- US region (UAT enabled)
- EU region (OGN enabled)
- Different Debian versions
- Custom kernel versions
```

**Deliverables:**
- CI/CD pipeline for device tests
- Test result dashboard
- Hardware compatibility matrix
- Regression tracking

---

## Phase 5: Advanced Testing (Future) 🔮

### Phase 5.1: Fuzzing and Security Testing
- Protocol fuzzing (malformed ADS-B, UAT, OGN messages)
- Network input validation
- Web interface security testing
- Configuration injection testing

### Phase 5.2: Performance and Load Testing
- Benchmark suite for critical paths
- Memory leak detection
- CPU usage profiling
- Throughput testing (messages/sec)

### Phase 5.3: Certification Support
- DO-178C compliance testing
- RTCA DO-160G environmental testing
- TSO-C154c conformance (if pursuing certification)
- Documentation for regulatory submission

---

## Testing Best Practices

### Code Coverage Targets
- **Critical Safety Functions:** 100% coverage with boundary testing
- **Protocol Parsers:** 90%+ coverage
- **Utility Functions:** 100% coverage
- **Hardware Interfaces:** 80%+ coverage (with mocking)
- **Overall Codebase:** 80% coverage (long-term goal)

### Test Design Principles
1. **Independence:** Tests must not depend on execution order
2. **Repeatability:** Same input always produces same output
3. **Isolation:** Use mocks/stubs for external dependencies
4. **Clarity:** Test names describe what is being tested
5. **Coverage:** Test happy path, edge cases, and error conditions
6. **Performance:** Tests should run quickly (<1 second each)

### Test Organization
```
main/
├── *_test.go              # Unit tests alongside source
├── testdata/              # Test fixtures and trace files
│   ├── protocol_name/
│   │   ├── *.trace.gz
│   │   └── generate_trace.go
│   └── README.md
└── integration_*_test.go  # Integration tests

test/
├── device/                # Physical device tests
│   ├── hardware/
│   ├── sdr/
│   └── system/
└── README.md              # Test utilities documentation
```

### Continuous Integration
- All tests run on every commit
- Coverage must not decrease
- Tests must pass before merge
- Nightly device tests on hardware
- Performance benchmarks tracked

---

## Current Status Summary

### Completed (Phases 1-2)
- ✅ 5,228 lines of test code
- ✅ 132+ test functions
- ✅ 18.9% main package coverage
- ✅ 90.2% common package coverage
- ✅ 29.7% uatparse package coverage
- ✅ 24 functions at 100% coverage
- ✅ Integration test framework established
- ✅ Trace file replay methodology proven

### In Progress (Phase 3)
- 🚧 Legacy test migration planning
- 🚧 /test/ directory audit
- 🚧 /test-data/ conversion strategy

### Future Work (Phase 4+)
- 🔮 Physical device test suite
- 🔮 Hardware sensor validation
- 🔮 End-to-end system tests
- 🔮 Flight test procedures

### Estimated Timeline
- **Phase 3.1-3.2:** 1-2 weeks (audit and planning)
- **Phase 3.3-3.6:** 4-6 weeks (integration test development)
- **Phase 4.1-4.2:** 2-3 weeks (device test framework)
- **Phase 4.3-4.5:** 4-8 weeks (hardware validation)
- **Phase 4.6:** 2-3 weeks (automation infrastructure)

### Coverage Projection
- After Phase 3 completion: **35-45%** coverage
- After Phase 4 completion: **60-75%** coverage
- Long-term goal: **80%+** coverage

---

## Contributing to Test Coverage

### Adding New Tests
1. Identify untested functions (use `go tool cover`)
2. Create test file following naming convention
3. Write comprehensive test cases (happy path + edge cases)
4. Ensure tests are independent and repeatable
5. Update this roadmap with completed work
6. Submit PR with test coverage improvement

### Converting Legacy Tests
1. Identify test scenario from `/test-data/`
2. Create trace file generator
3. Implement trace replay test
4. Verify behavior matches manual test
5. Document in testdata README
6. Remove obsolete test-data file

### Device Testing Setup
1. Acquire test hardware (Raspberry Pi + sensors)
2. Install Stratux in development mode
3. Set up test harness
4. Run device test suite
5. Report results and issues

---

## References

- **Coverage Summary:** [coverage_summary.md](./coverage_summary.md)
- **CI Status:** [CI_STATUS.md](./CI_STATUS.md)
- **Coding Standards:** [CODING_STANDARDS.md](./CODING_STANDARDS.md)
- **Build Instructions:** [BUILD_INSTRUCTIONS.md](./BUILD_INSTRUCTIONS.md)
- **Developer Guide:** [CLAUDE.md](./CLAUDE.md)

---

**Last Updated:** 2025-10-14
**Current Phase:** 3.1 (Legacy Test Migration - Audit)
**Next Milestone:** Complete /test/ and /test-data/ audit
