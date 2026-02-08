<script>
  import { onMount, onDestroy } from 'svelte';
  import {
    situation,
    connectSituation,
    disconnectSituation,
    gpsSolutionText,
    gpsPosition,
    gpsAltSpeed,
    ahrsData,
    gpsSatelliteSummary,
    ahrsStatus,
    cageAHRS,
    calibrateAHRS,
    resetGMeter,
    fetchSatellites
  } from '../stores/situation.js';

  let satellites = [];
  let satelliteInterval;
  let isCaging = false;
  let isCalibrating = false;

  // Fetch satellites periodically
  async function updateSatellites() {
    const data = await fetchSatellites();
    if (Array.isArray(data)) {
      satellites = data.sort((a, b) => (a.SatelliteNMEA || 0) - (b.SatelliteNMEA || 0));
    }
  }

  async function handleCage() {
    isCaging = true;
    await cageAHRS();
    setTimeout(() => { isCaging = false; }, 2000);
  }

  async function handleCalibrate() {
    isCalibrating = true;
    await calibrateAHRS();
    setTimeout(() => { isCalibrating = false; }, 5000);
  }

  async function handleResetG() {
    await resetGMeter();
  }

  onMount(() => {
    connectSituation();
    updateSatellites();
    satelliteInterval = setInterval(updateSatellites, 2000);
  });

  onDestroy(() => {
    disconnectSituation();
    if (satelliteInterval) clearInterval(satelliteInterval);
  });
</script>

