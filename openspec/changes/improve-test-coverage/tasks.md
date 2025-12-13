## 1. Phase 1: Low-Hanging Fruit (Current: 56.5%, Target: 60%)

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
- [x] Test getStratuxLogFiles (77.8% coverage)
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

### 1.8 Remaining to reach 60% (~3.5% needed)

**Blockers for remaining functions:**
- [ ] sendTrafficUpdates (45.1%) - blocked by network operations (WebSocket broadcasting)
- [ ] handleDownloadAHRSLogsRequest (36.1%) - requires files in /var/log/
- [ ] processNMEALineLow (84.2%) - already heavily tested (91 test calls)
- [ ] sdrKill (66.7%) - requires mock SDR device
- [ ] rotateLogs (33.3%) - reads from const logDir="/var/log/"
- [ ] deleteOldestLog (27.3%) - reads from const logDir="/var/log/"
- [ ] lookupOgnTailNumber (42.9%) - reads from STRATUX_HOME constant
- [ ] handleTilesets (50%) - reads from STRATUX_HOME constant
- [ ] loadTile (9.5%) - reads from STRATUX_HOME constant

**Note:** Many remaining functions have dependencies on:
1. Constants (STRATUX_HOME, varLogDir, logDir) that can't be overridden in tests
2. Hardware (SDR, GPS, sensors) that isn't available in test environment
3. Network operations (real sockets, DHCP leases)
4. Infinite loops (watchers, listeners)

**Recommendation:** To reach 60%+ coverage, consider Phase 2 interface mocking approach.

## 2. Phase 2: Interface Mocking (Target: 75%)

### 2.1 Create Hardware Interfaces
- [ ] Define GPSReader interface
- [ ] Define SDRDevice interface
- [ ] Define SensorReader interface

### 2.2 Create Network Interfaces
- [ ] Define UDPSender interface
- [ ] Define ConnectionManager interface

### 2.3 Refactor for Dependency Injection
- [ ] Refactor GPS functions to use interfaces
- [ ] Refactor network functions to use interfaces

## 3. Phase 3: Integration Tests (Target: 85%)

### 3.1 Traffic Flow Tests
- [ ] SDR input → parsing → traffic map → GDL90 output
- [ ] Multi-source fusion (1090ES + UAT + OGN)

### 3.2 GPS Flow Tests
- [ ] NMEA input → position update → ownship report

### 3.3 Network Output Tests
- [ ] Client connection lifecycle
- [ ] Message queuing and throttling

## 4. Phase 4: Edge Cases (Target: 90%)

### 4.1 Error Path Testing
- [ ] Network errors, timeouts
- [ ] Invalid input data
- [ ] Resource exhaustion

### 4.2 Boundary Conditions
- [ ] Max/min values
- [ ] Empty inputs
- [ ] Concurrent access

## Success Criteria
- [ ] Coverage reaches 90%
- [ ] All tests pass in CI
- [ ] No flaky tests
- [ ] Test execution time < 60 seconds
