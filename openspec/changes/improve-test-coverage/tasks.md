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

### 1.3 Message Parsing
- [x] Test parseDownlinkReport edge cases
- [x] Test parseDump1090Message edge cases
- [x] Test parseAprsMessage edge cases (91.5% coverage)

### 1.4 Logging & OGN Functions
- [x] Test logging functions (logInf/logErr/logDbg)
- [x] Test OGN/APRS functions

### 1.5 MBTiles/Map Functions
- [x] Test connectMbTilesArchive (92.9% coverage)
- [x] Test readMbTilesMetadata (89.3% coverage)

## 2. Phase 2: Configurable Paths (Complete: 57.1%)

### 2.1 Created config_paths.go
- [x] Added stratuxHome variable
- [x] Added logDirPath variable
- [x] Added varLogDirPath variable
- [x] Added resetPathsToDefaults() function (100% coverage)

### 2.2 Logging Function Improvements
- [x] getStratuxLogFiles: 77.8% → 100%
- [x] rotateLogs: 33.3% → 100%
- [x] deleteOldestLog: 27.3% → 81.8%

### 2.3 Management Interface Improvements
- [x] handleDownloadAHRSLogsRequest: 36.1% → 72.2%
- [x] viewLogs: 77.5% → 85.0%
- [x] handleDeleteAHRSLogFiles: 100%

### 2.4 GDL90 Heartbeat Functions (December 2025)
- [x] Test sendAllOwnshipInfo: 0% → 100%
  - Tests heartbeat message sending with valid GPS
  - Tests message sending without GPS fix
  - Tests detected ownship scenario

## 3. Phase 3: Hardware Interface Mocking (In Progress: 57.3%)

### 3.1 GPS Serial Input Interface (Complete)
- [x] Created `SerialReaderInterface` in gps.go
- [x] Extracted `processSerialInput(reader io.Reader)` from gpsSerialReader
- [x] Created `mockSerialReader` for tests in gps_test.go
- [x] Added `TestProcessSerialInput` with 6 test cases
- [x] processSerialInput: 0% → 87.5%

### 3.2 SDR Interface (Deferred)
- [ ] Create `SDRDeviceInterface` in sdr.go
- [ ] Define methods: Start, Stop, Kill, Status
- [ ] Create mock implementation for tests
- **Note**: SDR code heavily depends on RTL-SDR hardware (`rtl.GetDeviceCount()`) and external processes. Requires significant refactoring to make testable.

### 3.3 Network Interface (Partially Complete)
- [x] Core network functions already at 100% (`sendMsg`, `sendGDL90`, `connectionWriter`)
- [ ] `getDHCPLeases` - needs path configuration
- [ ] `refreshConnectedClients` - needs refactoring
- **Note**: Loop-based watchers (`monitorDHCPLeases`, `getNetworkStats`) require extraction like GPS.

### 3.4 Coverage Summary
- processSerialInput: 0% → 87.5%
- gpsSerialReader: 0% (wrapper only, logic now in processSerialInput)
- Network core functions: 100%
- SDR: Requires hardware, deferred to future work

## 4. Phase 4: Configurable Constants (Complete: 57.8%)

### 4.1 Make STRATUX_HOME Configurable
- [x] Added stratuxHome variable to config_paths.go
- [x] Added helper functions: getMapdataPath(), getMapdataStylesPath(), getOgnPath()
- [x] Modify handleTilesets to use getMapdataPath()
- [x] Modify loadTile to use getMapdataPath()
- [x] Modify readMbTilesMetadata to use getMapdataStylesPath()
- [x] Modify lookupOgnTailNumber to use getOgnPath()
- [x] Updated HTTP file server to use getMapdataStylesPath()

### 4.2 Test Tile Handlers
- [x] Created setupTileTestDir() helper for temp directories
- [x] Updated all tile tests to use temp directories
- [x] handleTilesets: 50% (unchanged - needs more work)
- [x] loadTile: 9.5% → 85.7%

### 4.3 Test OGN Functions
- [x] Created setupOgnTestDir() helper for temp directories
- [x] Updated all OGN tests to use temp directories
- [x] lookupOgnTailNumber: 42.9% → 100%
- [x] Test OGN device database parsing (full coverage)

## 5. Phase 5: Loop Extraction (In Progress: 58.4%)

