# Stratux Software Quality Assurance Plan

**Document Version**: 1.0
**Last Updated**: 2025-10-13
**Status**: Active
**Compliance**: DO-278A SAL-3

---

## 1. Introduction

### 1.1 Purpose

This Software Quality Assurance Plan (SQAP) defines the quality assurance activities, standards, and procedures for the Stratux ADS-B/UAT/OGN receiver system to ensure compliance with DO-278A Software Assurance Level 3 (SAL-3) requirements.

### 1.2 Scope

This plan applies to all software development, verification, and configuration management activities for the Stratux system.

### 1.3 Applicable Documents

| Document | Title | Location |
|----------|-------|----------|
| DO-278A | Software Integrity Assurance for CNS/ATM Systems | RTCA DO-278A (2011) |
| SDP | Software Development Process | `docs/SOFTWARE_DEVELOPMENT_PROCESS.md` |
| SRS | System Requirements Specification | `docs/REQUIREMENTS.md` |
| HLD | High-Level Design | `docs/HIGH_LEVEL_DESIGN.md` |
| CMP | Configuration Management Plan | `docs/CONFIGURATION_MANAGEMENT_PLAN.md` |

---

## 2. Quality Assurance Organization

### 2.1 Responsibilities

| Role | Responsibilities | Independence |
|------|------------------|--------------|
| **QA Lead** | Oversee QA activities, audit compliance, report quality metrics | Independent from development |
| **Process Auditor** | Verify process compliance, review documentation | Independent from development |
| **Test Lead** | Ensure test coverage, verify test results | May be development team member |
| **Peer Reviewers** | Code review, design review | Independent from original author |

**Current Status**: Open-source project leverages community peer review and automated CI/CD for QA activities. Formal QA Lead TBD for DO-278A compliance certification.

### 2.2 Authority

The QA organization has authority to:
- Review all software deliverables
- Identify non-conformances
- Halt release if critical defects or process violations found
- Report quality status to project leadership
- Request corrective action

---

## 3. Quality Assurance Activities

### 3.1 Process Assurance

**Objective**: Ensure the Software Development Process (SDP) is followed consistently.

**Activities**:
1. **Process Audits**: Quarterly review of development activities
2. **Tool Qualification**: Verify CI/CD tools produce correct results
3. **Process Compliance**: Check adherence to lifecycle, coding standards, review requirements
4. **Deviation Tracking**: Document and approve any process deviations

**Schedule**:
- Process audit: Quarterly
- Tool qualification: Once per tool version change
- Continuous monitoring via CI/CD checks

**Deliverables**:
- Process audit reports
- Tool qualification reports
- Non-conformance reports
- Corrective action tracking

### 3.2 Product Assurance

**Objective**: Ensure software products meet requirements and quality standards.

**Activities**:

#### 3.2.1 Requirements Review
- **Frequency**: For each new/modified requirement
- **Criteria**:
  - Requirements are unambiguous, testable, and traceable
  - Acceptance criteria defined
  - Source reference documented
  - Priority assigned
- **Reviewer**: Peer reviewer + QA

#### 3.2.2 Design Review
- **Frequency**: For architectural changes or safety-critical module modifications
- **Criteria**:
  - Design documented in HLD with Mermaid diagrams
  - Traceability to requirements maintained
  - Interfaces clearly defined
  - Error handling strategy documented
- **Participants**: Designer, peer reviewer, QA

#### 3.2.3 Code Review
- **Frequency**: Every pull request
- **Criteria** (per `.github/PULL_REQUEST_TEMPLATE.md`):
  - [ ] Implements documented requirements (Implements: FR-xxx)
  - [ ] Adheres to coding standards (Go style guide)
  - [ ] Error handling complete
  - [ ] Unit tests included (≥80% coverage of new/modified code)
  - [ ] No compiler warnings
  - [ ] Thread-safety verified (if concurrent)
  - [ ] Documentation updated
- **Reviewer**: At least one peer reviewer (not the author)

#### 3.2.4 Test Review
- **Frequency**: For each test case added/modified
- **Criteria**:
  - Test traces to requirements (Verifies: FR-xxx)
  - Test covers normal, boundary, and error cases
  - Test is automated and repeatable
  - Expected results clearly defined
- **Reviewer**: Test lead or peer reviewer

### 3.3 Conformance Reviews

**Objective**: Verify software artifacts conform to standards and plans.

**Review Types**:

#### 3.3.1 Documentation Reviews
- Requirements document (REQUIREMENTS.md)
- Design document (HIGH_LEVEL_DESIGN.md)
- Test procedures (TEST_PROCEDURES.md)
- Process documents (SDP, SQAP, CMP)

**Review Checklist**:
- [ ] Document follows template
- [ ] Content complete and accurate
- [ ] References to standards documented
- [ ] Version control information present
- [ ] Approval signatures (if required)

