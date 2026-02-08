<script>
  import { onMount, onDestroy } from 'svelte';
  import {
    status,
    settings,
    uptime,
    cpuTemp,
    diskSpace,
    gpsHardware,
    gpsProtocol,
    ognNoiseColor,
    ognRangeLossFactor,
    trafficSourceColors,
    connectStatus,
    disconnectStatus,
    fetchSettings
  } from '../stores/status.js';

  let uatTowers = 0;
  let towerInterval;

  // Fetch tower count periodically
  async function fetchTowers() {
    if (!$settings.UAT_Enabled) return;
    try {
      const response = await fetch('/getTowers');
      const towers = await response.json();
      let count = 0;
      for (const key in towers) {
        if (towers[key].Messages_last_minute > 0) count++;
      }
      uatTowers = count;
    } catch (e) {
      console.error('Failed to fetch towers:', e);
    }
  }

  // GPS position accuracy display
  $: gpsAccuracy = ['Disconnected', 'No Fix', 'Unknown'].includes($status.GPS_solution)
    ? ''
    : `, ${($status.GPS_position_accuracy || 0).toFixed(1)} m`;

  // Clamp helper for progress bars
  function clamp(num, min, max) {
    return Math.min(Math.max(num, min), max);
  }

  // Calculate bar width percentage
  function barWidth(current, max) {
    return max ? (100 * current / max) : 0;
  }

  onMount(() => {
    connectStatus();
    fetchSettings();
    fetchTowers();
    towerInterval = setInterval(fetchTowers, 5000);
  });

  onDestroy(() => {
    disconnectStatus();
    if (towerInterval) clearInterval(towerInterval);
  });
</script>

