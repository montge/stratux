# Change: Fix Web Interface Rendering After Library Upgrades

## Why

The web interface is not rendering correctly after recent upgrades to core JavaScript libraries:
- AngularJS: 1.4.6 → 1.8.3
- Angular UI Router: 0.2.15 → 1.0.30
- OpenLayers: 6.x → 10.7.0

These upgrades were made to address CodeQL security alerts (ReDoS vulnerabilities) but introduced breaking changes that prevent the web interface from functioning.

## What Changes

1. **AppCache manifest is outdated** - The `stratux.appcache` file is missing many new files (radar.js, map.js, developer.js, ol.js, olms.js, ol-layerswitcher.js, CSS files) which can cause browsers to serve stale/incomplete cached content
2. **AppCache deprecated** - Modern browsers have deprecated/removed AppCache support entirely; should migrate to Service Worker or remove caching
3. **UI Router state lifecycle hooks** - Controllers use `$state.get('stateName').onEnter/onExit` pattern which may behave differently in UI Router 1.0.x
4. **OpenLayers API changes** - OpenLayers 10.x may have breaking API changes from 6.x that affect map.js and radar.js

## Impact

- Affected specs: `web-interface` (if exists) or new capability
- Affected code:
  - `web/index.html` - AppCache manifest reference
  - `web/stratux.appcache` - Outdated file list
  - `web/plates/js/status.js` - Uses onEnter/onExit hooks
  - `web/plates/js/towers.js` - Uses onEnter/onExit hooks
  - `web/plates/js/weather.js` - Uses onEnter/onExit hooks
  - `web/plates/js/gps.js` - Uses onEnter/onExit hooks
  - `web/plates/js/map.js` - Uses onExit hook, OpenLayers APIs
  - `web/plates/js/radar.js` - Uses onEnter/onExit hooks, OpenLayers APIs
  - `web/plates/js/traffic.js` - Uses onEnter/onExit hooks