#### 3.3.2 Build and Release Reviews
- Verify version numbering correct
- Verify all tests passing
- Verify artifacts uploaded to GitHub Releases
- Verify release notes complete

### 3.4 Quality Metrics

**Tracked Metrics**:

| Metric | Target | Measurement Method | Frequency |
|--------|--------|-------------------|-----------|
| **Code Coverage (Statement)** | ≥80% | `go test -cover` | Every commit (CI) |
| **Code Coverage (Decision)** | ≥70% | `go test -cover` | Every commit (CI) |
| **Requirements Coverage** | 100% | Traceability matrix | Monthly |
| **Test Pass Rate** | ≥99% | CI/CD results | Every commit |
| **Build Success Rate** | ≥95% | GitHub Actions | Monthly |
| **Defect Density** | <0.5 defects/KLOC | Issue tracking | Per release |
| **Code Review Turnaround** | <48 hours | GitHub PR timestamps | Monthly |
| **Critical Defects Open** | 0 at release | GitHub issues (label: critical) | Per release |

**Reporting**:
- Metrics dashboard published in GitHub Actions artifacts
- Monthly quality report in GitHub Discussions
- Quarterly compliance review with project leadership

---

## 4. Standards and Procedures

### 4.1 Coding Standards

**Language**: Go (Golang)
**Standard**: Effective Go + Stratux-specific conventions

**Key Requirements**:
- `go fmt` formatting enforced (CI check)
- `go vet` static analysis passing (CI check)
- No golint warnings
- Package and exported function documentation complete
- Error handling: all errors checked and logged

**Safety-Critical Code** (traffic.go, gps.go, gen_gdl90.go, sdr.go, sensors.go):
- Enhanced documentation (algorithm, safety rationale, preconditions/postconditions)
- Input validation (range checks, null checks)
- No undefined behavior (overflow, divide-by-zero)
- Reference to applicable DO-260B/DO-282B sections

**Reference**: `docs/SOFTWARE_DEVELOPMENT_PROCESS.md` Section 5

### 4.2 Testing Standards

**Unit Testing**:
- Go testing framework (`testing` package)
- Test naming: `TestFunctionName_Scenario_ExpectedResult`
- Table-driven tests for multiple scenarios
- Coverage: ≥80% statement, ≥70% decision

**Integration Testing**:
- Replay-based regression tests (captured RF data)
- Hardware-in-loop testing on Raspberry Pi
- X-Plane simulation integration

**System Testing**:
- Functional: All requirements verified
- Performance: Latency, throughput, resource usage
- Reliability: 8-hour continuous operation
- Compatibility: Multiple EFB applications

**Test Documentation**: `docs/TEST_PROCEDURES.md` (in development)

### 4.3 Documentation Standards

**Format**: Markdown (`.md` files)
**Version Control**: Git (tracked with code)
**Required Sections**:
- Title and document metadata (version, date, status)
- Table of contents (for documents >5 pages)
- Sections numbered hierarchically
- References to source standards
- Change history table

**Technical Diagrams**: Mermaid (embedded in Markdown)

### 4.4 Configuration Management Standards

**Version Control System**: Git + GitHub
**Branching Strategy**: Feature branches merged to master via pull requests
**Commit Messages**: Structured format (type: description, Implements: FR-xxx)
**Tagging**: Semantic versioning for releases (vMAJOR.MINOR.PATCH)

**Reference**: `docs/CONFIGURATION_MANAGEMENT_PLAN.md`

---

## 5. Tool Qualification

### 5.1 Development Tools

| Tool | Version | Purpose | Qualification Required |
|------|---------|---------|------------------------|
| **Go Compiler** | 1.19+ | Code compilation | Yes (verification only) |
| **go test** | 1.19+ | Unit testing, coverage | Yes (verification only) |
| **Git** | 2.x | Version control | No (not in verification path) |
| **GitHub Actions** | N/A | CI/CD automation | Yes (verification only) |
| **go vet** | 1.19+ | Static analysis | Yes (verification only) |
| **go fmt** | 1.19+ | Code formatting | No (not in verification path) |

**Qualification Criteria** (DO-278A Section 11.4):
- **Verification Tools**: Require qualification data showing tool produces correct results
- **Non-Verification Tools**: No qualification required

**Qualification Method**:
- For verification tools: Comparison with known-good test cases
- Tool qualification reports stored in `docs/tool-qualification/`

### 5.2 Tool Validation

**Go Compiler and Test Framework**:
- Use official Go releases from https://go.dev/
- Verify release checksums (SHA256)
- Test on sample code with known expected behavior
- Document Go version used for each release

**GitHub Actions Runners**:
- Use ubuntu-24.04-arm official runners
- Verify build reproducibility (same input → same output)
- Document runner version and environment

