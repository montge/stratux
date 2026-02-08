<script>
  import { onMount } from 'svelte';
  import { settingsStore, loadSettings, saveSettings, hasSerialOutput, serialBaud } from '../stores/settings.js';

  let loading = $state(true);
  let error = $state(null);
  let gpsHardwareCode = $state(0);
  let hasTracker = $state(false);
  let showRebootConfirm = $state(false);
  let showShutdownConfirm = $state(false);
  let updateFile = $state(null);
  let uploading = $state(false);
  let wifiClientNetworks = $state([]);

  const channels = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11];

  const gainValues = [0.9, 1.4, 2.7, 3.7, 7.7, 8.7, 12.5, 14.4, 15.7, 16.6, 19.7, 20.7, 22.9, 25.4, 28.0, 29.7, 32.8, 33.8, 36.4, 37.2, 38.6, 40.2, 42.1, 43.4, 43.9, 44.5, 48.0, 49.6];

  const acftTypes = [
    { value: 1, label: 'Glider/Motorglider' },
    { value: 2, label: 'Tow plane', hideSoftRF: true },
    { value: 3, label: 'Helicopter' },
    { value: 4, label: 'Parachute', hideSoftRF: true, hideGXAir: true },
    { value: 5, label: 'Drop plane', hideSoftRF: true, hideGXAir: true },
    { value: 6, label: 'Hang glider' },
    { value: 7, label: 'Para glider' },
    { value: 8, label: 'Powered Aircraft' },
    { value: 9, label: 'Jet Aircraft', hideSoftRF: true, hideGXAir: true },
    { value: 10, label: 'UFO', hideSoftRF: true, hideGXAir: true },
    { value: 11, label: 'Balloon' },
    { value: 12, label: 'Airship', hideSoftRF: true, hideGXAir: true },
    { value: 13, label: 'UAV' },
    { value: 14, label: 'Ground support', hideGXAir: true },
    { value: 15, label: 'Static object', hideGXAir: true }
  ];

  let settings = $state({});

  // Subscribe to store
  const unsub = settingsStore.subscribe(val => { settings = val; });

  onMount(async () => {
    const data = await loadSettings();
    if (data) {
      wifiClientNetworks = data.WiFiClientNetworks || [];
      // Fetch status for GPS hardware detection
      try {
        const res = await fetch('/getStatus');
        const status = await res.json();
        gpsHardwareCode = status.GPS_detected_type & 0x0f;
        hasTracker = gpsHardwareCode === 3 || status.OGN_tx_enabled || gpsHardwareCode === 13 || gpsHardwareCode === 15;
      } catch (e) { /* ignore */ }
    }
    loading = false;
    return unsub;
  });

  function saveField(field, value) {
    saveSettings({ [field]: value });
  }

  function saveToggle(field) {
    saveSettings({ [field]: settings[field] });
  }

  async function updateWiFi() {
    await saveSettings({
      WiFiSSID: settings.WiFiSSID,
      WiFiPassphrase: settings.WiFiPassphrase,
      WiFiSecurityEnabled: settings.WiFiSecurityEnabled,
      WiFiChannel: settings.WiFiChannel,
      WiFiIPAddress: settings.WiFiIPAddress,
      WiFiMode: parseInt(settings.WiFiMode),
      WiFiDirectPin: settings.WiFiDirectPin,
      WiFiCountry: settings.WiFiCountry,
      WiFiClientNetworks: wifiClientNetworks,
      WiFiInternetPassThroughEnabled: settings.WiFiInternetPassThroughEnabled
    });
  }

  function addClientNetwork() {
    wifiClientNetworks = [...wifiClientNetworks, { SSID: '', Password: '' }];
  }

  function removeClientNetwork(idx) {
    wifiClientNetworks = wifiClientNetworks.filter((_, i) => i !== idx);
  }

  function updateOgnTracker() {
    saveSettings({
      OGNAddrType: parseInt(settings.OGNAddrType),
      OGNAddr: settings.OGNAddr,
      OGNAcftType: parseInt(settings.OGNAcftType),
      OGNPilot: settings.OGNPilot,
      OGNReg: settings.OGNReg,
      OGNTxPower: settings.OGNTxPower
    });
  }

  async function postReboot() {
    showRebootConfirm = false;
    try { await fetch('/reboot', { method: 'POST' }); } catch {}
    window.location.href = '/';
  }

  async function postShutdown() {
    showShutdownConfirm = false;
    try { await fetch('/shutdown', { method: 'POST' }); } catch {}
    window.location.href = '/';
  }

  async function uploadUpdate() {
    if (!updateFile) return;
    uploading = true;
    const fd = new FormData();
    fd.append('update_file', updateFile);
    try {
      await fetch('/updateUpload', { method: 'POST', body: fd });
      alert('Success. Wait 5 minutes and refresh to verify new version.');
      window.location.replace('/');
    } catch {
      alert('Upload failed.');
    }
    uploading = false;
  }

  function isAcftTypeVisible(type) {
    if (type.hideSoftRF && gpsHardwareCode === 13) return false;
    if (type.hideGXAir && gpsHardwareCode === 15) return false;
    return true;
  }
