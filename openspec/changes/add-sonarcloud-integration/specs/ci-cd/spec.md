## ADDED Requirements

### Requirement: SonarCloud Integration
The CI pipeline SHALL run SonarCloud analysis on every push to master and pull request to detect code quality issues, bugs, and security vulnerabilities.

#### Scenario: Analysis runs on push
- **WHEN** code is pushed to master branch
- **THEN** SonarCloud analysis is executed after tests complete
- **AND** coverage report is uploaded to SonarCloud

#### Scenario: Analysis runs on pull request
- **WHEN** a pull request is opened or updated
- **THEN** SonarCloud analysis is executed
- **AND** quality gate status is reported on the PR

### Requirement: Coverage Reporting
The CI pipeline SHALL upload Go test coverage reports to SonarCloud for tracking coverage metrics over time.

#### Scenario: Coverage report uploaded
- **WHEN** tests complete successfully
- **THEN** coverage.out is uploaded to SonarCloud
- **AND** coverage percentage is visible in SonarCloud dashboard

### Requirement: Workflow Security Permissions
All GitHub Actions workflows SHALL declare explicit permissions following the principle of least privilege to minimize security risk.

#### Scenario: Default deny permissions
- **WHEN** a workflow does not need special permissions
- **THEN** it SHALL declare `permissions: {}` at the workflow level
- **AND** only grant specific permissions at the job level where needed

#### Scenario: Release workflow permissions
- **WHEN** a workflow creates GitHub releases
- **THEN** it SHALL declare `permissions: contents: write` only on jobs that create releases
- **AND** other jobs SHALL have minimal or no permissions
