/**
 * Mock Stratux server for e2e testing without a real device.
 * Serves static files and mocks API endpoints.
 *
 * Usage: node e2e/mock-server.js
 */

const http = require('http');
const fs = require('fs');
const path = require('path');
const { WebSocketServer } = require('ws');

const PORT = process.env.PORT || 9090;
const WEB_ROOT = path.join(__dirname, '..');

// Mock data
const mockStatus = {
  Version: '0.0.test',
  Build: 'mock-build-12345',
  HardwareBuild: '',
  Devices: 2,
  Connected_Users: 1,
  DiskBytesFree: 1000000000,
  UAT_messages_last_minute: 0,
  UAT_messages_max: 0,
  UAT_messages_total: 0,
  ES_messages_last_minute: 10,
  ES_messages_max: 50,
  ES_messages_total: 100,
  OGN_messages_last_minute: 0,
  OGN_messages_max: 0,
  OGN_messages_total: 0,
  OGN_connected: false,
  AIS_messages_last_minute: 0,
  AIS_messages_max: 0,
  AIS_messages_total: 0,
  AIS_connected: false,
  GPS_satellites_locked: 8,
  GPS_satellites_tracked: 12,
  GPS_satellites_seen: 15,
  GPS_solution: '3D GPS',
  GPS_detected_type: 0x15, // u-blox with NMEA
  GPS_position_accuracy: 2.5,
  GPS_NetworkRemoteIp: '',
  Uptime: 3600000,
  CPUTemp: 45.0,
  CPUTempMin: 40.0,
  CPUTempMax: 55.0,
  Errors: [],
  Ping_connected: false,
  Pong_connected: false,
  Pong_Heartbeats: 0,
  UAT_METAR_total: 0,
  UAT_TAF_total: 0,
  UAT_NEXRAD_total: 0,
  UAT_SIGMET_total: 0,
  UAT_PIREP_total: 0,
  UAT_NOTAM_total: 0,
  UAT_OTHER_total: 0,
  OGN_noise_db: 5.0,
  OGN_gain_db: 40.0,
};

const mockSettings = {
  UAT_Enabled: true,
  ES_Enabled: true,
  OGN_Enabled: false,
  AIS_Enabled: false,
  GPS_Enabled: true,
  Ping_Enabled: false,
  Pong_Enabled: false,
  DeveloperMode: true,
  DarkMode: false,
  RegionSelect: 1,
};

const mockSituation = {
  GPSLatitude: 37.7749,
  GPSLongitude: -122.4194,
  GPSFixQuality: 1,
  GPSAltitudeMSL: 100,
  GPSVerticalAccuracy: 5,
  GPSHorizontalAccuracy: 2.5,
  GPSGroundSpeed: 0,
  GPSTrueCourse: 0,
  GPSTime: new Date().toISOString(),
};

const mockTowers = {};

const mockRegion = {
  IsSet: true,
  RegionSelect: 1,
};

// MIME types
const mimeTypes = {
  '.html': 'text/html',
  '.js': 'application/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.svg': 'image/svg+xml',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.ttf': 'font/ttf',
};

// HTTP server
const server = http.createServer((req, res) => {
  const url = new URL(req.url, `http://localhost:${PORT}`);
  const pathname = url.pathname;

  // CORS headers
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }

  // API endpoints
  if (pathname === '/getStatus') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(mockStatus));
    return;
  }

  if (pathname === '/getSettings') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(mockSettings));
    return;
  }

  if (pathname === '/getSituation') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(mockSituation));
    return;
  }

  if (pathname === '/getTowers') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(mockTowers));
    return;
  }

  if (pathname === '/getRegion') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(mockRegion));
    return;
  }

  if (pathname === '/getSatellites') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify([]));
    return;
  }

  if (pathname === '/tiles/tilesets') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({}));
    return;
  }

  // Serve static files
  let filePath = pathname === '/' ? '/index.html' : pathname;
  filePath = path.join(WEB_ROOT, filePath);

  // Security: prevent directory traversal
  if (!filePath.startsWith(WEB_ROOT)) {
    res.writeHead(403);
    res.end('Forbidden');
    return;
  }

  fs.readFile(filePath, (err, data) => {
    if (err) {
      if (err.code === 'ENOENT') {
        res.writeHead(404);
        res.end('Not Found');
      } else {
        res.writeHead(500);
        res.end('Server Error');
      }
      return;
    }

    const ext = path.extname(filePath);
    const contentType = mimeTypes[ext] || 'application/octet-stream';
    res.writeHead(200, { 'Content-Type': contentType });
    res.end(data);
  });
});

// WebSocket server for /status
const wss = new WebSocketServer({ server, path: '/status' });

wss.on('connection', (ws) => {
  console.log('WebSocket client connected to /status');

  // Send status updates every second
  const interval = setInterval(() => {
    mockStatus.Uptime += 1000;
    ws.send(JSON.stringify(mockStatus));
  }, 1000);

  ws.on('close', () => {
    clearInterval(interval);
    console.log('WebSocket client disconnected');
  });
});

// WebSocket for /traffic
const wssTraffic = new WebSocketServer({ server, path: '/traffic' });

wssTraffic.on('connection', (ws) => {
  console.log('WebSocket client connected to /traffic');

  // Send mock traffic occasionally
  const interval = setInterval(() => {
    const mockTraffic = {
      Icao_addr: 0xABCDEF,
      Addr_type: 0,
      Lat: 37.78 + (Math.random() - 0.5) * 0.1,
      Lng: -122.42 + (Math.random() - 0.5) * 0.1,
      Position_valid: true,
      Alt: 5000 + Math.floor(Math.random() * 5000),
      Track: Math.floor(Math.random() * 360),
      Speed: 150 + Math.floor(Math.random() * 100),
      Speed_valid: true,
      Vvel: 0,
      Tail: 'N12345',
      Last_source: 1,
      TargetType: 1,
      Emitter_category: 1,
      Age: 0,
    };
    ws.send(JSON.stringify(mockTraffic));
  }, 2000);

  ws.on('close', () => {
    clearInterval(interval);
  });
});

// WebSocket for /situation
const wssSituation = new WebSocketServer({ server, path: '/situation' });

wssSituation.on('connection', (ws) => {
  console.log('WebSocket client connected to /situation');

  const interval = setInterval(() => {
    ws.send(JSON.stringify(mockSituation));
  }, 1000);

  ws.on('close', () => {
    clearInterval(interval);
  });
});

server.listen(PORT, () => {
  console.log(`Mock Stratux server running at http://localhost:${PORT}`);
  console.log('Press Ctrl+C to stop');
});
