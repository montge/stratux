# Stratux Configuration Management Plan

**Document Version**: 1.0
**Last Updated**: 2025-10-13
**Status**: Active
**Compliance**: DO-278A SAL-3

---

## 1. Introduction

### 1.1 Purpose

This Configuration Management Plan (CMP) defines the configuration management activities, tools, and procedures for the Stratux ADS-B/UAT/OGN receiver system to ensure compliance with DO-278A Software Assurance Level 3 (SAL-3) configuration management requirements.

### 1.2 Scope

This plan covers:
- Source code version control
- Build and release management
- Change control procedures
- Configuration identification
- Baseline management
- Problem tracking and resolution

### 1.3 Applicable Documents

| Document | Title | Location |
|----------|-------|----------|
| DO-278A | Software Integrity Assurance for CNS/ATM Systems | RTCA DO-278A (2011) |
| SDP | Software Development Process | `docs/SOFTWARE_DEVELOPMENT_PROCESS.md` |
| SQAP | Software Quality Assurance Plan | `docs/SOFTWARE_QUALITY_ASSURANCE_PLAN.md` |
| SRS | System Requirements Specification | `docs/REQUIREMENTS.md` |

---

## 2. Configuration Management Organization

### 2.1 Roles and Responsibilities

| Role | Responsibilities | Authority |
|------|------------------|-----------|
| **Configuration Manager** | Oversee CM activities, maintain baselines, approve changes | Approve/reject change requests |
| **Release Manager** | Build process, version management, release coordination | Create release baselines |
| **Developer** | Implement changes, commit code, create pull requests | Read/write access to feature branches |
| **Reviewer** | Code review, change approval | Approve/reject pull requests |
| **Project Lead** | Approve significant changes, release approval | Final authority on master branch |

**Current Assignments**:
- **Configuration Manager**: Automated (Git + GitHub) + Project Lead oversight
- **Release Manager**: Automated (GitHub Actions) + Project Lead approval
- **Developers**: Open-source contributors with GitHub accounts
- **Reviewers**: Core team members with write access
- **Project Lead**: TBD (stratux organization maintainer)

### 2.2 CM Tools

| Tool | Version | Purpose | Qualification Status |
|------|---------|---------|---------------------|
| **Git** | 2.x | Version control | Not required (not in verification path) |
| **GitHub** | N/A | Hosting, CI/CD, issue tracking | Not required |
| **GitHub Actions** | N/A | Build automation | Required (verification tool) |
| **Make** | 4.x | Build orchestration | Not required |

---

## 3. Configuration Identification

### 3.1 Configuration Items (CIs)

**Software Configuration Items**:

| CI ID | Item Name | Description | Location |
|-------|-----------|-------------|----------|
| **CI-001** | Source Code | Go application code | `main/`, `selfupdate/`, `dump1090/`, `test/` |
| **CI-002** | Build Scripts | Makefile, build scripts | `Makefile`, `scripts/` |
| **CI-003** | CI/CD Workflows | GitHub Actions workflows | `.github/workflows/` |
| **CI-004** | Documentation | Requirements, design, process docs | `docs/` |
| **CI-005** | Web Interface | HTML, JavaScript, CSS | `web/` |
| **CI-006** | Configuration Files | System config, services | `image/`, `stratux.conf`, `config.txt` |
| **CI-007** | Test Data | Test fixtures, test cases | `test/`, `testdata/` |
| **CI-008** | Dependencies | Go modules | `go.mod`, `go.sum` |

**Hardware Configuration** (target platform):
- Raspberry Pi 3B, 3B+, 4, Zero 2W (ARM64)
- RTL-SDR dongles (R820T/R820T2 tuner)
- GPS receivers (u-blox, Prolific, NMEA-compatible)
- AHRS sensors (MPU-9250, ICM-20948, BMP-280/388)

### 3.2 Naming Conventions

#### File Naming
- Go source: `lowercase.go` (e.g., `traffic.go`, `gen_gdl90.go`)
- Go tests: `lowercase_test.go` (e.g., `traffic_test.go`)
- Documentation: `UPPERCASE.md` (e.g., `REQUIREMENTS.md`)
- Scripts: `lowercase.sh` (e.g., `getversion.sh`)

#### Version Numbering
**Semantic Versioning**: `MAJOR.MINOR.PATCH`

