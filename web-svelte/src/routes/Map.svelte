<script>
  import { onMount, onDestroy } from 'svelte';
  import Map from 'ol/Map';
  import View from 'ol/View';
  import TileLayer from 'ol/layer/Tile';
  import VectorLayer from 'ol/layer/Vector';
  import VectorTileLayer from 'ol/layer/VectorTile';
  import VectorSource from 'ol/source/Vector';
  import VectorTileSource from 'ol/source/VectorTile';
  import OSM from 'ol/source/OSM';
  import XYZ from 'ol/source/XYZ';
  import MVT from 'ol/format/MVT';
  import Feature from 'ol/Feature';
  import Point from 'ol/geom/Point';
  import LineString from 'ol/geom/LineString';
  import { fromLonLat, transformExtent } from 'ol/proj';
  import { Style, Icon, Text, Stroke, Fill } from 'ol/style';
  import { api } from '../services/websocket.js';

  const TRAFFIC_MAX_AGE_SECONDS = 15;
  const TRAFFIC_AIS_MAX_AGE_SECONDS = 60 * 15;
  const TARGET_TYPE_AIS = 5;

  // Traffic source colors
  const sourceColors = {
    1: 'cornflowerblue',
    2: '#FF8C00',
    4: 'green',
    5: '#0077be'
  };

  let mapContainer;
  let map = null;
  let trafficSocket = null;
  let gpsSocket = null;
  let updateInterval = null;
  let shouldReconnect = true;

  let connectState = $state('Disconnected');

  // OL sources for aircraft
  let aircraftSymbolsSource = new VectorSource();
  let aircraftTrailsSource = new VectorSource();

  // Aircraft tracking
  let aircraft = [];
  let gpsLayer = null;

  // --- Utility functions ---

  function toRadians(deg) {
    return deg * Math.PI / 180;
  }

  function toDegrees(rad) {
    return rad * 180 / Math.PI;
  }

  function bearing(startLng, startLat, destLng, destLat) {
    const sLat = toRadians(startLat);
    const sLng = toRadians(startLng);
    const dLat = toRadians(destLat);
    const dLng = toRadians(destLng);
    const y = Math.sin(dLng - sLng) * Math.cos(dLat);
    const x = Math.cos(sLat) * Math.sin(dLat) - Math.sin(sLat) * Math.cos(dLat) * Math.cos(dLng - sLng);
    return (toDegrees(Math.atan2(y, x)) + 360) % 360;
  }

  function distance(lon1, lat1, lon2, lat2) {
    const R = 6371;
    const dLat = toRadians(lat2 - lat1);
    const dLon = toRadians(lon2 - lon1);
    const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos(toRadians(lat1)) * Math.cos(toRadians(lat2)) *
      Math.sin(dLon / 2) * Math.sin(dLon / 2);
    return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  }

  function getTrafficColor(ac) {
    if (ac.TargetType === TARGET_TYPE_AIS) return '#0077be';
    return sourceColors[ac.Last_source] || 'gray';
  }

  function isTrafficAged(ac) {
    if (ac.TargetType === TARGET_TYPE_AIS) return ac.Age > TRAFFIC_AIS_MAX_AGE_SECONDS;
    return ac.Age > TRAFFIC_MAX_AGE_SECONDS;
  }

  function getAircraftIcon(ac) {
    if (ac.TargetType === TARGET_TYPE_AIS) return ['/img/actype/vessel.svg', true];
    switch (ac.Emitter_category) {
      case 1: case 6: return ['/img/actype/light.svg', true];
      case 2: case 3: case 4: case 5: return ['/img/actype/heavy.svg', true];
      case 7: return ['/img/actype/helicopter.svg', true];
      case 9: return ['/img/actype/glider.svg', true];
      case 10: return ['/img/actype/lighter-than-air.svg', false];
      case 11: case 12: return ['/img/actype/skydiver.svg', false];
      default: return ['/img/actype/undef.svg', true];
    }
  }

  function isSameAircraft(addr1, addrType1, addr2, addrType2) {
    if (addr1 !== addr2) return false;
    return (addrType1 === 1 && addrType2 === 1) || (addrType1 !== 1 && addrType2 !== 1);
  }

  function computeTrackFromPositions(ac) {
    let dist = 0;
    let prev = [ac.Lng, ac.Lat];
    let i;
    for (i = ac.posHistory.length - 1; i >= 0; i--) {
      dist += distance(prev[0], prev[1], ac.posHistory[i][0], ac.posHistory[i][1]);
      prev = ac.posHistory[i];
      if (dist >= 0.5) break;
    }
    if (dist !== 0 && i >= 0) {
      return bearing(ac.posHistory[i][0], ac.posHistory[i][1], ac.Lng, ac.Lat);
    }
    return 0;
  }

  function clipPosHistory(ac, maxLenKm) {
    let dist = 0;
    let i;
    for (i = ac.posHistory.length - 2; i >= 0; i--) {
      const prev = ac.posHistory[i + 1];
      const curr = ac.posHistory[i];
      dist += distance(prev[0], prev[1], curr[0], curr[1]);
      if (dist > maxLenKm) break;
    }
    if (i > 0) ac.posHistory = ac.posHistory.slice(i);
  }

  function updateOpacity(ac) {
    let opacity;
    if (isTrafficAged(ac)) {
      opacity = 0.0;
    } else if (ac.TargetType === TARGET_TYPE_AIS) {
      opacity = 1.0;
    } else {
      opacity = 1.0 - (ac.Age / TRAFFIC_MAX_AGE_SECONDS);
    }
    ac.marker?.getStyle()?.getImage()?.setOpacity(opacity);
  }

  function updateVehicleText(ac) {
    const text = [];
    if (ac.Tail?.length > 0) text.push(ac.Tail);
    if (ac.TargetType !== TARGET_TYPE_AIS) text.push(ac.Alt + 'ft');
    if (ac.Speed_valid && ac.Speed > 0.1) text.push(ac.Speed + 'kt');
    ac.marker?.getStyle()?.getText()?.setText(text.join('\n'));
  }

  function updateAircraftTrail(ac) {
    if (!ac.posHistory || ac.posHistory.length < 2) return;
    const coords = ac.posHistory.map(c => fromLonLat(c));
    coords.push(fromLonLat([ac.Lng, ac.Lat]));

    if (!ac.trail) {
      ac.trail = new Feature({ geometry: new LineString(coords) });
      aircraftTrailsSource.addFeature(ac.trail);
    } else {
      ac.trail.getGeometry().setCoordinates(coords);
    }
  }

  // --- Traffic message handler ---

  function onTrafficMessage(msg) {
    const ac = JSON.parse(msg.data);
    if (!ac.Position_valid || isTrafficAged(ac)) return;
    ac.receivedTs = Date.now();

    let prevColor;
    let prevEmitterCat;
    let updateIndex = -1;

    for (let i = 0; i < aircraft.length; i++) {
      if (isSameAircraft(aircraft[i].Icao_addr, aircraft[i].Addr_type, ac.Icao_addr, ac.Addr_type)) {
        const old = aircraft[i];
        prevColor = getTrafficColor(old);
        prevEmitterCat = old.Emitter_category;
        ac.marker = old.marker;
        ac.trail = old.trail;
        ac.posHistory = old.posHistory;

        const prevPos = ac.posHistory[ac.posHistory.length - 1];
        if (distance(prevPos[0], prevPos[1], ac.Lng, ac.Lat) > 0.1) {
          ac.posHistory.push([ac.Lng, ac.Lat]);
        }
        clipPosHistory(ac, 9.25);

        if (!ac.Speed_valid) {
          ac.Track = computeTrackFromPositions(ac);
        }
        aircraft[i] = ac;
        updateIndex = i;
        break;
      }
    }

    if (updateIndex < 0) {
      aircraft.push(ac);
      ac.posHistory = [[ac.Lng, ac.Lat]];
    }

    const acPosition = [ac.Lng, ac.Lat];

    if (!ac.marker) {
      const offsetY = ac.TargetType === TARGET_TYPE_AIS ? 20 : 40;
      const style = new Style({
        text: new Text({
          text: '',
          offsetY,
          font: 'bold 1em sans-serif',
          stroke: new Stroke({ color: 'white', width: 2 })
        })
      });
      const feature = new Feature({ geometry: new Point(fromLonLat(acPosition)) });
      feature.setStyle(style);
      ac.marker = feature;
      aircraftSymbolsSource.addFeature(feature);
    } else {
      ac.marker.getGeometry().setCoordinates(fromLonLat(acPosition));
      updateAircraftTrail(ac);
    }

    updateVehicleText(ac);

    if (!prevColor || prevColor !== getTrafficColor(ac) || prevEmitterCat !== ac.Emitter_category) {
      const [icon, rotatable] = getAircraftIcon(ac);
      const imageStyle = new Icon({
        opacity: 1.0,
        src: icon,
        rotation: rotatable ? toRadians(ac.Track) : 0,
        anchor: [0.5, 0.5],
        anchorXUnits: 'fraction',
        anchorYUnits: 'fraction',
        color: getTrafficColor(ac)
      });
      ac.marker.getStyle().setImage(imageStyle);
    }

    updateOpacity(ac);
    ac.marker.getStyle().getImage()?.setRotation(toRadians(ac.Track));
  }

  // --- Age updates & stale removal ---

  function updateAges() {
    const now = Date.now();
    for (const ac of aircraft) {
      if (!ac.ageReceived) ac.ageReceived = ac.Age;
      ac.Age = ac.ageReceived + (now - ac.receivedTs) / 1000.0;
      updateOpacity(ac);
    }
  }

  function removeStaleTraffic() {
    for (let i = aircraft.length - 1; i >= 0; i--) {
      if (isTrafficAged(aircraft[i])) {
        if (aircraft[i].marker) aircraftSymbolsSource.removeFeature(aircraft[i].marker);
        if (aircraft[i].trail) aircraftTrailsSource.removeFeature(aircraft[i].trail);
        aircraft.splice(i, 1);
      }
    }
  }

  function tick() {
    updateAges();
    removeStaleTraffic();
  }

  // --- GPS position ---

  function updateMyLocation(data) {
    const lat = data.GPSLatitude;
    const lon = data.GPSLongitude;
    if (data.GPSFixQuality <= 0) return;

    if (!gpsLayer) {
      const pos = fromLonLat([lon, lat]);
      map.setView(new View({
        center: pos,
        zoom: 10,
        enableRotation: false
      }));
      gpsLayer = new VectorLayer({
        source: new VectorSource({
          features: [new Feature({ geometry: new Point(pos), name: 'Your GPS position' })]
        }),
        style: new Style({
          text: new Text({
            text: '\u25CE',
            font: 'normal 28px sans-serif',
            fill: new Fill({ color: '#1976d2' }),
            stroke: new Stroke({ color: 'white', width: 2 })
          })
        })
      });
      map.addLayer(gpsLayer);
    } else {
      const geom = new Point(fromLonLat([lon, lat]));
      gpsLayer.getSource().getFeatures()[0].setGeometry(geom);
    }
  }

  // --- WebSocket connections ---

  function getWsUrl(endpoint) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host || '192.168.10.1';
    return `${protocol}//${host}${endpoint}`;
  }

  function connectTraffic() {
    if (trafficSocket) return;
    trafficSocket = new WebSocket(getWsUrl('/traffic'));

    trafficSocket.onopen = () => { connectState = 'Connected'; };
    trafficSocket.onclose = () => {
      connectState = 'Disconnected';
      trafficSocket = null;
      if (shouldReconnect) setTimeout(connectTraffic, 1000);
    };
    trafficSocket.onerror = () => { connectState = 'Problem'; };
    trafficSocket.onmessage = onTrafficMessage;
  }

  function connectGPS() {
    if (gpsSocket) return;
    gpsSocket = new WebSocket(getWsUrl('/situation'));

    gpsSocket.onclose = () => {
      gpsSocket = null;
      if (shouldReconnect) setTimeout(connectGPS, 1000);
    };
    gpsSocket.onmessage = (msg) => {
      try { updateMyLocation(JSON.parse(msg.data)); } catch (e) { /* ignore parse errors */ }
    };
  }

  // --- Dynamic MBTiles layers ---

  async function loadTilesets() {
    try {
      const tilesets = await api.get('/getTilesets');
      const aircraftSymbolsLayer = new VectorLayer({
        title: 'Aircraft symbols',
        source: aircraftSymbolsSource,
        zIndex: 10
      });
      const aircraftTrailsLayer = new VectorLayer({
        title: 'Aircraft trails 5NM',
        source: aircraftTrailsSource,
        zIndex: 9
      });

      for (const file in tilesets) {
        const meta = tilesets[file];
        const name = meta.name || file;
        const baselayer = meta.type === 'baselayer';
        const format = meta.format || 'png';
        const minzoom = meta.minzoom ? parseInt(meta.minzoom) : 1;
        const maxzoom = meta.maxzoom ? parseInt(meta.maxzoom) : 18;

        let ext = [-180, -85, 180, 85];
        if (meta.bounds) ext = meta.bounds.split(',').map(Number);
        ext = transformExtent(ext, 'EPSG:4326', 'EPSG:3857');

        let layer;
        if (format.toLowerCase() === 'pbf') {
          layer = new VectorTileLayer({
            title: name,
            type: baselayer ? 'base' : 'overlay',
            extent: ext,
            source: new VectorTileSource({
              url: '/getTile/' + file + '/{z}/{x}/{-y}.' + format,
              format: new MVT(),
              maxZoom: maxzoom,
              minZoom: minzoom
            })
          });
        } else {
          layer = new TileLayer({
            title: name,
            type: baselayer ? 'base' : 'overlay',
            extent: ext,
            source: new XYZ({
              url: '/getTile/' + file + '/{z}/{x}/{-y}.' + format,
              maxZoom: maxzoom,
              minZoom: minzoom
            })
          });
        }

        if (baselayer) {
          map.getLayers().insertAt(0, layer);
        } else {
          map.addLayer(layer);
        }
      }

      map.addLayer(aircraftSymbolsLayer);
      map.addLayer(aircraftTrailsLayer);

      // Restore layer visibility from localStorage
      map.getLayers().forEach((layer) => {
        const title = layer.get('title');
        if (!title) return;
        const key = 'stratux.map.layers.' + title + '.visible';
        const saved = localStorage.getItem(key);
        if (saved !== null) layer.setVisible(saved === 'true');
      });

      // Save layer visibility changes
      map.getLayers().forEach((layer) => {
        layer.on('change:visible', (ev) => {
          const title = ev.target.get('title');
          if (!title) return;
          localStorage.setItem('stratux.map.layers.' + title + '.visible', ev.target.get('visible'));
        });
      });
    } catch (e) {
      // No tilesets available — just add aircraft layers directly
      map.addLayer(new VectorLayer({
        title: 'Aircraft symbols',
        source: aircraftSymbolsSource,
        zIndex: 10
      }));
      map.addLayer(new VectorLayer({
        title: 'Aircraft trails 5NM',
        source: aircraftTrailsSource,
        zIndex: 9
      }));
    }
  }

  // --- Lifecycle ---

  onMount(() => {
    map = new Map({
      target: mapContainer,
      layers: [
        new TileLayer({ title: 'OSM', type: 'base', source: new OSM() }),
        new TileLayer({
          title: 'OpenAIP',
          type: 'overlay',
          visible: false,
          source: new XYZ({
            url: 'https://api.tiles.openaip.net/api/data/openaip/{z}/{x}/{y}.png?apiKey=f64474b4ab9d2f6bacb2f30d4680e8ae'
          })
        })
      ],
      view: new View({
        center: fromLonLat([10.0, 52.0]),
        zoom: 4,
        enableRotation: false
      })
    });

    loadTilesets();
    connectTraffic();
    connectGPS();
    updateInterval = setInterval(tick, 1000);
  });

  onDestroy(() => {
    shouldReconnect = false;
    if (trafficSocket) { trafficSocket.close(); trafficSocket = null; }
    if (gpsSocket) { gpsSocket.close(); gpsSocket = null; }
    if (updateInterval) { clearInterval(updateInterval); updateInterval = null; }
    if (map) { map.setTarget(null); map = null; }
  });
