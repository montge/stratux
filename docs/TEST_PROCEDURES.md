# Software Verification Plan (SVP)

**System**: Stratux ADS-B/UAT/OGN Receiver
**Document Type**: Software Verification Plan
**Standard**: DO-278A SAL-3 Compliance
**Document Version**: 1.0
**Date**: 2025-10-13
**Status**: Draft for Review

---

## 1. Introduction

### 1.1 Purpose

This Software Verification Plan (SVP) defines the verification strategy, procedures, and acceptance criteria for the Stratux system to achieve DO-278A Software Assurance Level 3 (SAL-3) compliance.

This document satisfies DO-278A Objective A-8: "Verification procedures are developed and recorded."

### 1.2 Scope

This plan covers:
- **Unit Testing**: Module-level verification
- **Integration Testing**: Subsystem interface verification
- **System Testing**: End-to-end functional verification
- **Structural Coverage Analysis**: Statement and decision coverage
- **Requirements-Based Testing**: Traceability from requirements to tests
- **Regression Testing**: Continuous verification of changes

### 1.3 Applicable Standards

- **DO-278A**: Guidelines for CNS/ATM Systems Software Integrity Assurance
- **DO-260B**: Minimum Operational Performance Standards for 1090 MHz ADS-B
- **DO-282B**: Minimum Operational Performance Standards for UAT ADS-B
- **RTCA/DO-178C**: Referenced for testing best practices (adapted for ground systems)

### 1.4 Software Assurance Level

**SAL-3 (Software Level C)** per DO-278A Section 2.3:
- **Failure Effect**: Major (incorrect advisory information)
- **Required Coverage**: Statement coverage + Decision coverage
- **MC/DC Coverage**: Not required for SAL-3

---

## 2. Verification Strategy

### 2.1 Verification Levels

```
┌─────────────────────────────────────────────────┐
│              Requirements (101)                  │
│  FR-101 to FR-604, NFR-101 to NFR-703, etc.     │
└─────────────────┬───────────────────────────────┘
                  │
                  ├──► Unit Tests (Target: 80% coverage)
                  │    - Function-level verification
                  │    - Mock external dependencies
                  │    - Automated via CI/CD
                  │
                  ├──► Integration Tests (Target: All interfaces)
                  │    - Subsystem interaction verification
                  │    - Hardware-in-loop where practical
                  │    - Replay-based regression
                  │
                  ├──► System Tests (Target: 100% requirements)
                  │    - End-to-end functional verification
                  │    - Performance and stress testing
                  │    - Field testing with real hardware
                  │
                  └──► Structural Coverage Analysis
                       - Statement coverage ≥80%
                       - Decision coverage ≥70%
                       - Automated via go test -cover
```

### 2.2 Test Categories

| Category | Purpose | Execution | Coverage Target |
|----------|---------|-----------|-----------------|
| **Unit Tests** | Verify individual functions/modules | Every commit (CI/CD) | 80% statement |
| **Integration Tests** | Verify subsystem interfaces | Nightly + pre-release | All interfaces |
| **System Tests** | Verify end-to-end functionality | Weekly + pre-release | 100% requirements |
| **Performance Tests** | Verify timing and throughput | Pre-release | All NFRs |
| **Regression Tests** | Verify no new defects | Every commit | Changed code |
| **Field Tests** | Verify real-world operation | Pre-release (alpha) | Key scenarios |

### 2.3 Verification Independence

Per DO-278A SAL-3 requirements:
- **Test Design**: May be performed by developers with peer review
- **Test Execution**: Automated (CI/CD) with manual verification for system tests
- **Test Review**: Independent review of test procedures and results
- **Test Coverage Analysis**: Automated tools (go test -cover, go tool cover)

---

## 3. Test Infrastructure

### 3.1 Test Framework

**Primary Framework**: Go built-in testing package
- `go test` for test execution
- `testing.T` for unit tests
- `testing.B` for benchmarks
- Table-driven tests for multiple scenarios

**Coverage Tools**:
```bash
# Generate coverage profile
go test -v -coverprofile=coverage.out -covermode=atomic ./main/... ./common/...

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Display function-level coverage
go tool cover -func=coverage.out
```

### 3.2 Test Data Management

**Location**: `main/testdata/` (per Go conventions)