- **MAJOR**: Incompatible changes (e.g., config file format breaking change)
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

**Examples**:
- `v1.6.0` - Major/minor release
- `v1.6.1` - Patch release
- `v2.0.0` - Major version (breaking changes)

**Development Versions**: `0.0.YYYYMMDD-<commit-hash>`
- Example: `0.0.20251013-baeb308`

#### Branch Naming
- `master` - Main development branch (always stable)
- `feature/<description>` - New features (e.g., `feature/add-ogn-support`)
- `bugfix/<description>` - Bug fixes (e.g., `bugfix/fix-gps-overflow`)
- `release/<version>` - Release preparation (e.g., `release/v1.6.0`)
- `hotfix/<description>` - Critical production fixes (e.g., `hotfix/security-fix`)

#### Tag Naming
- Release tags: `v<MAJOR>.<MINOR>.<PATCH>` (e.g., `v1.6.0`)
- Nightly tags: `nightly-<YYYYMMDD>` (e.g., `nightly-20251013`)
- Annotated tags preferred: `git tag -a v1.6.0 -m "Release 1.6.0"`

---

## 4. Baseline Management

### 4.1 Baseline Types

| Baseline Type | Description | Trigger | Contents | Tag Format |
|---------------|-------------|---------|----------|------------|
| **Development Baseline** | Each commit to master | Every commit | All CIs at commit | Commit SHA |
| **Test Baseline** | Nightly builds for testing | Nightly (2 AM UTC) | Source + build artifacts | `nightly-YYYYMMDD` |
| **Release Baseline** | Official releases | Version tag pushed | Source + .deb + .img | `vX.Y.Z` |

### 4.2 Baseline Creation

#### Development Baseline
**Automatic** - every commit to master is a baseline
- **Identification**: Git commit SHA (e.g., `baeb308a1b2c3d4e5f6g7h8i9j0k`)
- **Contents**: All tracked files at that commit
- **Reproducibility**: `git checkout <commit-sha>` recreates exact baseline

#### Test Baseline (Nightly)
**Automatic** - triggered by `.github/workflows/nightly.yml` at 2 AM UTC daily

**Process**:
1. Check if commits in last 24 hours (if none, skip build)
2. Build US and EU .deb packages
3. Create prerelease on GitHub with tag `nightly-YYYYMMDD`
4. Upload .deb artifacts
5. Clean up old nightlies (keep last 7 days)

**Artifacts**:
- `stratux-US-0.0.YYYYMMDD-<hash>-arm64.deb`
- `stratux-EU-0.0.YYYYMMDD-<hash>-arm64.deb`

**Retention**: 7 days (old nightlies automatically deleted)

#### Release Baseline
**Manual trigger** - developer pushes version tag to GitHub

**Process**:
1. Developer creates annotated tag: `git tag -a v1.6.0 -m "Release 1.6.0"`
2. Push tag: `git push origin v1.6.0`
3. GitHub Actions workflow `.github/workflows/release-images.yml` triggered
4. Build both US and EU configurations:
   - Full SD card images (stratux-lite-vX.Y.Z-{US,EU}.img)
   - Debian packages (stratux-{US,EU}-X.Y.Z-arm64.deb)
5. Create draft release on GitHub
6. Project lead reviews and publishes release

**Artifacts**:
- `stratux-lite-v1.6.0-US.img` (and .zip)
- `stratux-lite-v1.6.0-EU.img` (and .zip)
- `stratux-US-1.6.0-arm64.deb`
- `stratux-EU-1.6.0-arm64.deb`

**Retention**: Permanent (all releases preserved on GitHub)

### 4.3 Baseline Integrity

**Protection Mechanisms**:
1. **Git Immutability**: Commits cannot be modified (SHA-256 hash)
2. **Branch Protection**: Master branch requires PR reviews, no force push
3. **Tag Protection**: Release tags cannot be moved (delete/recreate if error)
4. **Checksums**: All release artifacts include SHA-256 checksums
5. **Build Reproducibility**: Same source → same binary (verified in CI)

**Verification**:
```bash
# Verify Git commit integrity
git fsck --full

# Verify artifact checksum
sha256sum stratux-US-1.6.0-arm64.deb
# Compare with published checksum in release notes
```

---

## 5. Change Control

### 5.1 Change Request Process

All changes follow this workflow:

