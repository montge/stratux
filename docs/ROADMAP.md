# Stratux Development Roadmap

**Document Version**: 1.0
**Last Updated**: 2025-10-13
**Status**: Draft

---

## Executive Summary

This roadmap outlines the path to achieving DO-278A SAL-3 compliance, 80% test coverage, and enhanced security for the Stratux system over the next 12 months.

**Current State**:
- ✅ Mature, functional codebase
- ✅ CI/CD with GitHub Actions
- ⚠️ Limited test coverage (<5%)
- ❌ No formal requirements documentation
- ❌ No unit test infrastructure

**Target State** (12 months):
- ✅ DO-278A SAL-3 compliant
- ✅ 80% test coverage
- ✅ Formal requirements and verification
- ✅ Enhanced security posture
- ✅ Continuous compliance infrastructure

---

## Phase 1: Foundation (Months 1-3)

### 1.1 Requirements Engineering (Months 1-2)

**Goal**: Establish formal requirements baseline

**Tasks**:
- [x] Create System Requirements Specification (SRS) - COMPLETE
- [ ] Review SRS with stakeholders
- [ ] Create High-Level Design (HLD) document
- [ ] Establish requirements traceability matrix
- [ ] Define acceptance criteria for each requirement
- [ ] Create requirements management process

**Deliverables**:
- ✅ `docs/REQUIREMENTS.md` (101 requirements documented)
- ⏳ `docs/HIGH_LEVEL_DESIGN.md`
- ⏳ `docs/TRACEABILITY_MATRIX.xlsx`
- ⏳ `docs/REQUIREMENTS_MANAGEMENT_PLAN.md`

**Success Criteria**:
- All major features documented as requirements
- Requirements reviewed and approved
- Traceability established from requirements to code

---

### 1.2 Test Infrastructure (Month 2-3)

**Goal**: Establish comprehensive testing framework

**Tasks**:
- [ ] Set up Go testing framework
- [ ] Configure code coverage tools (`go test -cover`)
- [ ] Integrate coverage reporting in CI/CD
- [ ] Create test data fixtures and mocks
- [ ] Establish testing standards and guidelines
- [ ] Set up test result tracking

**Deliverables**:
- ⏳ `.github/workflows/test.yml` - Test automation
- ⏳ `test/README.md` - Testing guide
- ⏳ `test/fixtures/` - Test data
- ⏳ `test/mocks/` - Mock implementations
- ⏳ Coverage reports in CI/CD

**Success Criteria**:
- All tests run automatically on every commit
- Coverage reports generated and published
- Zero-friction test execution for developers

**Implementation**:
```bash
# Enable coverage in CI
go test -v -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html

# Upload to codecov or coveralls
bash <(curl -s https://codecov.io/bash)
```

---

## Phase 2: Core Testing (Months 3-6)

### 2.1 Safety-Critical Module Testing (Months 3-4)

**Goal**: 80% coverage of safety-critical functions

**Priority Modules** (in order):
1. **Traffic Fusion** (`main/traffic.go`)
   - Target: 90% coverage
   - Critical: Ownship detection, alerting, position extrapolation

2. **GPS Processing** (`main/gps.go`)
   - Target: 85% coverage
   - Critical: Position validation, accuracy checking

3. **GDL90 Generation** (`main/gen_gdl90.go`)
   - Target: 85% coverage
   - Critical: Message format, CRC, traffic/ownship reports

4. **ADS-B Decoding** (`main/sdr.go`)
   - Target: 80% coverage
   - Critical: Message validation, CRC checking

5. **AHRS Processing** (`main/sensors.go`)
   - Target: 75% coverage
   - Critical: Attitude calculation, sensor validation

**Tasks**:
- [ ] Write unit tests for traffic fusion logic
- [ ] Write unit tests for GPS validation
- [ ] Write unit tests for GDL90 message generation
- [ ] Write unit tests for ADS-B/UAT decoding
- [ ] Write unit tests for AHRS calculations

**Success Criteria**:
- Safety-critical modules achieve target coverage
- All critical functions have test cases
- Tests cover normal, boundary, and error conditions

---

### 2.2 Integration Testing (Months 4-5)

**Goal**: Verify subsystem interactions

**Test Scenarios**:
1. **End-to-End Data Flow**
   - RF Input → Processing → GDL90 Output
   - GPS Input → Position → GDL90 Ownship
   - Multiple traffic sources → Fusion → Output

2. **Error Handling**
   - GPS loss → Graceful degradation
   - SDR failure → Reconnection
   - Invalid messages → Rejection

3. **Performance**
   - 500 msg/sec traffic load
   - Multiple simultaneous clients
   - 8-hour continuous operation

**Tasks**:
- [ ] Create integration test harness
- [ ] Develop test data sets (captured traffic)
- [ ] Implement replay-based regression tests
- [ ] Create performance test suite
- [ ] Document test procedures

**Deliverables**:
- ⏳ `test/integration/` - Integration tests
- ⏳ `test/testdata/` - Test data sets
- ⏳ `docs/TEST_PROCEDURES.md` - Test execution guide

