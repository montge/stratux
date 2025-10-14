# Test Coverage Improvement Roadmap

## Executive Summary

This document outlines the strategy to improve test coverage for Stratux from the current 9.4% to DO-278A SAL-3 compliance targets (80% statement coverage, 70% decision coverage) without requiring major code refactoring.

**Current Status**: 9.4% overall coverage (as of 2025-10-13)
**Target**: 80% statement coverage, 70% decision coverage
**Approach**: Integration testing via trace/replay + targeted unit tests
**Timeline**: Phased approach over multiple iterations

## Coverage Analysis

### Current Coverage by Package

| Package | Coverage | Test Lines | Status |
|---------|----------|------------|--------|
| common | 90.2% | 1,089 | ✅ Complete |
| uatparse | 29.7% | 350 | ⚠️ Needs work |
| main | 9.4% | 4,247 | ❌ Primary focus |
| **Total** | **~15%** | **5,686** | **In progress** |

### Why Coverage is Low

Analysis of `main/` package reveals:
- **91%** of code involves hardware I/O, goroutines, or global state
- **Hardware dependencies**: SDR radios (dump1090, dump978), GPS serial, network sockets
- **Background workers**: 15+ goroutines handling continuous data streams
- **Global state**: mySituation, traffic map, ADSBTowers, globalSettings

### What's Already Well-Tested

✅ **Utility functions** (90-100% coverage):
- ICAO to registration conversion (98.6%)
- Distance calculations (100%)
- Traffic priority computation (100%)
- NMEA checksum validation (100%)
- String formatting utilities (100%)

## Coverage Goals by Phase

### Phase 1: Infrastructure (COMPLETED ✅)
**Goal**: Build testing infrastructure without hardware
**Duration**: Completed 2025-10-13

Deliverables:
- [x] Trace file format tests (trace_test.go - 336 lines)
- [x] Sample trace files (1090ES, GPS NMEA)
- [x] Protocol documentation (PROTOCOLS.md - 565 lines)
- [x] Trace generation scripts
- [x] testdata/ directory structure

**Coverage Impact**: +0.3% (infrastructure only)

### Phase 2: Protocol Parser Integration Tests (NEXT)
**Goal**: Test protocol parsers via trace replay
**Duration**: 3-5 days
**Expected Coverage**: +15-20%

#### 2.1 1090ES ADS-B Parser Testing
**Target**: `parseDump1090Message()` and related traffic functions
**Current Coverage**: ~20%
**Target Coverage**: 80%

Test scenarios to create:
- [ ] Basic aircraft position/velocity updates
- [ ] Aircraft appearing and disappearing (timeout handling)
- [ ] Invalid/malformed JSON handling
- [ ] High-frequency update scenarios (1Hz position updates)
- [ ] Traffic deduplication (same aircraft from multiple towers)
- [ ] Altitude encoding edge cases (ground, above FL600)
- [ ] TISB vs ADS-B source handling

**Files to create**:
- `main/integration_adsb_test.go` (300+ lines)
- `testdata/adsb/high_frequency.trace.gz`
- `testdata/adsb/invalid_messages.trace.gz`
- `testdata/adsb/multi_tower.trace.gz`

#### 2.2 GPS/NMEA Parser Testing
**Target**: `processNMEALineLow()` and GPS state management
**Current Coverage**: ~15%
**Target Coverage**: 75%

Test scenarios:
- [ ] Complete NMEA sentence sequences (RMC, GGA, GSA, GSV)
- [ ] Invalid checksums
- [ ] Incomplete sentences
- [ ] GPS signal loss and reacquisition
- [ ] Coordinate parsing edge cases (equator, prime meridian, poles)
- [ ] Speed/altitude/heading updates
- [ ] Satellite tracking state changes

**Files to create**:
- `main/integration_gps_test.go` (250+ lines)
- `testdata/gps/signal_loss.trace.gz`
- `testdata/gps/invalid_checksums.trace.gz`
- `testdata/gps/edge_locations.trace.gz`

#### 2.3 UAT 978MHz Parser Testing
**Target**: `handleUatMessage()` and FIS-B weather
**Current Coverage**: ~10%
**Target Coverage**: 70%