</script>

{#if loading}
  <p>Loading settings...</p>
{:else}

<!-- AHRS Section -->
<div class="panel">
  <h3>AHRS</h3>
  <div class="field">
    <label>G Limits</label>
    <input type="text" bind:value={settings.GLimits}
      disabled={!settings.IMU_Sensor_Enabled}
      onblur={() => saveField('GLimits', settings.GLimits)}
      placeholder="Space-separated negative and positive G meter limits" />
  </div>
</div>

<!-- Commands Section -->
<div class="panel">
  <h3>Commands</h3>
  <div class="field">
    {#if !updateFile && !uploading}
      <label class="file-btn">
        Select System Update file
        <input type="file" onchange={(e) => updateFile = e.target.files[0]} />
      </label>
    {:else if updateFile && !uploading}
      <button class="btn" onclick={uploadUpdate}>Install {updateFile.name}</button>
    {:else}
      <button class="btn" disabled>Uploading... Please wait</button>
    {/if}
  </div>
  <div class="field">
    <button class="btn" onclick={() => showRebootConfirm = true}>Reboot</button>
  </div>
  <div class="field">
    <button class="btn" onclick={() => showShutdownConfirm = true}>Shutdown</button>
  </div>
</div>

<!-- Theme -->
<div class="panel">
  <h3>Theme</h3>
  <div class="toggle-row">
    <label>Dark Mode</label>
    <input type="checkbox" bind:checked={settings.DarkMode} onchange={() => saveToggle('DarkMode')} />
  </div>
</div>

<!-- OGN Tracker -->
{#if hasTracker}
<div class="panel">
  <h3>Tracker Configuration</h3>
  <div class="field">
    <label>Address Type</label>
    <select bind:value={settings.OGNAddrType}>
      {#if gpsHardwareCode !== 15}<option value="0">Random</option>{/if}
      <option value="1">ICAO</option>
      <option value="2">Flarm</option>
      {#if gpsHardwareCode !== 13 && gpsHardwareCode !== 15}<option value="3">OGN</option>{/if}
    </select>
  </div>
  {#if !(gpsHardwareCode === 13 && settings.OGNAddrType !== '1')}
  <div class="field">
    <label>Address (Hex)</label>
    <input type="text" maxlength="6" bind:value={settings.OGNAddr} />
  </div>
  {/if}
  <div class="field">
    <label>Aircraft Type</label>
    <select bind:value={settings.OGNAcftType}>
      {#each acftTypes.filter(isAcftTypeVisible) as type}
        <option value={type.value}>{type.label}</option>
      {/each}
    </select>
  </div>
  {#if gpsHardwareCode !== 13}
  <div class="field">
    <label>Pilot Name</label>
    <input type="text" maxlength="15" bind:value={settings.OGNPilot} />
  </div>
  {/if}
  {#if gpsHardwareCode !== 13 && gpsHardwareCode !== 15}
  <div class="field">
    <label>Registration</label>
    <input type="text" maxlength="15" bind:value={settings.OGNReg} />
  </div>
  <div class="field">
    <label>Transmit Power (dBm)</label>
    <input type="number" min="-32" max="31" bind:value={settings.OGNTxPower} />
  </div>
  {/if}
  <button class="btn" onclick={updateOgnTracker}>Configure Tracker</button>
</div>
{/if}

<!-- Configuration -->
<div class="panel">
  <h3>Configuration</h3>
  <div class="field">
    <label>Ownship Mode S / OGN Codes</label>
    <input type="text" bind:value={settings.OwnshipModeS}
      onblur={() => saveField('OwnshipModeS', (settings.OwnshipModeS || '').toUpperCase())}
      placeholder="Hex, comma separated" />
  </div>
  <div class="field">
    <label>Watch List</label>
    <input type="text" bind:value={settings.WatchList}
      onblur={() => saveField('WatchList', (settings.WatchList || '').toUpperCase())}
      placeholder="Space-delimited 4-letter identifiers" />
  </div>
  <div class="field">
    <label>PPM Correction</label>
    <input type="number" bind:value={settings.PPM}
      onblur={() => saveField('PPM', parseInt(settings.PPM) || 0)} />
  </div>
  {#if settings.DeveloperMode}
  <div class="field">
    <label>SDR Dump 1090 ES Gain</label>
    <select bind:value={settings.Dump1090Gain}
      onchange={() => saveField('Dump1090Gain', parseFloat(settings.Dump1090Gain))}>
      {#each gainValues as g}
        <option value={g}>{g}{g === 37.2 ? ' DEFAULT' : ''}</option>
      {/each}
    </select>
  </div>
  {/if}
  {#if $hasSerialOutput}
  <div class="field">
    <label>Serial Output Baudrate</label>
    <input type="number" bind:value={settings.Baud}
      onblur={() => saveField('Baud', parseInt(settings.Baud) || 0)} />
  </div>
  {/if}
  <div class="field">
    <label>Static IPs</label>
    <input type="text" value={(settings.StaticIps || []).join(' ')}
      onblur={(e) => saveField('StaticIps', e.target.value.trim().split(/\s+/).filter(Boolean))}
      placeholder="Space-delimited IPs" />
  </div>
  {#if settings.BMP_Sensor_Enabled}
  <div class="field">
    <label>Pressure Altitude Offset</label>
    <input type="number" bind:value={settings.AltitudeOffset}
      onblur={() => saveField('AltitudeOffset', parseInt(settings.AltitudeOffset) || 0)} />
  </div>
  {/if}
  <div class="toggle-row">
    <label>GDL90 MSL Altitude</label>
    <input type="checkbox" bind:checked={settings.GDL90MSLAlt_Enabled} onchange={() => saveToggle('GDL90MSLAlt_Enabled')} />
  </div>
  <div class="toggle-row">
    <label>Bearingless target distance estimation</label>
    <input type="checkbox" bind:checked={settings.EstimateBearinglessDist} onchange={() => saveToggle('EstimateBearinglessDist')} />
  </div>
</div>

<!-- WiFi Settings -->
<div class="panel">
  <h3>WiFi Settings</h3>
  <div class="field">
    <label>WiFi Mode</label>
    <select bind:value={settings.WiFiMode}>
      <option value="0">AccessPoint</option>
      <option value="1">WiFi-Direct</option>
      <option value="2">AP+Client</option>
    </select>
  </div>
  <div class="field">
    <label>WiFi SSID</label>
    <input type="text" bind:value={settings.WiFiSSID} placeholder="WiFi Network Name" />
  </div>
  {#if settings.WiFiMode === '0' || settings.WiFiMode === '2' || settings.WiFiMode === 0 || settings.WiFiMode === 2}
  <div class="toggle-row">
    <label>Network Security</label>
    <input type="checkbox" bind:checked={settings.WiFiSecurityEnabled} />
  </div>
  {/if}
  <div class="field">
    <label>WiFi Passphrase</label>
    <input type="text" bind:value={settings.WiFiPassphrase}
      disabled={!settings.WiFiSecurityEnabled} placeholder="WiFi Passphrase" />
  </div>
  {#if settings.WiFiMode === '1' || settings.WiFiMode === 1}
  <div class="field">
    <label>WiFi-Direct Pin</label>
    <input type="text" bind:value={settings.WiFiDirectPin} placeholder="WiFi-Direct PIN" />
  </div>
  {/if}
  {#if settings.WiFiMode === '0' || settings.WiFiMode === 0}
  <div class="field">
    <label>WiFi Channel</label>
    <select bind:value={settings.WiFiChannel}>
      {#each channels as ch}<option value={ch}>{ch}</option>{/each}
    </select>
  </div>
  {/if}
  <div class="field">
    <label>Stratux IP Address</label>
    <input type="text" bind:value={settings.WiFiIPAddress} placeholder="192.168.10.1" />
  </div>
  {#if settings.WiFiMode === '2' || settings.WiFiMode === 2}
  <div class="toggle-row">
    <label>Internet Passthrough</label>
    <input type="checkbox" bind:checked={settings.WiFiInternetPassThroughEnabled} />
  </div>
  <button class="btn btn-sm" onclick={addClientNetwork}>Add WiFi Client Network</button>
  {#each wifiClientNetworks as net, i}
    <div class="client-net">
      <div class="field">
        <label>Client SSID</label>
        <input type="text" bind:value={net.SSID} />
      </div>
      <div class="field">
        <label>Client Passphrase</label>
        <input type="text" bind:value={net.Password} />
      </div>
      <button class="btn btn-sm btn-danger" onclick={() => removeClientNetwork(i)}>Remove</button>
    </div>
  {/each}
  {/if}
  <button class="btn" onclick={updateWiFi}>Submit WiFi Changes</button>
</div>

<!-- Developer Mode: Hardware -->
{#if settings.DeveloperMode}
<div class="panel">
  <h3>Hardware</h3>
  {#each [
    ['GPS', 'GPS_Enabled'],
    ['978 MHz (UAT)', 'UAT_Enabled'],
    ['1090 MHz (ADS-B/Mode-S ES)', 'ES_Enabled'],
    ['868 MHz (OGN, ADS-L)', 'OGN_Enabled'],
    ['162 MHz (AIS)', 'AIS_Enabled'],
    ['Internet (OGN APRS)', 'APRS_Enabled'],
    ['AHRS Sensor', 'IMU_Sensor_Enabled'],
    ['Baro Sensor', 'BMP_Sensor_Enabled'],
    ['Ping ADS-B', 'Ping_Enabled'],
    ['Pong ADS-B Receiver', 'Pong_Enabled'],
    ['OGN TX via I2C/GPIO HAT', 'OGNI2CTXEnabled']
  ] as [label, key]}
    <div class="toggle-row">
      <label>{label}</label>
      <input type="checkbox" bind:checked={settings[key]} onchange={() => saveToggle(key)} />
    </div>
  {/each}
  {#if settings.IMU_Sensor_Enabled}
  <div class="field">
    <label>Min fan duty cycle %</label>
    <input type="number" min="0" max="100" bind:value={settings.PWMDutyMin}
      onchange={() => saveField('PWMDutyMin', parseInt(settings.PWMDutyMin) || 0)} />
  </div>
  {/if}
</div>

<!-- Developer Mode: Diagnostics -->
<div class="panel">
  <h3>Diagnostics</h3>
  {#each [
    ['Show Traffic Source in Callsign', 'DisplayTrafficSource'],
    ['Verbose Message Log', 'DEBUG'],
    ['Record Replay Logs', 'ReplayLog'],
    ['Record Trace Logs', 'TraceLog'],
    ['Record AHRS Logs', 'AHRSLog'],
    ['Persistent Logging', 'PersistentLogging']
  ] as [label, key]}
    <div class="toggle-row">
      <label>{label}</label>
      <input type="checkbox" bind:checked={settings[key]} onchange={() => saveToggle(key)} />
    </div>
  {/each}
</div>

<!-- Raw Configuration -->
<div class="panel">
  <h3>Raw Configuration</h3>
  <pre class="raw-config">{JSON.stringify(settings, null, 2)}</pre>
</div>
{/if}

<!-- Modals -->
{#if showRebootConfirm}
<div class="modal-overlay" onclick={() => showRebootConfirm = false}>
  <div class="modal-dialog" onclick={(e) => e.stopPropagation()}>
    <h4>Are you sure?</h4>
    <p>Do you wish to reboot the Stratux? It will stop responding during reboot.</p>
    <div class="modal-actions">
      <button class="btn" onclick={() => showRebootConfirm = false}>Cancel</button>
      <button class="btn btn-danger" onclick={postReboot}>Reboot</button>
    </div>
  </div>
</div>
{/if}
{#if showShutdownConfirm}
<div class="modal-overlay" onclick={() => showShutdownConfirm = false}>
  <div class="modal-dialog" onclick={(e) => e.stopPropagation()}>
    <h4>Are you sure?</h4>
    <p>Do you wish to shutdown the Stratux? Disconnect power after shutdown.</p>
    <div class="modal-actions">
      <button class="btn" onclick={() => showShutdownConfirm = false}>Cancel</button>
      <button class="btn btn-danger" onclick={postShutdown}>Shutdown</button>
    </div>
  </div>
</div>
{/if}

{/if}

<style>
  .panel { background: white; border-radius: 8px; padding: 16px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,.1); }
  .panel h3 { margin: 0 0 12px; font-size: 1.1em; border-bottom: 1px solid #eee; padding-bottom: 8px; }
  .field { margin-bottom: 12px; }
  .field label { display: block; font-weight: 600; font-size: 0.9em; margin-bottom: 4px; color: #555; }
  .field input, .field select { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; font-size: 0.95em; }
  .field input:disabled { background: #f0f0f0; color: #999; }
  .toggle-row { display: flex; justify-content: space-between; align-items: center; padding: 6px 0; border-bottom: 1px solid #f0f0f0; }
  .toggle-row label { font-weight: 500; font-size: 0.9em; }
  .toggle-row input[type=checkbox] { width: 20px; height: 20px; }
  .btn { display: block; width: 100%; padding: 10px; margin-top: 8px; border: none; border-radius: 4px; background: #1976d2; color: white; font-size: 0.95em; cursor: pointer; }
  .btn:hover { background: #1565c0; }
  .btn:disabled { background: #999; cursor: not-allowed; }
  .btn-sm { padding: 6px 12px; font-size: 0.85em; width: auto; }
  .btn-danger { background: #d32f2f; }
  .btn-danger:hover { background: #c62828; }
  .client-net { border: 1px solid #ddd; border-radius: 4px; padding: 8px; margin: 8px 0; }
  .file-btn { display: block; padding: 10px; background: #1976d2; color: white; border-radius: 4px; text-align: center; cursor: pointer; }
  .file-btn input { display: none; }
  .raw-config { background: #f5f5f5; padding: 12px; border-radius: 4px; overflow-x: auto; font-size: 0.8em; max-height: 400px; overflow-y: auto; }
  .modal-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
  .modal-dialog { background: white; border-radius: 8px; padding: 24px; max-width: 400px; width: 90%; }
  .modal-dialog h4 { margin: 0 0 12px; }
  .modal-actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 16px; }
  .modal-actions .btn { width: auto; }
</style>