**Success Criteria**:
- All major data paths tested
- Error handling verified
- Performance requirements validated

---

### 2.3 Remaining Module Coverage (Months 5-6)

**Goal**: 80% overall code coverage

**Remaining Modules**:
- Network management (`main/network.go`, `main/clientconnection.go`)
- Weather processing (`main/weather.go`)
- OGN processing (`main/ogn.go`)
- Configuration management (`main/managementinterface.go`)
- Logging (`main/datalog.go`)
- Utilities (`common/`)

**Tasks**:
- [ ] Write unit tests for network management
- [ ] Write unit tests for weather processing
- [ ] Write unit tests for OGN/FLARM
- [ ] Write unit tests for configuration
- [ ] Write unit tests for logging
- [ ] Write unit tests for utility functions

**Success Criteria**:
- Overall code coverage ≥ 80%
- All modules have basic test coverage
- Coverage gaps identified and documented

---

## Phase 3: Verification & Validation (Months 6-9)

### 3.1 Requirements Verification (Months 6-7)

**Goal**: Verify all requirements have tests

**Tasks**:
- [ ] Map each requirement to test cases
- [ ] Identify untested requirements
- [ ] Create tests for gaps
- [ ] Document test results
- [ ] Create verification report

**Deliverables**:
- ⏳ `docs/VERIFICATION_MATRIX.xlsx` - Requirement→Test mapping
- ⏳ `docs/VERIFICATION_REPORT.md` - Test results summary

**Success Criteria**:
- Every requirement has ≥1 test case
- All tests passing
- Traceability complete

---

### 3.2 System Testing (Months 7-8)

**Goal**: Validate system-level requirements

**Test Types**:
1. **Functional Testing**
   - All features work as specified
   - User workflows validated
   - EFB compatibility verified

2. **Non-Functional Testing**
   - Performance benchmarks
   - Reliability (MTBF)
   - Resource usage (CPU, memory, disk)

3. **Field Testing**
   - Real aircraft installation
   - Live traffic reception
   - Multiple EFB apps
   - Various hardware configurations

**Tasks**:
- [ ] Develop system test plan
- [ ] Execute functional test cases
- [ ] Execute performance tests
- [ ] Conduct field testing (alpha users)
- [ ] Document results and issues

**Success Criteria**:
- System meets all functional requirements
- Performance requirements validated
- No critical defects in field testing

---

### 3.3 Security Hardening (Month 8-9)

**Goal**: Enhanced security posture

**Security Improvements**:
1. **Authentication**
   - Web UI login (optional)
   - HTTP Basic Auth or token-based

2. **Input Validation**
   - All user inputs sanitized
   - SQL injection prevention
   - XSS prevention

3. **Secure Updates**
   - Cryptographic signature verification
   - Update integrity checking

4. **Credential Protection**
   - WiFi passwords encrypted
   - Secure storage of sensitive data

**Tasks**:
- [ ] Implement web authentication
- [ ] Audit all user inputs
- [ ] Implement update signing
- [ ] Encrypt stored credentials
- [ ] Conduct security testing (penetration test)

**Success Criteria**:
- Security requirements met
- No high-severity vulnerabilities
- Third-party security audit passed

---

## Phase 4: Compliance & Documentation (Months 9-12)

### 4.1 DO-278A Compliance Documentation (Months 9-10)

**Goal**: Complete DO-278A SAL-3 compliance package

**Documents Required**:
- [x] Software Accomplishment Summary (SAS) - Partial
- [x] Software Requirements Specification (SRS) - ✅ COMPLETE
- [ ] Software Design Description (SDD)
- [ ] Software Verification Plan (SVP)
- [ ] Software Verification Results (SVR)
- [ ] Software Configuration Index (SCI)
- [ ] Software Quality Assurance Records

**Tasks**:
- [ ] Complete High-Level Design document
- [ ] Create Software Verification Plan
- [ ] Document all verification results
- [ ] Create configuration management records
- [ ] Compile QA records
- [ ] Review for completeness

**Success Criteria**:
- All DO-278A required documents complete
- Documents reviewed and approved
- Ready for audit

---

### 4.2 Third-Party Audit (Month 11)

**Goal**: Independent verification of compliance

**Audit Scope**:
- DO-278A SAL-3 compliance
- Requirements traceability
- Test coverage adequacy
- Documentation completeness
- Security posture

**Tasks**:
- [ ] Select audit firm
- [ ] Prepare for audit
- [ ] Conduct audit
- [ ] Address findings
- [ ] Obtain certification

**Success Criteria**:
- Audit completed with no major findings
- Minor findings addressed
- Compliance certificate issued

---

### 4.3 Continuous Compliance Infrastructure (Month 12)

**Goal**: Sustain compliance as system evolves

**Infrastructure**:
1. **Automated Testing**
   - Unit tests run on every commit
   - Integration tests run nightly
   - System tests run weekly

2. **Continuous Coverage**
   - Coverage tracked in CI/CD
   - Coverage regression blocked
   - Coverage reports published

