<script>
  import { onMount, onDestroy } from 'svelte';
  import {
    status,
    settings,
    uptime,
    cpuTemp,
    connectStatus,
    disconnectStatus,
    fetchSettings
  } from '../stores/status.js';
  import { api } from '../services/websocket.js';

  let restartConfirm = false;
  let deleteLogConfirm = false;
  let deleteAhrsConfirm = false;

  function humanFileSize(size) {
    if (!size || size === 0) return '0 B';
    const i = Math.floor(Math.log(size) / Math.log(1024));
    return (size / Math.pow(1024, i)).toFixed(2) * 1 + ' ' + ['B', 'kB', 'MB', 'GB', 'TB'][i];
  }

  $: logfileSize = humanFileSize($status.Logfile_Size);
  $: ahrsLogSize = humanFileSize($status.AHRS_LogFiles_Size);

  async function postRestart() {
    if (!restartConfirm) {
      restartConfirm = true;
      return;
    }
    try {
      await api.post('/restart');
    } catch (e) {
      // Expected - connection will drop during restart
    }
    restartConfirm = false;
  }

  function cancelRestart() {
    restartConfirm = false;
  }

  async function deleteLogfile() {
    if (!deleteLogConfirm) {
      deleteLogConfirm = true;
      return;
    }
    try {
      await api.post('/deletelogfile');
    } catch (e) {
      console.error('Failed to delete logfile:', e);
    }
    deleteLogConfirm = false;
  }

  function cancelDeleteLog() {
    deleteLogConfirm = false;
  }

  async function deleteAhrsLogs() {
    if (!deleteAhrsConfirm) {
      deleteAhrsConfirm = true;
      return;
    }
    try {
      await api.post('/deleteahrslogfiles');
    } catch (e) {
      console.error('Failed to delete AHRS logs:', e);
    }
    deleteAhrsConfirm = false;
  }

  function cancelDeleteAhrs() {
    deleteAhrsConfirm = false;
  }

  function refreshUI() {
    window.location.reload();
  }

  onMount(() => {
    connectStatus();
    fetchSettings();
  });

  onDestroy(() => {
    disconnectStatus();
  });
</script>

