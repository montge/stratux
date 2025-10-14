# Coding Standards for Stratux

## Overview

Stratux is aviation safety-related software used in aircraft for traffic awareness and weather information. As such, we aim to follow industry-recognized coding standards for safety-critical systems.

## Applicable Standards

### Primary Standards

1. **MISRA C:2012** (Motor Industry Software Reliability Association)
   - Widely adopted coding standard for C in safety-critical systems
   - Focuses on code reliability, portability, and maintainability
   - Particularly relevant for our C-based SDR and sensor integration code
   - Future compliance target for all C code in dump1090, rtl-ais, and C libraries

2. **JSF AV C++ Coding Standards** (Joint Strike Fighter Air Vehicle)
   - Originally developed by Lockheed Martin for F-35 avionics
   - Applicable to both C and C++ code
   - Emphasizes predictable, verifiable code behavior
   - Relevant for aviation-specific safety considerations

3. **ISO 26262** (Functional Safety for Road Vehicles)
   - While automotive-focused, provides excellent framework for functional safety
   - ASIL (Automotive Safety Integrity Level) concepts applicable to aviation
   - Safety lifecycle management principles
   - Useful for hazard analysis and risk assessment

### Aviation-Specific Considerations

4. **DO-178C** (Software Considerations in Airborne Systems and Equipment Certification)
   - While Stratux is not certified avionics, DO-178C principles are valuable
   - Emphasis on requirements traceability, testing coverage, and verification
   - Structured testing and configuration management practices

5. **RTCA DO-278A** (Guidelines for Communication, Navigation, Surveillance and Air Traffic Management)
   - Guidelines for non-airborne CNS/ATM systems
   - More applicable to Stratux as ground/portable equipment

## Current Status

### Implementation Status: PLANNED

These standards are documented as future goals. Current codebase does not claim compliance.

### Roadmap

1. **Phase 1: Test Coverage** (Current)
   - Achieve comprehensive unit test coverage
   - Establish baseline code quality metrics
   - Document existing code behavior

2. **Phase 2: Static Analysis** (Future)
   - Integrate MISRA C checker tools (e.g., PC-lint, Cppcheck with MISRA addon)
   - Configure static analysis for MISRA compliance
   - Generate compliance reports with deviation rationales

3. **Phase 3: Code Refactoring** (Future)
   - Address MISRA violations systematically
   - Refactor for JSF AV compliance where applicable
   - Document deviations with safety justifications

4. **Phase 4: Functional Safety** (Future)
   - Perform hazard analysis (ISO 26262 concepts)
   - Implement safety mechanisms where needed
   - Establish safety case documentation

## Key Principles to Follow Now

While full compliance is a future goal, developers should keep these principles in mind:

### General Principles

1. **Predictability**: Code behavior should be deterministic and predictable
2. **Simplicity**: Prefer simple, clear code over clever optimizations
3. **Testability**: All code should be unit-testable where possible
4. **Documentation**: Document assumptions, constraints, and safety considerations
5. **Error Handling**: All error conditions must be explicitly handled

### Specific Guidelines

#### Memory Safety
- Avoid dynamic memory allocation in time-critical paths
- Check all pointer dereferences
- Validate array bounds
- Use const correctness

#### Type Safety
- Use explicit type conversions
- Avoid implicit type conversions that may lose precision
- Use sized types (int32_t, uint8_t) for portability

#### Control Flow
- Single entry, single exit for functions where practical
- Avoid deep nesting (max 4 levels recommended)
- No unreachable code
- All switch statements should have default case

#### Functions
- Keep functions focused and small
- Limit function complexity (cyclomatic complexity < 10)
- Document pre-conditions and post-conditions
- Validate all input parameters

#### Concurrency
- Document thread safety for all shared data
- Use appropriate synchronization primitives
- Avoid race conditions
- Document mutex acquisition order to prevent deadlocks

## Testing Standards

### Current Testing Approach

Following DO-178C and ISO 26262 testing principles:

1. **Requirements-Based Testing**: Tests verify specified behavior
2. **Boundary Testing**: Test edge cases and boundaries explicitly
3. **Error Injection**: Test error handling paths
4. **Coverage Goals**:
   - Statement coverage: Target 100% for non-hardware code
   - Branch coverage: Target 100% for decision points
   - MC/DC (Modified Condition/Decision Coverage): Future goal for safety-critical paths

### Test Code Quality

Test code should also follow high standards:
- Clear test naming (describe what is being tested)
- Comprehensive edge case coverage
- Independent, repeatable tests
- No side effects between tests
- Document expected behavior

## Tools and Automation

### Planned Tools

1. **Static Analysis**:
   - PC-lint Plus (MISRA C checker)
   - Cppcheck with MISRA addon (open source alternative)
   - Clang Static Analyzer
   - Go vet (for Go code)

2. **Dynamic Analysis**:
   - Valgrind (memory safety)
   - AddressSanitizer (memory errors)
   - ThreadSanitizer (race conditions)

3. **Coverage Tools**:
   - gcov/lcov for C code
   - go tool cover for Go code
   - Target: >90% coverage for non-hardware code

4. **Code Quality Metrics**:
   - Cyclomatic complexity analysis
   - Code duplication detection
   - Maintainability index

## Compliance Verification

### Future Verification Process

1. **Automated Checks**: CI/CD integration for MISRA/JSF compliance
2. **Manual Review**: Safety-critical code sections require peer review
3. **Deviation Process**: Document and justify any standard deviations
4. **Traceability**: Maintain requirements-to-test traceability matrix

## References

- MISRA C:2012: Guidelines for the use of the C language in critical systems
- JSF AV C++ Coding Standards (Document Number 2RDU00001 Rev C)
- ISO 26262:2018 Road vehicles — Functional safety
- DO-178C: Software Considerations in Airborne Systems and Equipment Certification
- DO-278A: Guidelines for Communication, Navigation, Surveillance and Air Traffic Management

## Notes

- These standards are aspirational goals, not current claims of compliance
- Stratux is not certified avionics equipment
- Standards serve as guidance for improving code quality and reliability
- Safety-critical principles help ensure robust, dependable operation
- Community contributions should be aware of these long-term goals

## Contact

For questions about coding standards compliance or safety considerations, please open an issue on the GitHub repository.

---

*This document establishes coding standards goals for the Stratux project. Implementation is ongoing and iterative. Last updated: 2025-01-13*
