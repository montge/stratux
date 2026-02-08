<script>
  import Router from 'svelte-spa-router';
  import Layout from './components/Layout.svelte';
  import Status from './routes/Status.svelte';
  import Traffic from './routes/Traffic.svelte';
  import Towers from './routes/Towers.svelte';
  import GPS from './routes/GPS.svelte';
  import Weather from './routes/Weather.svelte';
  import Radar from './routes/Radar.svelte';
  import Placeholder from './routes/Placeholder.svelte';

  // Route definitions
  const routes = {
    '/': Status,
    '/traffic': Traffic,
    '/towers': Towers,
    '/weather': Weather,
    '/gps': GPS,
    '/radar': Radar,
    '/map': Placeholder,
    '/logs': Placeholder,
    '/settings': Placeholder,
    '/developer': Placeholder,
    '*': Placeholder
  };
</script>

<Layout>
  <Router {routes} />
</Layout>

<style>
  :global(*) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    font-size: 16px;
    line-height: 1.5;
    color: #333;
    background: #f5f5f5;
  }

  :global(a) {
    color: #1976d2;
  }
</style>