<div class="gps-page">
  <div class="panel-grid">
    <!-- AHRS Panel -->
    <div class="panel">
      <div class="panel-heading">
        <span class="panel-label">AHRS</span>
        {#if $situation._connected}
          <span class="badge success">Connected</span>
        {:else}
          <span class="badge danger">Disconnected</span>
        {/if}
      </div>
      <div class="panel-body">
        <!-- Status Indicators -->
        <div class="indicator-row">
          <div class="indicator" class:on={!$ahrsStatus.isCalibrating}>
            {$ahrsStatus.isCalibrating ? 'Calibrating' : 'Ready'}
          </div>
          <div class="indicator" class:on={$ahrsStatus.hasGPS}>GPS</div>
          <div class="indicator" class:on={$ahrsStatus.hasIMU}>ATT</div>
          <div class="indicator" class:on={$ahrsStatus.hasBaro}>ALT</div>
        </div>

        <hr />

        <!-- Control Buttons -->
        <div class="control-row">
          <div class="buttons">
            <button
              class="btn btn-primary"
              on:click={handleCage}
              disabled={isCaging || !$ahrsStatus.hasIMU}
            >
              Set Level
            </button>
            <button
              class="btn btn-primary"
              on:click={handleCalibrate}
              disabled={isCalibrating || !$ahrsStatus.hasIMU}
            >
              Zero Drift
            </button>
          </div>

          <!-- AHRS Values -->
          <div class="ahrs-values">
            <div class="value-row">
              <div class="value-item">
                <span class="label">Heading</span>
                <span class="value">{$ahrsData.heading}&deg;</span>
              </div>
              <div class="value-item">
                <span class="label">Pitch</span>
                <span class="value">{$ahrsData.pitch}&deg;</span>
              </div>
              <div class="value-item">
                <span class="label">Roll</span>
                <span class="value">{$ahrsData.roll}&deg;</span>
              </div>
              <div class="value-item">
                <span class="label">P-Alt</span>
                <span class="value">{$ahrsData.pressureAlt}'</span>
              </div>
            </div>
            <div class="value-row">
              <div class="value-item">
                <span class="label">Mag Hdg</span>
                <span class="value">{$ahrsData.headingMag}&deg;</span>
              </div>
              <div class="value-item">
                <span class="label">Slip/Skid</span>
                <span class="value">{$ahrsData.slipSkid}&deg;</span>
              </div>
              <div class="value-item">
                <span class="label">Turn Rate</span>
                <span class="value">{$ahrsData.turnRate} min</span>
              </div>
              <div class="value-item">
                <span class="label">G Load</span>
                <span class="value">{$ahrsData.gLoad}G</span>
              </div>
            </div>
          </div>
        </div>

        <!-- G-Meter Summary -->
        <div class="g-meter-summary">
          <span>G Range: {$ahrsData.gLoadMin}G to {$ahrsData.gLoadMax}G</span>
          <button class="btn btn-small" on:click={handleResetG}>Reset G</button>
        </div>
      </div>
    </div>

    <!-- GPS Position Panel -->
    <div class="panel">
      <div class="panel-heading">
        <span class="panel-label">GPS Position</span>
      </div>
      <div class="panel-body">
        <div class="gps-data">
          <div class="data-section">
            <h4>Location</h4>
            <p class="coords">{$gpsPosition.lat}, {$gpsPosition.lng}</p>
            <p class="accuracy">&plusmn; {$gpsAltSpeed.hAccuracy} m</p>
          </div>

          <div class="data-section">
            <h4>Altitude MSL</h4>
            <p class="value-large">{$gpsAltSpeed.altMSL}'</p>
            <p class="accuracy">&plusmn; {$gpsAltSpeed.vAccuracy} ft @ {$gpsAltSpeed.vertSpeed} ft/min</p>
          </div>

          <div class="data-section">
            <h4>Height WGS-84</h4>
            <p class="value-large">{$gpsAltSpeed.altEllipsoid}'</p>
          </div>

          <div class="data-section">
            <h4>Track</h4>
            <p class="value-large">{$gpsAltSpeed.track}&deg; @ {$gpsAltSpeed.groundSpeed} KTS</p>
          </div>
        </div>
      </div>
    </div>
  </div>

  <div class="panel-grid">
    <!-- Satellites Panel -->
    <div class="panel">
      <div class="panel-heading">
        <span class="panel-label">Satellites</span>
      </div>
      <div class="panel-body">
        <!-- Satellite Header -->
        <div class="sat-header">
          <span class="col-name">Satellite</span>
          <span class="col-elev">Elevation</span>
          <span class="col-az">Azimuth</span>
          <span class="col-signal">Signal</span>
        </div>

        <!-- Satellite List -->
        <div class="sat-list">
          {#each satellites as sat}
            <div class="sat-row">
              <span class="col-name">
                {sat.SatelliteID}
                {#if sat.InSolution}
                  <span class="in-solution">&#x2705;</span>
                {/if}
              </span>
              <span class="col-elev">
                {sat.Elevation < -5 ? '---' : sat.Elevation}&deg;
              </span>
              <span class="col-az">
                {sat.Azimuth < 0 ? '---' : sat.Azimuth}&deg;
              </span>
              <span class="col-signal">
                {sat.Signal < 1 ? '---' : sat.Signal}<span class="unit">dB-Hz</span>
              </span>
            </div>
          {:else}
            <div class="empty-message">No satellites detected</div>
          {/each}
        </div>
      </div>
      <div class="panel-footer">
        <div class="sat-summary">
          <span><strong>Solution:</strong> {$gpsSolutionText} @ {$gpsSatelliteSummary.sampleRate} Hz</span>
          <span><strong>Summary:</strong> {$gpsSatelliteSummary.inSolution} in solution; {$gpsSatelliteSummary.seen} seen; {$gpsSatelliteSummary.tracked} tracked</span>
        </div>
      </div>
    </div>

    <!-- Placeholder for future AHRS/G-Meter visualizations -->
    <div class="panel">
      <div class="panel-heading">
        <span class="panel-label">Attitude Display</span>
      </div>
      <div class="panel-body visualization-placeholder">
        <p>AHRS artificial horizon visualization</p>
        <p class="note">Coming in a future update</p>
        <div class="attitude-preview">
          <div class="horizon-line" style="transform: rotate({$ahrsData.roll}deg)"></div>
          <div class="pitch-indicator">{$ahrsData.pitch}&deg;</div>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .gps-page {
    padding: 1rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .panel-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .panel {
    background: var(--bg-primary, #fff);
    border: 1px solid var(--border-color, #ddd);
    border-radius: 8px;
  }

  .panel-heading {
    padding: 0.75rem 1rem;
    background: var(--bg-secondary, #f5f5f5);
    border-bottom: 1px solid var(--border-color, #ddd);
    border-radius: 8px 8px 0 0;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .panel-label {
    font-weight: 600;
    font-size: 1.1rem;
  }

  .panel-body {
    padding: 1rem;
  }

  .panel-footer {
    padding: 0.75rem 1rem;
    background: var(--bg-secondary, #f5f5f5);
    border-top: 1px solid var(--border-color, #ddd);
    border-radius: 0 0 8px 8px;
  }

  .badge {
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
    font-weight: 500;
  }

  .badge.success { background: #28a745; color: white; }
  .badge.danger { background: #dc3545; color: white; }

  hr {
    border: none;
    border-top: 1px solid var(--border-color, #ddd);
    margin: 1rem 0;
  }

  /* Indicator row */
  .indicator-row {
    display: flex;
    gap: 0.5rem;
  }

  .indicator {
    flex: 1;
    text-align: center;
    padding: 0.5rem;
    background: #666;
    color: white;
    border-radius: 4px;
    font-size: 0.9rem;
    font-weight: 500;
  }

  .indicator.on {
    background: #28a745;
  }

  /* Control row */
  .control-row {
    display: flex;
    gap: 1rem;
  }

  .buttons {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .btn {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
  }

  .btn-primary {
    background: #1976d2;
    color: white;
  }

  .btn-primary:disabled {
    background: #ccc;
    cursor: not-allowed;
  }

  .btn-small {
    padding: 0.25rem 0.5rem;
    font-size: 0.85rem;
    background: #6c757d;
    color: white;
  }

  /* AHRS values */
  .ahrs-values {
    flex: 1;
  }

  .value-row {
    display: flex;
    margin-bottom: 0.5rem;
  }

  .value-item {
    flex: 1;
    text-align: center;
  }

  .value-item .label {
    display: block;
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-secondary, #666);
  }

  .value-item .value {
    font-size: 1rem;
  }

  .g-meter-summary {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 1rem;
    padding-top: 0.5rem;
    border-top: 1px solid var(--border-color, #eee);
    font-size: 0.9rem;
  }

  /* GPS data */
  .gps-data {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .data-section h4 {
    margin: 0 0 0.25rem;
    font-size: 0.9rem;
    color: var(--text-secondary, #666);
  }

  .coords {
    font-size: 0.95rem;
    margin: 0;
  }

  .value-large {
    font-size: 1.2rem;
    font-weight: 500;
    margin: 0;
  }

  .accuracy {
    font-size: 0.85rem;
    color: var(--text-secondary, #666);
    margin: 0;
  }

  /* Satellite list */
  .sat-header, .sat-row {
    display: flex;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border-color, #eee);
  }

  .sat-header {
    font-weight: 600;
    background: var(--bg-secondary, #f9f9f9);
    margin: 0 -1rem;
    padding: 0.5rem 1rem;
  }

  .sat-list {
    max-height: 300px;
    overflow-y: auto;
  }

  .col-name { flex: 0 0 30%; }
  .col-elev { flex: 0 0 23%; text-align: right; }
  .col-az { flex: 0 0 23%; text-align: right; }
  .col-signal { flex: 0 0 24%; text-align: right; }

  .in-solution {
    margin-left: 0.25rem;
  }

  .unit {
    font-size: 0.65rem;
    margin-left: 2px;
  }

  .sat-summary {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.9rem;
  }

  .empty-message {
    padding: 1rem;
    text-align: center;
    color: var(--text-secondary, #666);
    font-style: italic;
  }

  /* Visualization placeholder */
  .visualization-placeholder {
    text-align: center;
    color: var(--text-secondary, #666);
  }

  .visualization-placeholder .note {
    font-size: 0.85rem;
    font-style: italic;
  }

  .attitude-preview {
    position: relative;
    width: 200px;
    height: 200px;
    margin: 1rem auto;
    background: linear-gradient(to bottom, #6ec6ff 50%, #8d6e63 50%);
    border-radius: 50%;
    border: 3px solid #333;
    overflow: hidden;
  }

  .horizon-line {
    position: absolute;
    top: 50%;
    left: 0;
    right: 0;
    height: 2px;
    background: white;
    transform-origin: center;
  }

  .pitch-indicator {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background: rgba(0,0,0,0.5);
    color: white;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
  }

  /* Responsive */
  @media (max-width: 992px) {
    .panel-grid {
      grid-template-columns: 1fr;
    }

    .control-row {
      flex-direction: column;
    }

    .buttons {
      flex-direction: row;
    }

    .gps-data {
      grid-template-columns: 1fr;
    }
  }
</style>
