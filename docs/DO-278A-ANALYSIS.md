# DO-278A Software Integrity Assurance Analysis for Stratux

## Executive Summary

**System**: Stratux ADS-B/UAT/OGN Receiver
**Classification**: CNS/ATM Ground-Based System (Portable Aviation Receiver)
**Applicable Standard**: DO-278A (Software Integrity Assurance Considerations for Communication, Navigation, Surveillance and Air Traffic Management Systems)
**Certification Status**: Non-TSO (Non-certified, Portable Equipment)
**Intended Use**: Supplemental situational awareness for pilots
**Safety Class**: Class 3 (Equipment providing advisory information only)

---

## 1. System Classification

### 1.1 DO-278A vs DO-178C

Stratux is properly classified under **DO-278A** rather than DO-178C for the following reasons:

1. **Ground-Based Equipment**: Stratux is a ground/portable receiver, not permanently installed airborne equipment
2. **CNS System**: Provides Communication, Navigation, and Surveillance information
3. **Non-TSO**: Not TSO-C154c or TSO-C166b certified
4. **Advisory Only**: Provides situational awareness, not flight-critical functionality
5. **No ADS-B Out**: Receive-only system, does not transmit

### 1.2 Software Assurance Level (SAL)

Per DO-278A Section 2.3, Stratux is classified as:

**Software Assurance Level 3 (SAL-3)**

**Rationale**:
- **Class 3 Equipment**: "Equipment providing advisory information only to the air traffic controller or pilot"
- **Failure Impact**: Incorrect or missing information could lead to inappropriate pilot action but is not directly flight-critical
- **Redundancy**: Pilots have other sources of traffic/weather information (ATC, visual, other systems)
- **Time to Detect**: Pilots can detect system failure or erroneous information within operational timeframe
- **Consequence**: Failure results in **Major** effect (not Catastrophic or Hazardous)

Per DO-278A Table 2-2:
- SAL-3 requires **Software Level C** rigor
- Emphasizes: Requirements-based testing, structural coverage analysis, verification of integration

---

## 2. Safety Analysis

### 2.1 Failure Modes and Effects

| Failure Mode | Hazard | Severity | Detection | Mitigation |
|--------------|--------|----------|-----------|------------|
| **False Traffic Display** | Pilot takes unnecessary evasive action | Major | Visual confirmation, ATC | Multiple traffic sources, visual verification |
| **Missing Traffic Display** | Pilot unaware of proximate traffic | Major | Visual scan, ATC alerts | See-and-avoid, ATC services, TCAS (if equipped) |
| **Incorrect Ownship Position** | Inaccurate traffic relative position | Major | Cross-check with panel GPS | Panel GPS, visual confirmation |
| **False Weather** | Flight into weather | Major | Visual, ATC, other weather sources | Pilot weather briefing, visual meteorological conditions |
| **System Failure (Complete)** | Loss of ADS-B information | Major | Visual indication (LED), web UI | Pilot awareness, revert to non-ADS-B operations |
| **Incorrect AHRS Data** | Disorientation in IMC (if used as primary) | Hazardous* | Cross-check with panel instruments | **NOT FOR PRIMARY FLIGHT REFERENCE** |

*Note: AHRS functionality is explicitly marked as "supplemental" and not for primary flight reference

### 2.2 Hazard Classification

Per DO-278A Section 2.2.3, the following hazard classifications apply:

1. **Normal Operations**: No safety effect (system operating correctly)
2. **False Alerts**: Minor (nuisance, increased workload)
3. **Missed Traffic/Weather**: Major (potential unsafe situation, reduced safety margins)
4. **Erroneous Data with Action**: Major to Hazardous (depending on phase of flight)
5. **Complete System Failure**: Major (loss of supplemental information source)

### 2.3 Safety Requirements

