# Security Fixes Roadmap

**Generated:** 2025-10-17
**Source:** GitHub Dependabot + CodeQL Analysis
**Total Issues:** 58 (42 code scanning + 16 dependency alerts)

## Phase 1: Critical Security Fixes (PRIORITY 1)

### 1.1 Path Injection & XSS in Management Interface
**Files:** `main/managementinterface.go`
**Alerts:** #40, #41, #42 (path injection), #27, #28 (XSS)
**Severity:** CRITICAL
**Current Test Coverage:** 0% (NO TESTS EXIST)

**Attack Vector:**
- Path injection: `GET /logs/../../etc/passwd` could read arbitrary files
- XSS: Error messages reflect unsanitized user input

**Fix Strategy:**
1. **FIRST:** Write comprehensive tests for `viewLogs()` function
   - Test normal case (valid log file access)
   - Test path traversal attempts (`../`, absolute paths)
   - Test XSS injection in error messages
   - Test directory listing functionality

2. **THEN:** Implement fixes:
   - Add path validation using `filepath.Clean()` and prefix checking
   - HTML-escape all user-controlled output
   - Use `http.Error()` instead of direct `w.Write()` for errors

3. **VERIFY:** Run tests, verify all alerts are resolved

**Estimated Effort:** 4-6 hours (2h tests + 2h fixes + 2h verification)

---

### 1.2 Update High-Severity Go Dependencies
**Severity:** HIGH
**Current Test Coverage:** E2E tests exist (24.5% overall coverage)

**Dependencies to Update:**
1. `golang.org/x/net` - HTTP/2 rapid reset (CVE-2023-39325, CVE-2023-44487)
2. `golang.org/x/image` - Panic on invalid images
3. `github.com/prometheus/client_golang` - DoS vulnerability
4. `google.golang.org/protobuf` - Infinite loop

**Fix Strategy:**
1. **FIRST:** Run full test suite to establish baseline
   ```bash
   go test -coverprofile=coverage_baseline.out -coverpkg=./... ./...
   ```

2. **THEN:** Update dependencies one at a time:
   ```bash
   go get -u golang.org/x/net@latest
   go get -u golang.org/x/image@latest
   go get -u github.com/prometheus/client_golang@latest
   go get -u google.golang.org/protobuf@latest
   ```

3. **AFTER EACH UPDATE:**
   - Run full test suite
   - Run E2E tests
   - Check for breaking changes
   - Commit if successful, rollback if tests fail

4. **VERIFY:** Check Dependabot alerts are resolved

**Estimated Effort:** 2-4 hours (depending on breaking changes)

---

## Phase 2: Code Quality & Integer Safety (PRIORITY 2)

### 2.1 Fix Integer Conversion Issues
**Files:** `main/gen_gdl90.go`, `main/gps.go`, `test/replay.go`
**Alerts:** #29-#39 (13 warnings)
**Severity:** MEDIUM
**Current Test Coverage:**
- `gen_gdl90.go`: ~50% coverage (makeStratuxStatus, makeHeartbeat tested)
- `gps.go`: Unknown (need to check)

**Fix Strategy:**
1. **FIRST:** Identify affected functions and add tests
   - `gen_gdl90.go:537-540` - appears to be in `makeStratuxStatus()` (already 95.3% covered)
   - `gps.go:1177,1483,1661,1671,1677,1690` - need to identify functions
   - `test/replay.go:36,108` - test code, lower priority

2. **THEN:** Add bounds checking before conversions:
   ```go
   // Example fix:
   val, err := strconv.Atoi(s)
   if err != nil || val < 0 || val > 255 {
       return 0, fmt.Errorf("value out of range for uint8: %d", val)
   }
   return uint8(val), nil
   ```

3. **VERIFY:** Run tests, verify CodeQL alerts are resolved

**Estimated Effort:** 6-8 hours (3h investigation + 3h tests + 2h fixes)

---

### 2.2 Fix JavaScript Code Quality Issues
**Files:** `web/plates/js/settings.js`
**Alerts:** #15-#18 (useless regex escapes)
**Severity:** LOW (code quality)

**Fix Strategy:**
1. Remove unnecessary escapes from regex patterns
2. Test web interface manually (no automated web tests exist)

**Estimated Effort:** 1 hour

---

## Phase 3: Dependency Updates - Medium Risk (PRIORITY 3)

### 3.1 Update Remaining Go Dependencies
**Severity:** MEDIUM (12 alerts)