</script>

<div class="map-page">
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">Traffic Map</span>
      {#if connectState === 'Connected'}
        <span class="badge success">Connected</span>
      {:else}
        <span class="badge danger">{connectState}</span>
      {/if}
    </div>
    <div class="panel-body-fullsize">
      <div bind:this={mapContainer} class="map-container"></div>
    </div>
  </div>
</div>

<style>
  .map-page {
    padding: 0;
    height: calc(100vh - 56px);
    display: flex;
    flex-direction: column;
  }

  .panel {
    background: white;
    border-radius: 4px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12);
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
  }

  .panel-heading {
    padding: 10px 15px;
    background: #f8f9fa;
    border-bottom: 1px solid #dee2e6;
    display: flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;
  }

  .panel-label {
    font-weight: 600;
    font-size: 1.1em;
  }

  .badge {
    padding: 2px 8px;
    border-radius: 3px;
    font-size: 0.8em;
    font-weight: 600;
    color: white;
  }

  .badge.success {
    background: #28a745;
  }

  .badge.danger {
    background: #dc3545;
  }

  .panel-body-fullsize {
    flex: 1;
    min-height: 0;
    position: relative;
  }

  .map-container {
    width: 100%;
    height: 100%;
    position: absolute;
    top: 0;
    left: 0;
  }
</style>