**Organization**:
```
main/testdata/
├── README.md              # Test data documentation
├── adsb/                  # Sample ADS-B messages (1090 MHz)
│   ├── sample_traffic.bin
│   └── edge_cases.bin
├── uat/                   # Sample UAT messages (978 MHz)
│   ├── sample_uplink.bin
│   └── sample_downlink.bin
├── gps/                   # Sample GPS NMEA sentences
│   ├── valid_position.txt
│   └── invalid_data.txt
├── gdl90/                 # Expected GDL90 output
│   ├── heartbeat.bin
│   └── traffic_report.bin
└── ogn/                   # Sample OGN/FLARM messages
    └── sample_ogn.txt
```

### 3.3 Mock Objects

**Purpose**: Isolate units under test from external dependencies

**Mock Requirements**:
- GPS receivers (serial port I/O)
- SDR hardware (USB devices)
- Network connections (TCP/UDP sockets)
- SQLite database
- System clocks (time.Time for deterministic testing)

**Implementation**: Interface-based mocking in `main/*_test.go` files

### 3.4 CI/CD Integration

**GitHub Actions Workflow** (`.github/workflows/ci.yml`):

```yaml
- name: Run tests with coverage
  run: |
    go test -v -coverprofile=coverage.out -covermode=atomic ./main/...
    go test -v ./common/...

- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$COVERAGE < 80.0" | bc -l) )); then
      echo "::warning::Coverage ${COVERAGE}% is below target of 80%"
    fi

- name: Run static analysis
  run: |
    go vet ./main/...
    go vet ./common/...
```

---

## 4. Unit Testing Procedures

### 4.1 Unit Test Naming Conventions

**Test File Naming**: `<source_file>_test.go`
- Example: `traffic.go` → `traffic_test.go`

**Test Function Naming**: `Test<FunctionName>_<Scenario>`
- Examples:
  - `TestIsTrafficAlertable_WithinRange`
  - `TestIcao2reg_USCivil`
  - `TestExtrapolateTraffic_ValidHeading`

### 4.2 Test Structure (Table-Driven)

```go
func TestFunctionName_Scenario(t *testing.T) {
    // Verifies: FR-XXX (Requirement ID)
    testCases := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        // ... more test cases
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := FunctionName(tc.input)

            if (err != nil) != tc.wantErr {
                t.Errorf("unexpected error: %v", err)
            }

            if result != tc.expected {
                t.Errorf("got %v, want %v", result, tc.expected)
            }
        })
    }
}
```

### 4.3 Unit Test Coverage Requirements

Per DO-278A SAL-3 and industry best practices:

| Module | Statement Coverage | Decision Coverage | Priority |
|--------|-------------------|-------------------|----------|
| **traffic.go** | 90% | 85% | CRITICAL |
| **gps.go** | 85% | 80% | CRITICAL |
| **gen_gdl90.go** | 85% | 80% | CRITICAL |
| **sdr.go** | 80% | 75% | HIGH |
| **sensors.go** (AHRS) | 75% | 70% | HIGH |
| **network.go** | 75% | 70% | MEDIUM |
| **Other modules** | 70% | 65% | MEDIUM |
| **Overall Target** | **80%** | **70%** | - |

### 4.4 Test Conditions

Each unit test SHALL verify:
1. **Normal Operation**: Valid inputs, expected outputs
2. **Boundary Conditions**: Min/max values, edge cases
3. **Error Conditions**: Invalid inputs, error handling
4. **State Transitions**: Mode changes, status updates

**Example Test Matrix for Traffic Alerting** (FR-407):

| Test Case | Distance (m) | BearingValid | Expected Alertable | Requirement |
|-----------|-------------|--------------|-------------------|-------------|
| TC-407-01 | 3703 | true | true | Within 2nm threshold |
| TC-407-02 | 3705 | true | false | Beyond 2nm threshold |
| TC-407-03 | 10000 | false | true | Conservative (no bearing) |
| TC-407-04 | 0 | true | true | Collision threat |

---

## 5. Integration Testing Procedures

### 5.1 Integration Test Scenarios

#### 5.1.1 ADS-B Message Pipeline

**Test**: `TestADSBPipeline_EndToEnd`

**Data Flow**: RF Input → Demodulation → Decoding → Fusion → GDL90 Output

**Verification**:
1. Inject sample 1090ES message
2. Verify demodulation produces valid bits
3. Verify decoding extracts correct fields
4. Verify traffic fusion updates target
5. Verify GDL90 traffic report generated
6. Verify CRC correctness