```
┌─────────────────────────────────────────────────────┐
│  1. CHANGE REQUEST                                   │
│     • GitHub Issue created                           │
│     • Labels: bug, enhancement, security, etc.       │
│     • Priority assigned                              │
└──────────────┬──────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────┐
│  2. IMPACT ANALYSIS                                  │
│     • Requirements impact?                           │
│     • Design impact?                                 │
│     • Test impact?                                   │
│     • Documentation impact?                          │
└──────────────┬──────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────┐
│  3. APPROVAL                                         │
│     • Minor: Developer proceeds                      │
│     • Major: Project lead approval required          │
│     • Requirements change: Update REQUIREMENTS.md    │
└──────────────┬──────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────┐
│  4. IMPLEMENTATION                                   │
│     • Create feature/bugfix branch                   │
│     • Develop and test locally                       │
│     • Update documentation                           │
│     • Commit with structured message                 │
└──────────────┬──────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────┐
│  5. REVIEW                                           │
│     • Create pull request                            │
│     • Automated checks (CI/CD)                       │
│     • Peer review                                    │
│     • Approval required                              │
└──────────────┬──────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────┐
│  6. MERGE                                            │
│     • Merge to master                                │
│     • CI build and test                              │
│     • Deployment (nightly or release)                │
└──────────────┬──────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────┐
│  7. CLOSURE                                          │
│     • Verify tests passing                           │
│     • Update issue (Closes: #123)                    │
│     • Update traceability matrix                     │
└─────────────────────────────────────────────────────┘
```

### 5.2 Change Classification

| Change Type | Description | Approval Required | Examples |
|-------------|-------------|-------------------|----------|
| **Minor** | Bug fixes, refactoring, documentation | Peer review only | Fix typo, improve logging, refactor function |
| **Major** | New features, design changes | Project lead | Add OGN support, change traffic fusion algorithm |
| **Critical** | Safety-related, security fixes | Project lead + expedited review | GPS validation error, buffer overflow fix |
| **Requirements** | Changes to REQUIREMENTS.md | Project lead + traceability update | Add new requirement, modify acceptance criteria |

### 5.3 Emergency Changes

**Emergency Change**: Critical defect or security vulnerability requiring immediate fix

**Process**:
1. Create hotfix branch from master: `hotfix/<description>`
2. Implement fix with tests
3. Expedited review (same-day turnaround target)
4. Merge to master
5. Create patch release (if currently released version affected)
6. Post-review: Document change in next QA audit

**Approval**: Project lead (retroactive approval acceptable for true emergencies)

### 5.4 Commit Message Format

**Structure**:
```
<type>: <short summary (50 chars max)>

<detailed description (optional, wrap at 72 chars)>

Implements: FR-101, FR-102
Fixes: #123
Closes: #456
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `test`: Adding/updating tests
- `refactor`: Code restructuring (no behavior change)
- `perf`: Performance improvement
- `ci`: CI/CD changes
- `build`: Build system changes
- `style`: Code formatting (no logic change)

**Examples**:
```
feat: Add OGN FLARM support for EU region

Implements OGN 868 MHz reception for glider tracking in Europe.
Supports OGN-flavored ADS-B decoding and integration with traffic fusion.

Implements: FR-103, FR-310
Closes: #234
```

```
fix: Correct GPS overflow for high altitude

GPS altitude field overflowed at 32,767 feet due to int16 usage.
Changed to int32 to support altitudes up to 60,000 feet.