3. **Requirements Management**
   - New features require requirements
   - Requirements reviewed before implementation
   - Traceability maintained

4. **Change Control**
   - All changes reviewed
   - Impact analysis performed
   - Verification updated

**Tasks**:
- [ ] Enhance CI/CD with all test types
- [ ] Implement coverage gates
- [ ] Create requirements workflow
- [ ] Document change control process
- [ ] Train team on processes

**Success Criteria**:
- All compliance checks automated
- Coverage maintained above 80%
- Requirements up-to-date
- Change control followed

---

## Milestones

| Milestone | Target Date | Status |
|-----------|-------------|---------|
| **M1**: Requirements Documented | Month 2 | 🟢 COMPLETE |
| **M2**: Test Infrastructure Ready | Month 3 | 🔴 Not Started |
| **M3**: 50% Code Coverage | Month 4 | 🔴 Not Started |
| **M4**: Safety-Critical Modules 80%+ | Month 5 | 🔴 Not Started |
| **M5**: 80% Overall Coverage | Month 6 | 🔴 Not Started |
| **M6**: All Requirements Verified | Month 7 | 🔴 Not Started |
| **M7**: System Testing Complete | Month 8 | 🔴 Not Started |
| **M8**: Security Hardened | Month 9 | 🔴 Not Started |
| **M9**: Documentation Complete | Month 10 | 🔴 Not Started |
| **M10**: Third-Party Audit | Month 11 | 🔴 Not Started |
| **M11**: Continuous Compliance | Month 12 | 🔴 Not Started |

---

## Resource Requirements

### 4.1 Personnel

**Recommended Team**:
- **Requirements Engineer**: 0.5 FTE (Months 1-3)
- **Test Engineer**: 1.0 FTE (Months 2-9)
- **Software Developer**: 0.5 FTE (Months 3-9, test development)
- **QA Engineer**: 0.5 FTE (Months 6-12)
- **Security Engineer**: 0.5 FTE (Months 8-9)
- **Technical Writer**: 0.25 FTE (Months 9-10)

**Total Effort**: ~3.5 FTE-years

### 4.2 Budget Estimate

| Category | Cost Estimate |
|----------|--------------|
| Personnel (3.5 FTE-years @ $120K/yr) | $420,000 |
| Tools and Infrastructure | $10,000 |
| Third-Party Audit | $25,000 |
| Hardware for Testing | $5,000 |
| **TOTAL** | **$460,000** |

*Note: Costs can be significantly reduced using open-source community contributions*

---

## Success Metrics

### 5.1 Quantitative Metrics

- **Code Coverage**: ≥ 80% statement, ≥ 70% decision
- **Requirements Coverage**: 100% (all requirements have tests)
- **Test Pass Rate**: ≥ 99%
- **Defect Density**: < 0.5 defects/KLOC
- **Build Success Rate**: ≥ 95%

### 5.2 Qualitative Metrics

- DO-278A SAL-3 compliance achieved
- Third-party audit passed
- Security vulnerabilities addressed
- Documentation complete and approved
- Community feedback positive

---

## Risks and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Insufficient resources | MEDIUM | HIGH | Leverage open-source community |
| Schedule slip | MEDIUM | MEDIUM | Agile approach, prioritize critical items |
| Scope creep | MEDIUM | MEDIUM | Strict change control |
| Test infrastructure complexity | LOW | MEDIUM | Use proven tools (Go testing) |
| Third-party audit fails | LOW | HIGH | Internal audits before official |
| Team attrition | LOW | HIGH | Documentation, knowledge sharing |

---

## Dependencies

### 6.1 External Dependencies

- GitHub Actions availability (CI/CD)
- Go toolchain updates
- RTL-SDR driver availability
- Raspberry Pi hardware availability

### 6.2 Internal Dependencies

- Community support for testing
- Access to test hardware
- Subject matter expert availability for reviews

---

## Continuous Improvement

This roadmap will be reviewed and updated:
- **Monthly**: Progress review, adjust as needed
- **Quarterly**: Major milestone assessment
- **Annually**: Full roadmap refresh

**Roadmap Owner**: TBD
**Review Board**: TBD

---

## Communication Plan

### 7.1 Stakeholder Updates

- **Monthly**: Progress report to community
- **Quarterly**: Milestone achievements announced
- **Ad-hoc**: Major decisions communicated

### 7.2 Documentation

- Roadmap published on GitHub
- Progress tracked in GitHub Projects
- Milestones visible in GitHub Releases

---

## Conclusion

This roadmap provides a clear path from the current state to DO-278A SAL-3 compliance with 80% test coverage over 12 months. While ambitious, it is achievable with focused effort and community support.

**Immediate Next Steps** (This Week):
1. Review and approve `REQUIREMENTS.md`
2. Set up basic Go test framework
3. Write first 10 unit tests for traffic.go
4. Configure GitHub Actions for test execution
5. Establish project tracking (GitHub Projects)

**Let's build a world-class, verified aviation system together!**

---

**END OF ROADMAP**