**Requirements Verified**: FR-401, FR-402, FR-601, FR-602

#### 5.1.2 UAT Message Pipeline

**Test**: `TestUATPipeline_EndToEnd`

**Data Flow**: RF Input → Demodulation → Decoding → Fusion → GDL90 Output

**Requirements Verified**: FR-403, FR-404, FR-603, FR-604

#### 5.1.3 GPS Integration

**Test**: `TestGPSIntegration_PositionUpdate`

**Data Flow**: GPS NMEA → Parsing → Validation → Ownship Update → GDL90 Ownship Report

**Requirements Verified**: FR-201, FR-202, FR-203, FR-605

#### 5.1.4 Multi-Source Traffic Fusion

**Test**: `TestTrafficFusion_MultipleSources`

**Scenario**: Same aircraft reported by 1090ES, UAT, and OGN

**Verification**:
1. Three sources report same ICAO address
2. System correctly fuses into single target
3. Highest-quality data selected
4. No duplicate traffic displayed

**Requirements Verified**: FR-405, FR-407

### 5.2 Integration Test Execution

**Frequency**:
- Nightly automated execution
- Manual execution before each release

**Test Data**: Pre-captured message logs from real flights

**Test Environment**:
- CI/CD: Simulated inputs via mock interfaces
- Local: Replay captured data via rtl_sdr file input
- Field: Live hardware with test aircraft/ground stations

---

## 6. System Testing Procedures

### 6.1 Functional System Tests

#### 6.1.1 Traffic Display Test (ST-TRAFFIC-01)

**Objective**: Verify complete traffic alerting functionality

**Procedure**:
1. Power on Stratux with functional SDR
2. Connect EFB app (ForeFlight, iFly, etc.)
3. Verify traffic appears on EFB display
4. Verify relative position accuracy (compare to ATC radar)
5. Verify alerting logic (targets within 2nm)
6. Verify ownship filtering (own aircraft not shown)

**Pass Criteria**:
- All traffic within range displayed
- Position accuracy ±0.1nm
- Alert logic correct (FR-407)
- Ownship correctly filtered (FR-410)

**Requirements Verified**: FR-401 through FR-410

#### 6.1.2 Weather Display Test (ST-WEATHER-01)

**Objective**: Verify FIS-B weather reception

**Procedure**:
1. Power on Stratux in area with FIS-B coverage
2. Connect EFB app
3. Verify NEXRAD radar display
4. Verify METARs displayed
5. Verify TAFs displayed
6. Verify age indication (>15 min warning)

**Pass Criteria**:
- Weather data displayed correctly
- Age indication accurate (FR-503)
- Stale data indicated (NFR-601)

**Requirements Verified**: FR-501 through FR-507

#### 6.1.3 GPS Accuracy Test (ST-GPS-01)

**Objective**: Verify GPS position accuracy

**Procedure**:
1. Connect GPS receiver (VK-172 or similar)
2. Allow GPS lock (cold start up to 2 minutes)
3. Record ownship position
4. Compare to known surveyed position
5. Verify WAAS augmentation active
6. Verify altitude accuracy

**Pass Criteria**:
- Horizontal accuracy <10m with WAAS (FR-202)
- Altitude accuracy ±100ft (NFR-201)
- Lock time <2 minutes cold start (NFR-202)

**Requirements Verified**: FR-201, FR-202, FR-203, NFR-201, NFR-202

### 6.2 Performance System Tests

#### 6.2.1 High Traffic Load Test (ST-PERF-01)

**Objective**: Verify system handles high message rates

**Procedure**:
1. Inject 500 messages/second (simulated busy airspace)
2. Monitor CPU usage
3. Monitor memory usage
4. Verify no message drops
5. Verify GDL90 output latency <1 second

**Pass Criteria**:
- All messages processed (FR-405)
- CPU usage <75% (NFR-401)
- Memory usage stable (no leaks)
- Output latency <1 second (NFR-402)

**Requirements Verified**: FR-405, NFR-401, NFR-402

#### 6.2.2 Multiple Client Test (ST-PERF-02)

**Objective**: Verify system supports multiple simultaneous clients

**Procedure**:
1. Connect 5 EFB clients simultaneously
2. Verify all receive traffic updates
3. Verify no client starvation
4. Measure bandwidth usage

**Pass Criteria**:
- All clients receive data (FR-301)
- Update rate ≥1 Hz to each client (NFR-403)
- Bandwidth usage reasonable

