<script>
  import { onMount, onDestroy } from 'svelte';
  import {
    trafficStore,
    connectTraffic,
    disconnectTraffic
  } from '../stores/traffic.js';
  import {
    situation,
    connectSituation,
    disconnectSituation
  } from '../stores/situation.js';
  import { status, connectStatus, disconnectStatus } from '../stores/status.js';

  // Radar configuration
  const RADIUS_STEPS = [2, 5, 10, 20, 40];
  const ALT_STEPS = [5, 10, 20, 50, 100, 500];
  const RADAR_CUTOFF_SECONDS = 59;
  const EARTH_RADIUS_M = 6371008.8;

  let displayRadius = 10;    // NM
  let altThreshold = 20;     // in 100s of feet
  let radarContainer;

  // SVG viewBox dimensions
  const VB = 200; // half-width/height of the viewBox

  // Ownship altitude
  $: baroAltitude = getBaroAltitude($situation);
  $: baroAltValid = getBaroAltValid($situation);
  $: gpsValid = $situation.GPSHorizontalAccuracy < 19999;
  $: gpsCourse = $situation.GPSTrueCourse || 0;
  $: ownLat = $situation.GPSLatitude;
  $: ownLng = $situation.GPSLongitude;
  $: flightLevel = baroAltitude > -100000 ? 'FL' + Math.round(baroAltitude / 100) : '---';
  $: cpuTemp = ($status.CPUTemp || 0).toFixed(1);
  $: cpuTempClass = $status.CPUTemp < 70 ? 'temp-ok' : $status.CPUTemp < 80 ? 'temp-warn' : 'temp-danger';

  // Compute displayed traffic targets
  $: radarTargets = computeRadarTargets(
    $trafficStore.validTraffic,
    $trafficStore.invalidTraffic,
    ownLat, ownLng, gpsCourse, baroAltitude, displayRadius, altThreshold
  );

  function getBaroAltitude(sit) {
    const pressTime = Date.parse(sit.BaroLastMeasurementTime || 0);
    const gpsTime = Date.parse(sit.GPSLastGPSTimeStratuxTime || 0);
    if (gpsTime - pressTime < 1000 && sit.BaroPressureAltitude !== undefined) {
      return Math.round(sit.BaroPressureAltitude);
    }
    if (sit.GPSHorizontalAccuracy > 19999) return -100000;
    if (sit.BaroSourceType === 4) return Math.round(sit.BaroPressureAltitude || 0);
    return Math.round(sit.GPSAltitudeMSL || 0);
  }

  function getBaroAltValid(sit) {
    const pressTime = Date.parse(sit.BaroLastMeasurementTime || 0);
    const gpsTime = Date.parse(sit.GPSLastGPSTimeStratuxTime || 0);
    if (gpsTime - pressTime < 1000) {
      return sit.BaroSourceType !== 4 ? 'Baro' : 'AdsbEstimated';
    }
    if (sit.GPSHorizontalAccuracy > 19999) return 'Invalid';
    if (sit.BaroSourceType === 4) return 'AdsbEstimated';
    return 'GPS';
  }

  function radiansRel(angle) {
    if (angle > 180) angle -= 360;
    if (angle <= -180) angle += 360;
    return angle * Math.PI / 180;
  }

  function computeRadarTargets(validTraffic, invalidTraffic, lat, lng, course, baroAlt, radius, altThresh) {
    const targets = [];

    // Valid position traffic
    for (const t of validTraffic) {
      const alt = Math.round(t.Alt / 25) * 25;
      let altDiff;
      if (baroAlt > -100000 && alt > 0) {
        altDiff = Math.round((alt - baroAlt) / 100);
      } else {
        altDiff = alt > 0 ? alt : null;
      }

      if (altDiff !== null && Math.abs(altDiff) > altThresh) continue;

      // Calculate distance in NM from ownship
      const avgLat = radiansRel((lat + t.Lat) / 2);
      const distLat = (radiansRel(t.Lat - lat) * EARTH_RADIUS_M) / 1852;
      const distLng = ((radiansRel(t.Lng - lng) * EARTH_RADIUS_M) / 1852) * Math.abs(Math.cos(avgLat));
      const distNM = Math.sqrt(distLat * distLat + distLng * distLng);

      if (distNM > radius) continue;

      // Convert to radar coordinates (VB units)
      const x = (VB / radius) * distLng;
      const y = -(VB / radius) * distLat;

      const heading = t.Speed_valid ? Math.round(t.Track / 5) * 5 : 0;
      const speed = t.Speed_valid ? Math.round(t.Speed / 5) * 5 : null;
      const vspeed = Math.round(t.Vvel / 100) * 100;

      let altLabel = '';
      if (altDiff !== null) {
        const sign = altDiff >= 0 ? '+' : '-';
        const arrow = vspeed > 0 ? '\u2191' : vspeed < 0 ? '\u2193' : '';
        altLabel = sign + Math.abs(altDiff) + arrow;
      } else {
        altLabel = 'u/s';
      }

      targets.push({
        key: t.Icao_addr,
        x, y,
        heading,
        speed,
        tail: (t.Tail || '').trim() || '',
        altLabel,
        distNM,
        positionValid: true
      });
    }

    // Invalid position traffic (show as range circles)
    for (const t of invalidTraffic) {
      const alt = Math.round(t.Alt / 25) * 25;
      let altDiff;
      if (baroAlt > -100000 && alt > 0) {
        altDiff = Math.round((alt - baroAlt) / 100);
      } else {
        altDiff = alt > 0 ? alt : null;
      }

      if (altDiff !== null && Math.abs(altDiff) > altThresh) continue;

      const distNM = (t.DistanceEstimated || 0) / 1852;
      if (distNM > radius) continue;

      const circleR = Math.max((VB / radius) * distNM, 12);
      const vspeed = Math.round(t.Vvel / 100) * 100;

      let altLabel = '';
      if (altDiff !== null) {
        const sign = altDiff >= 0 ? '+' : '-';
        const arrow = vspeed > 0 ? '\u2191' : vspeed < 0 ? '\u2193' : '';
        altLabel = sign + Math.abs(altDiff) + arrow;
      } else {
        altLabel = 'u/s';
      }

      targets.push({
        key: t.Icao_addr,
        circleR,
        altLabel,
        tail: (t.Tail || '').trim() || '',
        positionValid: false
      });
    }

    return targets;
  }

  function cycleRadius(direction) {
    const idx = RADIUS_STEPS.indexOf(displayRadius);
    if (direction > 0 && idx < RADIUS_STEPS.length - 1) {
      displayRadius = RADIUS_STEPS[idx + 1];
    } else if (direction < 0 && idx > 0) {
      displayRadius = RADIUS_STEPS[idx - 1];
    }
  }

  function cycleAlt(direction) {
    const idx = ALT_STEPS.indexOf(altThreshold);
    if (direction > 0 && idx < ALT_STEPS.length - 1) {
      altThreshold = ALT_STEPS[idx + 1];
    } else if (direction < 0 && idx > 0) {
      altThreshold = ALT_STEPS[idx - 1];
    }
  }

  function requestFullScreen() {
    if (!radarContainer) return;
    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      radarContainer.requestFullscreen();
    }
  }

  onMount(() => {
    connectTraffic();
    connectSituation();
    connectStatus();
  });

  onDestroy(() => {
    disconnectTraffic();
    disconnectSituation();
    disconnectStatus();
  });
