<script>
  import { onMount, onDestroy } from 'svelte';
  import {
    weatherConnected,
    watchList,
    recentList,
    watching,
    connectWeather,
    disconnectWeather
  } from '../stores/weather.js';

  let watchExpanded = true;
  let recentExpanded = true;

  // Flight condition badge color
  function conditionClass(fc) {
    if (fc === 'VFR') return 'condition-vfr';
    if (fc === 'MVFR') return 'condition-mvfr';
    if (fc === 'IFR') return 'condition-ifr';
    if (fc === 'LIFR') return 'condition-lifr';
    return '';
  }

  // Type badge class (for non-METAR/SPECI types like TAF, PIREP, etc.)
  function typeClass(type) {
    if (type === 'TAF') return 'type-taf';
    if (type === 'PIREP') return 'type-pirep';
    if (type === 'SIGMET' || type === 'AIRMET') return 'type-sigmet';
    return 'type-default';
  }

  onMount(() => {
    connectWeather();
  });

  onDestroy(() => {
    disconnectWeather();
  });
</script>

<div class="weather-page">
  <!-- Header Panel -->
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">Weather</span>
      {#if $weatherConnected}
        <span class="badge success">Connected</span>
      {:else}
        <span class="badge danger">Disconnected</span>
      {/if}
    </div>
  </div>

  <!-- Watchlist Panel -->
  <div class="panel">
    <button class="panel-heading clickable" on:click={() => watchExpanded = !watchExpanded}>
      <span class="panel-label">
        Watching <span class="watching-list">{$watching}</span>
        ({$watchList.length})
      </span>
      <span class="expand-icon">{watchExpanded ? '\u25BC' : '\u25B6'}</span>
    </button>

    {#if watchExpanded}
      <div class="panel-body">
        <!-- Header Row -->
        <div class="weather-header">
          <span class="col-location">Location</span>
          <span class="col-type">Type</span>
          <span class="col-time">Age</span>
          <span class="col-report">Report</span>
        </div>

        {#each $watchList as weather (weather.location + '-' + weather.type)}
          <div class="weather-row">
            <div class="weather-meta">
              <span class="col-location"><strong>{weather.location}</strong></span>
              <span class="col-type">
                <span class="type-badge {weather.flightCondition ? conditionClass(weather.flightCondition) : typeClass(weather.type)}">
                  {weather.type}{#if weather.isAmended} (AMD){/if}
                </span>
              </span>
              <span class="col-time">{weather.time}</span>
            </div>
            <div class="weather-data">{weather.data}</div>
          </div>
        {:else}
          <div class="empty-message">No watched weather data received yet</div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Recent Reports Panel -->
  <div class="panel">
    <button class="panel-heading clickable" on:click={() => recentExpanded = !recentExpanded}>
      <span class="panel-label">
        Recent Reports ({$recentList.length})
      </span>
      <span class="expand-icon">{recentExpanded ? '\u25BC' : '\u25B6'}</span>
    </button>

    {#if recentExpanded}
      <div class="panel-body">
        <!-- Header Row -->
        <div class="weather-header">
          <span class="col-location">Location</span>
          <span class="col-type">Type</span>
          <span class="col-time">Age</span>
          <span class="col-report">Report</span>
        </div>

        {#each $recentList as weather, i (i)}
          <div class="weather-row">
            <div class="weather-meta">
              <span class="col-location"><strong>{weather.location}</strong></span>
              <span class="col-type">
                <span class="type-badge {weather.flightCondition ? conditionClass(weather.flightCondition) : typeClass(weather.type)}">
                  {weather.type}{#if weather.isAmended} (AMD){/if}
                </span>
              </span>
              <span class="col-time">{weather.time}</span>
            </div>
            <div class="weather-data">{weather.data}</div>
          </div>
        {:else}
          <div class="empty-message">No weather reports received yet</div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .weather-page {
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
    width: 100%;
    border: none;
    text-align: left;
    font-size: inherit;
    font-family: inherit;
    color: inherit;
  }

  .panel-heading.clickable {
    cursor: pointer;
  }

  .panel-heading.clickable:hover {
    background: var(--bg-hover, #ebebeb);
  }

  .panel-label {
    font-weight: 600;
    font-size: 1.1rem;
    flex: 1;
  }

  .watching-list {
    font-weight: 400;
    font-size: 0.95rem;
    color: var(--text-secondary, #666);
  }

  .expand-icon {
    font-size: 0.8rem;
    color: var(--text-secondary, #666);
  }

  .panel-body {
    padding: 0.5rem 1rem;
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

  .weather-header {
    display: flex;
    flex-wrap: wrap;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border-color, #ddd);
    font-weight: 600;
    background: var(--bg-secondary, #f9f9f9);
    margin: 0 -1rem;
    padding: 0.5rem 1rem;
  }

  .weather-row {
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border-color, #eee);
  }

  .weather-row:last-child {
    border-bottom: none;
  }

  .weather-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    margin-bottom: 0.25rem;
  }

  .col-location {
    flex: 0 0 15%;
    min-width: 80px;
  }

  .col-type {
    flex: 0 0 15%;
    min-width: 70px;
  }

  .col-time {
    flex: 0 0 20%;
    text-align: right;
    min-width: 100px;
  }

  .col-report {
    flex: 1;
    min-width: 100px;
  }

  .weather-data {
    font-family: 'Courier New', Courier, monospace;
    font-size: 0.9rem;
    line-height: 1.4;
    padding: 0.25rem 0;
    word-break: break-word;
    color: var(--text-primary, #333);
  }

  .type-badge {
    display: inline-block;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
    font-weight: 600;
    text-align: center;
  }

  /* Flight condition colors (FAA standard) */
  .condition-vfr {
    background: #28a745;
    color: white;
  }

  .condition-mvfr {
    background: #1976d2;
    color: white;
  }

  .condition-ifr {
    background: #dc3545;
    color: white;
  }

  .condition-lifr {
    background: #e040fb;
    color: white;
  }

  /* Report type colors */
  .type-taf {
    background: #6c757d;
    color: white;
  }

  .type-pirep {
    background: #ff9800;
    color: white;
  }

  .type-sigmet {
    background: #f44336;
    color: white;
  }

  .type-default {
    background: #607d8b;
    color: white;
  }

  .empty-message {
    padding: 1rem;
    text-align: center;
    color: var(--text-secondary, #666);
    font-style: italic;
  }

  /* Responsive */
  @media (max-width: 768px) {
    .col-location,
    .col-type,
    .col-time {
      flex: 0 0 33%;
    }

    .col-report {
      display: none;
    }
  }

  @media (max-width: 480px) {
    .col-time {
      flex: 1;
    }
  }
</style>