**Requirements Verified**: FR-301, FR-302, NFR-403

#### 6.2.3 Endurance Test (ST-PERF-03)

**Objective**: Verify system stability over extended operation

**Procedure**:
1. Power on Stratux
2. Run continuously for 8 hours
3. Monitor for crashes, hangs, errors
4. Monitor resource usage trends

**Pass Criteria**:
- No crashes or hangs (NFR-602)
- No memory leaks (stable memory usage)
- No error accumulation

**Requirements Verified**: NFR-602, NFR-701

### 6.3 Failure Mode Tests

#### 6.3.1 GPS Loss Test (ST-FAIL-01)

**Objective**: Verify graceful degradation when GPS lost

**Procedure**:
1. Start with GPS locked
2. Disconnect GPS receiver
3. Verify system indicates GPS loss (FR-203)
4. Verify ownship reports cease
5. Verify traffic continues to function

**Pass Criteria**:
- GPS loss indicated on UI
- Ownship reports stop (no false position)
- Traffic continues normally

**Requirements Verified**: FR-203, FR-302

#### 6.3.2 SDR Failure Test (ST-FAIL-02)

**Objective**: Verify recovery from SDR disconnection

**Procedure**:
1. Start with SDR functional
2. Disconnect SDR USB
3. Verify error indication (FR-302)
4. Reconnect SDR
5. Verify automatic recovery

**Pass Criteria**:
- Error indicated within 5 seconds
- Automatic reconnection occurs
- Traffic resumes after recovery

**Requirements Verified**: FR-302, NFR-602

---

## 7. Structural Coverage Analysis

### 7.1 Statement Coverage

**Definition**: Every executable statement executed at least once

**Measurement**: `go test -covermode=atomic`

**Target**: ≥80% overall, ≥90% for safety-critical modules

**Exclusions** (justified):
- Unreachable error handlers (defensive coding)
- Debug logging statements
- OS-specific code not relevant to target platform

### 7.2 Decision Coverage

**Definition**: Every decision (if/else, switch) evaluated to both TRUE and FALSE

**Measurement**: Manual analysis + review of branch coverage

**Target**: ≥70% overall, ≥85% for safety-critical modules

**Example Decision Coverage Analysis**:

```go
// Function: isTrafficAlertable
// Decisions: 2
//   D1: if !ti.BearingDist_valid
//   D2: if ti.Distance <= TRAFFIC_ALERT_DISTANCE
//
// Required Test Cases:
//   TC1: BearingDist_valid=false → D1=TRUE  (return true)
//   TC2: BearingDist_valid=true, Distance=3703 → D1=FALSE, D2=TRUE (return true)
//   TC3: BearingDist_valid=true, Distance=3705 → D1=FALSE, D2=FALSE (return false)
//
// Decision Coverage: 100% (all outcomes tested)
```

### 7.3 Coverage Reporting

**Automated Reports**: Generated by CI/CD on every commit

**Report Contents**:
1. Overall coverage percentage (statement + decision)
2. Per-file coverage breakdown
3. Uncovered lines highlighted
4. Coverage trend over time
5. Comparison to target thresholds

**Report Location**:
- GitHub Actions artifacts
- Coverage HTML: `coverage.html`
- Coverage text: `coverage.txt`

### 7.4 Coverage Gaps

**Handling of Uncovered Code**:
1. **Identify**: List all uncovered functions/branches
2. **Analyze**: Determine reason for non-coverage
3. **Justify or Fix**:
   - Write additional tests if testable
   - Document justification if untestable (e.g., hardware-specific, unreachable)
4. **Review**: Independent review of coverage gaps

**Coverage Gap Report** (example format):

| File | Line | Coverage | Justification |
|------|------|----------|---------------|
| sdr.go | 234-238 | 0% | Hardware-specific error path (RTL-SDR firmware bug) |
| network.go | 456 | 0% | Defensive error check (cannot trigger in test) |

---

## 8. Requirements-Based Testing

### 8.1 Traceability Matrix

**Purpose**: Map every requirement to one or more test cases

**Format**: Spreadsheet (Excel/CSV) with columns:
- Requirement ID (FR-XXX, NFR-XXX)
- Requirement Description
- Test Case IDs
- Test Level (Unit/Integration/System)
- Status (Pass/Fail/Not Run)
- Last Verified Date

