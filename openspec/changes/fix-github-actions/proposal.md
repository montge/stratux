# Change: Fix GitHub Actions CI/CD Pipeline

## Why
Multiple GitHub Actions workflows are failing due to several issues:
1. **Go version mismatch**: Nightly builds use system Go 1.22 but project requires Go 1.24+
2. **Failing test**: `TestLoadTile_InvalidGzippedData` fails in CI
3. **Workflow duplication**: Build logic is duplicated across 4 workflow files
4. **SonarCloud not working**: Coverage reports not being properly uploaded

These failures block nightly builds, waste CI resources, and prevent visibility into code quality metrics.

## What Changes
- **nightly.yml**: Add `actions/setup-go@v5` with Go 1.24 instead of using system `golang-go`
- **build-regions.yml**: Same Go setup fix
- **release-images.yml**: Same Go setup fix
- **ci.yml**: Fix web UI coverage path issue (tests run before npm install sometimes)
- **sonar-project.properties**: Update configuration for proper coverage reporting
- Consolidate common build steps into reusable composite action or workflow

## Impact
- Affected workflows: `ci.yml`, `nightly.yml`, `build-regions.yml`, `release-images.yml`
- Affected code: `.github/workflows/*`, `sonar-project.properties`
- Benefits:
  - Unblock nightly builds
  - Enable SonarCloud quality gate
  - Reduce workflow maintenance burden
  - Faster issue identification through reliable CI
