# Change: Fix SonarCloud Reliability and Maintainability Issues

## Why

SonarCloud analysis has identified 71 reliability issues (bugs) and 3800+ maintainability issues (code smells) in the codebase. While many are in third-party code that we should exclude, there are actionable issues in our own code that affect code quality, maintainability, and potentially reliability.

## What Changes

### Phase 1: Reliability Issues (Bugs)
- Fix transaction rollback issue in datalog.go
- Fix web UI bugs in web/plates/

### Phase 2: Critical Maintainability Issues
- Reduce cognitive complexity in production code
- Fix duplicated string literals

### Phase 3: Major Maintainability Issues
- Remove unused parameters
- Address FIXME comments
- Reduce switch statement branches
- Reduce function parameter counts

### Phase 4: Test Code Quality
- Fix test file issues (lower priority)

### Phase 5: Web UI Code Quality
- Fix JavaScript issues in web/plates/js/

### Phase 6: Script Quality
- Fix shell script issues in debian/

### Exclusions (Third-Party Code)
- dump978/ - UAT decoder library (45 bugs, 159 smells)
- web/maui/ - Mobile Angular UI (7 bugs, 689 smells)
- test/metar-to-text/ - METAR test utility (11 bugs, 338 smells)

## Impact

- **Affected code:** main/, web/plates/, uatparse/, common/, debian/
- **Estimated fixes:** ~200 actionable issues after excluding third-party code
- **Risk:** Low - mostly refactoring and cleanup
- **Benefits:** Improved code quality, reduced technical debt, better maintainability