**Dependencies:**
- `golang.org/x/net` - Additional vulnerabilities (beyond Phase 1.2)
- `golang.org/x/image` - TIFF decoder issues
- `github.com/golang/glog` - Insecure temp file

**Fix Strategy:** Same as Phase 1.2 (test → update → verify)

**Estimated Effort:** 2-3 hours

---

### 3.2 Update/Replace JavaScript Dependencies
**Files:** `web/maui/js/angular*.js`, `web/js/ol*.js`
**Alerts:** #13-#14 (ReDoS), #19-#26 (various)
**Severity:** MEDIUM-HIGH (ReDoS is concerning)

**Challenge:** These are vendored third-party libraries (Angular, OpenLayers)

**Fix Strategy (Options):**
1. **Option A:** Update to latest versions (PREFERRED)
   - Check if newer versions are available
   - Test web interface thoroughly after update

2. **Option B:** Replace with modern alternatives
   - Angular → Modern framework (React, Vue, Svelte)
   - This is a MAJOR refactor, out of scope for security fixes

3. **Option C:** Accept risk and document
   - If updates aren't available and app is only accessible on local WiFi
   - Add to security documentation

**Estimated Effort:** 4-8 hours (depending on option chosen)

---

## Phase 4: Test Coverage Expansion (ONGOING)

### 4.1 Priority Test Gaps
Based on security findings, these areas need tests:

1. **managementinterface.go** - 0% coverage
   - All HTTP handlers
   - Input validation
   - Authentication/authorization (if any)

2. **gps.go** - Coverage unknown
   - NMEA parsing functions (where integer conversions occur)

3. **gen_gdl90.go** - Expand from 50% to 80%+
   - Ownship report generation (0% coverage)
   - Geometric altitude report (0% coverage)

**Estimated Effort:** 20-30 hours (this is substantial work)

---

## Execution Plan

### Week 1: Critical Fixes
- [ ] Day 1-2: Phase 1.1 - Path injection & XSS (tests + fixes)
- [ ] Day 3: Phase 1.2 - High-severity dependency updates
- [ ] Day 4: Verify all Phase 1 fixes, run full test suite
- [ ] Day 5: Create PR, code review

### Week 2: Code Quality
- [ ] Day 1-3: Phase 2.1 - Integer conversion issues
- [ ] Day 4: Phase 2.2 - JavaScript code cleanup
- [ ] Day 5: Phase 3.1 - Remaining Go dependency updates

### Week 3: JavaScript & Long-term
- [ ] Day 1-2: Phase 3.2 - JavaScript dependencies
- [ ] Day 3-5: Begin Phase 4 - Test coverage expansion
- [ ] Ongoing: Continue test coverage work

---

## Success Metrics

- [ ] All CRITICAL security alerts resolved (5 alerts)
- [ ] All HIGH severity dependency alerts resolved (4 alerts)
- [ ] Test coverage for managementinterface.go > 80%
- [ ] Test coverage for affected gps.go functions > 80%
- [ ] Overall code coverage > 30% (up from 24.5%)
- [ ] All GitHub Actions passing
- [ ] No new security alerts introduced

---

## Risk Mitigation

### Before Making Changes:
1. ✅ Full test suite baseline established (current: 24.5% coverage)
2. Create feature branch for each phase
3. Test on real hardware (Raspberry Pi) before merging

### During Changes:
1. One fix at a time (no batching critical fixes)
2. Run tests after EVERY change
3. Commit working state frequently
4. Keep PRs small and focused

### Rollback Plan:
- Each phase is a separate branch
- Can cherry-pick successful fixes
- Can rollback failed dependency updates

---

## Notes

- **GitHub Actions:** Currently working, some Dependabot PRs failing (expected)
- **Current Coverage:** 24.5% overall, recent progress from 22.7%
- **Coverage Files Available:** Multiple coverage*.out files in main/ directory
- **No Breaking Changes Policy:** Must maintain backward compatibility for EFB apps

---

## Questions to Address Before Starting

1. **Access Control:** Does managementinterface.go have any authentication? (doesn't appear so)
   - If not, should we add authentication as part of security fixes?
   - Web interface is on private WiFi (192.168.10.1) - is this sufficient?

2. **JavaScript Libraries:**
   - What's the policy on updating vendored third-party libraries?
   - Are there automated web tests we should know about?

3. **Test Hardware:**
   - Do we have access to Raspberry Pi for testing?
   - Can we test SDR functionality without hardware?

4. **Breaking Changes:**
   - What's the policy for dependency updates that require code changes?
   - Can we accept minor API changes in internal code?
