import { writable, derived } from 'svelte/store';
import { createWebSocketConnection, api } from '../services/websocket.js';

// Constants
const TRAFFIC_MAX_AGE_SECONDS = 59;
const TRAFFIC_AIS_MAX_AGE_SECONDS = 60 * 15;
const TARGET_TYPE_AIS = 5;

// Traffic source colors
export const trafficSourceColors = {
  1: 'cornflowerblue', // ES
  2: '#FF8C00',        // UAT
  4: 'green',          // OGN
  5: '#0077be'         // AIS
};

/**
 * Traffic store - receives real-time updates from /traffic WebSocket
 * Contains two lists: valid (with position) and invalid (without position)
 */
export const trafficStore = writable({
  _connected: false,
  _error: null,
  validTraffic: [],    // Aircraft with valid positions
  invalidTraffic: []   // Aircraft without valid positions (Mode S, etc.)
});

/**
 * Format degrees to DMS string (without seconds for space)
 */
function dmsString(val) {
  const deg = Math.trunc(val);
  const min = Math.trunc(Math.abs(val % 1) * 60);
  const degStr = Math.abs(deg) < 10 ? `0${Math.abs(deg)}` : `${Math.abs(deg)}`;
  const minStr = min < 10 ? `0${min}` : `${min}`;
  return `${deg < 0 ? '-' : ''}${degStr}° ${minStr}'`;
}

/**
 * Format UTC time string from timestamp
 */
function utcTimeString(timestamp) {
  const d = new Date(timestamp);
  const pad = n => n.toString().padStart(2, '0');
  return `${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}:${pad(d.getUTCSeconds())}Z`;
}

/**
 * Get aircraft category description
 */
function getCategory(emitterCategory) {
  const categories = {
    0: 'No info',
    1: 'Light',
    2: 'Small',
    3: 'Large',
    4: 'HiVortex',
    5: 'Heavy',
    6: 'HiPerf',
    7: 'Rotor',
    9: 'Glider',
    10: 'Balloon',
    11: 'Skydiver',
    12: 'Ultralight',
    14: 'UAV',
    15: 'Space',
    17: 'Surface',
    18: 'SurfEmerg',
    19: 'SurfServ',
    20: 'PointObst',
    21: 'ClusterObst',
    22: 'LineObst'
  };
  return categories[emitterCategory] || '---';
}

/**
 * Get traffic color based on type and source
 */
function getTrafficColor(obj) {
  if (obj.TargetType === TARGET_TYPE_AIS) {
    return '#0077be'; // AIS blue
  }
  return trafficSourceColors[obj.Last_source] || 'gray';
}

/**
 * Check if traffic is aged (should be removed)
 */
function isTrafficAged(aircraft, field = 'Age') {
  const value = aircraft[field];
  if (aircraft.TargetType === TARGET_TYPE_AIS) {
    return value > TRAFFIC_AIS_MAX_AGE_SECONDS;
  }
  return value > TRAFFIC_MAX_AGE_SECONDS;
}

/**
 * Transform raw traffic message to display format
 */
function transformTraffic(obj) {
  const icaoHex = obj.Icao_addr.toString(16).toUpperCase();
  const alt = Math.round(obj.Alt / 25) * 25;
  const speed = Math.round(obj.Speed / 5) * 5;
  const timestamp = Date.parse(obj.Timestamp);

  return {
    icao_int: obj.Icao_addr,
    icao: icaoHex,
    isStratux: obj.IsStratux,
    signal: obj.SignalLevel,
    Last_source: obj.Last_source,
    Emitter_category: obj.Emitter_category,
    TargetType: obj.TargetType,
    tail: obj.Tail?.trim() || '[--N/A--]',
    reg: obj.Reg?.trim() || '[--N/A--]',
    squawk: obj.Squawk === 0 ? '----' : obj.Squawk,
    addr_type: obj.Addr_type,
    lat: dmsString(obj.Lat),
    lng: dmsString(obj.Lng),
    alt: alt.toLocaleString(),
    speed: obj.Speed_valid ? speed.toString() : '---',
    heading: obj.Speed_valid ? Math.round(obj.Track / 5) * 5 : '---',
    vspeed: Math.round(obj.Vvel / 100) * 100,
    time: utcTimeString(timestamp),
    Age: obj.Age,
    AgeLastAlt: obj.AgeLastAlt,
    bearing: Math.round(obj.Bearing),
    dist: obj.Distance / 1852,  // nautical miles
    distEst: obj.DistanceEstimated / 1852,
    trafficColor: getTrafficColor(obj),
    category: getCategory(obj.Emitter_category),
    Position_valid: obj.Position_valid,
    // Symbol based on target type
    addr_symb: obj.TargetType === TARGET_TYPE_AIS ? '\uD83D\uDEA2' : '\u2708'
  };
}

