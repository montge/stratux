# Stratux Test Data Directory

This directory contains legacy test data files used for manual testing with replay utilities in `/test/`.

**Status:** Phase 3.2 - Test Data Audit Complete
**Date:** 2025-10-14

---

## Migration Status

🚧 **IN PROGRESS**: Converting legacy test data to automated integration tests using trace file replay methodology.

**Migration Strategy:**
1. Extract representative test cases from each log file
2. Convert to gzipped CSV trace files in `/main/testdata/`
3. Create integration tests using trace replay functions
4. Verify automated tests match manual test behavior
5. Deprecate legacy test data files after successful conversion

---

## File Inventory

### 1. cyoung-09062015-noproblem-stratux-uat.log
**Size:** 6.7 MB (22,556 lines)
**Format:** CSV with timestamps and UAT messages
**Description:** "No problem" scenario - good UAT reception from September 2, 2015

**Message Format:**
```
START,Wed Sep 2 04:23:37 +0000 UTC 2015
162698355928,-0b2b4a513c417b8875f40d9610f227401105c4e6c4e6c40a32700300000000000000;rs=4;
162830235042,+3c2643887cdcab801f00067457403455014a02c1492830db2c75c9a8...;rs=13;
```

**Contents:**
- START marker: Session timestamp
- Timestamp (nanoseconds): 162698355928, 162830235042, ...
- Message type: `+` (uplink from ground station), `-` (downlink from aircraft)
- Hex data: UAT frame payload
- Reed-Solomon errors: `;rs=N;` where N is error count
- Signal strength: `;ss=N` (optional)

**Messages:**
- Uplink messages (`+`): FIS-B weather data (NEXRAD, METARs, TAFs, AIRMETs, NOTAMs)
- Downlink messages (`-`): UAT traffic reports (basic and long reports)
- Mix of valid and corrupted messages (some with Reed-Solomon errors)

**Migration Status:** ✅ **PARTIALLY COMPLETE**
- Basic UAT parsing tests created: `/main/integration_uat_test.go` (398 lines, 11 test functions)
- Trace file generated: `/main/testdata/uat/basic_uat.trace.gz` (8 sample messages)
- **TODO:** Extract additional weather-specific scenarios:
  - NEXRAD decoding test cases
  - METAR parsing sequences
  - TAF message handling
  - AIRMET/NOTAM extraction
  - Multi-frame FIS-B products

**Priority:** HIGH (UAT is core functionality)

---

### 2. cyoung-09062015-noproblem-stratux-es.log
**Size:** 2.2 MB (44,481 lines)
**Format:** CSV with timestamps and SBS/BaseStation messages
**Description:** "No problem" scenario - good 1090ES reception from September 2, 2015

**Message Format:**
```
START,Wed Sep 2 04:23:37 +0000 UTC 2015
2936535519,MSG,8,,,ABDF3F,,,,,,,,,,,,,,,,,
2938233280,MSG,5,,,ABDF3F,,,,,,,8525,,,,,,,0,0,0,0
2941779027,MSG,3,,,ABDF3F,,,,,,,,5850,,,,,,,0,0,0,0
```

**Contents:**
- START marker: Session timestamp
- Timestamp (nanoseconds): 2936535519, 2938233280, ...
- SBS format: `MSG,type,,,ICAO,,,,...`
  - Type 3: Airborne position
  - Type 4: Airborne velocity
  - Type 5: Surveillance alt
  - Type 6: Surveillance ID
  - Type 8: Onground position

**Message Types:**
- **MSG,3**: Airborne position (lat, lon, altitude, alert, emergency, SPI)
- **MSG,4**: Airborne velocity (ground speed, track, vertical rate)
- **MSG,5**: Surveillance altitude (squawk, altitude, alert, emergency, SPI)
- **MSG,6**: Surveillance ID (squawk, alert, emergency, SPI)
- **MSG,8**: On-ground position (lat, lon, on-ground flag)

**Aircraft Tracked:** 100+ unique ICAO addresses

**Migration Status:** ❌ **TODO - Phase 3.4**
- Target: Create `/main/integration_1090es_test.go`
- Trace file: `/main/testdata/1090es/basic_1090es.trace.gz`
- Expected coverage: +3-5% main package
- Functions to test:
  - parseInput() for 1090ES
  - esListen()
  - SBS message parsing
  - Position calculation
  - Velocity decoding

**Priority:** HIGH (1090ES is core functionality)

---

