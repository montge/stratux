import { writable, derived } from 'svelte/store';
import { api } from '../services/websocket.js';

const MAX_RECENT_ITEMS = 10;
const STALE_CUTOFF_MS = 30 * 60 * 1000; // 30 minutes
const DEFAULT_WATCHLIST = 'KBOS KATL KORD KLAX';

/**
 * Weather store - receives real-time updates from /weather WebSocket
 */
export const weatherStore = writable({
  _connected: false,
  _error: null,
  watchList: [],    // Watched airport reports
  recentList: [],   // Recent incoming reports (rolling buffer)
  watching: DEFAULT_WATCHLIST
});

/**
 * Parse METAR/SPECI flight condition from report body.
 * Returns VFR, MVFR, IFR, LIFR, or empty string.
 */
function parseFlightCondition(type, body) {
  if (type !== 'METAR' && type !== 'SPECI') return '';

  let visibility;

  // Check for mixed number format: "X X/XSM"
  let match = /(\d) (\d)\/(\d)SM/.exec(body);
  if (match && match.length === 4) {
    visibility = parseInt(match[1]) + (parseInt(match[2]) / parseInt(match[3]));
  } else {
    match = /([0-9/]{1,5}?)SM/.exec(body);
    if (!match) return '';
    if (match[1].length === 3) return 'LIFR';
    visibility = parseInt(match[1]);
    if (visibility === 0) return '';
  }

  // Ceiling from BKN or OVC layer
  let ceiling = 999;
  match = /BKN(\d{3})/.exec(body);
  if (!match) match = /OVC(\d{3})/.exec(body);
  if (match) ceiling = parseInt(match[1]);

  if (visibility > 5 && ceiling > 30) return 'VFR';
  if (visibility >= 3 && ceiling >= 10) return 'MVFR';
  if (visibility >= 1 && ceiling >= 5) return 'IFR';
  return 'LIFR';
}

/**
 * Format a time delta as "XXd XXh XXm" string
 */
function deltaTimeString(ms) {
  const d = new Date(ms);
  let time = '';
  const days = d.getUTCDate() - 1;
  if (days > 0) time += (days < 10 ? '0' + days : days) + 'd ';
  const hours = d.getUTCHours();
  if (hours > 0) {
    time += (hours < 10 ? '0' + hours : hours) + 'h ';
  } else if (time.length > 0) {
    time += '00h ';
  }
  const mins = d.getUTCMinutes();
  time += (mins < 10 ? '0' + mins : mins) + 'm';
  return time;
}

/**
 * Parse short datetime from weather message (DDHHMMz or DDHH/DDHH for TAF)
 */
function parseShortDatetime(sdt) {
  const s = String(sdt);
  if (s.length < 7) return new Date(0);
  const d = new Date();
  d.setUTCDate(parseInt(s.substring(0, 2)));
  d.setUTCHours(parseInt(s.substring(2, 4)));
  if (s.length > 7) {
    d.setUTCMinutes(0); // TAF datetime range
  } else {
    d.setUTCMinutes(parseInt(s.substring(4, 6)));
  }
  d.setUTCSeconds(0);
  d.setUTCMilliseconds(0);
  return d;
}

/**
 * Transform raw weather message to display format
 */
function transformWeatherItem(obj) {
  const type = obj.Type === 'TAF.AMD' ? 'TAF' : obj.Type;
  const isAmended = obj.Type === 'TAF.AMD';
  const flightCondition = parseFlightCondition(obj.Type, obj.Data);

  const now = new Date();
  const then = parseShortDatetime(obj.Time);
  const diffMs = Math.abs(then - now);

  let timeStr;
  if (diffMs > 1000 * 60 * 60 * 24 * 2) {
    timeStr = '?';
  } else if (then > now) {
    timeStr = deltaTimeString(then - now) + ' from now';
  } else {
    timeStr = deltaTimeString(now - then) + ' old';
  }

  return {
    type,
    isAmended,
    flightCondition,
    location: obj.Location,
    age: then.getTime(),
    time: timeStr,
    data: obj.Data
  };
}

// WebSocket connection
let socket = null;
let reconnectTimer = null;
let shouldReconnect = true;
let cleanupInterval = null;

function getWebSocketUrl() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host || '192.168.10.1';
  return `${protocol}//${host}/weather`;
}

export function connectWeather() {
  if (socket && socket.readyState === WebSocket.OPEN) return;
  shouldReconnect = true;

  // Fetch watchlist from settings
  fetchWatchList();

  try {
    socket = new WebSocket(getWebSocketUrl());

    socket.onopen = () => {
      weatherStore.update(s => ({ ...s, _connected: true, _error: null }));
    };

    socket.onclose = () => {
      weatherStore.update(s => ({ ...s, _connected: false }));
      if (shouldReconnect) {
        reconnectTimer = setTimeout(connectWeather, 1000);
      }
    };

    socket.onerror = () => {
      weatherStore.update(s => ({ ...s, _error: 'Connection error', _connected: false }));
    };

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        const item = transformWeatherItem(message);

        weatherStore.update(state => {
          let watchList = [...state.watchList];
          let recentList = [...state.recentList];

          // Check if location is in the watchlist
          if (state.watching && state.watching.includes(message.Location)) {
            const existingIdx = watchList.findIndex(
              w => w.type === item.type && w.location === item.location
            );
            if (existingIdx >= 0) {
              watchList[existingIdx] = item;
            } else {
              watchList = [item, ...watchList];
            }
          }

          // Add to recent list (rolling buffer)
          recentList = [item, ...recentList];
          if (recentList.length > MAX_RECENT_ITEMS) {
            recentList = recentList.slice(0, MAX_RECENT_ITEMS);
          }

          return { ...state, watchList, recentList };
        });
      } catch (e) {
        console.error('Failed to parse weather message:', e);
      }
    };
  } catch (e) {
    console.error('Failed to create weather WebSocket:', e);
    if (shouldReconnect) {
      reconnectTimer = setTimeout(connectWeather, 1000);
    }
  }

  // Start stale message cleanup (every 5 minutes)
  if (!cleanupInterval) {
    cleanupInterval = setInterval(cleanupStaleWeather, 5 * 60 * 1000);
  }
}

export function disconnectWeather() {
  shouldReconnect = false;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (socket) {
    socket.close();
    socket = null;
  }
  if (cleanupInterval) {
    clearInterval(cleanupInterval);
    cleanupInterval = null;
  }
  weatherStore.set({
    _connected: false,
    _error: null,
    watchList: [],
    recentList: [],
    watching: DEFAULT_WATCHLIST
  });
}

function cleanupStaleWeather() {
  const cutoff = Date.now() - STALE_CUTOFF_MS;
  weatherStore.update(state => ({
    ...state,
    watchList: state.watchList.filter(w => w.age >= cutoff)
  }));
}

async function fetchWatchList() {
  try {
    const data = await api.get('/getSettings');
    if (data.WatchList) {
      weatherStore.update(s => ({ ...s, watching: data.WatchList.toUpperCase() }));
    }
  } catch (e) {
    // Use default watchlist on error
  }
}

// Derived stores
export const weatherConnected = derived(weatherStore, $s => $s._connected);
export const watchList = derived(weatherStore, $s =>
  [...$s.watchList].sort((a, b) => b.age - a.age)
);
export const recentList = derived(weatherStore, $s => $s.recentList);
export const watching = derived(weatherStore, $s => $s.watching);