/**
 * Check if two aircraft are the same based on address and type
 */
function isSameAircraft(addr1, addrType1, addr2, addrType2) {
  if (addr1 !== addr2) return false;
  // Both aircraft have the same address and it is either an ICAO address for both,
  // or a non-icao address for both. (1 = non-icao, everything else = icao)
  return (addrType1 === 1 && addrType2 === 1) || (addrType1 !== 1 && addrType2 !== 1);
}

// WebSocket connection
let wsConnection = null;
let cleanupInterval = null;

/**
 * Custom WebSocket handler for traffic that accumulates data
 */
function createTrafficConnection() {
  let socket = null;
  let reconnectTimer = null;
  let shouldReconnect = true;

  const getWebSocketUrl = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host || '192.168.10.1';
    return `${protocol}//${host}/traffic`;
  };

  const connect = () => {
    if (socket && socket.readyState === WebSocket.OPEN) return;

    try {
      socket = new WebSocket(getWebSocketUrl());

      socket.onopen = () => {
        trafficStore.update(state => ({ ...state, _connected: true, _error: null }));
      };

      socket.onclose = () => {
        trafficStore.update(state => ({ ...state, _connected: false }));
        if (shouldReconnect) {
          reconnectTimer = setTimeout(connect, 1000);
        }
      };

      socket.onerror = () => {
        trafficStore.update(state => ({ ...state, _error: 'Connection error', _connected: false }));
      };

      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          const traffic = transformTraffic(message);

          trafficStore.update(state => {
            const validIdx = state.validTraffic.findIndex(
              t => isSameAircraft(t.icao_int, t.addr_type, message.Icao_addr, message.Addr_type)
            );
            const invalidIdx = state.invalidTraffic.findIndex(
              t => isSameAircraft(t.icao_int, t.addr_type, message.Icao_addr, message.Addr_type)
            );

            let validTraffic = [...state.validTraffic];
            let invalidTraffic = [...state.invalidTraffic];

            if (message.Position_valid) {
              // Valid position - add/update in valid list
              if (validIdx >= 0) {
                validTraffic[validIdx] = traffic;
              } else {
                validTraffic = [traffic, ...validTraffic];
              }
              // Remove from invalid list if it was there
              if (invalidIdx >= 0) {
                invalidTraffic.splice(invalidIdx, 1);
              }
            } else {
              // Invalid position - add/update in invalid list
              if (invalidIdx >= 0) {
                invalidTraffic[invalidIdx] = traffic;
              } else {
                invalidTraffic = [traffic, ...invalidTraffic];
              }
              // Remove from valid list if it was there
              if (validIdx >= 0) {
                validTraffic.splice(validIdx, 1);
              }
            }

            return { ...state, validTraffic, invalidTraffic };
          });
        } catch (e) {
          console.error('Failed to parse traffic message:', e);
        }
      };
    } catch (e) {
      console.error('Failed to create WebSocket:', e);
      if (shouldReconnect) {
        reconnectTimer = setTimeout(connect, 1000);
      }
    }
  };

  const disconnect = () => {
    shouldReconnect = false;
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (socket) {
      socket.close();
      socket = null;
    }
  };

  return { connect, disconnect };
}

/**
 * Clean up stale traffic entries
 */
function cleanupStaleTraffic() {
  trafficStore.update(state => {
    const validTraffic = state.validTraffic.filter(t => !isTrafficAged(t));
    const invalidTraffic = state.invalidTraffic.filter(
      t => !isTrafficAged(t) && !isTrafficAged(t, 'AgeLastAlt')
    );
    return { ...state, validTraffic, invalidTraffic };
  });
}

export function connectTraffic() {
  if (!wsConnection) {
    wsConnection = createTrafficConnection();
  }
  wsConnection.connect();

  // Start cleanup interval (every 10 seconds)
  if (!cleanupInterval) {
    cleanupInterval = setInterval(cleanupStaleTraffic, 10000);
  }
}

export function disconnectTraffic() {
  if (wsConnection) {
    wsConnection.disconnect();
    wsConnection = null;
  }
  if (cleanupInterval) {
    clearInterval(cleanupInterval);
    cleanupInterval = null;
  }
  // Clear traffic data
  trafficStore.set({
    _connected: false,
    _error: null,
    validTraffic: [],
    invalidTraffic: []
  });
}

// Derived stores
export const validTraffic = derived(trafficStore, $store =>
  [...$store.validTraffic].sort((a, b) => a.dist - b.dist)
);

export const invalidTraffic = derived(trafficStore, $store =>
  [...$store.invalidTraffic].sort((a, b) => a.icao.localeCompare(b.icao))
);

export const trafficConnected = derived(trafficStore, $store => $store._connected);