</script>

<div class="radar-page">
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">Traffic Radar</span>
      {#if $trafficStore._connected}
        <span class="badge success">Connected</span>
      {:else}
        <span class="badge danger">Disconnected</span>
      {/if}
      {#if baroAltValid === 'Baro'}
        <span class="badge success">Baro-Alt</span>
      {:else if baroAltValid === 'GPS'}
        <span class="badge warning">GPS-Alt</span>
      {:else if baroAltValid === 'AdsbEstimated'}
        <span class="badge warning">ADSB-Est</span>
      {:else}
        <span class="badge danger">No Alt</span>
      {/if}
      {#if !gpsValid}
        <span class="badge danger">No GPS</span>
      {/if}
      <span class="badge {cpuTempClass}">{cpuTemp}°C</span>
    </div>

    <div class="panel-body">
      <div class="radar-container" bind:this={radarContainer}>
        <svg viewBox="-210 -210 420 420" class="radar-svg">
          <!-- Background -->
          <rect x="-210" y="-210" width="420" height="420" rx="5" class="radar-bg" />

          <!-- Rotating group (rotated by GPS course so north is up relative to heading) -->
          <g transform="rotate({-gpsCourse}, 0, 0)">
            <!-- Range circles -->
            <circle cx="0" cy="0" r="{VB}" class="range-circle" />
            <circle cx="0" cy="0" r="{VB / 2}" class="range-circle" />

            <!-- Cardinal direction labels -->
            <text x="0" y="-190" class="dir-label" transform="rotate({gpsCourse}, 0, -190)">N</text>
            <text x="0" y="193" class="dir-label" transform="rotate({gpsCourse}, 0, 193)">S</text>
            <text x="-192" y="4" class="dir-label" transform="rotate({gpsCourse}, -192, 4)">W</text>
            <text x="192" y="4" class="dir-label" transform="rotate({gpsCourse}, 192, 4)">E</text>

            <!-- Valid position targets -->
            {#each radarTargets.filter(t => t.positionValid) as target (target.key)}
              <g transform="translate({target.x}, {target.y})">
                <!-- Aircraft icon (rotated to heading, counter-rotated for display) -->
                <g transform="rotate({target.heading}, 0, 0)">
                  <polygon points="0,-10 -4,6 0,3 4,6" class="aircraft-icon" />
                </g>
                <!-- Labels (counter-rotate so text reads normally) -->
                <g transform="rotate({gpsCourse}, 0, 0)">
                  <text x="12" y="-4" class="target-alt-outline">{target.altLabel}</text>
                  <text x="12" y="-4" class="target-alt">{target.altLabel}</text>
                  {#if target.speed !== null}
                    <text x="12" y="6" class="target-speed">{target.speed}kts</text>
                  {/if}
                  {#if target.tail}
                    <text x="12" y="16" class="target-tail">{target.tail}</text>
                  {/if}
                </g>
              </g>
            {/each}

            <!-- Invalid position targets (range circles) -->
            {#each radarTargets.filter(t => !t.positionValid) as target (target.key)}
              <circle cx="0" cy="0" r="{target.circleR}" class="estimated-circle" />
              <g transform="rotate({gpsCourse}, 0, 0)">
                <text x="0" y="{-target.circleR - 5}" class="target-alt" text-anchor="middle">{target.altLabel}</text>
                {#if target.tail}
                  <text x="0" y="{-target.circleR - 15}" class="target-tail" text-anchor="middle">{target.tail}</text>
                {/if}
              </g>
            {/each}
          </g>

          <!-- Ownship icon (always pointing up / fixed position) -->
          <polygon points="0,-12 -5,8 0,4 5,8" class="ownship-icon" />
          <circle cx="0" cy="0" r="2" class="ownship-center" />

          <!-- Flight level text -->
          <text x="7" y="15" class="fl-text">{flightLevel}</text>

          <!-- Controls: Range -->
          <g class="control-btn" on:click={() => cycleRadius(-1)} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && cycleRadius(-1)}>
            <circle cx="-145" cy="-185" r="18" class="btn-circle" />
            <text x="-145" y="-181" class="btn-text">Ra-</text>
          </g>
          <g class="control-btn" on:click={() => cycleRadius(1)} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && cycleRadius(1)}>
            <circle cx="-185" cy="-185" r="18" class="btn-circle" />
            <text x="-185" y="-181" class="btn-text">Ra+</text>
          </g>

          <!-- Controls: Altitude -->
          <g class="control-btn" on:click={() => cycleAlt(1)} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && cycleAlt(1)}>
            <circle cx="145" cy="-185" r="18" class="btn-circle" />
            <text x="145" y="-181" class="btn-text">Alt+</text>
          </g>
          <g class="control-btn" on:click={() => cycleAlt(-1)} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && cycleAlt(-1)}>
            <circle cx="185" cy="-185" r="18" class="btn-circle" />
            <text x="185" y="-181" class="btn-text">Alt-</text>
          </g>

          <!-- Controls: Fullscreen -->
          <g class="control-btn" on:click={requestFullScreen} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && requestFullScreen()}>
            <rect x="168" y="-140" width="34" height="28" rx="6" class="btn-circle" />
            <text x="185" y="-122" class="btn-text">F/S</text>
          </g>

          <!-- Range / Alt display text -->
          <text x="-200" y="-158" class="info-text left">{displayRadius} nm</text>
          <text x="200" y="-158" class="info-text right">\u00B1{altThreshold}00ft</text>
        </svg>
      </div>
    </div>
  </div>
</div>

<style>
  .radar-page {
    padding: 1rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .panel {
    background: var(--bg-primary, #fff);
    border: 1px solid var(--border-color, #ddd);
    border-radius: 8px;
    margin-bottom: 1rem;
  }

  .panel-heading {
    padding: 0.75rem 1rem;
    background: var(--bg-secondary, #f5f5f5);
    border-bottom: 1px solid var(--border-color, #ddd);
    border-radius: 8px 8px 0 0;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .panel-label {
    font-weight: 600;
    font-size: 1.1rem;
  }

  .panel-body {
    padding: 1rem;
  }

  .badge {
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
    font-weight: 500;
  }

  .badge.success { background: #28a745; color: white; }
  .badge.danger { background: #dc3545; color: white; }
  .badge.warning { background: #ffc107; color: #333; }
  .temp-ok { background: #17a2b8; color: white; }
  .temp-warn { background: #ffc107; color: #333; }
  .temp-danger { background: #dc3545; color: white; }

  .radar-container {
    width: 100%;
    max-width: 800px;
    margin: 0 auto;
    aspect-ratio: 1;
    background: #111;
    border-radius: 8px;
    overflow: hidden;
  }

  .radar-container:fullscreen {
    display: flex;
    align-items: center;
    justify-content: center;
    background: #000;
  }

  .radar-svg {
    width: 100%;
    height: 100%;
    display: block;
  }

  /* SVG styles */
  .radar-bg {
    fill: #111;
  }

  .range-circle {
    fill: none;
    stroke: #2a4a2a;
    stroke-width: 1;
  }

  .dir-label {
    fill: #6a6a6a;
    font-size: 14px;
    font-weight: 600;
    text-anchor: middle;
    dominant-baseline: central;
  }

  .ownship-icon {
    fill: #00cc00;
    stroke: #00ff00;
    stroke-width: 0.5;
  }

  .ownship-center {
    fill: #00ff00;
  }

  .aircraft-icon {
    fill: #ffcc00;
    stroke: #ffcc00;
    stroke-width: 0.5;
  }

  .estimated-circle {
    fill: none;
    stroke: #00cc00;
    stroke-width: 1;
    stroke-dasharray: 4 2;
  }

  .target-alt {
    fill: #00ff00;
    font-size: 10px;
    font-weight: 600;
  }

  .target-alt-outline {
    fill: none;
    stroke: #000;
    stroke-width: 2;
    font-size: 10px;
    font-weight: 600;
    paint-order: stroke;
  }

  .target-speed {
    fill: #aaa;
    font-size: 8px;
  }

  .target-tail {
    fill: #8888ff;
    font-size: 8px;
  }

  .fl-text {
    fill: #aaa;
    font-size: 10px;
  }

  .info-text {
    fill: #aaa;
    font-size: 10px;
    font-weight: 500;
  }

  .info-text.left {
    text-anchor: start;
  }

  .info-text.right {
    text-anchor: end;
  }

  .control-btn {
    cursor: pointer;
  }

  .control-btn:hover .btn-circle {
    fill: #444;
  }

  .btn-circle {
    fill: #333;
    stroke: #555;
    stroke-width: 1;
  }

  .btn-text {
    fill: #ccc;
    font-size: 9px;
    font-weight: 500;
    text-anchor: middle;
    dominant-baseline: central;
  }

  /* Responsive */
  @media (max-width: 600px) {
    .radar-container {
      max-width: 100%;
    }
  }
</style>