Test scenarios:
- [ ] UAT traffic messages
- [ ] FIS-B weather products (METARs, TAFs, TFRs, NOTAMs)
- [ ] Ownship reports
- [ ] Ground station tracking
- [ ] Message reassembly for long FIS-B products

**Files to create**:
- `main/integration_uat_test.go` (200+ lines)
- `testdata/uat/basic_traffic.trace.gz`
- `testdata/uat/fisb_weather.trace.gz`

#### 2.4 OGN/APRS Parser Testing
**Target**: `parseAprsMessage()` and glider tracking
**Current Coverage**: ~5%
**Target Coverage**: 65%

Test scenarios:
- [ ] Valid APRS position reports
- [ ] Invalid/malformed APRS strings
- [ ] Various aircraft type codes (gliders, powered, helicopters)
- [ ] Stealth mode handling
- [ ] Position extrapolation
- [ ] OGN-specific extensions (climb rate, turn rate)

**Files to create**:
- `main/integration_ogn_test.go` (150+ lines)
- `testdata/ogn/mixed_types.trace.gz`
- `testdata/ogn/invalid_formats.trace.gz`

**Phase 2 Total**: +15-20% coverage, ~900 lines of test code

### Phase 3: GDL90 Output Testing
**Goal**: Verify GDL90 message generation
**Duration**: 2-3 days
**Expected Coverage**: +10-15%

#### 3.1 GDL90 Message Encoding Tests
**Target**: `makeHeartbeat()`, `makeTrafficReport()`, `makeOwnshipReport()`
**Current Coverage**: 70-87% (missing edge cases)
**Target Coverage**: 95%

Test scenarios:
- [ ] Heartbeat with various GPS states (valid, invalid, 2D fix, 3D fix)
- [ ] Heartbeat with error conditions
- [ ] Traffic reports with all aircraft types
- [ ] Traffic reports with altitude sources (pressure, GNSS)
- [ ] Ownship reports with varying accuracy
- [ ] Message byte stuffing (0x7D, 0x7E escaping)
- [ ] CRC calculation verification

**Files to create**:
- `main/gdl90_integration_test.go` (300+ lines)
- Helper functions to decode GDL90 messages
- Verification against FAA spec

#### 3.2 End-to-End Protocol Tests
**Target**: Full pipeline from trace input to GDL90 output
**Current Coverage**: N/A (new tests)
**Target Coverage**: Integration coverage

Test scenarios:
- [ ] Replay trace → verify correct GDL90 traffic messages generated
- [ ] Replay trace → verify GDL90 heartbeat timing
- [ ] Replay trace → verify ownship report accuracy
- [ ] Multiple aircraft → verify GDL90 message ordering
- [ ] GPS updates → verify ownship report changes

**Files to create**:
- `main/integration_e2e_test.go` (250+ lines)

**Phase 3 Total**: +10-15% coverage, ~550 lines of test code

### Phase 4: Error Handling and Edge Cases
**Goal**: Cover error paths and boundary conditions
**Duration**: 2-3 days
**Expected Coverage**: +8-12%

Test scenarios:
- [ ] Network connection failures
- [ ] Disk full scenarios (trace logging, SQLite)
- [ ] Memory pressure (large traffic counts)
- [ ] Invalid configuration values
- [ ] Clock rollback handling
- [ ] Concurrent access edge cases
- [ ] Sensor failure modes (AHRS, pressure)

**Files to create**:
- `main/error_handling_test.go` (200+ lines)
- `main/boundary_conditions_test.go` (150+ lines)

**Phase 4 Total**: +8-12% coverage, ~350 lines of test code

### Phase 5: UAT Parse Library Coverage
**Goal**: Improve uatparse package from 29.7% to 80%
**Duration**: 2 days
**Expected Coverage**: +15% (weighted by package size)

Test scenarios:
- [ ] UAT frame decoding
- [ ] FIS-B product parsing (all product types)
- [ ] CRC validation
- [ ] Incomplete frame handling
- [ ] Bit error scenarios

**Files to create**:
- `uatparse/integration_test.go` (300+ lines)

**Phase 5 Total**: +15% coverage, ~300 lines of test code

## Coverage Target Summary