<div class="developer-page">
  <!-- Status Overview -->
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">System Status</span>
      {#if $status._connected}
        <span class="badge success">Connected</span>
      {:else}
        <span class="badge danger">Disconnected</span>
      {/if}
    </div>
    <div class="panel-body">
      <div class="status-grid">
        <div class="status-item">
          <span class="stat-label">Version</span>
          <span class="stat-value">{$status.Version} ({($status.Build || '').substring(0, 10)})</span>
        </div>
        <div class="status-item">
          <span class="stat-label">Uptime</span>
          <span class="stat-value">{$uptime}</span>
        </div>
        <div class="status-item">
          <span class="stat-label">CPU Temp</span>
          <span class="stat-value">{$cpuTemp.celsius}C / {$cpuTemp.fahrenheit}F</span>
        </div>
        <div class="status-item">
          <span class="stat-label">SDR Devices</span>
          <span class="stat-value">{$status.Devices}</span>
        </div>
        <div class="status-item">
          <span class="stat-label">Connected Users</span>
          <span class="stat-value">{$status.Connected_Users}</span>
        </div>
        <div class="status-item">
          <span class="stat-label">GPS Solution</span>
          <span class="stat-value">{$status.GPS_solution}</span>
        </div>
        <div class="status-item">
          <span class="stat-label">GPS Satellites</span>
          <span class="stat-value">{$status.GPS_satellites_locked} locked / {$status.GPS_satellites_seen} seen / {$status.GPS_satellites_tracked} tracked</span>
        </div>
      </div>

      {#if $settings.UAT_Enabled}
        <hr />
        <div class="section-label">UAT Weather Totals</div>
        <div class="status-grid compact">
          <div class="status-item"><span class="stat-label">METAR</span><span class="stat-value">{$status.UAT_METAR_total}</span></div>
          <div class="status-item"><span class="stat-label">TAF</span><span class="stat-value">{$status.UAT_TAF_total}</span></div>
          <div class="status-item"><span class="stat-label">NEXRAD</span><span class="stat-value">{$status.UAT_NEXRAD_total}</span></div>
          <div class="status-item"><span class="stat-label">SIGMET</span><span class="stat-value">{$status.UAT_SIGMET_total}</span></div>
          <div class="status-item"><span class="stat-label">PIREP</span><span class="stat-value">{$status.UAT_PIREP_total}</span></div>
          <div class="status-item"><span class="stat-label">NOTAM</span><span class="stat-value">{$status.UAT_NOTAM_total}</span></div>
          <div class="status-item"><span class="stat-label">Other</span><span class="stat-value">{$status.UAT_OTHER_total}</span></div>
        </div>
      {/if}

      {#if $status.Errors && $status.Errors.length > 0}
        <hr />
        <div class="section-label">Errors</div>
        <ul class="error-list">
          {#each $status.Errors as error}
            <li>{error}</li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>

  <div class="panel-grid">
    <!-- Restart Application -->
    <div class="panel">
      <div class="panel-heading">
        <span class="panel-label">Application Control</span>
      </div>
      <div class="panel-body action-panel">
        {#if restartConfirm}
          <p class="confirm-text">Are you sure you want to restart Stratux?</p>
          <div class="btn-group">
            <button class="btn btn-danger" on:click={postRestart}>Confirm Restart</button>
            <button class="btn btn-secondary" on:click={cancelRestart}>Cancel</button>
          </div>
        {:else}
          <button class="btn btn-primary btn-block" on:click={postRestart}>Restart Application</button>
        {/if}
        <button class="btn btn-secondary btn-block" on:click={refreshUI}>Refresh Web UI</button>
      </div>
    </div>

    <!-- Logfile Management -->
    <div class="panel">
      <div class="panel-heading">
        <span class="panel-label">Logfile</span>
        <span class="size-badge">{logfileSize}</span>
      </div>
      <div class="panel-body action-panel">
        <a href="/downloadlog" class="btn btn-primary btn-block" download>Download Logfile</a>
        {#if deleteLogConfirm}
          <p class="confirm-text">Delete the logfile?</p>
          <div class="btn-group">
            <button class="btn btn-danger" on:click={deleteLogfile}>Confirm Delete</button>
            <button class="btn btn-secondary" on:click={cancelDeleteLog}>Cancel</button>
          </div>
        {:else}
          <button class="btn btn-warning btn-block" on:click={deleteLogfile}>Delete Logfile</button>
        {/if}
      </div>
    </div>

    <!-- AHRS Logs -->
    <div class="panel">
      <div class="panel-heading">
        <span class="panel-label">AHRS Log Files</span>
        <span class="size-badge">{ahrsLogSize}</span>
      </div>
      <div class="panel-body action-panel">
        <a href="/downloadahrslogs" class="btn btn-primary btn-block" download="ahrs_logs.zip">Download AHRS Logs</a>
        {#if deleteAhrsConfirm}
          <p class="confirm-text">Delete all AHRS log files?</p>
          <div class="btn-group">
            <button class="btn btn-danger" on:click={deleteAhrsLogs}>Confirm Delete</button>
            <button class="btn btn-secondary" on:click={cancelDeleteAhrs}>Cancel</button>
          </div>
        {:else}
          <button class="btn btn-warning btn-block" on:click={deleteAhrsLogs}>Delete AHRS Logs</button>
        {/if}
      </div>
    </div>

    <!-- Database -->
    <div class="panel">
      <div class="panel-heading">
        <span class="panel-label">Database</span>
      </div>
      <div class="panel-body action-panel">
        <a href="/downloaddb" class="btn btn-primary btn-block" download>Download Database</a>
      </div>
    </div>
  </div>
</div>

<style>
  .developer-page {
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

  .panel-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .badge {
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
    font-weight: 500;
  }

  .badge.success { background: #28a745; color: white; }
  .badge.danger { background: #dc3545; color: white; }

  .size-badge {
    font-size: 0.85rem;
    color: var(--text-secondary, #666);
    font-weight: 400;
  }

  hr {
    border: none;
    border-top: 1px solid var(--border-color, #ddd);
    margin: 1rem 0;
  }

  .status-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 0.75rem;
  }

  .status-grid.compact {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 0.5rem;
  }

  .status-item {
    display: flex;
    flex-direction: column;
  }

  .stat-label {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-secondary, #666);
  }

  .stat-value {
    font-size: 0.95rem;
  }

  .section-label {
    font-weight: 600;
    margin-bottom: 0.5rem;
  }

  .error-list {
    margin: 0;
    padding-left: 1.5rem;
    color: #dc3545;
  }

  .error-list li {
    margin-bottom: 0.25rem;
  }

  .action-panel {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .btn {
    display: block;
    width: 100%;
    padding: 0.6rem 1rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
    font-size: 0.95rem;
    text-align: center;
    text-decoration: none;
  }

  .btn-primary {
    background: #1976d2;
    color: white;
  }

  .btn-primary:hover {
    background: #1565c0;
  }

  .btn-secondary {
    background: #6c757d;
    color: white;
  }

  .btn-secondary:hover {
    background: #5a6268;
  }

  .btn-warning {
    background: #ffc107;
    color: #333;
  }

  .btn-warning:hover {
    background: #e0a800;
  }

  .btn-danger {
    background: #dc3545;
    color: white;
  }

  .btn-danger:hover {
    background: #c82333;
  }

  .btn-block {
    width: 100%;
  }

  .btn-group {
    display: flex;
    gap: 0.5rem;
  }

  .btn-group .btn {
    flex: 1;
  }

  .confirm-text {
    margin: 0;
    font-weight: 500;
    color: #dc3545;
    text-align: center;
  }

  /* Responsive */
  @media (max-width: 768px) {
    .panel-grid {
      grid-template-columns: 1fr;
    }

    .status-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }
</style>
