## Why
1. SonarCloud is configured with `sonar-project.properties` and a `SONAR_TOKEN` secret exists, but the CI workflow does not run SonarCloud analysis. This means code quality issues, bugs, and vulnerabilities are not being automatically detected and reported.

2. GitHub Actions workflows are missing explicit `permissions` blocks, which is flagged as a security issue. Following the principle of least privilege, workflows should declare minimal required permissions.

## What Changes
- Add SonarCloud analysis step to CI workflow
- Configure coverage reporting to SonarCloud (coverage.out already generated)
- Enable automatic quality gate checks on PRs
- **SECURITY**: Add explicit `permissions` blocks to all workflows (least privilege)

## Impact
- Affected specs: ci-cd (new capability)
- Affected code:
  - `.github/workflows/ci.yml` - Add SonarCloud + permissions
  - `.github/workflows/release.yml` - Add permissions
  - `.github/workflows/nightly.yml` - Fix permissions (partially done)
  - `.github/workflows/build-regions.yml` - Add permissions
  - `.github/workflows/release-images.yml` - Fix permissions (partially done)
- Dependencies: Requires `SONAR_TOKEN` secret (already configured)
