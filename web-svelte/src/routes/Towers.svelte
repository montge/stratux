<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../services/websocket.js';

  let towers = [];
  let connected = false;
  let error = null;
  let pollInterval;

  /**
   * Format degrees to DMS string (full format with seconds)
   */
  function dmsString(val) {
    const deg = Math.trunc(val);
    const minRaw = Math.abs(val % 1) * 60;
    const min = Math.trunc(minRaw);
    const sec = Math.trunc((minRaw % 1) * 60);
    return `${deg}° ${min}' ${sec}"`;
  }

  /**
   * Transform tower data for display
   */
  function transformTower(obj) {
    return {
      lat: dmsString(obj.Lat),
      lng: dmsString(obj.Lng),
      power: obj.Signal_strength_now.toFixed(2),
      power_last_min: obj.Signal_strength_last_minute.toFixed(2),
      power_max: obj.Signal_strength_max.toFixed(2),
      messages: obj.Messages_last_minute
    };
  }

  /**
   * Fetch towers from API
   */
  async function fetchTowers() {
    try {
      const data = await api.get('/getTowers');
      // Filter to only towers with recent messages and transform
      towers = Object.values(data)
        .filter(t => t.Messages_last_minute > 0)
        .map(transformTower)
        .sort((a, b) => b.messages - a.messages); // Sort by messages descending
      connected = true;
      error = null;
    } catch (e) {
      connected = false;
      error = e.message;
      console.error('Failed to fetch towers:', e);
    }
  }

  onMount(() => {
    fetchTowers();
    // Poll every 2 seconds
    pollInterval = setInterval(fetchTowers, 2000);
  });

  onDestroy(() => {
    if (pollInterval) {
      clearInterval(pollInterval);
    }
  });
</script>

<div class="towers-page">
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">UAT Towers</span>
      {#if connected}
        <span class="badge success">Connected</span>
      {:else}
        <span class="badge danger">Disconnected</span>
      {/if}
    </div>

    <div class="panel-body">
      <!-- Header Row -->
      <div class="tower-header">
        <span class="col-location">Location</span>
        <span class="col-power">Current Power (dB)</span>
        <span class="col-power">Avg Power (dB)</span>
        <span class="col-power">Max Power (dB)</span>
        <span class="col-messages">Msgs Last Minute</span>
      </div>

      <!-- Tower Rows -->
      {#each towers as tower, i}
        <div class="tower-row">
          <span class="col-location">{tower.lat} {tower.lng}</span>
          <span class="col-power">{tower.power}</span>
          <span class="col-power">{tower.power_last_min}</span>
          <span class="col-power">{tower.power_max}</span>
          <span class="col-messages">{tower.messages}</span>
        </div>
      {:else}
        <div class="empty-message">
          {#if error}
            Error loading towers: {error}
          {:else}
            No towers detected. Make sure UAT is enabled and you are within range of a ground station.
          {/if}
        </div>
      {/each}
    </div>

    <div class="panel-footer">
      <p>Showing {towers.length} tower{towers.length !== 1 ? 's' : ''} with messages in the last minute</p>
    </div>
  </div>
</div>

<style>
  .towers-page {
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
  }

  .panel-label {
    font-weight: 600;
    font-size: 1.1rem;
  }

  .panel-body {
    padding: 0.5rem 1rem;
  }

  .panel-footer {
    padding: 0.75rem 1rem;
    background: var(--bg-secondary, #f5f5f5);
    border-top: 1px solid var(--border-color, #ddd);
    border-radius: 0 0 8px 8px;
  }

  .panel-footer p {
    margin: 0;
    font-size: 0.9rem;
    color: var(--text-secondary, #666);
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

  .tower-header,
  .tower-row {
    display: flex;
    flex-wrap: wrap;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border-color, #eee);
  }

  .tower-header {
    font-weight: 600;
    background: var(--bg-secondary, #f9f9f9);
    margin: 0 -1rem;
    padding: 0.5rem 1rem;
  }

  .tower-row:last-child {
    border-bottom: none;
  }

  .col-location {
    flex: 0 0 30%;
    min-width: 150px;
  }

  .col-power {
    flex: 0 0 17.5%;
    text-align: right;
    min-width: 80px;
  }

  .col-messages {
    flex: 0 0 17.5%;
    text-align: right;
    min-width: 80px;
  }

  .empty-message {
    padding: 2rem 1rem;
    text-align: center;
    color: var(--text-secondary, #666);
    font-style: italic;
  }

  /* Responsive adjustments */
  @media (max-width: 768px) {
    .col-location {
      flex: 0 0 100%;
      margin-bottom: 0.5rem;
    }

    .col-power,
    .col-messages {
      flex: 0 0 25%;
    }

    .tower-header .col-power,
    .tower-header .col-messages {
      font-size: 0.8rem;
    }
  }

  @media (max-width: 480px) {
    .col-power,
    .col-messages {
      flex: 0 0 50%;
      margin-bottom: 0.25rem;
    }
  }
</style>