**Example Traceability**:

| Req ID | Description | Test Cases | Level | Status | Date |
|--------|-------------|------------|-------|--------|------|
| FR-407 | Traffic Alerting | TC-407-01, TC-407-02, TC-407-03, ST-TRAFFIC-01 | Unit, System | Pass | 2025-10-13 |
| FR-410 | Ownship Filtering | TC-410-01, TC-410-02, ST-TRAFFIC-01 | Unit, System | Pass | 2025-10-13 |
| NFR-401 | CPU Usage <75% | ST-PERF-01 | System | Pass | 2025-10-13 |

**Traceability Requirements**:
- ✅ Every requirement has ≥1 test case
- ✅ Every test case traces to ≥1 requirement
- ✅ Bidirectional traceability maintained

### 8.2 Test Case Documentation

**Test Case Template**:

```
Test Case ID: TC-XXX-NN
Requirement: FR-XXX
Test Level: Unit / Integration / System
Priority: Critical / High / Medium / Low

Objective:
  One-sentence description of what is being verified

Preconditions:
  - System state before test
  - Required hardware/software setup

Test Data:
  - Input values
  - Expected outputs
  - Boundary conditions

Procedure:
  1. Step 1
  2. Step 2
  ...

Pass Criteria:
  - Specific measurable criteria
  - Acceptable tolerances

Dependencies:
  - Other test cases that must pass first
  - External dependencies
```

---

## 9. Regression Testing

### 9.1 Regression Test Strategy

**Trigger**: Every code commit (automated via CI/CD)

**Scope**:
1. All unit tests
2. All integration tests
3. Subset of system tests (smoke tests)

**Execution Time Target**: <10 minutes for unit tests, <1 hour for full suite

### 9.2 Regression Test Selection

**Criteria for Inclusion**:
- Previously failing tests (defect regression)
- Tests covering changed code
- Safety-critical module tests (always run)
- Smoke tests for core functionality

**Test Prioritization**:
1. **P0 (Critical)**: Safety-critical functions, must pass to merge
2. **P1 (High)**: Core functionality, must pass before release
3. **P2 (Medium)**: Extended functionality, run nightly
4. **P3 (Low)**: Edge cases, run weekly

### 9.3 Continuous Integration

**GitHub Actions Workflow**:
```yaml
on:
  push:
    branches: [ master ]
  pull_request:
    branches: [ master ]

jobs:
  test:
    runs-on: ubuntu-24.04-arm
    steps:
      - uses: actions/checkout@v4
      - name: Run unit tests
        run: go test -v -coverprofile=coverage.out ./main/... ./common/...
      - name: Run static analysis
        run: go vet ./main/... ./common/...
      - name: Check formatting
        run: gofmt -l ./main ./common
      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage.out
```

---

## 10. Test Execution and Reporting

### 10.1 Test Execution Schedule

| Test Type | Frequency | Trigger | Duration |
|-----------|-----------|---------|----------|
| Unit Tests | Every commit | Push/PR | 5 min |
| Integration Tests | Nightly | Scheduled | 30 min |
| System Tests (Smoke) | Nightly | Scheduled | 1 hour |
| System Tests (Full) | Weekly | Scheduled | 4 hours |
| Performance Tests | Pre-release | Manual | 2 hours |
| Field Tests | Pre-release | Manual | Varies |

### 10.2 Test Result Recording

**Automated Tests**: Results stored in GitHub Actions artifacts
- Test logs (stdout/stderr)
- Coverage reports (HTML + text)
- Failure screenshots (if applicable)
- Timing data

**Manual Tests**: Results recorded in test log spreadsheet
- Test Case ID
- Execution Date
- Tester Name
- Result (Pass/Fail/Blocked)
- Notes/Observations
- Defect IDs (if failed)

### 10.3 Test Metrics

**Collected Metrics**:
1. **Test Pass Rate**: (Passed / Total) × 100%
2. **Code Coverage**: Statement and decision coverage %
3. **Defect Detection Rate**: Defects found per test hour
4. **Test Execution Time**: Time to run full suite
5. **Mean Time To Fix**: Average time from defect detection to fix

**Target Metrics** (DO-278A SAL-3):
- Test Pass Rate: ≥99%
- Code Coverage: ≥80% statement, ≥70% decision
- Defect Detection Rate: Tracked but no target (higher is better early)
- Test Execution Time: <1 hour for full suite
- Mean Time To Fix: <48 hours for critical defects

