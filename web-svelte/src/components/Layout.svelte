<script>
  import { link, location } from 'svelte-spa-router';

  let menuOpen = false;

  function toggleMenu() {
    menuOpen = !menuOpen;
  }

  function closeMenu() {
    menuOpen = false;
  }

  const navItems = [
    { path: '/', label: 'Status', icon: 'home' },
    { path: '/traffic', label: 'Traffic', icon: 'plane' },
    { path: '/towers', label: 'Towers', icon: 'broadcast' },
    { path: '/weather', label: 'Weather', icon: 'cloud' },
    { path: '/gps', label: 'GPS/AHRS', icon: 'satellite' },
    { path: '/radar', label: 'Radar', icon: 'radar' },
    { path: '/map', label: 'Map', icon: 'map' },
    { path: '/logs', label: 'Logs', icon: 'file' },
    { path: '/settings', label: 'Settings', icon: 'gear' },
    { path: '/developer', label: 'Developer', icon: 'code' }
  ];

  // Check if path is active
  function isActive(path, currentLocation) {
    if (path === '/') return currentLocation === '/';
    return currentLocation.startsWith(path);
  }
</script>

<div class="app-container" class:menu-open={menuOpen}>
  <!-- Header -->
  <header class="app-header">
    <button class="menu-toggle" on:click={toggleMenu} aria-label="Toggle menu">
      <span class="hamburger"></span>
    </button>
    <h1 class="app-title">Stratux</h1>
  </header>

  <!-- Sidebar Navigation -->
  <nav class="sidebar" class:open={menuOpen}>
    <ul class="nav-list">
      {#each navItems as item}
        <li>
          <a
            href={item.path}
            use:link
            class:active={isActive(item.path, $location)}
            on:click={closeMenu}
          >
            <span class="nav-icon">{item.icon === 'home' ? '🏠' : item.icon === 'plane' ? '✈️' : item.icon === 'broadcast' ? '📡' : item.icon === 'cloud' ? '🌤️' : item.icon === 'satellite' ? '🛰️' : item.icon === 'radar' ? '📍' : item.icon === 'map' ? '🗺️' : item.icon === 'file' ? '📄' : item.icon === 'gear' ? '⚙️' : '💻'}</span>
            <span class="nav-label">{item.label}</span>
          </a>
        </li>
      {/each}
    </ul>
  </nav>

  <!-- Overlay for mobile -->
  {#if menuOpen}
    <div class="sidebar-overlay" on:click={closeMenu} on:keydown={closeMenu} role="button" tabindex="0"></div>
  {/if}

  <!-- Main Content -->
  <main class="main-content">
    <slot />
  </main>
</div>

<style>
  :root {
    --header-height: 56px;
    --sidebar-width: 220px;
    --primary-color: #1976d2;
    --bg-primary: #ffffff;
    --bg-secondary: #f5f5f5;
    --text-primary: #333;
    --text-secondary: #666;
    --border-color: #ddd;
  }

  .app-container {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }

  .app-header {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: var(--header-height);
    background: var(--primary-color);
    color: white;
    display: flex;
    align-items: center;
    padding: 0 1rem;
    z-index: 1000;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .menu-toggle {
    background: none;
    border: none;
    padding: 0.5rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .hamburger {
    display: block;
    width: 24px;
    height: 2px;
    background: white;
    position: relative;
  }

  .hamburger::before,
  .hamburger::after {
    content: '';
    position: absolute;
    width: 24px;
    height: 2px;
    background: white;
    left: 0;
  }

  .hamburger::before {
    top: -7px;
  }

  .hamburger::after {
    top: 7px;
  }

  .app-title {
    margin: 0 0 0 1rem;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .sidebar {
    position: fixed;
    top: var(--header-height);
    left: 0;
    bottom: 0;
    width: var(--sidebar-width);
    background: var(--bg-primary);
    border-right: 1px solid var(--border-color);
    transform: translateX(-100%);
    transition: transform 0.3s ease;
    z-index: 999;
    overflow-y: auto;
  }

  .sidebar.open {
    transform: translateX(0);
  }

  .sidebar-overlay {
    position: fixed;
    top: var(--header-height);
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0,0,0,0.5);
    z-index: 998;
  }

  .nav-list {
    list-style: none;
    padding: 0.5rem 0;
    margin: 0;
  }

  .nav-list a {
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    color: var(--text-primary);
    text-decoration: none;
    transition: background 0.2s;
  }

  .nav-list a:hover {
    background: var(--bg-secondary);
  }

  .nav-list a.active {
    background: var(--primary-color);
    color: white;
  }

  .nav-icon {
    margin-right: 0.75rem;
    font-size: 1.25rem;
  }

  .nav-label {
    font-weight: 500;
  }

  .main-content {
    margin-top: var(--header-height);
    padding: 1rem;
    flex: 1;
  }

  /* Desktop layout */
  @media (min-width: 1024px) {
    .menu-toggle {
      display: none;
    }

    .sidebar {
      transform: translateX(0);
    }

    .sidebar-overlay {
      display: none;
    }

    .main-content {
      margin-left: var(--sidebar-width);
    }
  }
</style>