<div class="status-page">
  <!-- Version Header -->
  <div class="version-header">
    <strong>Version: <code>{$status.Version} ({$status.Build?.substring(0, 10)})</code></strong>
  </div>

  <!-- Main Status Panel -->
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">Status</span>
      {#if $status._connected}
        <span class="badge success">Connected</span>
      {:else}
        <span class="badge danger">Disconnected</span>
      {/if}
      {#if $settings.DeveloperMode}
        <span class="badge warning">Developer Mode</span>
      {/if}
    </div>

    <div class="panel-body">
      <!-- Basic Info -->
      <div class="row">
        <div class="col-6">
          <strong>Recent Clients:</strong>
          <span>{$status.Connected_Users}</span>
        </div>
        {#if $settings.Pong_Enabled && $status.Pong_connected}
          <div class="col-6">
            <strong>Pong Heartbeats:</strong>
            <span>{$status.Pong_Heartbeats}</span>
          </div>
        {/if}
      </div>

      <div class="row">
        <div class="col-6">
          <strong>SDR devices:</strong>
          <span>{$status.Devices}</span>
        </div>
        {#if $settings.Ping_Enabled}
          <div class="col-6">
            <strong>Ping device:</strong>
            {#if $status.Ping_connected}
              <span class="badge success">Connected</span>
            {:else}
              <span class="badge danger">Disconnected</span>
            {/if}
          </div>
        {/if}
        {#if $settings.Pong_Enabled}
          <div class="col-6">
            <strong>Pong device:</strong>
            {#if $status.Pong_connected}
              <span class="badge success">Connected</span>
            {:else}
              <span class="badge danger">Disconnected</span>
            {/if}
          </div>
        {/if}
      </div>

      <hr />

      <!-- Message Statistics -->
      <div class="row header-row">
        <label class="col-4">Messages</label>
        <label class="col-6">Current</label>
        <label class="col-2 text-right">Peak</label>
      </div>

      {#if $settings.UAT_Enabled}
        <div class="row">
          <label class="col-4">UAT:</label>
          <div class="col-6">
            <div class="bar-container">
              <div class="bar-fill" style="background-color: {trafficSourceColors.UAT}; width: {barWidth($status.UAT_messages_last_minute, $status.UAT_messages_max)}%"></div>
              <div class="bar-text">{$status.UAT_messages_last_minute}</div>
            </div>
          </div>
          <span class="col-2 text-right">{$status.UAT_messages_max}</span>
        </div>
      {/if}

      {#if $settings.ES_Enabled}
        <div class="row">
          <label class="col-4">1090ES:</label>
          <div class="col-6">
            <div class="bar-container">
              <div class="bar-fill" style="background-color: {trafficSourceColors.ES}; width: {barWidth($status.ES_messages_last_minute, $status.ES_messages_max)}%"></div>
              <div class="bar-text">{$status.ES_messages_last_minute}</div>
            </div>
          </div>
          <span class="col-2 text-right">{$status.ES_messages_max}</span>
        </div>
      {/if}

      {#if $settings.AIS_Enabled}
        <div class="row">
          <label class="col-4">AIS{$status.AIS_connected ? '' : ' (disconnected)'}:</label>
          <div class="col-6">
            <div class="bar-container">
              <div class="bar-fill" style="background-color: {trafficSourceColors.AIS}; width: {barWidth($status.AIS_messages_last_minute, $status.AIS_messages_max)}%"></div>
              <div class="bar-text">{$status.AIS_messages_last_minute}</div>
            </div>
          </div>
          <span class="col-2 text-right">{$status.AIS_messages_max}</span>
        </div>
      {/if}

      {#if $settings.OGN_Enabled}
        <div class="row">
          <label class="col-4">OGN{$status.OGN_connected ? '' : ' (disconnected)'}:</label>
          <div class="col-6">
            <div class="bar-container">
              <div class="bar-fill" style="background-color: {trafficSourceColors.OGN}; width: {barWidth($status.OGN_messages_last_minute, $status.OGN_messages_max)}%"></div>
              <div class="bar-text">{$status.OGN_messages_last_minute}</div>
            </div>
          </div>
          <span class="col-2 text-right">{$status.OGN_messages_max}</span>
        </div>

        <div class="row">
          <label class="col-4">OGN Noise@Gain 50:<br /><small>Half range for each 6 dB noise</small></label>
          <div class="col-6">
            <div class="bar-container">
              <div class="bar-fill" style="background-color: {$ognNoiseColor}; width: {clamp($status.OGN_noise_db, 0, 25) / 25 * 100}%"></div>
              <div class="bar-text">{$status.OGN_noise_db} dB, range loss: {$ognRangeLossFactor}x</div>
            </div>
          </div>
          <span class="col-2 text-right">25 dB</span>
        </div>
      {/if}

      <!-- OGN Statistics -->
      {#if $settings.OGN_Enabled}
        <div class="section-header">OGN Statistics</div>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-label">Background Noise@Gain 50</div>
            <div class="stat-value">{$status.OGN_noise_db} dB</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">Receiver Gain</div>
            <div class="stat-value">{$status.OGN_gain_db} dB</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">RF Spectrogram</div>
            <div class="stat-value"><a href="http://{window.location.hostname}:8082/rf-spectro.jpg">Download</a></div>
          </div>
        </div>
      {/if}

      <!-- UAT Statistics -->
      {#if $settings.UAT_Enabled}
        <div class="section-header">UAT Statistics</div>
        <div class="stats-grid four-col">
          <div class="stat-item">
            <div class="stat-label">Towers</div>
            <div class="stat-value">{uatTowers}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">METARS</div>
            <div class="stat-value">{$status.UAT_METAR_total}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">TAFS</div>
            <div class="stat-value">{$status.UAT_TAF_total}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">NEXRAD</div>
            <div class="stat-value">{$status.UAT_NEXRAD_total}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">PIREP</div>
            <div class="stat-value">{$status.UAT_PIREP_total}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">SIGMET</div>
            <div class="stat-value">{$status.UAT_SIGMET_total}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">NOTAMS</div>
            <div class="stat-value">{$status.UAT_NOTAM_total}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">Other</div>
            <div class="stat-value">{$status.UAT_OTHER_total}</div>
          </div>
        </div>
      {/if}

      <hr />

      <!-- GPS Info -->
      {#if $settings.GPS_Enabled}
        <div class="row">
          <label class="col-6">GPS hardware:</label>
          <span class="col-6">
            {$gpsHardware}
            {#if $gpsHardware === 'Network'}
              <a href="http://{$status.GPS_NetworkRemoteIp}">{$status.GPS_NetworkRemoteIp}</a>
            {/if}
            ({$gpsProtocol})
          </span>
        </div>
        <div class="row">
          <label class="col-6">GPS solution:</label>
          <span class="col-6">{$status.GPS_solution}{gpsAccuracy}</span>
        </div>
        <div class="row">
          <label class="col-6">GPS satellites:</label>
          <span class="col-6">{$status.GPS_satellites_locked} in solution; {$status.GPS_satellites_seen} seen; {$status.GPS_satellites_tracked} tracked</span>
        </div>
        <hr />
      {/if}

      <!-- System Info -->
      <div class="row system-info">
        <div class="col-4">
          <strong>Uptime:</strong>
          <span>{$uptime}</span>
        </div>
        <div class="col-4">
          <strong>CPU Temp:</strong>
          <span>{$cpuTemp.celsius}°C / {$cpuTemp.fahrenheit}°F</span>
          {#if $settings.DeveloperMode}
            <span class="temp-range">[{$cpuTemp.min} - {$cpuTemp.max}]</span>
          {/if}
        </div>
        <div class="col-4">
          <strong>Free Storage:</strong>
          <span>{$diskSpace} MiB</span>
        </div>
      </div>

      <hr />

      <!-- Total Frames -->
      <div class="row">
        {#if $settings.UAT_Enabled}
          <div class="col-3">
            <strong>UAT Frames</strong>
            <span>{$status.UAT_messages_total}</span>
          </div>
        {/if}
        {#if $settings.ES_Enabled}
          <div class="col-3">
            <strong>ES Frames</strong>
            <span>{$status.ES_messages_total}</span>
          </div>
        {/if}
        {#if $settings.AIS_Enabled}
          <div class="col-3">
            <strong>AIS Frames</strong>
            <span>{$status.AIS_messages_total}</span>
          </div>
        {/if}
        {#if $settings.OGN_Enabled}
          <div class="col-3">
            <strong>OGN Frames</strong>
            <span>{$status.OGN_messages_total}</span>
          </div>
        {/if}
      </div>
    </div>
  </div>

  <!-- Errors Panel -->
  {#if $status.Errors && $status.Errors.length > 0}
    <div class="panel error-panel">
      <div class="panel-heading">
        <span class="panel-label">Errors</span>
      </div>
      <div class="panel-body">
        <ul class="error-list">
          {#each $status.Errors as error}
            <li class="error-item">
              <span class="error-icon">⚠</span>
              <span class="error-text">{error}</span>
            </li>
          {/each}
        </ul>
      </div>
    </div>
  {/if}
</div>

<style>
  .status-page {
    padding: 1rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .version-header {
    text-align: center;
    margin-bottom: 1rem;
  }

  .version-header code {
    background: var(--bg-secondary, #f5f5f5);
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
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

  .badge.success {
    background: #28a745;
    color: white;
  }

  .badge.danger {
    background: #dc3545;
    color: white;
  }

  .badge.warning {
    background: #ffc107;
    color: #333;
  }

  .row {
    display: flex;
    flex-wrap: wrap;
    margin-bottom: 0.5rem;
    align-items: center;
  }

  .header-row {
    font-weight: 600;
    border-bottom: 1px solid var(--border-color, #ddd);
    padding-bottom: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .col-2 { flex: 0 0 16.66%; }
  .col-3 { flex: 0 0 25%; }
  .col-4 { flex: 0 0 33.33%; }
  .col-6 { flex: 0 0 50%; }

  .text-right {
    text-align: right;
  }

  hr {
    border: none;
    border-top: 1px solid var(--border-color, #ddd);
    margin: 1rem 0;
  }

  .bar-container {
    position: relative;
    height: 24px;
    background: var(--bg-secondary, #f0f0f0);
    border-radius: 4px;
    overflow: hidden;
  }

  .bar-fill {
    position: absolute;
    top: 0;
    left: 0;
    height: 100%;
    transition: width 0.3s ease;
  }

  .bar-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 0.85rem;
    font-weight: 500;
    color: #333;
    text-shadow: 0 0 2px white;
  }

  .section-header {
    font-weight: 600;
    margin: 1rem 0 0.5rem;
    padding-top: 0.5rem;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0.5rem;
  }

  .stats-grid.four-col {
    grid-template-columns: repeat(4, 1fr);
  }

  .stat-item {
    text-align: center;
  }

  .stat-label {
    font-size: 0.8rem;
    color: var(--text-secondary, #666);
  }

  .stat-value {
    font-weight: 500;
  }

  .system-info .col-4 {
    display: flex;
    flex-direction: column;
  }

  .temp-range {
    font-size: 0.8rem;
    color: var(--text-secondary, #666);
  }

  .error-panel .panel-heading {
    background: #f8d7da;
    border-color: #f5c6cb;
  }

  .error-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .error-item {
    padding: 0.5rem 0;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .error-icon {
    color: #dc3545;
  }

  .error-text {
    color: #721c24;
  }

  /* Responsive adjustments */
  @media (max-width: 768px) {
    .col-4, .col-6 {
      flex: 0 0 100%;
    }

    .stats-grid,
    .stats-grid.four-col {
      grid-template-columns: repeat(2, 1fr);
    }

    .system-info .col-4 {
      flex: 0 0 100%;
      margin-bottom: 0.5rem;
    }
  }
</style>
