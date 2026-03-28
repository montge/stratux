<script>
  import { onMount, onDestroy } from 'svelte';
  import { status, connectStatus, disconnectStatus } from '../stores/status.js';

  let userAgent = '';
  let deviceViewport = '';

  onMount(() => {
    connectStatus();
    userAgent = navigator.userAgent;
    deviceViewport = `screen = ${window.screen.width} x ${window.screen.height}`;
  });

  onDestroy(() => {
    disconnectStatus();
  });
</script>

<div class="logs-page">
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">Logs</span>
    </div>

    <div class="panel-body">
      <div class="log-links">
        <div class="log-item">
          <a href="/logs/stratux.log" target="_blank" rel="noopener" class="log-link">
            stratux.log
          </a>
          <span class="log-desc">Main application log</span>
        </div>

        <div class="log-item">
          <a href="/logs/" target="_blank" rel="noopener" class="log-link">
            System, AHRS, and replay logs
          </a>
          <span class="log-desc">Browse all log files</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Device Info -->
  <div class="panel">
    <div class="panel-heading">
      <span class="panel-label">Device Info</span>
    </div>
    <div class="panel-body">
      <div class="info-item">
        <span class="info-label">User Agent</span>
        <code class="info-value">{userAgent}</code>
      </div>
      <div class="info-item">
        <span class="info-label">Viewport</span>
        <code class="info-value">{deviceViewport}</code>
      </div>
    </div>
  </div>
</div>

<style>
  .logs-page {
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
    padding: 1rem;
  }

  .log-links {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .log-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .log-link {
    font-size: 1.1rem;
    font-weight: 500;
    color: #1976d2;
    text-decoration: none;
  }

  .log-link:hover {
    text-decoration: underline;
  }

  .log-desc {
    font-size: 0.85rem;
    color: var(--text-secondary, #666);
  }

  .info-item {
    margin-bottom: 0.75rem;
  }

  .info-item:last-child {
    margin-bottom: 0;
  }

  .info-label {
    display: block;
    font-weight: 600;
    font-size: 0.85rem;
    color: var(--text-secondary, #666);
    margin-bottom: 0.25rem;
  }

  .info-value {
    display: block;
    font-size: 0.85rem;
    background: var(--bg-secondary, #f5f5f5);
    padding: 0.5rem;
    border-radius: 4px;
    word-break: break-all;
  }
</style>
