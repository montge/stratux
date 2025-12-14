## Why
Test coverage has improved from 26% to 57.0% through Phase 1 and Phase 2 work. The target is 90%+ coverage to ensure code reliability and catch regressions. SonarCloud is integrated to track coverage metrics over time.

**Related Issue**: GitHub #6 - Improve Go test coverage from 26% to 90%

## Current Status (December 2025)
- **Starting coverage**: 26.3%
- **Current coverage**: 57.3%
- **Target coverage**: 90%+
- **Remaining gap**: ~33%

## What Changes
- ✅ Phase 1: Add unit tests for pure logic functions (Complete - 56.5%)
- ✅ Phase 2: Add configurable paths for testability (Complete - 57.1%)
- [ ] Phase 3: Create interfaces for hardware dependencies to enable mocking
- [ ] Phase 4: Add integration tests for complete data flows
- [ ] Phase 5: Complete error path and edge case testing

## Impact
- Affected specs: testing (new capability)
- Affected code: `main/*_test.go`, new test files, interface definitions
- Coverage files tracked by SonarCloud

## Coverage Analysis by Category

### High-Value Targets (3+ percentage points each)
| Category | Current | Target | Strategy |
|----------|---------|--------|----------|
| Network watchers | 0% | 80% | Interface mocking |
| GPS serial I/O | 0% | 80% | Mock serial ports |
| SDR operations | 0-66% | 80% | Mock SDR device |
| Tile/map handlers | 50-80% | 100% | Make STRATUX_HOME configurable |

### Blocked Functions Requiring Interface Work
| Function | Coverage | Blocker | Solution |
|----------|----------|---------|----------|
| networkOutWatcher | 0% | Infinite loop + UDP | Extract logic, mock interface |
| gpsSerialReader | 0% | Serial port hardware | Mock serial.Port interface |
| sdrWatcher | 0% | SDR hardware | Mock SDR interface |
| heartBeatSender | 0% | Infinite loop + timing | Extract logic, mock clock |
| icmpEchoSender | 0% | Raw socket operations | Mock ICMP interface |

### Functions with Path Dependencies
| Function | Coverage | Blocker | Solution |
|----------|----------|---------|----------|
| handleTilesets | 50% | STRATUX_HOME constant | Make configurable |
| loadTile | 9.5% | STRATUX_HOME constant | Make configurable |
| lookupOgnTailNumber | 42.9% | STRATUX_HOME constant | Make configurable |

## Revised Phase Plan

### Phase 3: Hardware Interface Mocking (Target: 70%)
**Goal**: Create interfaces for hardware-dependent code

1. **GPS Interface**
   - Define `SerialPort` interface for gps.go
   - Mock serial operations in tests
   - Test processNMEALineLow without hardware

2. **SDR Interface**
   - Define `SDRDevice` interface for sdr.go
   - Mock SDR start/stop/kill operations
   - Test SDR management logic

3. **Network Interface**
   - Define `ConnectionWriter` interface
   - Mock UDP/TCP connections
   - Test message sending logic

### Phase 4: Configurable Constants (Target: 75%)
**Goal**: Make remaining constants configurable

1. Make `STRATUX_HOME` configurable (like we did with logDirPath)
2. Test handleTilesets, loadTile, lookupOgnTailNumber
3. Test OGN device database loading

### Phase 5: Loop Extraction (Target: 82%)
**Goal**: Extract testable logic from infinite loops

1. Extract `heartBeatSender` logic into testable function
2. Extract `networkOutWatcher` logic into testable function
3. Extract `gpsSerialReader` logic into testable function
4. Test extracted functions with mocked dependencies

### Phase 6: Integration Tests (Target: 88%)
**Goal**: Test complete data flows

1. **Traffic Flow Tests**
   - SDR input → parsing → traffic map → GDL90 output
   - Multi-source fusion (1090ES + UAT + OGN)

2. **GPS Flow Tests**
   - NMEA input → position update → ownship report
   - Position validation and filtering

3. **Network Output Tests**
   - Client connection lifecycle
   - Message queuing and throttling

### Phase 7: Edge Cases & Error Paths (Target: 90%+)
**Goal**: Complete coverage of error handling

1. Network errors, timeouts, disconnections
2. Invalid input data handling
3. Resource exhaustion scenarios
4. Concurrent access testing
5. Boundary condition testing (max/min values)

## Success Criteria
- [ ] Coverage reaches 90%+
- [x] All tests pass in CI
- [ ] No flaky tests
- [ ] Test execution time < 120 seconds
- [ ] No hardware dependencies in tests
