## ADDED Requirements

### Requirement: Reusable Build Setup Action
The CI/CD system SHALL provide a reusable composite action for consistent build environment setup.

#### Scenario: Build setup in CI workflow
- **WHEN** the CI workflow runs
- **THEN** it SHALL use the composite action to set up Go, submodules, and dependencies

#### Scenario: Build setup in nightly workflow
- **WHEN** the nightly workflow runs
- **THEN** it SHALL use the composite action to set up the build environment

#### Scenario: Go version consistency
- **WHEN** any build workflow runs
- **THEN** the Go version SHALL match the version specified in go.mod

### Requirement: SonarCloud Coverage Reporting
The CI system SHALL report test coverage to SonarCloud for quality tracking.

#### Scenario: Coverage upload on CI pass
- **WHEN** tests pass in the CI workflow
- **THEN** coverage data SHALL be uploaded to SonarCloud

#### Scenario: Coverage format
- **WHEN** coverage is generated
- **THEN** it SHALL be in the format expected by SonarCloud (Go coverage profile)

## MODIFIED Requirements

### Requirement: Nightly Build Reliability
The nightly build workflow SHALL produce US and EU region Debian packages reliably.

#### Scenario: Successful nightly build
- **WHEN** there are new commits in the last 24 hours
- **THEN** the nightly workflow SHALL build both US and EU packages

#### Scenario: Go toolchain availability
- **WHEN** the nightly workflow runs on ubuntu-24.04-arm
- **THEN** it SHALL install Go using actions/setup-go to ensure correct version

#### Scenario: Build failure handling
- **WHEN** a build step fails
- **THEN** the workflow SHALL report the failure clearly with actionable error messages

### Requirement: Web UI Test Coverage
The CI workflow SHALL run web UI tests and report coverage.

#### Scenario: Web UI tests run
- **WHEN** the CI workflow runs
- **THEN** web UI tests SHALL run after npm dependencies are installed

#### Scenario: Coverage artifact upload
- **WHEN** web UI tests complete (pass or fail)
- **THEN** coverage artifacts SHALL be uploaded if they exist

## REMOVED Requirements

### Requirement: System Go Package Dependency
**Reason**: System packages provide outdated Go versions
**Migration**: Use `actions/setup-go@v5` with `go-version-file: 'go.mod'`