### 10.4 Defect Management

**Defect Classification**:
- **Critical**: System crash, data corruption, safety impact
- **High**: Feature non-functional, major incorrect behavior
- **Medium**: Feature degraded, minor incorrect behavior
- **Low**: Cosmetic, documentation, minor usability

**Defect Workflow**:
1. **Detection**: Test fails or issue reported
2. **Recording**: Create GitHub Issue with defect template
3. **Triage**: Assign severity and priority
4. **Assignment**: Developer assigned to fix
5. **Fix**: Code change with unit test
6. **Verification**: Original test re-executed
7. **Closure**: Issue closed with verification note

**Defect Tracking Fields**:
- Issue ID
- Title
- Severity (Critical/High/Medium/Low)
- Status (New/Assigned/Fixed/Verified/Closed)
- Detected By (test case or reporter)
- Detected In (version)
- Fixed In (version)
- Root Cause (coding error, requirement defect, etc.)

---

## 11. Test Environment

### 11.1 Development Test Environment

**Hardware**:
- Developer workstation (Linux/macOS/Windows with WSL2)
- Raspberry Pi 4 (2GB+ RAM) for target testing
- RTL-SDR USB dongle (RTL2832U chipset)
- GPS receiver (VK-172 or equivalent)

**Software**:
- Go 1.21 or later
- Git for version control
- GitHub Actions for CI/CD
- Text editor / IDE (VSCode, GoLand, etc.)

**Network**:
- Internet access for go mod download
- Local network for multi-client testing

### 11.2 CI/CD Test Environment

**GitHub Actions Runner**: `ubuntu-24.04-arm`

**Installed Tools**:
- Go 1.21
- gcc (for CGO)
- make
- libusb-1.0-0-dev
- librtlsdr-dev (custom build)

**Limitations**:
- No actual SDR hardware (use mocks)
- No actual GPS hardware (use mocks)
- Network simulation limited

### 11.3 Field Test Environment

**Hardware**:
- Stratux complete build (RPi + SDR + GPS + battery)
- Aircraft installation (portable or semi-permanent)
- Multiple EFB devices (iPad, Android tablet)

**Software**:
- Latest release candidate build
- EFB apps: ForeFlight, iFly GPS, WingX, FltPlan Go, Garmin Pilot

**Conditions**:
- Various phases of flight (ground, taxi, takeoff, cruise, landing)
- Various traffic densities (light, moderate, heavy)
- Various weather conditions (VFR, MVFR, IMC)

---

## 12. Test Documentation

### 12.1 Test Plan (This Document)

**Purpose**: Define overall verification strategy

**Review**: Annually or when major changes to test approach

### 12.2 Test Cases

**Location**:
- Unit tests: Embedded in `*_test.go` files
- Integration/System tests: `docs/test_cases/` directory

**Format**: Markdown files with structured test case sections

### 12.3 Test Results

**Location**:
- Automated: GitHub Actions artifacts (30-day retention)
- Manual: `docs/test_results/` directory

**Format**:
- Automated: TAP (Test Anything Protocol) or JSON
- Manual: Spreadsheet (Excel/CSV)

### 12.4 Coverage Reports

**Location**: GitHub Actions artifacts, `coverage.html`

**Contents**:
- Overall coverage percentage
- Per-file coverage with color coding
- Uncovered lines highlighted
- Historical trend graph

### 12.5 Verification Report

**Purpose**: Summarize verification results for compliance audit

**Contents**:
1. Verification approach summary
2. Test execution statistics
3. Coverage analysis results
4. Traceability matrix summary
5. Open defects and status
6. Compliance statement

**Review**: Quarterly and before each major release

---

## 13. Roles and Responsibilities

### 13.1 Test Engineer

**Responsibilities**:
- Develop test plan and test cases
- Execute manual system tests
- Analyze test results and coverage
- Maintain traceability matrix
- Report defects

**Skills Required**:
- Aviation domain knowledge (ADS-B, GDL90)
- Go programming language
- Test automation
- Requirements analysis

### 13.2 Software Developer

**Responsibilities**:
- Write unit tests for new code
- Fix defects
- Review test cases
- Achieve coverage targets
- Support test automation

**Skills Required**:
- Go programming language
- Software testing principles
- Git version control

### 13.3 QA Reviewer (Independent)