### 3. gms5002-09072015-problem-stratux-uat.log
**Size:** 356 KB (1,373 lines)
**Format:** CSV with timestamps and UAT messages
**Description:** "Problem" scenario - poor UAT reception from September 7, 2015

**Characteristics:**
- Fewer messages than "no problem" scenario
- Higher Reed-Solomon error rates
- More corrupted/invalid messages
- Good for testing error handling

**Migration Status:** ❌ **TODO - Phase 3.3 Extension**
- Add as error scenario test case
- Test corrupt message handling
- Verify Reed-Solomon error reporting
- Check that invalid messages don't crash parser

**Priority:** MEDIUM (error handling validation)

---

### 4. gms5002-09072015-problem-stratux-es.log
**Size:** 51 KB (1,061 lines)
**Format:** CSV with timestamps and SBS messages
**Description:** "Problem" scenario - poor 1090ES reception from September 7, 2015

**Characteristics:**
- Sparse message rate
- Good for testing low-traffic scenarios
- Edge cases for position extrapolation

**Migration Status:** ❌ **TODO - Phase 3.4 Extension**
- Add as low-traffic test case
- Test sparse update handling
- Verify extrapolation limits

**Priority:** LOW (edge case testing)

---

### 5. example.dump978
**Size:** 1.8 MB (2,327 lines)
**Format:** Raw UAT messages (no timestamps)
**Description:** UAT uplink message samples for decoder testing

**Message Format:**
```
+3cc0978aa66ca1a0158000213c5d2082102c22cc00082eec1e012c22cc000000000...
+3c62ab89c854b370308000353f59682210000000ff005685d07c4d5060cb9c72d35833...
```

**Contents:**
- Uplink messages only (`+` prefix)
- 432-byte FIS-B frames (864 hex characters)
- Various weather products embedded
- No timestamps (static test data)

**Use Case:**
- UAT decoder unit testing
- FIS-B product extraction testing
- No need for timing simulation

**Migration Status:** ❌ **TODO - Phase 3.3 Extension**
- Extract unique weather product examples
- Create dedicated FIS-B product tests
- Test NEXRAD block extraction
- Test text report (METAR/TAF) parsing

**Priority:** MEDIUM (FIS-B product testing)

---

### 6. example.radar
**Size:** 2.7 MB (3,111 lines)
**Format:** Raw UAT NEXRAD messages (no timestamps)
**Description:** NEXRAD radar data blocks for weather display testing

**Message Format:**
```
+3d1583886136a0c0040000fc59e004157c10040000fc59e004c38300040000fc59e004...
```

**Contents:**
- NEXRAD uplink messages only
- 432-byte frames with radar blocks
- Compressed radar data
- Block location encoding

**Use Case:**
- NEXRAD block extraction testing
- Weather radar decoding
- Block location calculation (lat/lon to block ID)
- Intensity level parsing

**Migration Status:** ❌ **TODO - Phase 3.3 Extension**
- Create NEXRAD-specific integration tests
- Test block location calculation (already 100% coverage in uatparse)
- Test intensity level extraction
- Test weather proximity calculations

**Priority:** MEDIUM (weather display functionality)

---

## Conversion Methodology

### Trace File Format (gzipped CSV)

Automated integration tests use gzipped CSV trace files with this format:

```csv
timestamp,protocol,message_data
2025-10-14T12:00:00.000Z,uat,+3cc0978aa66ca1a0158000213c5d2082102c22cc00...
2025-10-14T12:00:00.500Z,uat,-0b2b4a513c417b8875f40d9610f227401105c4e6c4e6c4...
2025-10-14T12:00:01.000Z,1090es,MSG,3,,,ABDF3F,,,,,,,,5850,,,,,,,0,0,0,0
```

**Columns:**
1. **timestamp**: RFC3339Nano format (ISO 8601 with nanoseconds)
2. **protocol**: Message type (`uat`, `1090es`, `ogn`, `aprs`, `gps`)
3. **message_data**: Protocol-specific message payload

### Conversion Process

1. **Extract Representative Messages**
   - Select diverse message types
   - Include edge cases and errors
   - Keep trace files small (<20 messages)

2. **Create Trace Generator**
   - Write `generate_trace.go` in testdata subdirectory
   - Embed sample messages with relative timestamps
   - Generate gzipped CSV output

3. **Implement Replay Function**
   - Read trace file
   - Parse CSV records
   - Inject messages into parser
   - Track message counts

