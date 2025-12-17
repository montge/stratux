## ADDED Requirements

### Requirement: SonarCloud Analysis Exclusions
The SonarCloud analysis configuration SHALL exclude third-party code directories that are not maintained by the Stratux project.

#### Scenario: Third-party directories excluded
- **WHEN** SonarCloud analyzes the codebase
- **THEN** the following directories are excluded from analysis:
  - dump978/ (UAT decoder library)
  - web/maui/ (Mobile Angular UI framework)
  - test/metar-to-text/ (METAR test utility)

### Requirement: Database Transaction Safety
Database transactions SHALL be properly rolled back on failure to prevent resource leaks and data corruption.

#### Scenario: Transaction rollback on error
- **WHEN** a database transaction is started with `db.Begin()`
- **AND** an error occurs during the transaction
- **THEN** the transaction MUST be rolled back using `defer tx.Rollback()`

### Requirement: Code Complexity Limits
Production code functions SHALL maintain cognitive complexity at or below 15 to ensure maintainability.

#### Scenario: Function complexity check
- **WHEN** a function is analyzed by SonarCloud
- **THEN** its cognitive complexity SHOULD be 15 or less
- **AND** exceptions may be made for complex algorithms with documented justification

### Requirement: Unused Code Removal
Unused parameters and variables SHALL be removed from production code to improve clarity and maintainability.

#### Scenario: HTTP handler unused parameters
- **WHEN** an HTTP handler function does not use its request or response writer parameter
- **THEN** the parameter SHOULD be replaced with `_` or the function signature should be updated

### Requirement: HTML Accessibility Standards
HTML pages SHALL include proper language attributes for accessibility compliance.

#### Scenario: HTML lang attribute
- **WHEN** an HTML document is created
- **THEN** the `<html>` element MUST include a `lang` attribute (e.g., `lang="en"`)
