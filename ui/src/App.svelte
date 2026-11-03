<script>
  import { onMount } from 'svelte';
  import { url, go, VIEWS } from './lib/url.svelte.js';
  import { live, connect, resource } from './lib/api.svelte.js';
  import Attention from './views/Attention.svelte';
  import Work from './views/Work.svelte';
  import Sessions from './views/Sessions.svelte';
  import Pools from './views/Pools.svelte';
  import Deploys from './views/Deploys.svelte';
  import ChatView from './views/ChatView.svelte';
  import Drawer from './views/Drawer.svelte';

  onMount(connect);

  const attention = resource(() => '/api/attention');
  const count = $derived(attention.data?.length ?? 0);

  /** @type {Record<string, string>} */
  const names = { attention: 'Attention', chat: 'Chat', work: 'Work', sessions: 'Sessions', pools: 'Pools', deploys: 'Deploys' };

  let light = $state(localStorage.getItem('desk-theme') === 'light');
  $effect(() => {
    document.documentElement.dataset.theme = light ? 'light' : 'dark';
    localStorage.setItem('desk-theme', light ? 'light' : 'dark');
  });
</script>

<div class="shell">
  <nav>
    <div class="brand">Jobshop <span class="muted">Desk</span></div>
    {#each VIEWS as v (v)}
      <a
        href={'?view=' + v}
        class:active={url.view === v}
        data-view={v}
        onclick={(e) => { e.preventDefault(); go({ view: v, ticket: '', session: '' }); }}
      >
        {names[v]}
        {#if v === 'attention' && count > 0}<span class="pill warn">{count}</span>{/if}
      </a>
    {/each}
    <div class="foot">
      <span class="dot" class:on={live.connected} title={live.connected ? 'live' : 'disconnected'}></span>
      <span class="faint">{live.connected ? 'live' : 'offline'}</span>
      <button class="theme" onclick={() => (light = !light)} aria-label="toggle theme">{light ? '☾' : '☀'}</button>
    </div>
  </nav>
  <main>
    {#if url.view === 'attention'}<Attention items={attention} />
    {:else if url.view === 'chat'}<ChatView />
    {:else if url.view === 'work'}<Work />
    {:else if url.view === 'sessions'}<Sessions />
    {:else if url.view === 'pools'}<Pools />
    {:else if url.view === 'deploys'}<Deploys />
    {/if}
  </main>
  {#if url.ticket}<Drawer id={url.ticket} />{/if}
</div>

<style>
  .shell { display: grid; grid-template-columns: 200px 1fr; min-height: 100vh; }
  nav {
    border-right: 1px solid var(--border); padding: 14px 10px; display: flex; flex-direction: column; gap: 2px;
    position: sticky; top: 0; height: 100vh; background: var(--bg);
  }
  .brand { font-weight: 600; padding: 4px 10px 14px; }
  nav a {
    color: var(--muted); padding: 5px 10px; border-radius: 6px; display: flex; justify-content: space-between; align-items: center;
  }
  nav a:hover { background: var(--raised); text-decoration: none; color: var(--text); }
  nav a.active { background: var(--accent-soft); color: var(--text); }
  .foot { margin-top: auto; display: flex; align-items: center; gap: 6px; padding: 0 10px; }
  .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--bad); }
  .dot.on { background: var(--ok); }
  .theme { margin-left: auto; padding: 0 7px; }
  main { padding: 20px 28px; min-width: 0; }
  @media (max-width: 720px) {
    .shell { grid-template-columns: 1fr; }
    nav {
      position: sticky; height: auto; flex-direction: row; overflow-x: auto; border-right: 0;
      border-bottom: 1px solid var(--border); padding: 8px; z-index: 5;
    }
    .brand { display: none; }
    nav a { white-space: nowrap; gap: 6px; }
    .foot { margin: 0 0 0 auto; }
    main { padding: 14px 12px; }
  }
</style>