4. **Write Test Functions**
   - Test basic parsing
   - Test message type detection
   - Test error handling
   - Test state updates

5. **Verify Coverage**
   - Run `go test -cover`
   - Check coverage improvement
   - Ensure no regression

### Example: UAT Trace Generator

See `/main/testdata/uat/generate_trace.go` for reference implementation:

```go
messages := []struct {
    offsetMs int64
    data     string
}{
    {0, `+3cc0978aa66ca1a0158000...;rs=16;ss=128`},     // Uplink
    {500, `-000000000000000000000...;rs=12;ss=94`},    // Basic report
    {1000, `-00000000000000000000...;rs=14;ss=102`},   // Long report
}

for _, msg := range messages {
    ts := baseTime.Add(time.Duration(msg.offsetMs) * time.Millisecond)
    w.Write([]string{ts.Format(time.RFC3339Nano), "uat", msg.data})
}
```

---

## Phase 3 Roadmap

### Phase 3.1: `/test/` Directory Audit ✅ COMPLETE
- Documented all test utilities in `/test/README.md`
- Identified migration candidates
- Created comprehensive roadmap

### Phase 3.2: `/test-data/` Directory Audit ✅ COMPLETE (this document)
- Documented all test data files
- Analyzed message formats
- Planned conversion strategy

### Phase 3.3: UAT Integration Tests 🚧 PARTIAL
**Status:** Basic tests complete, weather products TODO

**Completed:**
- ✅ Basic UAT parsing tests (11 test functions)
- ✅ Message type detection (uplink/downlink/basic/long)
- ✅ Signal strength parsing
- ✅ Invalid message handling

**TODO:**
- ❌ NEXRAD block extraction tests
- ❌ METAR/TAF parsing tests
- ❌ AIRMET/NOTAM extraction tests
- ❌ Multi-frame FIS-B product assembly

**Expected Coverage:** +2-4% additional (beyond current 18.9%)

### Phase 3.4: 1090ES Integration Tests ❌ TODO
**Status:** Not started

**Tasks:**
- Create `/main/testdata/1090es/` directory
- Create `generate_trace.go` for 1090ES messages
- Generate `basic_1090es.trace.gz` with diverse message types
- Create `/main/integration_1090es_test.go`
- Implement replay function for SBS format
- Add test functions for:
  - SBS message parsing
  - Airborne position messages (MSG,3)
  - Velocity messages (MSG,4)
  - Surveillance altitude (MSG,5)
  - On-ground position (MSG,8)
  - Multi-target tracking

**Expected Coverage:** +3-5% main package

### Phase 3.5: GPS Integration Tests ❌ TODO
**Status:** Not started (no GPS logs in test-data yet)

**TODO:**
- Collect GPS NMEA sequences
- Create GPS trace files
- Test NMEA sentence parsing
- Test fix quality transitions

**Expected Coverage:** +2-4% main package

### Phase 3.6: End-to-End Integration Tests ❌ TODO
**Status:** Not started

**TODO:**
- Multi-protocol trace files (UAT + 1090ES + GPS)
- Traffic fusion scenarios
- Ownship detection accuracy
- Network output validation

**Expected Coverage:** +5-8% main package

---

## Deprecation Plan

Once automated integration tests are complete and verified:

1. **Move to Archive**
   ```bash
   mkdir -p test-data-archive
   mv test-data/*.log test-data-archive/
   mv test-data/example.* test-data-archive/
   ```

2. **Update Documentation**
   - Note deprecation in this README
   - Reference new automated tests
   - Keep archive for reference

3. **Update Replay Utilities**
   - Mark `/test/replay.go` as deprecated
   - Direct users to automated tests
   - Keep utility for debugging if needed

---

## References

- **Test Coverage Roadmap:** [/TEST_COVERAGE_ROADMAP.md](../TEST_COVERAGE_ROADMAP.md)
- **Test Utilities:** [/test/README.md](../test/README.md)
- **UAT Integration Tests:** [/main/integration_uat_test.go](../main/integration_uat_test.go)
- **UAT Trace Generator:** [/main/testdata/uat/generate_trace.go](../main/testdata/uat/generate_trace.go)
- **OGN Integration Tests:** [/main/integration_ogn_aprs_test.go](../main/integration_ogn_aprs_test.go)

---

**Last Updated:** 2025-10-14
**Phase:** 3.2 Complete - Test Data Audit
**Next Phase:** 3.4 - 1090ES Integration Tests
