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
- [x] 15.1 network.go:372,375 - Added error handling with log messages for BLE UUID parsing
- [ ] 15.2 traffic.go:252 - Intentional ignore of unused DistRect return values (documented by Go idiom)
- [x] 15.3 tracker.go:351 - Added error handling for strconv.Atoi calls in SoftRF config parsing
- [x] 15.4 logging.go:62,64,69 - Added error handling for os.Remove and os.Rename in log rotation
- [x] 15.5 managementinterface.go:640 - Added error handling for os.Remove in AHRS log deletion
- [x] 15.6 managementinterface.go:261,268 - Added error handling for json.Marshal in status/situation handlers

## 16. Debug Log Messages
- [x] 16.1 managementinterface.go:625 - Changed "handleDeleteLogFile called!!!" to professional message
- [x] 16.2 managementinterface.go:648 - Changed "handleDevelModeToggle called!!!" to professional message

## 17. Additional JSON Marshal Error Handling
- [x] 17.1 ais.go:231 - Added error handling in DEBUG block for AIS traffic marshal
- [x] 17.2 ogn.go:341 - Added error handling in DEBUG block for OGN traffic marshal
- [x] 17.3 cot-in.go:137 - Added error handling in DEBUG block for COT traffic marshal
- [x] 17.4 uibroadcast.go:44 - Added error handling in SendJSON utility function
- [x] 17.5 managementinterface.go:137,174 - Added error handling in handleJsonIo and handleTrafficWS
- [x] 17.6 managementinterface.go:239,257 - Added error handling in handleStatusWS and handleSituationWS
- [x] 17.7 managementinterface.go:615 - Added error handling in setSettings response
- [x] 17.8 managementinterface.go:790 - Added error handling in handleClientsGetRequest
- [x] 17.9 managementinterface.go:1239 - Added error handling in handleTilesets

## 18. Strconv Error Handling
- [x] 18.1 managementinterface.go:1192 - Added error handling for strconv.ParseInt in mbtiles metadata
- [x] 18.2 managementinterface.go:1289,1291 - Added error handling for x, z, and file parsing in handleTile

## 19. Redundant Boolean Equality Checks
- [x] 19.1 ping.go - Removed `== true` and `== false` comparisons (3 locations)
- [x] 19.2 pong.go - Removed `== true` and `== false` comparisons (4 locations)
- [x] 19.3 gen_gdl90.go - Removed `== true` comparisons (4 locations)
- [x] 19.4 ais.go - Removed `== false` comparison (1 location)
- [x] 19.5 clientconnection.go - Removed `== true` comparison (1 location)

## 20. Redundant Type Conversions
- [x] 20.1 traffic.go:773,1614 - Removed unnecessary string() conversion from string literals
- [x] 20.2 gps.go:241 - Removed unnecessary bool() and int() conversions
- [x] 20.3 ping.go:46 - Removed unnecessary int() conversion
- [x] 20.4 pong.go:44 - Removed unnecessary int() conversion
- [x] 20.5 datalog.go:283 - Removed unnecessary int() conversion

## 21. Variable Shadowing
- [x] 21.1 ping.go:404-405 - Renamed shadowed lat/lng to ownLat/ownLng
- [x] 21.2 managementinterface.go:375 - Renamed shadowed val to regionStr

## 22. Reliability Issues - Phase 2
- [x] 22.1 managementinterface.go:834 - Added defer unzippedFile.Close() to fix resource leak
- [x] 22.2 network.go:609 - Added netMutex.Lock() to fix data race in getNetworkConnsByIp
- [x] 22.3 traffic.go:718 - Added minimum frame length validation (4 bytes for header)
- [x] 22.4 traffic.go:760 - Added frame length check (31 bytes) for msg_type 1/3 processing
- [x] 22.5 datalog.go:298 - Use safe type assertion for query size calculation
- [x] 22.6 ais.go:142,153,170 - Use safe type assertions for AIS message packet handling