**Responsibilities**:
- Review test plan and procedures
- Review test results
- Verify traceability completeness
- Approve verification report
- Support audit activities

**Skills Required**:
- DO-278A knowledge
- Quality assurance principles
- Aviation systems

### 13.4 Project Manager

**Responsibilities**:
- Allocate resources for testing
- Monitor test progress
- Approve verification report
- Coordinate third-party audit

---

## 14. Configuration Management

### 14.1 Test Code Configuration

**Version Control**: All test code stored in Git repository

**Branching**:
- Test code on same branch as source code
- No separate test repository

**Tagging**:
- Release tags include all test code
- Test results archived per release

### 14.2 Test Data Configuration

**Version Control**: Test data in `main/testdata/` under version control

**Naming Convention**:
- `<module>_<scenario>_<date>.bin`
- Example: `adsb_high_traffic_20251013.bin`

**Documentation**: `testdata/README.md` describes each test data file

### 14.3 Test Environment Configuration

**Documentation**: `docs/dev_setup.md` describes test environment setup

**Containerization**: Consider Docker for reproducible test environments (future)

---

## 15. Schedule and Milestones

### 15.1 Phase 1: Foundation (Current - Month 2)

- [x] Test plan created (this document)
- [x] Go test framework established
- [x] CI/CD configured for automated testing
- [x] First 10 unit tests written (traffic.go)
- [ ] Test data fixtures created
- [ ] Mock objects implemented

**Target Date**: 2025-11 (2 weeks remaining)

### 15.2 Phase 2: Safety-Critical Modules (Months 3-4)

- [ ] Traffic fusion: 90% coverage
- [ ] GPS processing: 85% coverage
- [ ] GDL90 generation: 85% coverage
- [ ] ADS-B decoding: 80% coverage
- [ ] AHRS processing: 75% coverage

**Target Date**: 2025-12 to 2026-01

### 15.3 Phase 3: Remaining Modules (Months 5-6)

- [ ] Network management: 75% coverage
- [ ] Weather processing: 75% coverage
- [ ] OGN processing: 75% coverage
- [ ] Configuration: 75% coverage
- [ ] Utilities: 75% coverage
- [ ] Overall target: 80% coverage achieved

**Target Date**: 2026-02 to 2026-03

### 15.4 Phase 4: Verification Complete (Months 6-7)

- [ ] Traceability matrix complete
- [ ] All requirements verified
- [ ] Verification report published
- [ ] Ready for third-party audit

**Target Date**: 2026-03 to 2026-04

---

## 16. Compliance Statement

This Software Verification Plan satisfies the following DO-278A SAL-3 objectives:

- **A-8**: Verification procedures are developed and recorded ✅
- **A-9**: Test coverage of requirements (plan defined, execution in progress)
- **A-10**: Test coverage of structure (plan defined, target 80% statement, 70% decision)
- **A-11**: Verification of integration (integration test plan defined)

**Current Verification Status**: Plan Complete, Execution Beginning (1.8% coverage baseline established)

**Target Verification Status**: 100% requirements verified, 80% code coverage, all tests passing

---

## 17. References

1. **DO-278A**: Guidelines for Communication, Navigation, Surveillance and Air Traffic Management (CNS/ATM) Systems Software Integrity Assurance, RTCA, Inc., 2011
2. **DO-260B**: Minimum Operational Performance Standards for 1090 MHz Extended Squitter ADS-B, RTCA, Inc., 2009
3. **DO-282B**: Minimum Operational Performance Standards for UAT ADS-B, RTCA, Inc., 2009
4. **DO-178C**: Software Considerations in Airborne Systems and Equipment Certification (referenced for testing best practices)
5. **Stratux Requirements Specification**: `docs/REQUIREMENTS.md`
6. **Stratux High-Level Design**: `docs/HIGH_LEVEL_DESIGN.md`
7. **Stratux Software Development Process**: `docs/SOFTWARE_DEVELOPMENT_PROCESS.md`
8. **Stratux Configuration Management Plan**: `docs/CONFIGURATION_MANAGEMENT_PLAN.md`

---

## 18. Document Control

**Document Owner**: QA/Test Team
**Approval Authority**: Project Manager
**Review Cycle**: Annually or when major changes
**Next Review Date**: 2026-10-13

**Revision History**:

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-13 | Requirements Engineering | Initial creation, DO-278A SAL-3 compliant test plan |

---

**END OF SOFTWARE VERIFICATION PLAN**
