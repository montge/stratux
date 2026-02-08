import { writable, derived } from 'svelte/store';
import { api } from '../services/websocket.js';

/**
 * Full settings store - loaded from GET /getSettings, saved via POST /setSettings
 */
export const settingsStore = writable({
  _loaded: false,
  _error: null,
  // Toggles
  UAT_Enabled: false,
  ES_Enabled: false,
  OGN_Enabled: false,
  AIS_Enabled: false,
  APRS_Enabled: false,
  Ping_Enabled: false,
  Pong_Enabled: false,
  GPS_Enabled: false,
  OGNI2CTXEnabled: false,
  IMU_Sensor_Enabled: false,
  BMP_Sensor_Enabled: false,
  DisplayTrafficSource: false,
  DEBUG: false,
  ReplayLog: false,
  TraceLog: false,
  AHRSLog: false,
  PersistentLogging: false,
  GDL90MSLAlt_Enabled: false,
  EstimateBearinglessDist: false,
  DarkMode: false,
  // Values
  PPM: 0,
  Dump1090Gain: 37.2,
  AltitudeOffset: 0,
  WatchList: '',
  OwnshipModeS: '',
  DeveloperMode: false,
  GLimits: '',
  StaticIps: [],
  Baud: 0,
  PWMDutyMin: 0,
  // WiFi
  WiFiCountry: '',
  WiFiSSID: '',
  WiFiPassphrase: '',
  WiFiSecurityEnabled: false,
  WiFiChannel: 1,
  WiFiIPAddress: '192.168.10.1',
  WiFiMode: 0,
  WiFiDirectPin: '',
  WiFiClientNetworks: [],
  WiFiInternetPassThroughEnabled: false,
  // OGN Tracker
  OGNAddrType: 0,
  OGNAddr: '',
  OGNAcftType: 1,
  OGNPilot: '',
  OGNReg: '',
  OGNTxPower: 0,
  // Serial
  SerialOutputs: null
});

/**
 * Load settings from the server
 */
export async function loadSettings() {
  try {
    const data = await api.get('/getSettings');
    settingsStore.update(s => ({ ...s, ...data, _loaded: true, _error: null }));
    return data;
  } catch (e) {
    settingsStore.update(s => ({ ...s, _error: 'Failed to load settings' }));
    console.error('Failed to load settings:', e);
    return null;
  }
}

/**
 * Save partial settings to the server, then reload
 */
export async function saveSettings(partial) {
  try {
    const response = await fetch('/setSettings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(partial)
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data = await response.json();
    settingsStore.update(s => ({ ...s, ...data, _loaded: true, _error: null }));
    return data;
  } catch (e) {
    settingsStore.update(s => ({ ...s, _error: 'Failed to save settings' }));
    console.error('Failed to save settings:', e);
    return null;
  }
}

/**
 * Derived: whether serial output is visible
 */
export const hasSerialOutput = derived(settingsStore, $s => {
  return $s.SerialOutputs !== null && $s.SerialOutputs !== undefined &&
    Object.keys($s.SerialOutputs).length > 0;
});

/**
 * Derived: serial baud rate from first serial output
 */
export const serialBaud = derived(settingsStore, $s => {
  if ($s.SerialOutputs) {
    for (const k in $s.SerialOutputs) {
      return $s.SerialOutputs[k].Baud || 0;
    }
  }
  return 0;
});
