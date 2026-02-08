import { writable, derived } from 'svelte/store';
import { createWebSocketConnection, api } from '../services/websocket.js';

/**
 * Situation store - receives real-time updates from /situation WebSocket
 * Contains GPS position, AHRS attitude, and satellite data
 */
export const situation = writable({
  _connected: false,
  _error: null,
  // GPS data
  GPSLatitude: 0,
  GPSLongitude: 0,
  GPSAltitudeMSL: 0,
  GPSHeightAboveEllipsoid: 0,
  GPSGroundSpeed: 0,
  GPSTrueCourse: 0,
  GPSVerticalSpeed: 0,
  GPSHorizontalAccuracy: 0,
  GPSVerticalAccuracy: 0,
  GPSFixQuality: 0,
  GPSSatellites: 0,
  GPSSatellitesTracked: 0,
  GPSSatellitesSeen: 0,
  GPSPositionSampleRate: 0,
  // AHRS data
  AHRSPitch: 0,
  AHRSRoll: 0,
  AHRSGyroHeading: 0,
  AHRSMagHeading: 0,
  AHRSSlipSkid: 0,
  AHRSTurnRate: 0,
  AHRSGLoad: 0,
  AHRSGLoadMin: 0,
  AHRSGLoadMax: 0,
  AHRSStatus: 0,
  // Barometric
  BaroPressureAltitude: 0,
  BaroVerticalSpeed: 0,
  // Satellite list
  GPSSatellitesData: []
});

// WebSocket connection
let wsConnection = null;

export function connectSituation() {
  if (!wsConnection) {
    wsConnection = createWebSocketConnection('/situation', situation);
  }
  wsConnection.connect();
}

export function disconnectSituation() {
  if (wsConnection) {
    wsConnection.disconnect();
    wsConnection = null;
  }
}

// Derived stores for computed values

/**
 * GPS solution text based on fix quality
 */
export const gpsSolutionText = derived(situation, $s => {
  const quality = $s.GPSFixQuality;
  if (quality === 0) return 'No Fix';
  if (quality === 1) return '3D GPS';
  if (quality === 2) return '3D GPS + SBAS';
  if (quality === 6) return 'Dead Reckoning';
  return 'Unknown';
});

/**
 * Format latitude/longitude for display
 */
export const gpsPosition = derived(situation, $s => {
  const formatCoord = (val, pos, neg) => {
    const abs = Math.abs(val);
    const deg = Math.floor(abs);
    const min = ((abs - deg) * 60).toFixed(4);
    const dir = val >= 0 ? pos : neg;
    return `${deg}° ${min}' ${dir}`;
  };

  return {
    lat: formatCoord($s.GPSLatitude, 'N', 'S'),
    lng: formatCoord($s.GPSLongitude, 'E', 'W'),
    latRaw: $s.GPSLatitude,
    lngRaw: $s.GPSLongitude
  };
});

/**
 * GPS altitude and speed formatted
 */
export const gpsAltSpeed = derived(situation, $s => ({
  altMSL: Math.round($s.GPSAltitudeMSL),
  altEllipsoid: Math.round($s.GPSHeightAboveEllipsoid),
  vertSpeed: Math.round($s.GPSVerticalSpeed),
  groundSpeed: Math.round($s.GPSGroundSpeed),
  track: Math.round($s.GPSTrueCourse),
  hAccuracy: $s.GPSHorizontalAccuracy.toFixed(1),
  vAccuracy: $s.GPSVerticalAccuracy.toFixed(1)
}));

/**
 * AHRS data formatted
 */
export const ahrsData = derived(situation, $s => ({
  pitch: $s.AHRSPitch.toFixed(1),
  roll: $s.AHRSRoll.toFixed(1),
  heading: Math.round($s.AHRSGyroHeading),
  headingMag: Math.round($s.AHRSMagHeading),
  slipSkid: $s.AHRSSlipSkid.toFixed(1),
  turnRate: $s.AHRSTurnRate.toFixed(1),
  gLoad: $s.AHRSGLoad.toFixed(2),
  gLoadMin: $s.AHRSGLoadMin.toFixed(2),
  gLoadMax: $s.AHRSGLoadMax.toFixed(2),
  pressureAlt: Math.round($s.BaroPressureAltitude)
}));

/**
 * Satellite summary
 */
export const gpsSatelliteSummary = derived(situation, $s => ({
  inSolution: $s.GPSSatellites,
  tracked: $s.GPSSatellitesTracked,
  seen: $s.GPSSatellitesSeen,
  sampleRate: $s.GPSPositionSampleRate.toFixed(1)
}));

/**
 * AHRS/IMU status
 */
export const ahrsStatus = derived(situation, $s => {
  const status = $s.AHRSStatus || 0;
  return {
    hasGPS: (status & 0x01) !== 0,
    hasIMU: (status & 0x02) !== 0,
    hasBaro: (status & 0x04) !== 0,
    isCalibrating: (status & 0x10) !== 0,
    code: status
  };
});

// API functions for AHRS control

export async function cageAHRS() {
  try {
    await api.post('/cageAHRS');
    return { success: true };
  } catch (e) {
    return { success: false, error: e.message };
  }
}

export async function calibrateAHRS() {
  try {
    await api.post('/calibrateAHRS');
    return { success: true };
  } catch (e) {
    return { success: false, error: e.message };
  }
}

export async function resetGMeter() {
  try {
    await api.post('/resetGMeter');
    return { success: true };
  } catch (e) {
    return { success: false, error: e.message };
  }
}

// Fetch satellites separately (not in WebSocket)
export async function fetchSatellites() {
  try {
    const data = await api.get('/getSatellites');
    situation.update(s => ({ ...s, GPSSatellitesData: data || [] }));
    return data;
  } catch (e) {
    console.error('Failed to fetch satellites:', e);
    return [];
  }
}
