<script>
  import { onMount } from 'svelte';
  import { url, go, VIEWS } from './lib/url.svelte.js';
  import { live, connect, resource, get } from './lib/api.svelte.js';
  import { plugins, loadPlugins, contributions, dispatch, who, toasts, confirmBox } from './lib/plugins.svelte.js';
  import PluginSlot from './views/PluginSlot.svelte';
  import Attention from './views/Attention.svelte';
  import Artifacts from './views/Artifacts.svelte';
  import Work from './views/Work.svelte';
  import Sessions from './views/Sessions.svelte';
  import Pools from './views/Pools.svelte';
  import Activity from './views/Activity.svelte';
  import ChatView from './views/ChatView.svelte';
  import Drawer from './views/Drawer.svelte';

  onMount(() => connect(dispatch));
  onMount(loadPlugins);

  const attention = resource(() => 'api/attention');
  const count = $derived(attention.data?.length ?? 0);
  onMount(() => { get('api/whoami').then((w) => (who.user = w.user || '')).catch(() => {}); });
  const unread = resource(() => 'api/artifacts?unread=1');
  const unreadCount = $derived(unread.data?.length ?? 0);

  /** @type {Record<string, string>} */
  const names = { attention: 'Attention', artifacts: 'Artifacts', chat: 'Chat', work: 'Work', activity: 'Activity', sessions: 'Sessions', pools: 'Pools' };

  // Plugin views (#517): one with a built-in's id replaces it.
  const pviews = $derived(contributions('view'));
  const pview = $derived(pviews.find((v) => v.c.id === url.view));
  const nav = $derived([
    ...VIEWS.map((v) => ({ id: v, title: pviews.find((p) => p.c.id === v)?.c.title || names[v] })),
    ...pviews.filter((p) => !VIEWS.includes(p.c.id ?? '')).map((p) => ({ id: p.c.id ?? '', title: p.c.title || p.c.id || '' })),
  ]);
  const headers = $derived(contributions('header'));

  let light = $state(localStorage.getItem('desk-theme') === 'light');
  $effect(() => {
    document.documentElement.dataset.theme = light ? 'light' : 'dark';
    localStorage.setItem('desk-theme', light ? 'light' : 'dark');
  });
</script>

<div class="shell">
  <nav>
    <div class="brand">Jobshop <span class="muted">Desk</span></div>
    {#if who.user}<div class="user faint" title="signed in through the proxy">{who.user}</div>{/if}
    {#each nav as v (v.id)}
      <a
        href={'?view=' + v.id}
        class:active={url.view === v.id}
        data-view={v.id}
        onclick={(e) => { e.preventDefault(); go({ view: v.id, ticket: '', session: '', artifact: '' }); }}
      >
        {v.title}
        {#if v.id === 'attention' && count > 0}<span class="pill warn">{count}</span>{/if}
        {#if v.id === 'artifacts' && unreadCount > 0}<span class="pill accent" data-testid="artifacts-unread">{unreadCount}</span>{/if}
      </a>
    {/each}
    {#if plugins.error}<div class="faint perr" data-testid="plugins-error">plugins: {plugins.error}</div>{/if}
    {#each headers as h (h.plugin.name + h.c.element)}
      <div class="hdr"><PluginSlot plugin={h.plugin} element={h.c.element} /></div>
    {/each}
    <div class="foot">
      <span class="dot" class:on={live.connected} title={live.connected ? 'live' : 'disconnected'}></span>
      <span class="faint">{live.connected ? 'live' : 'offline'}</span>
      <button class="theme" onclick={() => (light = !light)} aria-label="toggle theme">{light ? '☾' : '☀'}</button>
    </div>
  </nav>
  <main>
    {#if pview}{#key pview.plugin.name + pview.c.element}<PluginSlot plugin={pview.plugin} element={pview.c.element} />{/key}
    {:else if url.view === 'artifacts'}<Artifacts />
    {:else if url.view === 'chat'}<ChatView />
    {:else if url.view === 'work'}<Work />
    {:else if url.view === 'activity'}<Activity />
    {:else if url.view === 'sessions'}<Sessions />
    {:else if url.view === 'pools'}<Pools />
    {:else if url.view !== 'attention' && !plugins.loaded}<p class="faint">Loading…</p>
    {:else}<Attention items={attention} />
    {/if}
  </main>
  {#if url.ticket}<Drawer id={url.ticket} />{/if}
  {#if toasts.items.length}
    <div class="toasts" data-testid="toasts">{#each toasts.items as t (t.id)}<div class="card toast">{t.msg}</div>{/each}</div>
  {/if}
  {#if confirmBox.open}
    {@const c = confirmBox.open}
    {@const done = (/** @type {boolean} */ v) => { confirmBox.open = null; c.resolve(v); }}
    <div class="backdrop" role="presentation" onclick={() => done(false)}></div>
    <div class="card confirm" role="dialog" aria-modal="true" data-testid="confirm">
      <p>{c.msg}</p>
      <div class="row"><button onclick={() => done(false)}>Cancel</button><button class="ok" onclick={() => done(true)}>OK</button></div>
    </div>
  {/if}
</div>

<style>
  .shell { display: grid; grid-template-columns: 200px 1fr; min-height: 100vh; }
  nav {
    border-right: 1px solid var(--border); padding: 14px 10px; display: flex; flex-direction: column; gap: 2px;
    position: sticky; top: 0; height: 100vh; background: var(--bg);
  }
  .brand { font-weight: 600; padding: 4px 10px 14px; }
  .user { padding: 0 10px 10px; margin-top: -8px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  nav a {
    color: var(--muted); padding: 5px 10px; border-radius: 6px; display: flex; justify-content: space-between; align-items: center;
  }
  nav a:hover { background: var(--raised); text-decoration: none; color: var(--text); }
  nav a.active { background: var(--accent-soft); color: var(--text); }
  .foot { margin-top: auto; display: flex; align-items: center; gap: 6px; padding: 0 10px; }
  .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--bad); }
  .dot.on { background: var(--ok); }
  .theme { margin-left: auto; padding: 0 7px; }
  .perr { padding: 4px 10px; color: var(--bad); }
  .hdr { padding: 4px 10px; }
  .toasts { position: fixed; right: 16px; bottom: 16px; display: flex; flex-direction: column; gap: 6px; z-index: 30; }
  .toast { padding: 8px 12px; }
  .backdrop { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.4); z-index: 40; }
  .confirm { position: fixed; left: 50%; top: 30%; transform: translateX(-50%); padding: 16px; z-index: 41; min-width: 280px; }
  .confirm .row { display: flex; gap: 8px; justify-content: flex-end; }
  .confirm .ok { border-color: var(--accent); }
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
