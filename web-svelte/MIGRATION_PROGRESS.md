# Settings + Map Migration Progress

## Completed
- Read all AngularJS source files (settings.js, settings.html, map.js, map.html)
- Read all reference Svelte pages (Weather, Radar, Developer, Traffic, Status)
- Read all stores (traffic.js, status.js, situation.js, weather.js, websocket.js)
- Installed OpenLayers: `npm install ol` (in package.json)
- Created settings store: `src/stores/settings.js`
  - settingsStore writable with all fields
  - loadSettings() and saveSettings(partial) functions
  - hasSerialOutput and serialBaud derived stores

## Next Steps
- Create `src/routes/Settings.svelte` - the largest page with sections:
  - AHRS (orientation, G limits)
  - Commands (update upload, reboot, shutdown)
  - Theme (dark mode)
  - OGN Tracker config (when hasTracker)
  - Configuration (ownship, watchlist, PPM, gain, baud, static IPs, altitude offset)
  - WiFi Settings (mode, country, SSID, passphrase, channel, IP, client networks)
  - Developer mode: Hardware toggles, Diagnostics, Raw config
  - Modals for confirmations
- Create `src/routes/Map.svelte` using OpenLayers:
  - OSM + OpenAIP base layers
  - Dynamic MBTiles layers from /tiles/tilesets
  - Traffic WebSocket overlay with aircraft symbols/trails
  - GPS WebSocket for ownship position marker
  - Layer switcher with localStorage persistence
- Update App.svelte: replace Placeholder with Settings and Map imports
- Build and commit

## Key API Endpoints
- GET /getSettings, POST /setSettings
- GET /getStatus (for GPS_detected_type → hasTracker)
- POST /orientAHRS, /reboot, /shutdown
- POST /updateUpload (multipart), /updatePong (multipart)
- GET /tiles/tilesets, GET /tiles/{file}/{z}/{x}/{-y}.{format}
- WS /traffic, WS /situation

## Country codes
- Full list in settings.js lines 7-258 (250+ countries)
- Used for WiFi country selection dropdown
