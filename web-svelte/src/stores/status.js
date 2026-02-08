import { writable, derived } from 'svelte/store';
import { createWebSocketConnection, api } from '../services/websocket.js';

/**
 * Status store - receives real-time updates from /status WebSocket
 */
export const status = writable({
  _connected: false,
  _error: null,
  Version: '',
  Build: '',
  Devices: 0,
  Connected_Users: 0,
  Uptime: 0,
  CPUTemp: 0,
  CPUTempMin: 0,
  CPUTempMax: 0,
  DiskBytesFree: 0,
  // UAT (978 MHz)
  UAT_messages_last_minute: 0,
  UAT_messages_max: 0,
  UAT_messages_total: 0,
  // 1090ES
  ES_messages_last_minute: 0,
  ES_messages_max: 0,
  ES_messages_total: 0,
  // OGN (868 MHz)
  OGN_messages_last_minute: 0,
  OGN_messages_max: 0,
  OGN_messages_total: 0,
  OGN_connected: false,
  OGN_noise_db: 0,
  OGN_gain_db: 0,
  // AIS (marine)
  AIS_messages_last_minute: 0,
  AIS_messages_max: 0,
  AIS_messages_total: 0,
  AIS_connected: false,
  // GPS
  GPS_satellites_locked: 0,
  GPS_satellites_tracked: 0,
  GPS_satellites_seen: 0,
  GPS_solution: 'Disconnected',
  GPS_position_accuracy: 0,
  GPS_detected_type: 0,
  GPS_NetworkRemoteIp: '',
  // Weather (UAT)
  UAT_METAR_total: 0,
  UAT_TAF_total: 0,
  UAT_NEXRAD_total: 0,
  UAT_SIGMET_total: 0,
  UAT_PIREP_total: 0,
  UAT_NOTAM_total: 0,
  UAT_OTHER_total: 0,
  // Ping/Pong
  Ping_connected: false,
  Pong_connected: false,
  Pong_Heartbeats: 0,
  // Errors
  Errors: []
});

/**
 * Settings store - fetched via HTTP, controls visibility
 */
export const settings = writable({
  loaded: false,
  DeveloperMode: false,
  UAT_Enabled: true,
  ES_Enabled: true,
  OGN_Enabled: false,
  AIS_Enabled: false,
  GPS_Enabled: true,
  Ping_Enabled: false,
  Pong_Enabled: false,
  RegionSelect: 0
});

// Derived stores for computed values
export const uptime = derived(status, $status => {
  const ms = $status.Uptime || 0;
  const totalSeconds = Math.floor(ms / 1000);
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  const pad = n => n.toString().padStart(2, '0');
  return `${days}/${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
});

export const cpuTemp = derived(status, $status => {
  const c = $status.CPUTemp || 0;
  const f = (c * 9 / 5) + 32;
  return {
    celsius: c.toFixed(1),
    fahrenheit: f.toFixed(1),
    min: ($status.CPUTempMin || 0).toFixed(1),
    max: ($status.CPUTempMax || 0).toFixed(1)
  };
});

export const diskSpace = derived(status, $status => {
  const mib = ($status.DiskBytesFree || 0) / 1048576;
  return mib.toFixed(1);
});

export const gpsHardware = derived(status, $status => {
  const code = ($status.GPS_detected_type || 0) & 0x0f;
  // SYNC WARNING: Must match GPS_TYPE_* constants in main/gen_gdl90.go
  const types = {
    1: 'Generic GPS device',
    2: 'Prolific USB-serial bridge',
    3: 'OGN Tracker',
    4: 'generic u-blox device',
    5: 'u-blox 10 GNSS receiver',
    7: 'u-blox 6 or 7 GNSS receiver',
    8: 'u-blox 8 GNSS receiver',
    9: 'u-blox 9 GNSS receiver',
    10: 'USB/Serial IN',
    11: 'SoftRF Dongle',
    12: 'Network',
    13: 'SoftRF',
    15: 'GxAirCom'
  };
  return types[code] || 'Not installed';
});

export const gpsProtocol = derived(status, $status => {
  const protocol = ($status.GPS_detected_type || 0) >> 4;
  return protocol === 1 ? 'NMEA protocol' : 'Not communicating';
});

export const ognNoiseColor = derived(status, $status => {
  const db = $status.OGN_noise_db || 0;
  if (db <= 6) return 'green';
  if (db < 12) return '#fc0';
  if (db < 18) return 'orange';
  return 'red';
});

export const ognRangeLossFactor = derived(status, $status => {
  return Math.pow(10, 0.05 * ($status.OGN_noise_db || 0)).toFixed(2);
});

// WebSocket connection controller
let wsConnection = null;

export function connectStatus() {
  if (!wsConnection) {
    wsConnection = createWebSocketConnection('/status', status);
  }
  wsConnection.connect();
}

export function disconnectStatus() {
  if (wsConnection) {
    wsConnection.disconnect();
  }
}

// Fetch settings via HTTP
export async function fetchSettings() {
  try {
    const data = await api.get('/getSettings');
    settings.update(s => ({ ...s, ...data, loaded: true }));
  } catch (e) {
    console.error('Failed to fetch settings:', e);
  }
}

// Traffic source colors (matches craftService)
export const trafficSourceColors = {
  ES: '#4169e1',    // 1090ES - Royal Blue
  UAT: '#ff8c00',   // UAT - Dark Orange
  OGN: '#32cd32',   // OGN - Lime Green
  AIS: '#9370db'    // AIS - Medium Purple
};