**SR-1**: System SHALL clearly indicate when GPS position is invalid or inaccurate
**SR-2**: System SHALL indicate when traffic information is unavailable
**SR-3**: System SHALL indicate when weather information is stale (>15 minutes)
**SR-4**: System SHALL provide visual indication of system errors
**SR-5**: AHRS data SHALL be marked as "advisory only" and "not for primary flight reference"
**SR-6**: System SHALL NOT display traffic or weather data if source integrity cannot be verified
**SR-7**: System SHALL detect and filter ownship from traffic display
**SR-8**: System SHALL validate all received ADS-B messages per DO-260B/DO-282B

---

## 3. DO-278A Compliance Requirements

### 3.1 SAL-3 (Software Level C) Objectives

Based on DO-278A Table A-1, the following objectives apply:

| Objective | Description | Current Status | Gap |
|-----------|-------------|----------------|-----|
| **A-1** | Requirements are developed | ❌ Not Documented | **HIGH** |
| **A-2** | Requirements are traceable | ❌ No Traceability | **HIGH** |
| **A-3** | Software design is developed | ⚠️ Implicit in code | **MEDIUM** |
| **A-4** | Design is traceable to requirements | ❌ No Traceability | **HIGH** |
| **A-5** | Source code is developed | ✅ Complete | None |
| **A-6** | Code is traceable to design | ⚠️ Partial (comments) | **MEDIUM** |
| **A-7** | Integration procedures developed | ❌ Not Documented | **MEDIUM** |
| **A-8** | Verification procedures developed | ❌ Minimal testing | **HIGH** |
| **A-9** | Test coverage of requirements | ❌ No unit tests | **HIGH** |
| **A-10** | Test coverage of structure | ❌ No coverage analysis | **HIGH** |
| **A-11** | Verification of integration | ⚠️ Manual only | **MEDIUM** |
| **A-12** | Reviews and analyses performed | ❌ Not Documented | **MEDIUM** |

**Current Compliance**: ~20%
**Target Compliance for SAL-3**: 100% of objectives with independence

### 3.2 Structural Coverage Analysis (SAL-3)

Per DO-278A Section 6.4.4.2, SAL-3 requires:

1. **Statement Coverage**: Every statement executed at least once
2. **Decision Coverage**: Every decision (if/else, switch) evaluated to both TRUE and FALSE
3. **MC/DC Not Required**: Modified Condition/Decision Coverage not required for SAL-3

**Current Status**: No coverage measurement infrastructure
**Target**: 80% statement coverage minimum (industry best practice exceeds DO-278A minimum)

---

## 4. Development Assurance Level (DAL) Recommendation

### 4.1 Recommended DAL

**DAL-C (Design Assurance Level C)**

**Justification**:
- Maps to SAL-3 per DO-278A
- Appropriate for "Major" hazard classification
- Balances safety assurance with open-source development model
- Achievable with focused testing and documentation effort

### 4.2 DAL-C Requirements Summary

1. **Requirements**: Must be documented, reviewed, traceable
2. **Design**: High-level design documented, traceable to requirements
3. **Code**: Coding standards, traceable to design
4. **Testing**: Requirements-based testing, structural coverage analysis
5. **Verification**: Independent verification of requirements and tests
6. **Configuration Management**: Version control, change tracking (✅ existing via Git)
7. **Quality Assurance**: Process audits, documentation reviews
8. **Certification Liaison**: Not applicable (non-TSO equipment)

---

## 5. Gap Analysis

### 5.1 Critical Gaps

1. **No Requirements Specification** ❌
   - Impact: Cannot verify system correctness
   - Effort: 4-6 weeks (1 person)
   - Priority: **CRITICAL**

2. **No Unit Tests** ❌
   - Impact: Cannot verify code correctness, no regression detection
   - Effort: 8-12 weeks (1-2 persons)
   - Priority: **CRITICAL**

3. **No Test Coverage Analysis** ❌
   - Impact: Cannot demonstrate adequacy of testing
   - Effort: 2 weeks (setup tooling)
   - Priority: **HIGH**

4. **No Formal Verification Procedures** ❌
   - Impact: No repeatable verification process
   - Effort: 2-3 weeks
   - Priority: **HIGH**

