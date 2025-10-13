# Stratux Software Development Process

**Document Version**: 1.0
**Last Updated**: 2025-10-13
**Status**: Active
**Compliance**: DO-278A SAL-3

---

## 1. Introduction

### 1.1 Purpose

This document defines the Software Development Process for the Stratux ADS-B/UAT/OGN receiver system. It establishes the practices, procedures, and standards used to develop, verify, and maintain software in compliance with DO-278A Software Assurance Level 3 (SAL-3) requirements.

### 1.2 Scope

This process applies to all software components of the Stratux system, including:
- Core application software (Go)
- System configuration scripts (Shell)
- Web interface (HTML/JavaScript)
- Build and deployment automation
- Documentation

### 1.3 Applicable Standards

| Standard | Title | Version | Reference |
|----------|-------|---------|-----------|
| DO-278A | Software Integrity Assurance Considerations for CNS/ATM Systems | 2011 | RTCA DO-278A |
| DO-260B | 1090 MHz Extended Squitter ADS-B MOPS | 2009 | RTCA DO-260B, Section 2 |
| DO-282B | UAT ADS-B MOPS | 2009 | RTCA DO-282B, Section 2 |
| GDL90 | GDL 90 Data Interface Specification | Rev A, 2007 | FAA 560-1058-00 Rev A |
| ESASSP | EUROCONTROL Specification for ATM Surveillance System Performance | Vol 1-2, 2023 | EUROCONTROL |
| ASTERIX | Category 021 ADS-B Target Reports | v2.6, 2021 | EUROCONTROL |

