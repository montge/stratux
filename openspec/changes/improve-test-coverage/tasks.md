## 1. Phase 1: Low-Hanging Fruit (Complete: 56.5%)

### 1.1 HTTP API Handlers
- [x] Test remaining managementinterface.go handlers
- [x] Use httptest.NewRecorder() pattern consistently

### 1.2 Protocol Encoding (gen_gdl90.go)
- [x] Test makeOwnshipReport edge cases (98.9% coverage)
- [x] Test makeFFIDMessage edge cases (93.8% coverage)
- [x] Test makeOwnshipGeometricAltitudeReport (100% coverage)
- [x] Test makeTrafficReportMsg variants (100% coverage)
- [x] Test makeAHRS* functions (all 100% coverage)
  - makeAHRSSimReport: 0% → 100%
  - makeFFAHRSMessage: 0% → 100%
  - makeAHRSGDL90Report: 0% → 100%

### 1.3 Message Parsing
- [x] Test parseDownlinkReport edge cases
- [x] Test parseDump1090Message edge cases
- [x] Test parseAprsMessage edge cases (91.5% coverage)

### 1.4 Logging Functions (NEW)
- [x] Test getStratuxLogFiles (77.8% → 100% via configurable paths)
- [x] Test logFileSize (100% coverage)
- [x] Test clearDebugLogFile (100% coverage)
- [x] Test openLogFile (90.9% coverage)
- [x] Test logInf/logErr/logDbg (100% coverage each)

### 1.5 OGN/APRS Functions (NEW)
- [x] Test importOgnTrafficMessage (100% coverage)
- [x] Test getTailNumber (100% coverage)
- [x] Test lookupOgnTailNumber (42.9% - limited by file I/O)
- [x] Test importOgnStatusMessage (100% coverage)

### 1.6 Traffic/Demo Functions (NEW)
- [x] Test updateDemoTraffic (100% coverage) - ADSR upgrade path
- [x] Test pingKill (100% coverage) - wait loop path
- [x] Test pongKill (100% coverage) - wait loop path

### 1.7 MBTiles/Map Functions (NEW)
- [x] Test connectMbTilesArchive (92.9% coverage)
- [x] Test handleTilesets (50% coverage)
- [x] Test handleTile validation
- [x] Test readMbTilesMetadata (82.1% → 89.3% coverage)
  - Bounds calculation when not present
  - PBF format detection
  - Empty value filtering

## 2. Phase 2: Configurable Paths (Current: 57.0%, Target: 60%)

**Status: IN PROGRESS**

### 2.1 Created config_paths.go for Testability
- [x] Added stratuxHome variable (replaces STRATUX_HOME constant)
- [x] Added logDirPath variable (replaces logDir constant)
- [x] Added varLogDirPath variable (replaces varLogDir constant)
- [x] Added resetPathsToDefaults() function (100% coverage)

### 2.2 Logging Function Improvements
- [x] Refactored getStratuxLogFiles to use logDirPath (77.8% → 100%)
- [x] Refactored rotateLogs to use configurable paths (33.3% → 100%)
- [x] Refactored deleteOldestLog to use configurable paths (27.3% → 81.8%)

### 2.3 Management Interface Improvements
- [x] Refactored managementinterface.go to use varLogDirPath
- [x] handleDeleteAHRSLogFiles tests with configurable paths (100%)
- [x] handleDownloadAHRSLogsRequest tests (36.1% → 72.2%)
- [x] viewLogs tests with configurable paths (77.5% → 85.0%)

### 2.4 Remaining to reach 60% (~3% needed)

**Blockers for remaining functions:**
- [ ] sendTrafficUpdates (45.1%) - blocked by network operations (WebSocket broadcasting)
- [ ] processNMEALineLow (84.2%) - already heavily tested (91 test calls)
- [ ] sdrKill (66.7%) - requires mock SDR device
- [ ] lookupOgnTailNumber (42.9%) - reads from STRATUX_HOME constant
- [ ] handleTilesets (50%) - reads from STRATUX_HOME constant
- [ ] loadTile (9.5%) - reads from STRATUX_HOME constant

**Note:** Many remaining functions have dependencies on:
1. Hardware (SDR, GPS, sensors) that isn't available in test environment
2. Network operations (real sockets, DHCP leases)
3. Infinite loops (watchers, listeners)

**Next steps:** Consider making STRATUX_HOME configurable or adding more HTTP handler tests.

## 3. Phase 3: Interface Mocking (Target: 75%)

### 3.1 Create Hardware Interfaces
- [ ] Define GPSReader interface
- [ ] Define SDRDevice interface
- [ ] Define SensorReader interface

### 3.2 Create Network Interfaces
- [ ] Define UDPSender interface
- [ ] Define ConnectionManager interface

### 3.3 Refactor for Dependency Injection
- [ ] Refactor GPS functions to use interfaces
- [ ] Refactor network functions to use interfaces

## 4. Phase 4: Integration Tests (Target: 85%)

### 4.1 Traffic Flow Tests
- [ ] SDR input → parsing → traffic map → GDL90 output
- [ ] Multi-source fusion (1090ES + UAT + OGN)

### 4.2 GPS Flow Tests
- [ ] NMEA input → position update → ownship report

### 4.3 Network Output Tests
- [ ] Client connection lifecycle
- [ ] Message queuing and throttling

## 5. Phase 5: Edge Cases (Target: 90%)

### 5.1 Error Path Testing
- [ ] Network errors, timeouts
- [ ] Invalid input data
- [ ] Resource exhaustion

### 5.2 Boundary Conditions
- [ ] Max/min values
- [ ] Empty inputs
- [ ] Concurrent access

## Success Criteria
- [ ] Coverage reaches 60% (minimum)
- [x] All tests pass in CI
- [ ] No flaky tests
- [ ] Test execution time < 120 seconds