---

## 6. Problem Reporting and Corrective Action

### 6.1 Problem Reporting

**Problem Categories**:
1. **Defect**: Software does not meet requirements
2. **Non-Conformance**: Process not followed
3. **Deviation**: Intentional deviation from process (requires approval)

**Reporting Mechanism**: GitHub Issues with labels:
- `bug` - Functional defect
- `security` - Security vulnerability
- `process-nonconformance` - Process violation
- `requirements-change` - Requirement modification request

**Severity Levels**:

| Severity | Description | Response Time | Example |
|----------|-------------|---------------|---------|
| **Critical** | Safety impact, system unusable | Immediate | False traffic alert, GPS position error >1 NM |
| **High** | Major function broken | 1 week | GPS acquisition fails, SDR crashes |
| **Medium** | Minor function broken, workaround exists | 1 month | Web UI glitch, cosmetic issue |
| **Low** | Enhancement request, nice-to-have | Best effort | UI improvement, documentation typo |

### 6.2 Root Cause Analysis

**Required for**: Critical and High severity defects

**Process**:
1. **Problem Description**: What happened?
2. **Impact Assessment**: What systems/functions affected?
3. **Root Cause**: Why did it happen? (not just symptoms)
4. **Corrective Action**: How to fix the root cause?
5. **Preventive Action**: How to prevent recurrence?
6. **Verification**: How to verify fix is effective?

**Documentation**: GitHub issue with label `root-cause-analysis`

### 6.3 Corrective Action Tracking

**Corrective Action Process**:
1. Problem identified and reported (GitHub issue)
2. Root cause analysis performed (for critical/high)
3. Corrective action implemented (pull request)
4. Verification performed (tests updated, passing)
5. Closure verification (QA review)
6. Trend analysis (monthly review of problem patterns)

**Tracking**: GitHub Projects board "Corrective Actions"

---

## 7. Verification Activities

### 7.1 Verification Methods

Per DO-278A Section 6.3, four verification methods are used:

| Method | Description | Applicable To | Example |
|--------|-------------|---------------|---------|
| **Test** | Execute code with inputs, observe outputs | Functional requirements | Unit test: FR-101 (1090 MHz ADS-B reception) |
| **Analysis** | Examine code/design without execution | Performance, algorithms | Analysis: NFR-101 (position accuracy) |
| **Inspection** | Visual examination of artifacts | Documentation, interfaces | Review: FR-601 (GDL90 message format) |
| **Demonstration** | Operator uses system under normal conditions | User workflows | Demo: Web UI configuration |

### 7.2 Verification Schedule

| Activity | Timing | Responsibility | Deliverable |
|----------|--------|----------------|-------------|
| **Unit Test Execution** | Every commit | Developer | Test results (CI log) |
| **Integration Test** | Nightly | CI/CD | Test results (nightly build) |
| **Code Review** | Every PR | Peer reviewer | Approval in GitHub PR |
| **Design Review** | Major changes | Design team | Review minutes |
| **System Test** | Pre-release | Test team | Test report |
| **Requirements Verification** | Per requirement | Test lead | Traceability matrix |
| **Coverage Analysis** | Every commit | CI/CD | Coverage report |

### 7.3 Verification Closure

**Verification Complete When**:
- All requirements have ≥1 verification method executed
- All tests passing
- Code coverage ≥80% statement, ≥70% decision
- All critical and high-severity defects resolved
- Traceability matrix 100% complete
- Verification report approved by QA

**Deliverable**: Software Verification Report (SVR) per DO-278A Section 10.4

---

## 8. Records and Documentation

### 8.1 Quality Records

**Maintained Records**:
- Process audit reports
- Review minutes (design, code, test)
- Test results and coverage reports
- Non-conformance reports
- Corrective action tracking
- Tool qualification reports
- Verification results
- Release approvals

**Storage**: GitHub repository (`docs/` directory) and GitHub Issues/Projects

**Retention**: Permanent (Git history preserved)

### 8.2 Document Control

**Version Control**: All documents tracked in Git
**Approval**: Pull request review and approval
**Distribution**: Public (open-source project)
**Change History**: Git commit history + change history table in document

**Document Reviews**: Annually or when significant changes occur

---

## 9. Training and Competency

### 9.1 Required Training

**For Developers**:
- DO-278A SAL-3 overview
- Stratux architecture and requirements
- Coding standards and best practices
- Test-driven development
- Git and GitHub workflow

**For Reviewers**:
- Code review best practices
- Safety-critical code review checklist
- Requirements traceability verification

**For QA Personnel**:
- DO-278A compliance requirements
- Process audit techniques
- Metrics collection and analysis

