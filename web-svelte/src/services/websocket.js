/**
 * WebSocket service for Stratux real-time data connections
 * Provides automatic reconnection and Svelte store integration
 */

/**
 * Creates a reconnecting WebSocket connection that updates a Svelte store
 * @param {string} endpoint - WebSocket endpoint path (e.g., '/status')
 * @param {import('svelte/store').Writable} store - Svelte writable store to update
 * @param {function} [transform] - Optional transform function for incoming data
 * @returns {object} - Connection controller with connect/disconnect methods
 */
export function createWebSocketConnection(endpoint, store, transform = null) {
  let socket = null;
  let reconnectTimer = null;
  let isConnected = false;
  let shouldReconnect = true;

  const getWebSocketUrl = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host || '192.168.10.1';
    return `${protocol}//${host}${endpoint}`;
  };

  const connect = () => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      return;
    }

    try {
      socket = new WebSocket(getWebSocketUrl());

      socket.onopen = () => {
        isConnected = true;
        store.update(state => ({ ...state, _connected: true, _error: null }));
      };

      socket.onclose = () => {
        isConnected = false;
        store.update(state => ({ ...state, _connected: false }));

        if (shouldReconnect) {
          reconnectTimer = setTimeout(connect, 1000);
        }
      };

      socket.onerror = (error) => {
        store.update(state => ({ ...state, _error: 'Connection error', _connected: false }));
      };

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          const transformed = transform ? transform(data) : data;
          store.update(state => ({ ...state, ...transformed, _connected: true }));
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e);
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
    isConnected = false;
  };

  const reconnect = () => {
    shouldReconnect = true;
    connect();
  };

  return {
    connect,
    disconnect,
    reconnect,
    get isConnected() { return isConnected; }
  };
}

/**
 * Creates a simple HTTP API client
 */
export const api = {
  async get(endpoint) {
    const response = await fetch(endpoint);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
  },

  async post(endpoint, data = null) {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: data ? { 'Content-Type': 'application/json' } : {},
      body: data ? JSON.stringify(data) : null
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json().catch(() => ({}));
  }
};
