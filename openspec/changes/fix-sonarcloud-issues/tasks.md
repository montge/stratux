# Tasks: Fix SonarCloud Issues

## 1. Configuration & Exclusions
- [x] 1.1 Configure SonarCloud to exclude third-party directories (dump978/, web/maui/, test/metar-to-text/)
- [ ] 1.2 Verify exclusion reduces issue count appropriately

## 2. Reliability Issues (Bugs) - HIGH PRIORITY
- [x] 2.1 Fix datalog.go:439 - Add `defer tx.Rollback()` after `db.Begin()` error check
- [x] 2.2 Fix web/plates/js/map.js - Remove use of output from clipPosHistory (void function)
- [x] 2.3 Fix web/plates/status.html - Remove deprecated `<tt>` element
- [x] 2.4 Fix web/index.html - Add lang attribute to html element
- [x] 2.5 Fix web/css/main.css - Remove duplicate border-bottom

## 3. Critical Code Smells - Production Code
- [x] 3.1 Fix gps.go - Reduce cognitive complexity by extracting processNMEAMessages()
- [x] 3.2 Fix managementinterface.go - Define constant for "/mapdata/styles/" (done in previous session)
- [ ] 3.3 Fix managementinterface.go:401 - Reduce switch branches from 48 to 30 (deferred - significant refactor)

## 4. Major Code Smells - Production Code (Unused Parameters)
- [x] 4.1 Fix managementinterface.go - handleRegionGet unused 'r'
- [x] 4.2 Fix managementinterface.go - handleDeleteLogFile unused 'w', 'r'
- [x] 4.3 Fix managementinterface.go - handleDeleteAHRSLogFiles unused 'r'
- [x] 4.4 Fix managementinterface.go - handleDevelModeToggle unused 'w', 'r'
- [x] 4.5 Fix managementinterface.go - handleRestartRequest unused 'w', 'r'
- [x] 4.6 Fix managementinterface.go - handleDownloadAHRSLogsRequest unused 'r'
- [x] 4.7 Fix managementinterface.go - handleTilesets unused 'r'
- [x] 4.8 Fix managementinterface.go - handleSatellitesRequest unused 'r'
- [x] 4.9 Fix managementinterface.go - handleSettingsGetRequest unused 'r'
- [x] 4.10 Fix managementinterface.go - handleStatusRequest unused 'r'
- [x] 4.11 Fix managementinterface.go - handleSituationRequest unused 'r'
- [x] 4.12 Fix managementinterface.go - handleTowersRequest unused 'r'
- [x] 4.13 Fix gen_gdl90.go:1042 - Remove unused parameter 'uatMsg'
- [x] 4.14 Fix clientconnection.go:110 - Remove unused parameter 'err'
- [x] 4.15 Fix tracker.go - OgnTracker.initNewConnection unused 'serialPort'
- [x] 4.16 Fix tracker.go - OgnTracker.onNmea unused 'serialPort'
- [x] 4.17 Fix tracker.go - GxAirCom.initNewConnection unused 'serialPort'
- [x] 4.18 Fix tracker.go - GxAirCom.onNmea unused 'serialPort'
- [x] 4.19 Fix tracker.go - SoftRF.initNewConnection unused 'serialPort'

## 5. Major Code Smells - FIXME Comments
- [x] 5.1 Address network.go:137 FIXME comment - converted to descriptive note
- [x] 5.2 Address sensors.go:217 FIXME comment - converted to descriptive note
- [x] 5.3 Address gps.go:889 FIXME comment - clarified as debug output
- [x] 5.4 Address gps.go:1002 FIXME comment - clarified as debug output
- [x] 5.5 Address gps.go:1725 FIXME comment - converted to descriptive comment
- [x] 5.6 Address managementinterface.go:1031 FIXME comment - converted to note about future enhancement
- [x] 5.7 Address sdr.go:764 FIXME comment - converted to descriptive comment

## 6. Major Code Smells - Function Parameters
- [x] 6.1 Fix xplane.go:38 - Reduce function parameters from 9 to 1 (using XPlaneTrafficData struct)

