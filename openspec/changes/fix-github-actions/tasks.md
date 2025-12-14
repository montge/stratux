# Tasks: Fix GitHub Actions CI/CD Pipeline

## 1. Fix Go Version in Build Workflows

- [x] 1.1 Update `nightly.yml` to use `actions/setup-go@v5` with Go from go.mod
- [x] 1.2 Update `build-regions.yml` to use `actions/setup-go@v5` with Go from go.mod
- [x] 1.3 Update `release-images.yml` to use `actions/setup-go@v5` with Go from go.mod
- [x] 1.4 Update `ci.yml` to use `go-version-file: 'go.mod'` (was using hardcoded Go 1.21)

## 2. Fix Failing CI Test

- [x] 2.1 Investigate `TestLoadTile_InvalidGzippedData` failure - caused by Go version mismatch
- [x] 2.2 Verify test passes locally with correct Go version (1.24.1)

## 3. Fix Web UI Coverage Upload

- [x] 3.1 Web UI tests work correctly when Go tests pass (npm steps weren't reached due to Go test failure)
- [x] 3.2 Artifact upload already has `if: always()`
- [x] 3.3 Verify web UI tests pass and generate coverage locally (77 tests passed)

## 4. Fix SonarCloud Configuration

- [x] 4.1 Verify SONAR_TOKEN secret is configured in repository (confirmed)
- [x] 4.2 `sonar-project.properties` already has correct source paths
- [x] 4.3 coverage.out is in correct format for SonarCloud
- [x] 4.4 Verify SonarCloud scan runs successfully in CI (confirmed on master)

## 5. Add Gitflow Support

- [x] 5.1 Update `ci.yml` to trigger on both `master` and `develop` branches
- [x] 5.2 Create develop branch from master
- [x] 5.3 Push develop branch to origin
- [x] 5.4 Note: SonarCloud free tier only scans default branch (master). Branch analysis expected 2026.

## 6. Fix Go Formatting

- [x] 6.1 Run gofmt on test files with formatting issues
- [x] 6.2 Verify formatting check passes in CI

## 7. Consolidate Workflow Duplication (Future)

- [ ] 7.1 Create reusable composite action for Go build setup
- [ ] 7.2 Update all workflows to use the composite action
- [ ] 7.3 Remove duplicated dependency installation steps

## 8. Verification

- [x] 8.1 Run all Go tests locally (all pass with Go 1.24.1)
- [x] 8.2 Run web UI tests locally (77 tests pass)
- [x] 8.3 Verify all workflows pass in CI after changes (master CI passed)
- [x] 8.4 Verify SonarCloud receives coverage data (scan completed successfully)