**Official Sources**:
- FAA GDL90 ICD: https://www.faa.gov/sites/faa.gov/files/air_traffic/technology/adsb/archival/GDL90_Public_ICD_RevA.PDF
- EUROCONTROL ESASSP: https://www.eurocontrol.int/publication/eurocontrol-specification-atm-surveillance-system-performance-esassp
- DO-278A: Available from RTCA, Inc. (https://www.rtca.org/)
- DO-260B/DO-282B: Available from RTCA, Inc., incorporated by reference in 14 CFR 91.225/91.227

### 1.4 Definitions

- **SAL-3**: Software Assurance Level 3 (equivalent to Software Level C per DO-178C)
- **CNS/ATM**: Communication, Navigation, Surveillance / Air Traffic Management
- **SRS**: System Requirements Specification
- **HLD**: High-Level Design
- **ADS-B**: Automatic Dependent Surveillance-Broadcast
- **UAT**: Universal Access Transceiver (978 MHz)
- **GDL90**: Garmin Data Link 90 protocol

---

## 2. Software Development Lifecycle

### 2.1 Lifecycle Model

Stratux follows an **Iterative Incremental** development model with continuous integration:

```
┌─────────────────────────────────────────────────────────┐
│                    PLANNING PHASE                        │
│  • Requirements Analysis                                 │
│  • Risk Assessment                                       │
│  • Sprint Planning                                       │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│                  DEVELOPMENT PHASE                       │
│  • Design (High-Level & Detailed)                       │
│  • Implementation (Coding)                               │
│  • Unit Testing                                          │
│  • Code Review                                           │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│                 VERIFICATION PHASE                       │
│  • Integration Testing                                   │
│  • System Testing                                        │
│  • Requirements Verification                             │
│  • Coverage Analysis                                     │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│                   RELEASE PHASE                          │
│  • Release Candidate Build                               │
│  • Field Testing (Beta)                                  │
│  • Documentation Update                                  │
│  • Release Approval                                      │
└───────────────┬─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│                 MAINTENANCE PHASE                        │
│  • Bug Fixes                                             │
│  • Security Updates                                      │
│  • Performance Improvements                              │
│  • Change Impact Analysis                                │
└─────────────────────────────────────────────────────────┘
```

### 2.2 Phase Entry/Exit Criteria

#### Planning Phase
**Entry Criteria**:
- Feature request or issue identified
- Stakeholder approval (if new feature)

**Exit Criteria**:
- Requirements documented in `docs/REQUIREMENTS.md`
- Requirement IDs assigned (FR-xxx, NFR-xxx, SR-xxx, SEC-xxx)
- Acceptance criteria defined
- Design impact assessed

#### Development Phase
**Entry Criteria**:
- Requirements approved
- Design review completed (for significant changes)
- Developer assigned

**Exit Criteria**:
- Code implementation complete
- Unit tests written (target: 80% coverage of new/modified code)
- Code adheres to coding standards
- Peer review completed
- All compiler warnings resolved

#### Verification Phase
**Entry Criteria**:
- Development phase complete
- Unit tests passing
- Build successful

**Exit Criteria**:
- Integration tests passing
- Requirements traced to test cases
- Code coverage ≥80% (statement), ≥70% (decision)
- No unresolved critical or high-severity defects
- Regression tests passing

#### Release Phase
**Entry Criteria**:
- Verification phase complete
- Release notes prepared
- Version number assigned

**Exit Criteria**:
- Release candidate tested on target hardware
- Field testing completed (for major releases)
- Documentation updated
- Release artifacts published (GitHub Releases)

---

## 3. Requirements Management

### 3.1 Requirements Development

All system requirements are documented in `docs/REQUIREMENTS.md` with the following attributes:

| Attribute | Description | Example |
|-----------|-------------|---------|
| **ID** | Unique requirement identifier | FR-101 |
| **Category** | FR (Functional), NFR (Non-Functional), SR (Safety), SEC (Security) | FR |
| **Title** | Short descriptive name | "1090 MHz ADS-B Reception" |
| **Priority** | CRITICAL, HIGH, MEDIUM, LOW | CRITICAL |
| **Source** | Origin of requirement | DO-260B Section 2.2.3.2 |
| **Verification** | Test, Analysis, Inspection, Demo | Test |
| **Acceptance Criteria** | Measurable criteria for compliance | "Decode 1090ES messages per DO-260B" |
| **Traceability** | Links to design/code/tests | Implemented in main/sdr.go:450-520 |

### 3.2 Requirements Change Control

All changes to requirements follow this process:

1. **Change Request**: Submit issue on GitHub with label `requirements-change`
2. **Impact Analysis**: Assess impact on design, code, tests, documentation
3. **Approval**: Requires approval from project maintainer
4. **Update Documentation**: Modify `docs/REQUIREMENTS.md` with change history
5. **Update Traceability**: Update `docs/TRACEABILITY_MATRIX.xlsx`
6. **Verification**: Update test cases to reflect new/modified requirements

### 3.3 Requirements Traceability

Requirements are traced bidirectionally:

- **Forward Traceability**: Requirements → Design → Code → Tests
- **Backward Traceability**: Tests → Code → Design → Requirements

Traceability is maintained in:
- `docs/TRACEABILITY_MATRIX.xlsx` - Master traceability spreadsheet
- Source code comments: `// Implements: FR-101, FR-102`
- Test code comments: `// Verifies: FR-101`

---

## 4. Design Process

### 4.1 Design Standards

#### High-Level Design (HLD)
All architectural decisions are documented in `docs/HIGH_LEVEL_DESIGN.md` including:
- System architecture diagrams (Mermaid)
- Data flow diagrams
- State machines
- Module interfaces
- Concurrency model
- Error handling strategy

**HLD Review Triggers**:
- New subsystem added
- Major architectural change
- Change to safety-critical module interfaces
- Performance optimization requiring architectural changes

#### Detailed Design
Detailed design is documented in code through:
- Package-level documentation (Go doc comments)
- Struct and interface definitions
- Function signatures with parameter descriptions
- Algorithm explanations for complex logic

### 4.2 Design Reviews

| Review Type | Trigger | Participants | Deliverable |
|-------------|---------|--------------|-------------|
| **Architecture Review** | New feature requiring HLD update | Lead developer, 1+ peer | Approved HLD update |
| **Code Review** | All pull requests | Author + 1 reviewer minimum | Approved PR |
| **Safety Review** | Changes to safety-critical modules* | Lead + safety-aware reviewer | Safety checklist completed |

*Safety-critical modules: `traffic.go`, `gps.go`, `gen_gdl90.go`, `sdr.go`, `sensors.go`

### 4.3 Design Patterns and Standards

**Go Design Principles**:
- Effective Go guidelines (https://go.dev/doc/effective_go)
- Clear, simple, idiomatic Go code
- Minimize use of global variables
- Use channels for goroutine communication
- Error handling: always check errors, propagate with context

**Concurrency**:
- Use goroutines for concurrent operations
- Synchronize with channels (preferred) or mutexes (when necessary)
- Document all shared data structures
- Avoid data races (verified with `go test -race`)

**Error Handling**:
- All errors logged to `globalStatus.Errors[]`
- Critical errors trigger automatic recovery
- Error context includes timestamp, module, severity

---

## 5. Implementation (Coding)

### 5.1 Coding Standards

#### General Principles
1. **Readability First**: Code is read far more often than written
2. **Self-Documenting**: Use clear variable/function names
3. **DRY (Don't Repeat Yourself)**: Extract common logic into functions
4. **YAGNI (You Aren't Gonna Need It)**: Don't add functionality prematurely
5. **Fail Fast**: Detect and report errors as early as possible

#### Go-Specific Standards
```go
// Package comment describes purpose
package main

import (
	// Standard library first
	"fmt"
	"time"

	// Third-party packages second
	"github.com/example/package"

	// Local packages last
	"stratux/common"
)

// Exported functions have complete doc comments
// ProcessADSBMessage decodes and validates a 1090 MHz ADS-B message.
// Implements: FR-101, FR-301
// Returns error if message fails CRC or validation.
func ProcessADSBMessage(msg []byte) (*TrafficInfo, error) {
	if len(msg) < 14 {
		return nil, fmt.Errorf("message too short: %d bytes", len(msg))
	}

	// Verify CRC per DO-260B Section 2.2.3.2.4
	if !verifyCRC(msg) {
		return nil, errors.New("CRC check failed")
	}

	// Continue processing...
	return traffic, nil
}
```

#### Naming Conventions
| Type | Convention | Example |
|------|------------|---------|
| Package | lowercase, single word | `main`, `common` |
| Exported | CamelCase | `TrafficInfo`, `ProcessMessage` |
| Unexported | camelCase | `globalStatus`, `processInternal` |
| Constants | MixedCaps or SCREAMING_SNAKE | `MaxClients`, `DEFAULT_TIMEOUT` |
| Interface | -er suffix (if applicable) | `Reader`, `Writer`, `Processor` |

#### File Organization
- One primary type per file (traffic.go for TrafficInfo, gps.go for GPSData)
- Group related functions near their type definitions
- Initialization functions (`init()`) at top of file
- Test files: `*_test.go` in same directory

### 5.2 Safety-Critical Code Requirements

For modules identified as safety-critical (see `docs/DO-278A-ANALYSIS.md` Section 2.3):

1. **Defensive Programming**:
   - Validate all inputs (range checks, null checks)
   - Check array bounds explicitly
   - Handle all error conditions
   - Use defensive copies for mutable data

2. **No Undefined Behavior**:
   - No arithmetic overflow (check before operations)
   - No divide-by-zero (check denominator)
   - No out-of-bounds array access

3. **Predictable Behavior**:
   - Avoid reliance on undefined behavior
   - Minimize use of external dependencies in critical paths
   - Deterministic execution where possible

4. **Enhanced Documentation**:
   - Algorithm description
   - Safety considerations
   - Preconditions and postconditions
   - Reference to requirements and specifications

Example:
```go
// ExtrapolatePosition computes predicted position based on track and speed.
// Safety-Critical: Used for traffic alerting (FR-305, SR-7)
// Algorithm: Dead reckoning per DO-260B Section 2.2.3.2.5
// Preconditions: ti.Lat, ti.Lng, ti.Track, ti.Speed must be valid
// Postconditions: Returns position or error if extrapolation exceeds 30 seconds
func ExtrapolatePosition(ti *TrafficInfo, deltaTime float64) (lat, lng float64, err error) {
	// Validate inputs
	if deltaTime < 0 || deltaTime > 30.0 {
		return 0, 0, fmt.Errorf("invalid deltaTime: %.2f seconds", deltaTime)
	}
	if ti.Speed < 0 || ti.Speed > 1500 { // Max reasonable speed: Mach 2+
		return 0, 0, fmt.Errorf("invalid speed: %.1f kt", ti.Speed)
	}

	// Compute extrapolation
	// ... implementation ...

	return lat, lng, nil
}
```

### 5.3 Version Control

**Git Workflow**:
1. **Master Branch**: Always stable, releasable
2. **Feature Branches**: For new features (`feature/add-ogn-support`)
3. **Bugfix Branches**: For bug fixes (`bugfix/fix-gps-overflow`)
4. **Release Branches**: For release preparation (`release/v1.6.0`)

**Commit Messages**:
```
<type>: <short summary (50 chars max)>

<detailed description (if needed)>

Implements: FR-101, FR-102
Fixes: #123
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `ci`, `build`

**Commit Frequency**:
- Commit logical units of work
- Each commit should build successfully
- Squash WIP commits before merging to master

---

## 6. Verification Process

### 6.1 Unit Testing

**Requirements**:
- All new functions SHALL have unit tests
- Target: ≥80% statement coverage, ≥70% decision coverage
- Tests SHALL verify both normal and error conditions
- Tests SHALL be automated and repeatable

**Test Organization**:
```
main/
  traffic.go          # Implementation
  traffic_test.go     # Unit tests
  testdata/
    adsb_samples.bin  # Test fixtures
```

**Test Naming**:
```go
// TestFunctionName_Scenario_ExpectedResult
func TestProcessADSBMessage_ValidMessage_ReturnsTraffic(t *testing.T) { ... }
func TestProcessADSBMessage_InvalidCRC_ReturnsError(t *testing.T) { ... }
func TestProcessADSBMessage_MessageTooShort_ReturnsError(t *testing.T) { ... }
```

**Running Tests**:
```bash
# Run all tests with coverage
go test -v -coverprofile=coverage.out -covermode=atomic ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Run tests with race detector
go test -race ./...
```

### 6.2 Integration Testing

**Test Scenarios** (documented in `docs/TEST_PROCEDURES.md`):
1. End-to-end data flow (RF → Processing → GDL90 output)
2. Multi-source traffic fusion (ADS-B + UAT + OGN)
3. GPS acquisition and position validation
4. Client connection/disconnection
5. Error handling and recovery
6. Performance under load (500 msg/sec)

**Test Harness**:
- Replay captured RF data
- Simulated GPS input
- Mock client connections
- X-Plane integration testing

### 6.3 System Testing

**Test Types**:
1. **Functional Testing**: All requirements verified
2. **Performance Testing**: Latency, throughput, CPU/memory usage
3. **Reliability Testing**: 8-hour continuous operation
4. **Compatibility Testing**: Multiple EFB applications
5. **Hardware Testing**: Raspberry Pi 3B, 3B+, 4, Zero 2W

**Field Testing**:
- Alpha/beta testers for pre-release builds
- Real aircraft installations
- Feedback collected via GitHub issues

### 6.4 Code Review

**All pull requests require**:
1. Automated tests passing (CI/CD)
2. Code coverage maintained or improved
3. No new compiler warnings
4. At least one peer review approval
5. Compliance with coding standards

**Review Checklist**:
- [ ] Requirements traced (Implements: FR-xxx)
- [ ] Error handling complete
- [ ] Logging appropriate
- [ ] No code duplication
- [ ] Performance acceptable
- [ ] Thread-safety verified (if concurrent)
- [ ] Documentation updated
- [ ] Tests adequate

---

## 7. Build and Integration

### 7.1 Build Process

**Build Types**:
1. **Development Build**: Local developer build (`make`)
2. **CI Build**: Automated build on every commit (GitHub Actions)
3. **Nightly Build**: Automated .deb packages (2 AM UTC daily)
4. **Release Build**: Full SD card images (triggered by version tag)

**Build Configuration**:
- **US Region**: UAT enabled, OGN disabled (978 MHz)
- **EU Region**: OGN enabled, UAT disabled (868 MHz)

**Build Artifacts**:
- `.deb` packages for upgrades (~80-90 MB)
- `.img` files for new installations (~2-4 GB, compressed to ~500 MB-1 GB)

### 7.2 Continuous Integration (CI/CD)

**GitHub Actions Workflows**:
- `.github/workflows/build-regions.yml` - On-demand .deb builds
- `.github/workflows/nightly.yml` - Nightly automated builds
- `.github/workflows/release-images.yml` - Full release images

**CI Pipeline**:
```
Push to master
    ↓
Build (native ARM64)
    ↓
Unit Tests + Coverage
    ↓
Integration Tests
    ↓
Static Analysis (go vet, go fmt)
    ↓
Race Detection (go test -race)
    ↓
Artifact Generation
    ↓
Deployment (GitHub Releases)
```

**Quality Gates**:
- All tests must pass
- Code coverage ≥ 80% (enforced in CI once baseline established)
- No high-severity security vulnerabilities (go sec)
- Build must succeed for both US and EU regions

---

## 8. Configuration Management

### 8.1 Version Control System

**System**: Git + GitHub
**Repository**: https://github.com/stratux/stratux

**Branch Protection**:
- Master branch requires pull request reviews
- Status checks must pass before merge
- No force pushes to master
- Linear history preferred (rebase before merge)

### 8.2 Version Numbering

**Semantic Versioning**: `MAJOR.MINOR.PATCH`

- **MAJOR**: Incompatible changes (e.g., config file format change)
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

**Development Versions**: `0.0.YYYYMMDD-<commit-hash>`

**Tagging**:
```bash
git tag -a v1.6.0 -m "Release version 1.6.0"
git push origin v1.6.0
```

### 8.3 Baseline Management

**Baseline Types**:
1. **Development Baseline**: Each commit on master
2. **Test Baseline**: Nightly builds (prerelease tags)
3. **Release Baseline**: Version tags (v1.x.x)

**Baseline Contents**:
- Source code (Git commit hash)
- Build scripts (Makefile, GitHub Actions workflows)
- Documentation (all docs/ files)
- Requirements and design documents
- Test procedures and results

---

## 9. Roles and Responsibilities

| Role | Responsibilities | Current Assignment |
|------|------------------|-------------------|
| **Project Lead** | Requirements approval, architecture decisions, release approval | TBD |
| **Developer** | Design, implementation, unit testing, code review | Open source contributors |
| **Tester** | Test plan development, integration/system testing, test automation | Community + CI/CD |
| **Reviewer** | Code review, design review, requirements review | Peer reviewers (GitHub) |
| **QA** | Process compliance, metrics tracking, audit preparation | TBD |
| **Release Manager** | Build process, version management, release coordination | Automated (GitHub Actions) |

---

## 10. Compliance and Audit

### 10.1 DO-278A Compliance Activities

**SAL-3 Objectives** (per DO-278A Table A-1):
- [x] A-1: Requirements are developed (`docs/REQUIREMENTS.md`)
- [x] A-2: Requirements are traceable (`docs/TRACEABILITY_MATRIX.xlsx`)
- [x] A-3: Software design is developed (`docs/HIGH_LEVEL_DESIGN.md`)
- [x] A-4: Design is traceable to requirements (Mermaid diagrams reference requirements)
- [x] A-5: Source code is developed (Complete implementation)
- [ ] A-6: Code is traceable to design (In progress)
- [ ] A-7: Integration procedures developed (Planned: `docs/TEST_PROCEDURES.md`)
- [ ] A-8: Verification procedures developed (Planned: test cases)
- [ ] A-9: Test coverage of requirements (Target: 100%)
- [ ] A-10: Test coverage of structure (Target: 80% statement, 70% decision)
- [ ] A-11: Verification of integration (Integration test suite)
- [x] A-12: Reviews and analyses performed (Code reviews via GitHub)

### 10.2 Process Metrics

**Tracked Metrics**:
- Code coverage percentage (statement, decision, branch)
- Requirements coverage (% of requirements with tests)
- Defect density (defects per KLOC)
- Build success rate
- Test pass rate
- Code review turnaround time

**Reporting**:
- Metrics dashboard (GitHub Actions artifacts)
- Monthly progress reports (GitHub Discussions)
- Quarterly roadmap reviews

---

## 11. Training and Competency

### 11.1 Developer Training

**Required Knowledge**:
- Go programming language (effective Go practices)
- Aviation concepts (ADS-B, GDL90, GNSS)
- DO-278A SAL-3 objectives and verification requirements
- Git and GitHub workflow
- Stratux architecture (read `docs/HIGH_LEVEL_DESIGN.md`)

**Recommended Reading**:
- DO-278A (RTCA, available for purchase)
- DO-260B Section 2 (incorporated by reference in 14 CFR 91.227)
- FAA GDL90 Public ICD
- Effective Go (https://go.dev/doc/effective_go)
- Stratux Requirements (`docs/REQUIREMENTS.md`)

### 11.2 Onboarding Process

**New Contributors**:
1. Read `CLAUDE.md` for project overview
2. Read `docs/REQUIREMENTS.md` to understand system requirements
3. Read `docs/HIGH_LEVEL_DESIGN.md` to understand architecture
4. Set up development environment (Go toolchain, RTL-SDR libraries)
5. Build and run tests locally
6. Start with "good first issue" label on GitHub

---

## 12. Document Control

**Document Owner**: Project Lead
**Review Cycle**: Quarterly or when significant process changes occur
**Approval**: Project Lead
**Distribution**: Public (GitHub repository)

**Change History**:

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-13 | Requirements Engineering | Initial version for DO-278A SAL-3 compliance |

---

## 13. References

1. **RTCA DO-278A**, "Guidelines for Communication, Navigation, Surveillance and Air Traffic Management (CNS/ATM) Systems Software Integrity Assurance," December 2011
2. **RTCA DO-260B**, "Minimum Operational Performance Standards for 1090 MHz Extended Squitter ADS-B," December 2, 2009 (incorporated by reference in 14 CFR 91.227)
3. **RTCA DO-282B**, "Minimum Operational Performance Standards for UAT ADS-B," December 2, 2009 (incorporated by reference in 14 CFR 91.227)
4. **FAA GDL90 Public ICD**, "GDL 90 Data Interface Specification," Revision A, June 5, 2007 (https://www.faa.gov/sites/faa.gov/files/air_traffic/technology/adsb/archival/GDL90_Public_ICD_RevA.PDF)
5. **EUROCONTROL ESASSP**, "EUROCONTROL Specification for ATM Surveillance System Performance," Volumes 1-2, 2023 (https://www.eurocontrol.int/publication/eurocontrol-specification-atm-surveillance-system-performance-esassp)
6. **EUROCONTROL ASTERIX Category 021**, "ADS-B Target Reports," v2.6, December 2021
7. **Stratux REQUIREMENTS.md**, System Requirements Specification, Version 1.0
8. **Stratux HIGH_LEVEL_DESIGN.md**, High-Level Design Document, Version 1.0
9. **Stratux DO-278A-ANALYSIS.md**, DO-278A Compliance Analysis, Version 1.0
10. **Effective Go**, The Go Programming Language (https://go.dev/doc/effective_go)

---

**END OF SOFTWARE DEVELOPMENT PROCESS DOCUMENT**