## 7. Test Code Quality (Lower Priority)
- [ ] 7.1 Fix unused parameters in test mock functions (gps_test.go, managementinterface_test.go, etc.)
- [ ] 7.2 Define constants for duplicated test strings
- [ ] 7.3 Reduce cognitive complexity in test functions (optional - tests can be more complex)

## 8. Web UI Code Quality
- [x] 8.1 Fix web/plates/gps.html - Add keyboard event handler to span with onclick (added ng-keydown, tabindex, role)
- [ ] 8.2 Review and fix web/plates/js/*.js issues (deferred - low priority)

## 9. Shell Script Quality
- [x] 9.1 Fix debian/*.sh shell script issues (fixed undefined variable, backtick substitution, quoting)
- [ ] 9.2 Review and fix scripts/*.sh issues (deferred - low priority)

## 10. Validation
- [ ] 10.1 Run SonarCloud analysis to verify fixes (requires CI push)
- [x] 10.2 Ensure no regressions in tests - all tests pass
- [x] 10.3 Update documentation if needed - tasks.md updated

---

# Phase 2: Additional Code Quality Improvements

## 11. Magic Numbers - Define Constants
- [x] 11.1 traffic.go:254 - Define constant for `99999.0` (altitudeInvalidFeet)
- [x] 11.2 traffic.go:351 - Define constant for `2000` feet threshold (bearinglessAltDiffMaxFeet)
- [x] 11.3 traffic.go:423 - Define constant for `15000` distance threshold (bearinglessDistMaxMeters)
- [ ] 11.4 traffic.go:468 - Define constants for altitude buckets `[2500.0, 2800.0, 3000.0]` (deferred - used as tunable array)
- [x] 11.5 traffic.go:491 - Define constants for `50000` and `1500` distance thresholds (estimateDistMaxMeters, estimateDistMinMeters)
- [x] 11.6 gps.go:1145 - Define constant for `3` knots threshold (minSpeedForCourseUpdateKnots)
- [x] 11.7 gps.go:1262 - Define constant for `299` GPS perf array size (gpsPerfStatsMaxEntries)
- [x] 11.8 traffic.go:392-394 - Define constants for radar margin (radarDisplayMargin, metersPerNauticalMile)
- [x] 11.9 traffic.go:293 - Define constant for `500` ft altitude diff (ownshipAltDiffMaxFeet)

## 12. TODO/FIXME Comments in Production Code
- [x] 12.1 clientconnection.go:261 - Defined bleMTUPayloadSize constant (20 bytes)
- [x] 12.2 cot-in.go:111 - Added TRAFFIC_SOURCE_COT constant (16), updated cot-in.go to use it
- [x] 12.3 datalog.go:362 - Converted FIXME to descriptive Note about recursive insertion pattern
- [ ] 12.4 flarm-nmea.go - Address 5 TODOs (lines 75, 86, 236, 458, 460) - deferred
- [ ] 12.5 gen_gdl90.go - Address 13 TODOs/FIXMEs - deferred (many are feature requests)
- [ ] 12.6 gps.go - Address 8 TODOs (lines 133, 143, 294, 672-673, 735, 1145, 1151) - deferred (feature enhancements)

## 13. High Cognitive Complexity Functions (Major Refactors)
- [ ] 13.1 traffic.go:697 parseDownlinkReport() - 352 lines, ~44 conditionals (extract bit manipulation helpers)
- [ ] 13.2 traffic.go:1084 parseDump1090Message() - 321 lines, ~35 conditionals (extract type-code handlers)
- [ ] 13.3 gps.go:1084 processNMEALineLow() - 800+ lines, ~50 conditionals (extract NMEA sentence handlers)
- [ ] 13.4 gen_gdl90.go:1095 parseInput() - 306 lines (extract message type handlers)

## 14. Code Synchronization Issues
- [ ] 14.1 gen_gdl90.go:93-115 / web/plates/js/status.js - GPS type enumeration sync (add validation or generation)

## 15. Ignored Error Returns
- [ ] 15.1 network.go:372,375 - Document or handle ignored returns
- [ ] 15.2 traffic.go:252 - Document or handle `trafficDist, _, _, _`
- [ ] 15.3 tracker.go:351 - Handle ignored error from OGNAddrType