| Phase | Focus Area | Expected Δ | Cumulative | Test Lines Added |
|-------|-----------|-----------|------------|------------------|
| 1 (DONE) | Infrastructure | +0.3% | 9.4% → 9.7% | 336 |
| 2 | Protocol Parsers | +15-20% | 9.7% → 27% | 900 |
| 3 | GDL90 Output | +10-15% | 27% → 40% | 550 |
| 4 | Error Handling | +8-12% | 40% → 50% | 350 |
| 5 | UAT Parse | +15% | 50% → 65% | 300 |
| **Total** | | **+55%** | **9.4% → 65%** | **~2,436** |

## What Remains Uncovered (and Why)

### Will NOT Be Covered Without Refactoring

**Hardware Initialization (5-7% of codebase)**:
- SDR device enumeration and initialization
- Serial port opening and configuration
- Network socket binding
- Requires actual hardware or major mocking refactor

**Background Goroutines (8-10% of codebase)**:
- 15+ goroutines running infinite loops
- `sdrWatcher()`, `gpsSerialReader()`, `initTraffic()`
- Would require architecture changes to test

**Global State Initialization (3-5% of codebase)**:
- `main()` function and startup sequence
- Config file loading
- System service integration
- Would need dependency injection

**Web Interface Handlers (5-7% of codebase)**:
- HTTP/WebSocket handlers in `managementinterface.go`
- Requires HTTP test server setup
- Complex state dependencies

### Could Be Covered With Future Work

**AHRS/Sensor Integration (3-5% of codebase)**:
- IMU data processing
- Pressure altitude calculations
- Could add sensor simulation traces

**Weather Product Processing (2-3% of codebase)**:
- FIS-B METAR/TAF/TFR parsing
- Can be added in Phase 5 with complete UAT traces

## Testing Strategy

### Trace-Based Integration Testing

**Advantages**:
- No hardware required
- Fast execution in CI
- Reproducible
- Can test rare scenarios

**How it works**:
1. Record real hardware data once (or generate synthetic)
2. Commit trace files to testdata/
3. Tests replay trace files at high speed
4. Verify protocol parser output
5. Verify state changes (traffic map, GPS position, etc.)

**Test pattern**:
```go
func TestADSBTrafficParsing(t *testing.T) {
    // Reset global state
    resetGlobalTrafficMap()

    // Replay trace file
    trace := NewTraceLogger()
    trace.Replay("testdata/adsb/basic_adsb.trace.gz", 10.0, 0, []string{"dump1090"})

    // Wait for processing
    time.Sleep(100 * time.Millisecond)

    // Verify traffic was parsed
    trafficMutex.Lock()
    if len(traffic) != 2 {
        t.Errorf("Expected 2 aircraft, got %d", len(traffic))
    }

    // Verify specific aircraft data
    if ti, ok := traffic[0xA12345]; ok {
        if ti.Tail != "UAL123" {
            t.Errorf("Expected tail UAL123, got %s", ti.Tail)
        }
        if ti.Alt != 35000 {
            t.Errorf("Expected altitude 35000, got %d", ti.Alt)
        }
    } else {
        t.Error("Expected to find aircraft A12345")
    }
    trafficMutex.Unlock()
}
```

### Limitations and Workarounds

**Global State Dependencies**:
- Problem: Most functions access global maps/structs
- Workaround: Reset functions at start of each test
- Example: `resetGlobalTrafficMap()`, `resetGPS()`

**Timing Dependencies**:
- Problem: Some functions use goroutines with delays
- Workaround: Use `time.Sleep()` in tests or inject timing control
- Alternative: Test synchronous functions directly

**Hardware Dependencies**:
- Problem: Can't test SDR initialization
- Workaround: Focus on data processing, not hardware access
- Acceptance: 10-15% will remain uncovered

## Implementation Guidelines