### 5.1 Extract Heartbeat Logic (Complete)
- [x] Create heartBeatOnce() function with all heartbeat logic
- [x] heartBeatSender calls heartBeatOnce in loop
- [x] Test heartBeatOnce with 7 test cases
- [x] heartBeatOnce: 0% → tested (LED status, ownship info, FLARM, traffic)

### 5.2 Extract Network Output Logic
- [ ] Create processNetworkOutput() for single iteration
- [ ] networkOutWatcher calls processNetworkOutput in loop
- [ ] Test processNetworkOutput with mock connections
- [ ] networkOutWatcher: 0% → 80%+
- **Note**: networkOutWatcher is already minimal (channel read + WebSocket send)

### 5.3 Extract GPS Serial Logic
- [ ] Create processSerialData() for single read
- [ ] gpsSerialReader calls processSerialData in loop
- [ ] Test processSerialData with mock serial port
- [ ] gpsSerialReader: 0% → 80%+
- **Note**: processSerialInput already extracted in Phase 3 (87.5% coverage)

### 5.4 Extract Other Watchers (Partial)
- [x] logFileWatcher: Extracted logFileWatcherOnce() with 4 test cases
- [ ] dataLogWatchdog: Extract single iteration
- [ ] monitorDHCPLeases: Extract single iteration

## 6. Phase 6: Integration Tests (Target: 88%)

### 6.1 Traffic Flow Tests
- [ ] Create test_integration.go file
- [ ] Test: SDR message → parseInput → traffic map update
- [ ] Test: Traffic map → makeTrafficReport → GDL90 message
- [ ] Test: Multi-source traffic fusion (1090ES + UAT + OGN)
- [ ] Test: Traffic aging and expiration

### 6.2 GPS Flow Tests
- [ ] Test: NMEA sentence → processNMEALineLow → mySituation update
- [ ] Test: Position update → makeOwnshipReport → GDL90 message
- [ ] Test: GPS fix validation and filtering

### 6.3 Network Output Tests
- [ ] Test: Message queuing per client
- [ ] Test: Client throttling behavior
- [ ] Test: Client connection/disconnection lifecycle
- [ ] Test: Broadcast to multiple clients

### 6.4 WebSocket Tests
- [ ] Test: Traffic WebSocket message format
- [ ] Test: Weather WebSocket message format
- [ ] Test: Radar WebSocket message format
- [ ] Test: Client subscription lifecycle

## 7. Phase 7: Edge Cases & Error Paths (Target: 90%+)

### 7.1 Network Error Handling
- [ ] Test UDP write failure recovery
- [ ] Test TCP connection timeout
- [ ] Test WebSocket disconnect handling
- [ ] Test ICMP permission errors

### 7.2 Input Validation
- [ ] Test malformed NMEA sentences
- [ ] Test invalid ADS-B messages
- [ ] Test corrupted MBTiles databases
- [ ] Test invalid JSON settings

### 7.3 Resource Limits
- [ ] Test max traffic targets limit
- [ ] Test message queue overflow
- [ ] Test log rotation under pressure
- [ ] Test database write failures

### 7.4 Concurrent Access
- [ ] Test concurrent traffic map updates
- [ ] Test concurrent settings changes
- [ ] Test concurrent client connections
- [ ] Test mutex deadlock prevention

### 7.5 Boundary Conditions
- [ ] Test latitude/longitude at limits (±90°, ±180°)
- [ ] Test altitude at extremes (0, FL600)
- [ ] Test speed at limits (0, Mach 3)
- [ ] Test timestamp edge cases

## Success Criteria
- [ ] Coverage reaches 90%+
- [x] All tests pass in CI
- [ ] No flaky tests (TestMonotonicTimeAdvancement needs fixing)
- [ ] Test execution time < 120 seconds
- [ ] No hardware dependencies in unit tests

## Estimated Coverage by Phase

| Phase | Target | Actual Coverage | LOC Change |
|-------|--------|-----------------|------------|
| Phase 1 | 60% | 56.5% (Complete) | +6,000 |
| Phase 2 | 60% | 57.0% (Complete) | +500 |
| Phase 3 | 70% | 57.3% (Complete) | +300 |
| Phase 4 | 75% | 57.8% (Complete) | +200 |
| Phase 5 | 65% | ~65% | +1,200 |
| Phase 6 | 75% | ~75% | +2,000 |
| Phase 7 | 90%+ | ~90%+ | +3,000 |

Total estimated new test code: ~7,200 lines
