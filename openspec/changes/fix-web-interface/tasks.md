## 1. Diagnosis

- [x] 1.1 Get browser console errors from the device at http://10.0.1.53 (F12 → Console)
- [x] 1.2 Verify if AppCache is blocking fresh content (check Network tab for cached responses)
- [x] 1.3 Test each page route to identify which pages fail

## 2. AppCache Fixes

- [x] 2.1 Remove AppCache entirely - deleted `stratux.appcache` and removed `manifest` attribute from `index.html` (AppCache deprecated in all modern browsers)

## 3. CSP Fixes

- [x] 3.1 Remove `upgrade-insecure-requests` CSP meta tag that was blocking HTTP-only device access

## 4. UI Router Migration

- [x] 4.1 Add `$locationProvider.hashPrefix('')` to fix URL hash prefix mismatch between UI Router 1.0.x default (`#!/`) and existing menu links (`#/`)
- [x] 4.2 Verified all pages navigate correctly with Playwright tests

## 5. Playwright E2E Testing

- [x] 5.1 Set up Playwright testing framework with `@playwright/test`
- [x] 5.2 Create navigation tests for all pages (home, traffic, gps, settings, logs, radar, map)
- [x] 5.3 Create menu navigation test
- [x] 5.4 Create mock server for CI testing without device
- [x] 5.5 All 9 navigation tests passing against device

## 6. Validation

- [x] 6.1 Test all navigation routes work (home, traffic, gps, logs, settings, radar, map)
- [x] 6.2 Verify WebSocket connections work (status WebSocket tested)
- [ ] 6.3 Verify settings can be saved and retrieved
- [ ] 6.4 Test on mobile browser (Safari iOS, Chrome Android)

## Summary of Changes

Files modified:
- `web/index.html` - Removed AppCache manifest, removed CSP upgrade-insecure-requests, made addToHomescreen() call safer
- `web/js/main.js` - Added `$locationProvider.hashPrefix('')` for UI Router 1.0.x compatibility
- `web/stratux.appcache` - DELETED

Files added:
- `web/package.json` - Added Playwright and ws dependencies
- `web/playwright.config.js` - Playwright configuration
- `web/e2e/navigation.spec.js` - Navigation tests
- `web/e2e/mock-server.js` - Mock server for CI testing
- `web/e2e/debug-*.spec.js` - Debug tests (can be removed)
