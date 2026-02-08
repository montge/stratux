# Design: Fix GitHub Actions CI/CD Pipeline

## Context

The Stratux project uses GitHub Actions for CI/CD with multiple workflows:
- **ci.yml**: Run on every push/PR - tests, coverage, SonarCloud scan, builds package
- **nightly.yml**: Scheduled daily - builds US and EU region packages
- **build-regions.yml**: Manual trigger - builds US/EU packages on demand
- **release-images.yml**: On tags - builds full Pi images and packages
- **release.yml**: On tags - legacy full image build

Current issues:
- Go version mismatch (system Go 1.22 vs required Go 1.24)
- Duplicated build setup across 4 workflows (~50 lines each)
- Flaky test failing in CI
- SonarCloud not receiving coverage data

## Goals / Non-Goals

**Goals:**
- Fix all failing workflows
- Enable reliable nightly builds
- Get SonarCloud working for coverage tracking
- Reduce maintenance burden through code reuse

**Non-Goals:**
- Adding new CI/CD features
- Changing build architecture
- Modifying release process

## Decisions

### Decision 1: Use `actions/setup-go@v5` consistently

**What:** Replace system `golang-go` package with `actions/setup-go@v5` in all workflows.

**Why:**
- System package is outdated (Go 1.22 on Ubuntu 24.04)
- Project now requires Go 1.24+ (specified in go.mod)
- `actions/setup-go` provides caching and consistent versions

**Alternative considered:** Pin go.mod to Go 1.22
- Rejected: Would limit access to new Go features and security fixes

### Decision 2: Create reusable composite action

**What:** Create `.github/actions/setup-stratux-build/action.yml` that handles:
- Go setup with version from go.mod
- Submodule initialization
- RTL-SDR library installation
- Build dependencies

**Why:**
- Eliminates ~50 lines of duplication per workflow
- Single point of maintenance
- Consistent build environment across workflows

**Alternative considered:** GitHub reusable workflows
- Rejected: Composite actions are simpler and don't require separate workflow files

### Decision 3: Fix test instead of skipping

**What:** Fix `TestLoadTile_InvalidGzippedData` to handle the gzip error gracefully.

**Why:**
- Test is catching a real issue (panic on invalid gzip data)
- Skipping would hide potential bugs
- Proper fix improves robustness

## Risks / Trade-offs

**Risk:** Composite action adds abstraction layer
- Mitigation: Keep action focused and well-documented

**Risk:** Go version bump might break builds
- Mitigation: Go 1.24 is already specified in go.mod, so this is just fixing the mismatch

## Migration Plan

1. Fix workflows one at a time (ci.yml first since it runs most often)
2. Create composite action and migrate workflows incrementally
3. Verify each workflow passes before proceeding
4. Update SonarCloud configuration last (depends on CI passing)

## Open Questions

- Should we add explicit Go version to all workflow files or read from go.mod?
  - Recommendation: Use go.mod as source of truth via `go-version-file: 'go.mod'`
