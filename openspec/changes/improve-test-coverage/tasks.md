## 1. Phase 1: Low-Hanging Fruit (Current: ~53.7%, Target: 60%)

### 1.1 HTTP API Handlers
- [ ] Test remaining managementinterface.go handlers
- [ ] Use httptest.NewRecorder() pattern consistently

### 1.2 Protocol Encoding (gen_gdl90.go)
- [ ] Test makeOwnshipReport edge cases
- [ ] Test makeOwnshipGeometricAltitudeReport
- [ ] Test makeTrafficReportMsg variants
- [ ] Test makeAHRS* functions

### 1.3 Message Parsing
- [ ] Test parseDownlinkReport edge cases
- [ ] Test parseDump1090Message edge cases
- [ ] Test parseAprsMessage edge cases

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