5. **No Traceability Matrix** ❌
   - Impact: Cannot demonstrate requirements coverage
   - Effort: 1-2 weeks (after requirements documented)
   - Priority: **MEDIUM**

### 5.2 Strengths

1. **Mature Error Handling** ✅
   - Comprehensive error tracking system
   - Automatic recovery mechanisms
   - Visual error indication

2. **Version Control** ✅
   - Git-based configuration management
   - Clear commit history
   - Branching strategy

3. **Integration Testing** ⚠️
   - Manual testing with real hardware
   - Replay capability for regression testing
   - X-Plane simulation support

4. **Code Quality** ⚠️
   - Generally well-structured
   - Good use of Go idioms
   - Room for improvement in documentation

---

## 6. Compliance Roadmap

### Phase 1: Foundation (Months 1-2)
- [ ] Document System Requirements Specification (SRS)
- [ ] Document High-Level Design (HLD)
- [ ] Establish coding standards
- [ ] Set up test coverage tooling
- [ ] Create requirements traceability matrix

### Phase 2: Verification Infrastructure (Months 3-4)
- [ ] Develop unit test framework
- [ ] Create integration test suite
- [ ] Implement automated testing in CI/CD
- [ ] Achieve 50% code coverage

### Phase 3: Compliance Achievement (Months 5-6)
- [ ] Achieve 80% code coverage
- [ ] Complete verification procedures
- [ ] Perform design and code reviews
- [ ] Document verification results
- [ ] Create compliance documentation package

### Phase 4: Continuous Compliance (Ongoing)
- [ ] Regression testing for all changes
- [ ] Requirements updates for new features
- [ ] Periodic design reviews
- [ ] Maintain traceability

**Total Effort Estimate**: 6 months (2 FTE)
**Cost Estimate**: $120,000 - $180,000 (contractor rates)

---

## 7. Quality Metrics

### 7.1 Current Metrics (Estimated)

- **Requirements Coverage**: 0% (no requirements document)
- **Code Coverage**: Unknown (no measurement)
- **Test Automation**: 5% (only CI build testing)
- **Defect Density**: Unknown (no tracking)
- **Documentation Coverage**: 30% (some inline comments, README, development docs)

### 7.2 Target Metrics (SAL-3 Compliance)

- **Requirements Coverage**: 100% (all requirements have tests)
- **Code Coverage**: ≥80% statement, ≥70% decision
- **Test Automation**: ≥95% (automated unit + integration tests)
- **Defect Density**: <0.5 defects per KLOC (post-release)
- **Documentation Coverage**: 100% (all modules documented)

---

## 8. Testing Strategy

### 8.1 Test Levels Required for SAL-3

1. **Unit Testing**
   - Test each function/module in isolation
   - Mock external dependencies
   - Target: 80% statement coverage

2. **Integration Testing**
   - Test subsystem interfaces
   - Hardware-in-loop testing where practical
   - Replay-based regression testing

3. **System Testing**
   - End-to-end functional testing
   - Performance testing (throughput, latency)
   - Stress testing (multiple clients, high message rates)

4. **Requirements-Based Testing**
   - Each requirement SHALL have at least one test
   - Test procedures SHALL be documented
   - Test results SHALL be recorded

### 8.2 Test Coverage Requirements

Per DO-278A Table 6-3 (SAL-3):
- ✅ Requirements-based test coverage
- ✅ Structural coverage analysis (statement + decision)
- ❌ MC/DC coverage (not required for SAL-3)

---

## 9. Security Considerations

While DO-278A focuses on safety, modern CNS systems must address security:

### 9.1 Threat Model

1. **Spoofed ADS-B Messages**
   - Threat: Attacker broadcasts false traffic
   - Mitigation: User awareness that ADS-B is not authenticated, cross-check with visual/ATC

2. **Malicious Configuration**
   - Threat: Attacker modifies settings via web UI
   - Mitigation: WiFi authentication, HTTPS for management interface

3. **Compromised Software Updates**
   - Threat: Malicious OTA update
   - Mitigation: Signed updates, version verification