Implements: FR-204
Fixes: #345
```

---

## 6. Version Control

### 6.1 Repository Structure

```
stratux/
├── .github/
│   └── workflows/           # CI/CD automation
│       ├── build-regions.yml
│       ├── nightly.yml
│       └── release-images.yml
├── docs/                    # Documentation (CIs)
│   ├── REQUIREMENTS.md
│   ├── HIGH_LEVEL_DESIGN.md
│   ├── SOFTWARE_DEVELOPMENT_PROCESS.md
│   ├── SOFTWARE_QUALITY_ASSURANCE_PLAN.md
│   └── CONFIGURATION_MANAGEMENT_PLAN.md
├── main/                    # Core application (Go)
│   ├── traffic.go
│   ├── gps.go
│   ├── gen_gdl90.go
│   ├── sdr.go
│   └── sensors.go
├── test/                    # Test code and fixtures
│   └── testdata/
├── web/                     # Web interface
│   ├── plates/
│   └── css/
├── image/                   # System configuration for image builds
├── scripts/                 # Build and utility scripts
│   └── getversion.sh
├── Makefile                 # Build orchestration
├── go.mod                   # Go dependencies
├── go.sum                   # Dependency checksums
├── CLAUDE.md                # Project overview for AI assistants
├── README.md                # User documentation
└── LICENSE                  # GPL v3
```

### 6.2 Branch Management

#### Master Branch
- **Protection**: Requires pull request reviews, status checks must pass
- **Quality**: Always buildable and testable
- **History**: Linear history preferred (rebase before merge)
- **Access**: Write access restricted to maintainers

#### Feature Branches
- **Naming**: `feature/<description>` (e.g., `feature/add-bluetooth`)
- **Lifetime**: Short-lived (merge within 1-2 weeks)
- **Origin**: Branch from master
- **Destination**: Merge back to master via PR

#### Bugfix Branches
- **Naming**: `bugfix/<description>` (e.g., `bugfix/gps-timeout`)
- **Lifetime**: Short-lived (merge within days)
- **Origin**: Branch from master
- **Destination**: Merge back to master via PR

#### Release Branches
- **Naming**: `release/<version>` (e.g., `release/v1.6.0`)
- **Purpose**: Final testing and preparation before release
- **Lifetime**: 1-2 weeks
- **Origin**: Branch from master when feature-complete
- **Allowed changes**: Bug fixes, documentation updates, version number updates
- **Destination**: Tag for release, merge back to master

#### Hotfix Branches
- **Naming**: `hotfix/<description>` (e.g., `hotfix/security-cve-2025-1234`)
- **Purpose**: Emergency fixes for released versions
- **Lifetime**: Hours to 1 day
- **Origin**: Branch from release tag
- **Destination**: Merge to master, create patch release

### 6.3 Merge Strategy

**Preferred**: Rebase and merge (linear history)
```bash
git checkout feature/my-feature
git rebase master
# Resolve conflicts if any
git checkout master
git merge --ff-only feature/my-feature
```

**Alternative**: Squash and merge (for messy feature branches)
- Combine all feature commits into single commit on master
- Preserves clean history
- Use for branches with many WIP commits

**Prohibited**: Merge commits with multiple parents (creates complex history)

### 6.4 Pull Request Workflow

**PR Creation**:
1. Push feature branch to GitHub: `git push origin feature/my-feature`
2. Create PR on GitHub (master ← feature/my-feature)
3. Fill out PR template with:
   - Description of changes
   - Requirements implemented (Implements: FR-xxx)
   - Issues closed (Closes: #123)
   - Testing performed
   - Checklist completed

**PR Review**:
- Automated checks run (CI build, tests, coverage)
- Peer reviewer assigned
- Reviewer comments on code
- Author addresses feedback
- Reviewer approves

**PR Merge**:
- All checks passing (green checkmarks)
- At least 1 approval
- No unresolved review comments
- Merge button enabled
- Merged by author or reviewer

---

## 7. Build and Release Management

### 7.1 Build Types

| Build Type | Trigger | Artifacts | Distribution | Retention |
|------------|---------|-----------|--------------|-----------|
| **Development** | Manual (`make`) | Local binary | Developer machine | N/A |
| **CI** | Every commit to master | N/A (verification only) | N/A | Logs: 90 days |
| **Nightly** | Daily (2 AM UTC) | .deb packages (US, EU) | GitHub Releases (prerelease) | 7 days |
| **On-Demand** | Manual workflow trigger | .deb packages (US, EU) | Artifacts download only | 90 days |
| **Release** | Version tag | .deb + .img (US, EU) | GitHub Releases | Permanent |

### 7.2 Build Process

**Native ARM64 Build** (preferred):
- GitHub Actions: `ubuntu-24.04-arm` runners
- Native compilation (no emulation)
- Fast build time (~2 minutes for .deb)

**Build Steps**:
1. Check out source code (Git)
2. Install dependencies (RTL-SDR libs, Go toolchain)
3. Modify Makefile for region configuration (US or EU)
4. Run `make dpkg` (Debian package build)
5. Generate checksums (SHA-256)
6. Upload artifacts to GitHub

**Build Reproducibility**:
- Same source code → same binary output
- Verified by comparing checksums across multiple builds
- Ensured by pinning dependency versions (go.mod)

### 7.3 Release Process

**Pre-Release Checklist**:
- [ ] All features for release complete
- [ ] All tests passing
- [ ] Code coverage ≥80% (target)
- [ ] No critical or high-severity open issues
- [ ] Documentation updated (README, CHANGELOG)
- [ ] Requirements traceability up-to-date
- [ ] Field testing completed (for major releases)

**Release Steps**:
1. **Prepare Release Branch** (optional, for major releases):
   ```bash
   git checkout -b release/v1.6.0
   # Update version numbers in documentation
   git commit -m "Prepare release v1.6.0"
   git push origin release/v1.6.0
   ```

2. **Create Release Tag**:
   ```bash
   git checkout master
   git pull
   git tag -a v1.6.0 -m "Release version 1.6.0"
   git push origin v1.6.0
   ```

3. **Automated Build**: GitHub Actions builds full release (45-60 minutes)

4. **Review Draft Release**:
   - Check artifacts uploaded correctly
   - Review auto-generated release notes
   - Add detailed changelog
   - Add known issues (if any)
   - Add installation instructions

5. **Publish Release**: Click "Publish release" button on GitHub

6. **Announcement**: Post in GitHub Discussions, notify community

**Post-Release**:
- Monitor for issues (first 48 hours critical)
- Prepare hotfix process if critical defect found
- Update roadmap and metrics

### 7.4 Version Identification in Software

**Runtime Version Display**:
- Web UI: Shows version on main page and settings
- Command-line: `stratux -version`
- Status output: `mySituation.Version`

**Implementation**:
```go
// Version information compiled into binary
var (
	Version   = "dev"     // Set by build process
	BuildDate = "unknown" // Set by build process
	GitCommit = "unknown" // Set by build process
)
```

**Build-time Injection**:
```makefile
VERSION := $(shell ./scripts/getversion.sh)
LDFLAGS := -X main.Version=$(VERSION) \
           -X main.BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
           -X main.GitCommit=$(shell git rev-parse --short HEAD)
