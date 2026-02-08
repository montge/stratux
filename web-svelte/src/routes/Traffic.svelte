<script>
  import { onMount, onDestroy } from 'svelte';
  import {
    trafficStore,
    validTraffic,
    invalidTraffic,
    trafficConnected,
    connectTraffic,
    disconnectTraffic
  } from '../stores/traffic.js';
  import { status, settings, fetchSettings, connectStatus, disconnectStatus } from '../stores/status.js';

  // Display options
  let showReg = false;
  let showSquawk = false;
  let showCategory = false;
  let showDistance = true;

  // Format squawk with leading zeros
  function formatSquawk(squawk) {
    if (squawk === '----') return squawk;
    return squawk.toString().padStart(4, '0');
  }

  // Format heading with leading zeros
  function formatHeading(heading) {
    if (heading === '---') return heading;
    return heading.toString().padStart(3, '0');
  }

  onMount(() => {
    connectStatus();
    connectTraffic();
    fetchSettings();
  });

  onDestroy(() => {
    disconnectStatus();
    disconnectTraffic();
  });
</script>

<div class="traffic-page">
  <!-- Valid Traffic Panel -->
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">ADS-B and TIS-B Traffic</span>
      {#if $trafficConnected}
        <span class="badge success">Connected</span>
      {:else}
        <span class="badge danger">Disconnected</span>
      {/if}
    </div>

    <div class="panel-body">
      <!-- Header Row -->
      <div class="traffic-header">
        <div class="col-half">
          <span class="col-name">{showReg ? 'Tail Num' : 'Callsign'}</span>
          {#if showSquawk}
            <span class="col-code">Squawk</span>
          {:else}
            <span class="col-code">Code</span>
          {/if}
          {#if showCategory}
            <span class="col-categ">Categ</span>
          {/if}
          {#if $status.GPS_connected && showDistance}
            <span class="col-dist">Dist</span>
            <span class="col-bearing">Bearing</span>
          {:else}
            <span class="col-location">Location</span>
          {/if}
        </div>
        <div class="col-half">
          <span class="col-alt">Altitude</span>
          <span class="col-vspeed"></span>
          <span class="col-speed">Speed</span>
          <span class="col-course">Course</span>
          <span class="col-power">Power</span>
          <span class="col-age">Age</span>
        </div>
      </div>

      <!-- Traffic Rows -->
      {#each $validTraffic as aircraft (aircraft.icao_int)}
        <div class="traffic-row">
          <div class="col-half">
            <span class="col-name">
              <span class="traffic-badge" style="background-color: {aircraft.trafficColor}">
                {#if aircraft.isStratux}
                  <img src="img/logo-transparent.png" alt="Stratux" class="stratux-logo" />
                {:else}
                  <span>{aircraft.addr_symb}</span>
                {/if}
                <strong>{showReg ? aircraft.reg : aircraft.tail}</strong>
              </span>
            </span>
            {#if showSquawk}
              <span class="col-code">{formatSquawk(aircraft.squawk)}</span>
            {:else}
              <span class="col-code code-small">
                {aircraft.icao}{#if aircraft.addr_type === 3}<span class="tfid">(TFID)</span>{/if}
              </span>
            {/if}
            {#if showCategory}
              <span class="col-categ code-small">{aircraft.category}</span>
            {/if}
            {#if $status.GPS_connected && showDistance}
              <span class="col-dist">{aircraft.dist.toFixed(1)}<span class="unit">NM</span></span>
              <span class="col-bearing">{aircraft.bearing}&deg;</span>
            {:else}
              <span class="col-location">{aircraft.lat} {aircraft.lng}</span>
            {/if}
          </div>
          <div class="col-half">
            <span class="col-alt">{aircraft.alt}'</span>
            <span class="col-vspeed">
              {#if aircraft.vspeed > 0}
                <span class="vspeed-up">&#9650;{aircraft.vspeed}</span>
              {:else if aircraft.vspeed < 0}
                <span class="vspeed-down">&#9660;{Math.abs(aircraft.vspeed)}</span>
              {/if}
            </span>
            <span class="col-speed">{aircraft.speed}<span class="unit">KTS</span></span>
            <span class="col-course">{formatHeading(aircraft.heading)}&deg;</span>
            <span class="col-power">{aircraft.signal.toFixed(2)}<span class="unit">dB</span></span>
            <span class="col-age">{aircraft.Age.toFixed(1)}<span class="unit">s</span></span>
          </div>
        </div>
      {:else}
        <div class="empty-message">No traffic with valid positions</div>
      {/each}
    </div>

    <!-- Footer with options -->
    <div class="panel-footer">
      <div class="option-grid">
        <label class="option">
          <span>Show Tail Number</span>
          <input type="checkbox" bind:checked={showReg} />
        </label>
        <label class="option">
          <span>Show Squawk</span>
          <input type="checkbox" bind:checked={showSquawk} />
        </label>
        <label class="option">
          <span>Show Category</span>
          <input type="checkbox" bind:checked={showCategory} />
        </label>
        <label class="option" class:disabled={!$status.GPS_connected}>
          <span>Show Distance {$status.GPS_connected ? '' : '(N/A)'}</span>
          <input type="checkbox" bind:checked={showDistance} disabled={!$status.GPS_connected} />
        </label>
      </div>
    </div>
  </div>

  <!-- Invalid Traffic Panel -->
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">Basic Mode S and No-Position Messages</span>
    </div>

    <div class="panel-body">
      <!-- Header Row -->
      <div class="traffic-header">
        <div class="col-half">
          <span class="col-name">{showReg ? 'Tail Num' : 'Callsign'}</span>
          <span class="col-code">Code</span>
          {#if showCategory}
            <span class="col-categ">Categ</span>
          {/if}
          <span class="col-dist">DistEst</span>
          <span class="col-squawk">Squawk</span>
        </div>
        <div class="col-half">
          <span class="col-alt">Altitude</span>
          <span class="col-vspeed"></span>
          <span class="col-speed">Speed</span>
          <span class="col-course">Course</span>
          <span class="col-power">Power</span>
          <span class="col-age">Age</span>
        </div>
      </div>

      <!-- Invalid Traffic Rows -->
      {#each $invalidTraffic as aircraft (aircraft.icao_int)}
        <div class="traffic-row">
          <div class="col-half">
            <span class="col-name">
              <span class="traffic-badge" style="background-color: {aircraft.trafficColor}">
                <span>{aircraft.addr_symb}</span>
                <strong>{showReg ? aircraft.reg : aircraft.tail}</strong>
              </span>
            </span>
            <span class="col-code code-small">
              {aircraft.icao}{#if aircraft.addr_type === 3}<span class="tfid">(TFID)</span>{/if}
            </span>
            {#if showCategory}
              <span class="col-categ code-small">{aircraft.category}</span>
            {/if}
            <span class="col-dist">{aircraft.distEst.toFixed(1)}<span class="unit">NM</span></span>
            <span class="col-squawk">{formatSquawk(aircraft.squawk)}</span>
          </div>
          <div class="col-half">
            <span class="col-alt">{aircraft.alt}</span>
            <span class="col-vspeed">
              {#if aircraft.vspeed > 0}
                <span class="vspeed-up">&#9650;{aircraft.vspeed}</span>
              {:else if aircraft.vspeed < 0}
                <span class="vspeed-down">&#9660;{Math.abs(aircraft.vspeed)}</span>
              {/if}
            </span>
            <span class="col-speed">{aircraft.speed}<span class="unit">KTS</span></span>
            <span class="col-course">{formatHeading(aircraft.heading)}&deg;</span>
            <span class="col-power">{aircraft.signal.toFixed(2)}<span class="unit">dB</span></span>
            <span class="col-age">{aircraft.Age.toFixed(1)}<span class="unit">s</span></span>
          </div>
        </div>
      {:else}
        <div class="empty-message">No traffic without positions</div>
      {/each}
    </div>

    <div class="panel-footer info-footer">
      <p>Stratux has not received valid ADS-B position transmissions from the aircraft in this section. They will only appear estimated on the Radar or if your EFB is able to display bearingless targets.</p>
    </div>
  </div>
</div>

<style>
  .traffic-page {
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

  .traffic-header,
  .traffic-row {
    display: flex;
    flex-wrap: wrap;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border-color, #eee);
  }

  .traffic-header {
    font-weight: 600;
    background: var(--bg-secondary, #f9f9f9);
    margin: 0 -1rem;
    padding: 0.5rem 1rem;
  }

  .traffic-row:last-child {
    border-bottom: none;
  }

  .col-half {
    flex: 0 0 50%;
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .col-name { flex: 0 0 25%; min-width: 100px; }
  .col-code { flex: 0 0 15%; min-width: 50px; }
  .col-categ { flex: 0 0 12%; min-width: 45px; }
  .col-location { flex: 0 0 35%; text-align: right; }
  .col-dist { flex: 0 0 15%; text-align: right; }
  .col-bearing { flex: 0 0 15%; text-align: right; }
  .col-squawk { flex: 0 0 15%; }
  .col-alt { flex: 0 0 20%; text-align: right; }
  .col-vspeed { flex: 0 0 12%; text-align: left; font-size: 0.75rem; }
  .col-speed { flex: 0 0 17%; text-align: right; }
  .col-course { flex: 0 0 17%; text-align: right; }
  .col-power { flex: 0 0 17%; text-align: right; }
  .col-age { flex: 0 0 17%; text-align: right; }

  .traffic-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.2rem 0.4rem;
    border-radius: 4px;
    color: white;
    font-size: 0.85rem;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .stratux-logo {
    height: 1em;
    vertical-align: middle;
  }

  .code-small {
    font-size: 0.8rem;
  }

  .tfid {
    font-size: 0.6rem;
    margin-left: 2px;
  }

  .unit {
    font-size: 0.6rem;
    margin-left: 1px;
  }

  .vspeed-up {
    color: #28a745;
  }

  .vspeed-down {
    color: #dc3545;
  }

  .empty-message {
    padding: 1rem;
    text-align: center;
    color: var(--text-secondary, #666);
    font-style: italic;
  }

  .option-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
  }

  .option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    cursor: pointer;
  }

  .option.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .option input[type="checkbox"] {
    width: 2rem;
    height: 1.2rem;
    cursor: pointer;
  }

  .info-footer p {
    margin: 0;
    font-size: 0.85rem;
    color: var(--text-secondary, #666);
  }

  /* Responsive adjustments */
  @media (max-width: 992px) {
    .col-half {
      flex: 0 0 100%;
    }

    .traffic-row .col-half:first-child {
      margin-bottom: 0.25rem;
    }

    .option-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }

  @media (max-width: 576px) {
    .option-grid {
      grid-template-columns: 1fr;
    }

    .col-location,
    .col-dist,
    .col-bearing,
    .col-categ {
      display: none;
    }
  }
</style>
