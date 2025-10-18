# Third-Party Library Upgrade Plan

## Overview

This document outlines the plan to upgrade outdated third-party JavaScript libraries in the Stratux web interface to address 10 remaining CodeQL security alerts.

**Current Status:** 10/22 remaining CodeQL alerts are in 3rd party libraries
**Impact:** Medium to High - web interface functionality
**Estimated Effort:** 4-8 hours (testing is the main concern)

---

## Libraries to Update

### 1. AngularJS (High Priority)

**Current:** v1.4.6 (October 2015 - ~10 years old)
**Target:** v1.8.3 (final AngularJS 1.x release, April 2020)
**Location:** `web/maui/js/angular.js` (1.1MB), `angular.min.js` (144KB)

**Security Issues (4 alerts):**
- js/incomplete-sanitization (2 alerts)
- js/unsafe-html-expansion (1 alert)
- js/prototype-pollution-utility (1 alert)

**Breaking Changes:** v1.4.6 → v1.8.3
- v1.5: Component-based architecture introduced (shouldn't affect us)
- v1.6: `$location` hash-prefix defaults to `!` (may affect routing)
- v1.7: jQuery 3.x support, deprecations
- v1.8: Final security fixes

**Migration Steps:**
1. Download AngularJS 1.8.3 from https://code.angularjs.org/1.8.3/
2. Replace `angular.js` and `angular.min.js`
3. Test all pages: settings, status, GPS, traffic, weather, radar, towers, AHRS
4. Fix any routing issues (check for `$location.hash()` usage)
5. Verify form validation still works
6. Check WebSocket updates on status page

**Rollback Plan:** Keep v1.4.6 files as `.bak` for quick restore

---

### 2. Angular UI Router (High Priority)

**Current:** v0.2.15 (June 2015)
**Target:** @uirouter/angularjs v1.0.30 (latest for AngularJS 1.x)
**Location:** `web/maui/js/angular-ui-router.js` (160KB), `angular-ui-router.min.js` (30KB)

**Security Issues (2 alerts):**
- js/redos (Regular Expression Denial of Service) - 2 alerts

**Breaking Changes:** v0.2.x → v1.0.x
- API is mostly compatible, but namespace changed from `ui.router` to `@uirouter/angularjs`
- State transition hooks use new syntax
- May need to update module imports in HTML

**Migration Steps:**
1. Download from https://unpkg.com/@uirouter/angularjs@1.0.30/release/
2. Replace `angular-ui-router.js` and `.min.js`
3. Update HTML script tags if needed (check namespace)
4. Test navigation between all pages
5. Verify state transitions work (home → settings → GPS, etc.)
6. Check browser history/back button

**Testing Focus:**
- Page navigation
- Deep linking (direct URL access)
- Browser back/forward buttons

---

### 3. OpenLayers (Medium Priority)

**Current:** Unknown version (appears to be v4.x or v5.x based on file size)
**Target:** v9.2.4 (latest stable, March 2024)
**Location:** `web/js/ol.js` (1.3MB), `web/js/olms.js` (617KB)

**Security Issues (4 alerts):**
- js/incomplete-sanitization - 4 alerts in ol.js and olms.js

**Breaking Changes:** v5.x → v9.x
- Map rendering API changes
- Layer creation syntax updates
- Source configuration may differ
- olms (OpenLayers + Mapbox Style) may need separate update

**Migration Steps:**
1. Determine current version: check `web/js/ol.js` header
2. Download OpenLayers 9.2.4 from https://openlayers.org/download/
3. Download olms (ol-mapbox-style) compatible version
4. Replace `ol.js` and `olms.js`
5. Test map display on map.html and radar.html
6. Verify layer switching works
7. Test traffic overlay on map
8. Check zoom/pan controls

**Testing Focus:**
- Map rendering
- Traffic target display
- Layer switching (map tiles, overlays)
- Touch gestures on mobile

**Risk:** HIGH - Map library updates can significantly break rendering

---

## Testing Strategy

### Pre-Upgrade Checklist
- [ ] Create full backup of `web/` directory
- [ ] Document current behavior with screenshots
- [ ] Test on multiple browsers (Chrome, Firefox, Safari)
- [ ] Test on mobile devices (iOS, Android)
- [ ] Note any existing bugs/quirks to avoid false regressions

### Test Plan (Per Library)

**Phase 1: Smoke Test**
- [ ] Web UI loads without JavaScript errors (check browser console)
- [ ] All pages accessible via navigation
- [ ] No white screen / blank page errors

**Phase 2: Functional Test**
- [ ] Settings page: All toggles work, settings save/load
- [ ] Status page: WebSocket updates work, real-time data displays
- [ ] GPS page: Satellite display updates
- [ ] Traffic page: Traffic list populates
- [ ] Weather page: Weather data displays
- [ ] Radar page: Map renders, traffic overlays work
- [ ] Towers page: Tower list displays
- [ ] AHRS page: Attitude indicator works (if applicable)

**Phase 3: Regression Test**
- [ ] Form validation works (WiFi SSID, IP addresses, etc.)
- [ ] File uploads work (firmware updates)
- [ ] Modal dialogs open/close correctly
- [ ] Dark mode toggle works
- [ ] Responsive layout on mobile

### Post-Upgrade Verification
- [ ] No new JavaScript console errors
- [ ] CodeQL scan shows alerts resolved
- [ ] Performance is same or better (page load times)
- [ ] Memory usage hasn't increased significantly

---

## Implementation Order

**Recommended sequence:**

1. **Angular UI Router** (Lowest risk, easiest to test)
   - Clear migration path
   - Mostly compatible API
   - Fast to verify (just test navigation)

2. **AngularJS** (Medium risk, extensive testing needed)
   - Well-documented upgrade path
   - Many minor version increments to review
   - Affects all pages but API is stable

3. **OpenLayers** (Highest risk, complex testing)
   - Major version jump
   - Rendering behavior changes
   - Requires map/geo testing expertise

**Alternative sequence (if issues found):**
- Do AngularJS first since UI Router depends on it
- Skip OpenLayers if map testing resources unavailable

---

## Rollback Procedures

**For each library:**

1. **Immediate rollback** (if web UI completely broken):
   ```bash
   cd web/maui/js  # or web/js for OpenLayers
   mv angular.js angular.js.new
   mv angular.js.bak angular.js
   # Repeat for .min.js files
   make -C web  # Rebuild web UI
   ```

2. **Partial rollback** (if specific features broken):
   - Keep detailed git commits (one library per commit)
   - Use `git revert <commit>` for clean rollback
   - Document which tests failed for future attempts

3. **Testing in isolation:**
   - Test each library update in separate git branch
   - Merge only after full test pass
   - Use `git worktree` for parallel testing if needed

---

## Success Criteria

**Must have (required for merge):**
- ✅ All pages load without JavaScript errors
- ✅ Core functionality works: settings save/load, real-time updates
- ✅ No regressions in existing features
- ✅ CodeQL alerts for updated libraries are resolved

**Nice to have (can defer):**
- ⭐ Improved performance
- ⭐ Better error messages
- ⭐ Fixed minor UI quirks
- ⭐ Mobile responsiveness improvements

---

## Timeline Estimate

| Task | Duration | Notes |
|------|----------|-------|
| Angular UI Router upgrade | 1-2 hours | Includes testing |
| AngularJS upgrade | 2-3 hours | Extensive testing needed |
| OpenLayers upgrade | 3-4 hours | Complex, may need multiple attempts |
| Documentation | 1 hour | Update CLAUDE.md, commit messages |
| **Total** | **7-10 hours** | Assumes no major issues found |

---

## Risk Assessment

| Library | Risk Level | Impact if Broken | Mitigation |
|---------|-----------|------------------|------------|
| Angular UI Router | LOW | Navigation fails | Easy rollback, clear API docs |
| AngularJS | MEDIUM | Entire UI broken | Extensive testing, incremental update |
| OpenLayers | HIGH | Maps don't render | Test in branch first, have expert review |

---

## Open Questions

1. **Do we have a test Stratux device for real-world testing?**
   - Web UI testing is best done with actual hardware feeding data
   - Simulator/replay mode can help but may miss edge cases

2. **What browsers must we support?**
   - Modern browsers: Chrome/Edge, Firefox, Safari?
   - Mobile browsers: iOS Safari, Android Chrome?
   - Older browsers: IE11? (probably not needed for aviation use)

3. **Is there a test suite for the web UI?**
   - Currently: `web/package.json` has Jest configured
   - Tests exist in `web/tests/` but coverage may be limited
   - Should we add integration tests before upgrading?

4. **Can we update in production incrementally?**
   - Option 1: Update all at once (risky but faster)
   - Option 2: Canary deployment (test on subset of users)
   - Option 3: Feature flag (toggle between old/new libs)

---

## References

- AngularJS 1.8.3: https://code.angularjs.org/1.8.3/
- AngularJS Migration Guide: https://docs.angularjs.org/guide/migration
- UI Router 1.x Migration: https://ui-router.github.io/guide/ng1/migrate-to-1_0
- OpenLayers 9.x: https://openlayers.org/
- CodeQL JavaScript Alerts: https://codeql.github.com/codeql-query-help/javascript/

---

## Notes

- All library files are currently vendored (committed to git), not fetched via CDN
- This is good for offline deployment but means manual updates
- Consider adding build step to fetch from CDN in future
- Current approach: Keep vendored for Raspberry Pi deployment reliability