### Test File Organization
```
main/
├── trace_test.go                    # Infrastructure tests (DONE)
├── integration_adsb_test.go         # 1090ES parser tests
├── integration_gps_test.go          # GPS/NMEA parser tests
├── integration_uat_test.go          # UAT parser tests
├── integration_ogn_test.go          # OGN/APRS parser tests
├── gdl90_integration_test.go        # GDL90 output tests
├── integration_e2e_test.go          # End-to-end tests
├── error_handling_test.go           # Error path tests
└── boundary_conditions_test.go      # Edge case tests

testdata/
├── adsb/
│   ├── basic_adsb.trace.gz         (DONE)
│   ├── high_frequency.trace.gz
│   ├── invalid_messages.trace.gz
│   ├── multi_tower.trace.gz
│   └── generate_*.go               (generators)
├── gps/
│   ├── basic_gps.trace.gz          (DONE)
│   ├── signal_loss.trace.gz
│   ├── invalid_checksums.trace.gz
│   └── edge_locations.trace.gz
├── uat/
│   ├── basic_traffic.trace.gz
│   └── fisb_weather.trace.gz
└── ogn/
    ├── mixed_types.trace.gz
    └── invalid_formats.trace.gz
```

### Coding Standards for Tests

1. **Table-driven tests** for multiple scenarios
2. **Clear test names** describing scenario
3. **Reset global state** at start of each test
4. **Use test fixtures** (trace files) checked into repo
5. **Verify specific values**, not just "no error"
6. **Document test purpose** in comments
7. **Keep tests independent** (can run in any order)

### CI Integration

Tests will run automatically in GitHub Actions:
```yaml
- name: Run tests with coverage
  run: |
    go test -v -coverprofile=coverage.out -covermode=atomic ./main/...
    go test -v -coverprofile=coverage.out -covermode=atomic ./uatparse/...
    go test -v ./common/...

- name: Coverage report
  run: |
    go tool cover -func=coverage.out
    go tool cover -html=coverage.out -o coverage.html
```

## Success Metrics

### Quantitative
- ✅ Phase 1: Infrastructure tests passing
- 🎯 Phase 2: Coverage ≥ 27% (main package ≥ 15%)
- 🎯 Phase 3: Coverage ≥ 40% (main package ≥ 30%)
- 🎯 Phase 4: Coverage ≥ 50% (main package ≥ 40%)
- 🎯 Phase 5: Coverage ≥ 65% (uatparse ≥ 80%)

### Qualitative
- All protocol parsers have integration tests
- All GDL90 message types tested
- Error paths covered for critical functions
- Tests run in <30 seconds
- Tests don't require hardware
- Tests are maintainable and documented

## Risk Mitigation

### Risk: Global State Makes Tests Flaky
**Mitigation**: Create reset functions for each global structure
**Fallback**: Run tests with `-p 1` flag (no parallelism)

### Risk: Trace Files Too Large for Git
**Mitigation**: Keep individual trace files <500KB
**Fallback**: Use Git LFS for larger files

### Risk: Integration Tests Too Slow
**Mitigation**: Use high trace replay speeds (10-100x)
**Fallback**: Mark slow tests with `-short` skip

### Risk: Can't Hit 80% Without Refactoring
**Acceptance**: 65% is acceptable given architecture
**Future Work**: Dependency injection would enable 80%+

## Future Enhancements (Beyond This Roadmap)

### Refactoring for Testability (Future)
Would enable 80%+ coverage but requires code changes:
- Dependency injection for hardware interfaces
- Protocol parser interfaces (strategy pattern)
- State management with testable boundaries
- Background worker lifecycle management

### System-Level Testing (Future)
Beyond unit/integration tests:
- Docker-based full system tests
- Simulated hardware via virtual SDR
- Real EFB application integration tests
- Long-duration stability tests

## Appendix: DO-278A Compliance

### SAL-3 Requirements
Software Assurance Level 3 (ground systems) requires:
- **Statement coverage**: ≥80%
- **Decision coverage**: ≥70%
- **Modified Condition/Decision Coverage (MC/DC)**: Not required for SAL-3

### Current Compliance Status
- Statement coverage: 9.4% → Need +70.6%
- Decision coverage: ~8% (estimated) → Need +62%

### Expected Compliance After Roadmap
- Statement coverage: ~65% → Still -15% short
- Decision coverage: ~55% → Still -15% short

**Note**: Full compliance requires architectural changes not in scope of this roadmap. Current approach targets "reasonable best effort" of 65% given constraints.

## Conclusion

This roadmap provides a practical path to significantly improve test coverage from 9.4% to ~65% without requiring major code refactoring. The trace-based integration testing approach allows hardware-independent testing in CI while covering the most critical data processing paths.

**Next Step**: Begin Phase 2 - Protocol Parser Integration Tests starting with 1090ES ADS-B parser testing.