```

---

## 8. Configuration Status Accounting

### 8.1 Status Tracking

**Git Provides**:
- **What**: All files in repository, with full content history
- **When**: Commit timestamps (UTC)
- **Who**: Commit author and committer
- **Why**: Commit messages
- **Where**: Repository URL and commit SHA

**GitHub Provides**:
- **Issue Tracking**: Problems, enhancements, requirements changes
- **Pull Requests**: Code review history, approval records
- **Releases**: Published baselines with artifacts
- **Actions**: Build history, test results, deployment records

### 8.2 Traceability

**Forward Traceability** (Requirements → Implementation):
```
Requirement FR-101 (REQUIREMENTS.md)
    ↓
Design: Radio subsystem (HIGH_LEVEL_DESIGN.md, Section 4.2)
    ↓
Code: main/sdr.go lines 450-520
    ↓
Tests: main/sdr_test.go TestProcess1090Message
    ↓
Verification: Test passed (CI log)
```

**Backward Traceability** (Code → Requirements):
```go
// Implements: FR-101 (1090 MHz ADS-B Reception)
// Verifies message per DO-260B Section 2.2.3.2
func Process1090Message(msg []byte) (*TrafficInfo, error) {
    // ...
}
```

**Traceability Matrix**: `docs/TRACEABILITY_MATRIX.xlsx`
- Columns: Requirement ID, Design Reference, Code Reference, Test Reference, Verification Status
- Updated after each requirement/design/code/test change
- Reviewed monthly for completeness

### 8.3 Audit Trail

**Available Information**:
- **Git Log**: Complete history of all changes (`git log --all --graph --decorate`)
- **GitHub Issues**: Problem reports and resolutions
- **Pull Requests**: Code reviews and approvals
- **Actions Logs**: Build and test results
- **Release Notes**: Published changes for each release

**Retention**: Permanent (Git history never deleted)

---

## 9. Configuration Audits

### 9.1 Audit Types

#### Functional Configuration Audit (FCA)
**Purpose**: Verify that software meets requirements
**Frequency**: Before each major release
**Scope**: Requirements verification, test results, performance validation
**Deliverable**: FCA report documenting compliance

**Checklist**:
- [ ] All requirements implemented
- [ ] All requirements verified (tests passing)
- [ ] Traceability matrix complete
- [ ] Performance requirements met
- [ ] Documentation complete

#### Physical Configuration Audit (PCA)
**Purpose**: Verify that deliverables match documented configuration
**Frequency**: Before each release
**Scope**: Build artifacts, checksums, version identification, baseline contents
**Deliverable**: PCA report with artifact inventory

**Checklist**:
- [ ] Artifact checksums match
- [ ] Version numbers correct
- [ ] Git tag matches release version
- [ ] All required artifacts present
- [ ] Build reproducible

#### Process Configuration Audit (QA Audit)
**Purpose**: Verify configuration management process followed
**Frequency**: Quarterly
**Scope**: CM procedures, change control, baseline management, access control
**Deliverable**: Audit report with findings and corrective actions

**Checklist**:
- [ ] Change control followed for all changes
- [ ] Branch protection enforced
- [ ] Code reviews completed
- [ ] Baselines properly tagged
- [ ] Documentation up-to-date

### 9.2 Audit Records

**Maintained Records**:
- Audit reports (FCA, PCA, QA)
- Findings and corrective actions
- Audit schedules and completion records

**Storage**: `docs/audits/` directory in Git repository

---

## 10. Backup and Recovery

### 10.1 Repository Backup

**Primary Repository**: https://github.com/stratux/stratux (GitHub)

**Backup Strategy**:
- **GitHub Redundancy**: GitHub maintains multiple replicas of all repositories
- **Local Clones**: Each developer has full repository clone
- **Mirror Repository** (optional): GitLab or Bitbucket mirror for redundancy

**Recovery**:
```bash
# Any developer can restore from local clone
git clone /path/to/local/stratux stratux-restored
cd stratux-restored
git remote set-url origin https://github.com/stratux/stratux
git push --mirror
```

### 10.2 Release Artifact Backup

**Primary Storage**: GitHub Releases

**Backup Storage** (recommended):
- Archive important releases to external storage (cloud or local)
- Retain at least last 3 major releases permanently
- Verify checksums periodically

**Recovery**:
- Download from GitHub Releases page
- Or restore from backup storage
- Verify SHA-256 checksums before use

### 10.3 Documentation Backup

**Primary**: Git repository (docs/ directory)
**Backup**: Same as repository backup (Git distributed nature provides redundancy)

---

## 11. Access Control

### 11.1 Repository Access Levels

| Role | Access Level | Permissions |
|------|-------------|-------------|
| **Public** | Read | Clone, view code, download releases |
| **Contributor** | Read | Same as public (external contributors submit PRs from forks) |
| **Developer** | Write | Create branches, submit PRs (not directly to master) |
| **Maintainer** | Admin | Merge PRs, create releases, manage settings |

**Master Branch Protection**:
- Direct commits prohibited
- Requires pull request review (at least 1 approval)
- Status checks must pass (CI build, tests)
- No force push
- No deletion

### 11.2 Access Request Process

**New Contributors**:
1. Create GitHub account
2. Fork stratux repository
3. Submit pull request from fork
4. After 3+ successful contributions, may request write access

**Write Access Request**:
1. Submit issue with label `access-request`
2. Maintainer reviews contribution history
3. If approved, invite sent to GitHub organization

**Access Revocation**:
- Inactive for >6 months: Access reviewed, may be revoked
- Code of conduct violation: Immediate revocation
- Project departure: Access removed upon request

---

## 12. Supplier and Third-Party Software Control

### 12.1 Third-Party Dependencies

**Management**: Go modules (`go.mod`, `go.sum`)

**Critical Dependencies**:
- **RTL-SDR libraries**: https://github.com/stratux/rtlsdr
  - License: GPL 2.0
  - Version: 2.0.2-2
  - Hosted: GitHub Releases (Stratux fork)
  - Validation: SHA-256 checksum verification in CI

**Go Modules**:
- All dependencies declared in `go.mod`
- Checksums locked in `go.sum` (integrity verification)
- Updated intentionally (not automatically)

### 12.2 Dependency Update Process

**Security Updates** (Critical):
1. GitHub Dependabot alerts on vulnerability
2. Assess impact and urgency
3. Update dependency in `go.mod`
4. Run full test suite
5. Merge as hotfix if critical

**Feature Updates** (Non-Critical):
1. Evaluate benefit vs. risk
2. Update in feature branch
3. Full testing and validation
4. Merge via standard PR process

**Review Frequency**: Monthly check for security advisories

### 12.3 License Compliance

**Stratux License**: GPL v3

**Allowed Dependency Licenses**:
- MIT, BSD, Apache 2.0 (permissive, GPL-compatible)
- LGPL (library use acceptable)
- GPL v2/v3 (compatible with Stratux license)

**Prohibited**:
- Proprietary licenses
- Non-GPL-compatible licenses

**Verification**: Review LICENSE files in all dependencies

---

## 13. Data Rights and Deliverables

### 13.1 Open Source Licensing

**Stratux License**: GNU General Public License v3.0 (GPL v3)

**Rights Granted to Users**:
- Freedom to use for any purpose
- Freedom to study and modify source code
- Freedom to distribute copies
- Freedom to distribute modified versions
- Copyleft: Derivative works must be GPL v3

**Obligations**:
- Provide source code with binary distributions
- Preserve copyright and license notices
- Distribute modifications under GPL v3

### 13.2 Contributor License

**Contribution Agreement**: By submitting pull request, contributors agree:
- Grant copyright license to Stratux project
- Contributions distributed under GPL v3
- Contributor retains copyright to their contributions
- No patent claims against project or users

**DCO (Developer Certificate of Origin)**: Contributors certify origin of contributions (optional but recommended)

---

## 14. Training and Competency

### 14.1 CM Training for Contributors

**Required Knowledge**:
- Git basics (clone, branch, commit, push, pull, merge)
- GitHub workflow (fork, PR, review)
- Commit message conventions
- Branch naming conventions
- Change control process

**Training Resources**:
- Git documentation: https://git-scm.com/doc
- GitHub guides: https://guides.github.com/
- Stratux CONTRIBUTING.md (if exists)
- This document (CONFIGURATION_MANAGEMENT_PLAN.md)

### 14.2 Competency Verification

**Method**: Successful completion of first pull request with guidance

**Mentorship**: Experienced contributors guide new contributors through first PR

---

## 15. Continuous Improvement

### 15.1 CM Process Improvement

**Review Triggers**:
- Quarterly process review
- Audit findings
- Significant CM issues (e.g., incorrect baseline release)
- Tool updates (e.g., new GitHub features)

**Improvement Process**:
1. Identify improvement opportunity
2. Propose change (GitHub issue)
3. Review by CM team
4. Approval by project lead
5. Update this document (CMP)
6. Communicate changes to team
7. Monitor effectiveness

### 15.2 Metrics

**Tracked Metrics**:
- Build success rate (target: ≥95%)
- PR merge time (target: <48 hours)
- Baseline integrity violations (target: 0)
- Unauthorized master commits (target: 0)
- Change control violations (target: 0)

**Reporting**: Monthly CM metrics in GitHub Discussions

---

## 16. Configuration Management Plan Approval

**Plan Owner**: Configuration Manager (TBD)
**Reviewed By**: Project Lead, QA Lead
**Approved By**: Project Lead
**Effective Date**: 2025-10-13

**Change History**:

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-13 | Requirements Engineering | Initial version for DO-278A SAL-3 compliance |

---

## 17. References

1. **RTCA DO-278A**, "Guidelines for Communication, Navigation, Surveillance and Air Traffic Management (CNS/ATM) Systems Software Integrity Assurance," Section 8 (Configuration Management), December 2011
2. **Stratux SOFTWARE_DEVELOPMENT_PROCESS.md**, Software Development Process, Version 1.0
3. **Stratux SOFTWARE_QUALITY_ASSURANCE_PLAN.md**, Software Quality Assurance Plan, Version 1.0
4. **Stratux REQUIREMENTS.md**, System Requirements Specification, Version 1.0
5. **Stratux HIGH_LEVEL_DESIGN.md**, High-Level Design Document, Version 1.0
6. **Stratux RELEASE_PROCESS.md**, Release Process Documentation, Version 1.0
7. **Git Documentation**, https://git-scm.com/doc
8. **GitHub Guides**, https://guides.github.com/
9. **Semantic Versioning 2.0.0**, https://semver.org/

---

**END OF CONFIGURATION MANAGEMENT PLAN**
