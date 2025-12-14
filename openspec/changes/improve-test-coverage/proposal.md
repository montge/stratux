## Why
Test coverage has improved from 26% to 57.8% through Phase 1-4 work. The target is 90%+ coverage to ensure code reliability and catch regressions. SonarCloud is integrated to track coverage metrics over time.

**Related Issue**: GitHub #6 - Improve Go test coverage from 26% to 90%

## Current Status (December 2025)
- **Starting coverage**: 26.3%
- **Current coverage**: 57.8%
- **Target coverage**: 90%+
- **Remaining gap**: ~32%

## What Changes
- ✅ Phase 1: Add unit tests for pure logic functions (Complete - 56.5%)
- ✅ Phase 2: Add configurable paths for testability (Complete - 57.1%)
- ✅ Phase 3: Create interfaces for hardware dependencies (Complete - 57.3%)
- ✅ Phase 4: Configurable constants for testability (Complete - 57.8%)
- [ ] Phase 5: Loop extraction for testability
- [ ] Phase 6: Integration tests for complete data flows
- [ ] Phase 7: Complete error path and edge case testing

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

### Functions with Path Dependencies (Phase 4 Complete)
| Function | Before | After | Solution |
|----------|--------|-------|----------|
| handleTilesets | 50% | 50% | Made stratuxHome configurable |
| loadTile | 9.5% | 85.7% | Tests use temp directories |
| lookupOgnTailNumber | 42.9% | 100% | Tests use temp directories |

## Revised Phase Plan

### Phase 3: Hardware Interface Mocking (Complete - 57.3%)
**Goal**: Create interfaces for hardware-dependent code

1. **GPS Interface** ✅
   - Created `SerialReaderInterface` in gps.go
   - Extracted `processSerialInput()` for testability
   - Added mockSerialReader in tests

2. **SDR Interface** (Deferred)
   - SDR code too hardware-dependent
   - Requires significant refactoring

3. **Network Interface** (Partially Complete)
   - Core network functions at 100%
   - Loop-based watchers need extraction

### Phase 4: Configurable Constants (Complete - 57.8%)
**Goal**: Make remaining constants configurable

1. ✅ Made `stratuxHome` configurable in config_paths.go
2. ✅ Added helper functions: `getMapdataPath()`, `getMapdataStylesPath()`, `getOgnPath()`
3. ✅ Updated tests to use temporary directories
4. ✅ lookupOgnTailNumber: 42.9% → 100%
5. ✅ loadTile: 9.5% → 85.7%

### Phase 5: Loop Extraction (Target: 65%)
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