**Resources**:
- `CLAUDE.md` - Project overview
- `docs/REQUIREMENTS.md` - System requirements
- `docs/HIGH_LEVEL_DESIGN.md` - Architecture
- `docs/SOFTWARE_DEVELOPMENT_PROCESS.md` - Development process
- DO-278A standard (available from RTCA)

### 9.2 Competency Verification

**Method**: Contribution history and peer review feedback
**Criteria**: Successful completion of ≥3 contributions with positive reviews

---

## 10. Supplier Oversight (if applicable)

**Current Status**: No external software suppliers. All third-party dependencies are open-source libraries.

**Third-Party Library Management**:
- Use Go modules (`go.mod`) for dependency management
- Pin dependency versions
- Review security advisories (GitHub Dependabot)
- Verify license compatibility (all must be open-source compatible)

**Critical Dependencies**:
- RTL-SDR libraries (GPL 2.0)
- Go standard library (BSD-style license)

**Acceptance Criteria**:
- Library actively maintained
- No known critical security vulnerabilities
- License compatible with GPL v3 (Stratux license)

---

## 11. Compliance Assessment

### 11.1 DO-278A SAL-3 Compliance

**Current Status**: In progress toward full compliance

**Compliance Checklist** (DO-278A Table A-1):

| Objective | Description | Status | Evidence |
|-----------|-------------|--------|----------|
| **A-1** | Requirements developed | ✅ Complete | `docs/REQUIREMENTS.md` |
| **A-2** | Requirements traceable | ✅ Complete | `docs/TRACEABILITY_MATRIX.xlsx` |
| **A-3** | Design developed | ✅ Complete | `docs/HIGH_LEVEL_DESIGN.md` |
| **A-4** | Design traceable | ✅ Complete | Design references requirements |
| **A-5** | Source code developed | ✅ Complete | `main/` directory |
| **A-6** | Code traceable | 🟡 In progress | Code comments reference requirements |
| **A-7** | Integration procedures | 🟡 Planned | `docs/TEST_PROCEDURES.md` (TBD) |
| **A-8** | Verification procedures | 🟡 Planned | Test cases (in progress) |
| **A-9** | Requirements coverage | 🔴 Incomplete | Target: 100%, Current: ~5% |
| **A-10** | Structural coverage | 🔴 Incomplete | Target: 80%, Current: <5% |
| **A-11** | Integration verification | 🟡 Partial | Manual testing only |
| **A-12** | Reviews performed | ✅ Complete | GitHub PR reviews |

**Target Compliance Date**: Per `docs/ROADMAP.md`, Month 10 (M9 milestone)

### 11.2 Audit Preparation

**Internal Audits**: Quarterly (self-assessment)
**External Audit**: Month 11 per roadmap (third-party certification)

**Audit Artifacts**:
- All process documents (SDP, SQAP, CMP)
- Requirements and design documents
- Traceability matrix
- Test results and coverage reports
- Review records
- Non-conformance and corrective action records
- Tool qualification reports
- Release baselines (Git tags)

---

## 12. Continuous Improvement

### 12.1 Process Improvement

**Feedback Sources**:
- Process audit findings
- Metrics trends (defect density, coverage, build success rate)
- Developer feedback (retrospectives)
- External audit findings

**Improvement Process**:
1. Identify improvement opportunity
2. Propose process change (GitHub issue with label `process-improvement`)
3. Impact assessment (effort, benefit, risk)
4. Approval by project lead
5. Update process documents
6. Train team on new process
7. Monitor effectiveness (metrics)

### 12.2 Lessons Learned

**Captured After**:
- Each major release
- Significant issues (critical defects, security vulnerabilities)
- Process audits

**Documentation**: `docs/lessons-learned/` directory (Markdown files)

**Review**: Quarterly team review of lessons learned database

---

## 13. Quality Assurance Plan Approval

**Plan Owner**: QA Lead (TBD)
**Reviewed By**: Project Lead
**Approved By**: Project Lead
**Effective Date**: 2025-10-13

**Change History**:

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-13 | Requirements Engineering | Initial version for DO-278A SAL-3 compliance |

---

## 14. References

1. **RTCA DO-278A**, "Guidelines for Communication, Navigation, Surveillance and Air Traffic Management (CNS/ATM) Systems Software Integrity Assurance," December 2011
2. **Stratux SOFTWARE_DEVELOPMENT_PROCESS.md**, Software Development Process, Version 1.0
3. **Stratux REQUIREMENTS.md**, System Requirements Specification, Version 1.0
4. **Stratux HIGH_LEVEL_DESIGN.md**, High-Level Design Document, Version 1.0
5. **Stratux DO-278A-ANALYSIS.md**, DO-278A Compliance Analysis, Version 1.0
6. **Stratux ROADMAP.md**, Development Roadmap, Version 1.0

---

**END OF SOFTWARE QUALITY ASSURANCE PLAN**