4. **Denial of Service**
   - Threat: Overwhelm system with messages
   - Mitigation: Message rate limiting, queue management

### 9.2 Security Requirements

**SEC-1**: Web management interface SHALL use authentication
**SEC-2**: OTA updates SHALL be cryptographically signed
**SEC-3**: Configuration changes SHALL be logged
**SEC-4**: System SHALL rate-limit incoming messages
**SEC-5**: System SHALL validate all input data

---

## 10. Regulatory Considerations

### 10.1 FAA Position

- Stratux is **NOT TSO-certified** equipment
- Stratux is **NOT approved** for meeting ADS-B Out mandate
- Stratux **MAY BE USED** as supplemental situational awareness
- No FAA approval required for portable, non-transmitting receivers

### 10.2 Limitations and Disclaimers

**IMPORTANT**: Stratux is an **uncertified** portable receiver intended for **supplemental situational awareness only**.

- NOT for primary navigation
- NOT for terrain avoidance (no TAWS/EGPWS)
- NOT for collision avoidance (no TCAS)
- NOT a replacement for ATC services
- NOT a replacement for visual scanning
- AHRS NOT for attitude reference in IMC

**Required Disclaimer**: All Stratux distributions and documentation SHALL include prominent disclaimers regarding limitations and non-certified status.

---

## 11. Recommendations

### 11.1 Immediate Actions (Next 30 Days)

1. **Create Basic Requirements Document**: Start with high-level functional requirements
2. **Implement Unit Test Framework**: Set up Go testing infrastructure
3. **Enable Coverage Reporting**: Integrate `go test -cover` in CI/CD
4. **Document Safety-Critical Functions**: Identify and document the 20-30 most critical functions

### 11.2 Short-Term Goals (3-6 Months)

1. **Achieve 50% Code Coverage**: Focus on safety-critical modules first
2. **Create Verification Procedures**: Document how to verify each requirement
3. **Establish Baseline Metrics**: Measure current quality metrics
4. **Security Hardening**: Implement authentication and input validation

### 11.3 Long-Term Goals (6-12 Months)

1. **Achieve 80% Code Coverage**: Meet industry best practices
2. **Complete DO-278A Compliance Documentation**: Prepare full compliance package
3. **Third-Party Audit**: Consider independent verification audit
4. **Continuous Compliance**: Maintain compliance as system evolves

---

## 12. Conclusion

Stratux is a mature, feature-rich aviation receiver that provides valuable situational awareness to pilots. While not currently compliant with DO-278A, achieving SAL-3 compliance is feasible and would significantly enhance the system's reliability and safety posture.

The primary gaps are:
1. Lack of formal requirements documentation
2. Absence of comprehensive unit testing
3. No structural coverage analysis

These gaps can be addressed with a focused 6-month effort, bringing Stratux to a level of software assurance appropriate for its role as supplemental aviation equipment.

**Recommended Next Step**: Begin Phase 1 (Foundation) of the compliance roadmap with requirements documentation and test framework setup.

---

## References

1. RTCA DO-278A: Guidelines for Communication, Navigation, Surveillance and Air Traffic Management (CNS/ATM) Systems Software Integrity Assurance
2. RTCA DO-260B: Minimum Operational Performance Standards for 1090 MHz Extended Squitter ADS-B
3. RTCA DO-282B: Minimum Operational Performance Standards for UAT ADS-B
4. FAA TSO-C154d: Universal Access Transceiver (UAT) Automatic Dependent Surveillance-Broadcast (ADS-B) Equipment
5. FAA TSO-C166b: Extended Squitter Automatic Dependent Surveillance-Broadcast (ADS-B) and Traffic Information Service-Broadcast (TIS-B) Equipment
6. FAA ADS-B FAQ: https://www.faa.gov/air_traffic/technology/adsb/faq
7. GRT Avionics Stratux Equipment Supplement

---

**Document Version**: 1.0
**Date**: 2025-10-13
**Author**: Requirements Engineering Analysis
**Status**: Draft for Review
