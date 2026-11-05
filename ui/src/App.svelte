<script>
  import { onMount } from 'svelte';
  import { url, go, closeSession, VIEWS } from './lib/url.svelte.js';
  import { live, connect, resource, get } from './lib/api.svelte.js';
  import { panel } from './lib/panel.svelte.js';
  import { initPush } from './lib/push.svelte.js';
  import Inbox from '@lucide/svelte/icons/inbox';
  import Package from '@lucide/svelte/icons/package';
  import MessageSquare from '@lucide/svelte/icons/message-square';
  import ListChecks from '@lucide/svelte/icons/list-checks';
  import ActivityIcon from '@lucide/svelte/icons/activity';
  import Terminal from '@lucide/svelte/icons/terminal';
  import Boxes from '@lucide/svelte/icons/boxes';
  import Rocket from '@lucide/svelte/icons/rocket';
  import Sun from '@lucide/svelte/icons/sun';
  import Moon from '@lucide/svelte/icons/moon';
  import Puzzle from '@lucide/svelte/icons/puzzle';
  import { plugins, loadPlugins, contributions, dispatch, who, toasts, confirmBox } from './lib/plugins.svelte.js';
  import PluginSlot from './views/PluginSlot.svelte';
  import Artifacts from './views/Artifacts.svelte';
  import Work from './views/Work.svelte';
  import Sessions from './views/Sessions.svelte';
  import Pools from './views/Pools.svelte';
  import Activity from './views/Activity.svelte';
  import Drawer from './views/Drawer.svelte';
  import SidePanel from './views/SidePanel.svelte';

  onMount(() => connect(dispatch));
  onMount(loadPlugins);
  // Registers the service worker, so the PWA installs and pushes arrive
  // even before a plugin asks for api.push.
  onMount(() => { initPush(); });

  // Keep the shell as tall as the visible viewport so the chat composer
  // stays above the on-screen keyboard (iOS does not resize the layout viewport).
  onMount(() => {
    const vv = window.visualViewport;
    if (!vv) return;
    const root = document.documentElement.style;
    const fit = () => {
      root.setProperty('--vvh', vv.height + 'px');
      root.setProperty('--vvtop', vv.offsetTop + 'px');
    };
    fit();
    vv.addEventListener('resize', fit);
    vv.addEventListener('scroll', fit);
    return () => { vv.removeEventListener('resize', fit); vv.removeEventListener('scroll', fit); };
  });

  const attention = resource(() => 'api/attention');
  const count = $derived(attention.data?.length ?? 0);
  onMount(() => { get('api/whoami').then((w) => (who.user = w.user || '')).catch(() => {}); });
  const unread = resource(() => 'api/artifacts?unread=1');
  const unreadCount = $derived(unread.data?.length ?? 0);

  /** @type {Record<string, { name: string, icon: any, key: string }>} */
  const builtin = {
    attention: { name: 'Attention', icon: Inbox, key: 'a' },
    artifacts: { name: 'Artifacts', icon: Package, key: 'r' },
    chat: { name: 'Chat', icon: MessageSquare, key: 'c' },
    work: { name: 'Work', icon: ListChecks, key: 'w' },
    activity: { name: 'Activity', icon: ActivityIcon, key: 'y' },
    sessions: { name: 'Sessions', icon: Terminal, key: 's' },
    pools: { name: 'Pools', icon: Boxes, key: 'p' },
    // Built-in plugin views (#517) keep their icon and key.
    deploys: { name: 'Deploys', icon: Rocket, key: 'd' },
  };

  // Plugin views (#517): one with a built-in's id replaces it.
  const pviews = $derived(contributions('view'));
  const pview = $derived(pviews.find((v) => v.c.id === url.view));
  const nav = $derived([
    ...VIEWS.map((v) => ({ id: v, title: pviews.find((p) => p.c.id === v)?.c.title || builtin[v].name })),
    ...pviews.filter((p) => !VIEWS.includes(p.c.id ?? '')).map((p) => ({ id: p.c.id ?? '', title: p.c.title || p.c.id || '' })),
  ].map((v) => ({ ...v, icon: builtin[v.id]?.icon || Puzzle, key: builtin[v.id]?.key || '' })));
  const headers = $derived(contributions('header'));

  let light = $state(localStorage.getItem('desk-theme') === 'light');
  $effect(() => {
    document.documentElement.dataset.theme = light ? 'light' : 'dark';
    localStorage.setItem('desk-theme', light ? 'light' : 'dark');
  });

  // "g" then a letter jumps to a view, as in Linear.
  let armed = 0;
  /** @param {KeyboardEvent} e */
  function onkey(e) {
    // The ticket drawer sits on top and closes first; it marks its Escape handled.
    if (e.key === 'Escape' && url.session && !url.ticket && !e.defaultPrevented) { closeSession(); return; }
    const el = /** @type {HTMLElement} */ (e.target);
    if (e.metaKey || e.ctrlKey || e.altKey || el.closest('input, textarea, select, [contenteditable]')) return;
    if (armed && Date.now() - armed < 1200) {
      armed = 0;
      const v = nav.find((x) => x.key && x.key === e.key);
      if (v) { e.preventDefault(); go({ view: v.id, ticket: '' }); }
    } else if (e.key === 'g') {
      armed = Date.now();
    }
  }
</script>

<svelte:window onkeydown={onkey} />

<div class="shell" class:withpanel={!!url.session} style:--pw={panel.w + 'px'}>
  <div class="brand"><span class="mark"></span>Jobshop <span class="faint">Desk</span></div>
  <nav aria-label="views">
    {#if who.user}<div class="user faint" title="signed in through the proxy">{who.user}</div>{/if}
    {#each nav as v (v.id)}
      {@const Icon = v.icon}
      <a
        href={'?view=' + v.id}
        class:active={url.view === v.id}
        data-view={v.id}
        title={'Go to ' + v.title + (v.key ? ' (g ' + v.key + ')' : '')}
        onclick={(e) => { e.preventDefault(); go({ view: v.id, ticket: '', artifact: '' }); }}
      >
        <span class="ico"><Icon size={14} />{#if v.id === 'attention' && count > 0}<span class="tabdot"></span>{/if}</span>
        <span class="lbl">{v.title}</span>
        {#if v.id === 'attention' && count > 0}<span class="badge">{count}</span>{/if}
        {#if v.id === 'artifacts' && unreadCount > 0}<span class="badge" data-testid="artifacts-unread">{unreadCount}</span>{/if}
      </a>
    {/each}
    {#if plugins.error}<div class="faint perr" data-testid="plugins-error">plugins: {plugins.error}</div>{/if}
    {#each headers as h (h.plugin.name + h.c.element)}
      <div class="hdr"><PluginSlot plugin={h.plugin} element={h.c.element} /></div>
    {/each}
  </nav>
  <div class="foot">
    <span class="live" title={live.connected ? 'live' : 'disconnected'}>
      <span class="dot" class:ok={live.connected} class:bad={!live.connected}></span>
      <span class="faint">{live.connected ? 'live' : 'offline'}</span>
    </span>
    <span class="tools">
      <button class="ghost icon" onclick={() => (light = !light)} aria-label="toggle theme" title="toggle theme">
        {#if light}<Moon size={14} />{:else}<Sun size={14} />{/if}
      </button>
    </span>
  </div>
  <main>
    {#if pview}{#key pview.plugin.name + pview.c.element}<PluginSlot plugin={pview.plugin} element={pview.c.element} />{/key}
    {:else if url.view === 'artifacts'}<Artifacts />
    {:else if url.view === 'work'}<Work />
    {:else if url.view === 'activity'}<Activity />
    {:else if url.view === 'sessions'}<Sessions />
    {:else if url.view === 'pools'}<Pools />
    {:else if !plugins.loaded}<p class="faint">Loading…</p>
    {:else}<p class="faint pad" data-testid="view-missing">No plugin shows <span class="mono">{url.view}</span>{plugins.error ? ': ' + plugins.error : ''}.</p>
    {/if}
  </main>
  {#if url.session}{#key url.session}<SidePanel id={url.session} />{/key}{/if}
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
  .shell {
    position: fixed; left: 0; right: 0; top: var(--vvtop, 0px); height: var(--vvh, 100dvh);
    display: grid; grid-template-columns: 212px minmax(0, 1fr); grid-template-rows: auto 1fr auto;
    grid-template-areas: 'brand main' 'nav main' 'foot main';
    align-content: start;
  }
  .shell.withpanel {
    grid-template-columns: 212px minmax(0, 1fr) var(--pw);
    grid-template-areas: 'brand main panel' 'nav main panel' 'foot main panel';
  }
  .shell > * { min-width: 0; }
  .brand, nav, .foot { background: var(--side); border-right: 1px solid var(--border); }
  .brand { grid-area: brand; display: flex; align-items: center; gap: 8px; padding: 14px 16px 10px; font-weight: 600; font-size: 13px; }
  .mark { width: 14px; height: 14px; border-radius: 4px; background: var(--accent); }
  nav { grid-area: nav; padding: 4px 8px; display: flex; flex-direction: column; gap: 1px; overflow-y: auto; }
  .user { padding: 0 8px 6px; font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  nav a {
    display: flex; align-items: center; gap: 9px; height: 28px; padding: 0 8px; border-radius: var(--radius);
    color: var(--muted); font-size: 13px;
  }
  nav a:hover { background: var(--hover); text-decoration: none; color: var(--text); }
  nav a.active { background: var(--raised); color: var(--text); }
  nav a.active .ico { color: var(--accent); }
  .ico { position: relative; display: inline-flex; flex: none; }
  .lbl { flex: 1; min-width: 0; }
  .badge {
    min-width: 18px; height: 16px; padding: 0 5px; border-radius: 999px; font-size: 11px; line-height: 16px; text-align: center;
    background: var(--accent-soft); color: var(--accent);
  }
  .tabdot { display: none; }
  .foot { grid-area: foot; display: flex; align-items: center; gap: 4px; padding: 8px 12px; border-top: 1px solid var(--border); }
  .live { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; }
  .tools { margin-left: auto; display: inline-flex; align-items: center; gap: 2px; }
  main { grid-area: main; min-height: 0; overflow-y: auto; overscroll-behavior: contain; display: flex; flex-direction: column; background: var(--bg); }
  /* The sidebar rows stay compact; the main column takes the full height. */
  .shell { grid-template-rows: auto 1fr auto; }
  .perr { padding: 4px 8px; color: var(--bad); }
  .hdr { padding: 4px 8px; }
  .toasts { position: fixed; right: 16px; bottom: 16px; display: flex; flex-direction: column; gap: 6px; z-index: 30; }
  .toast { padding: 8px 12px; }
  .backdrop { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.4); z-index: 40; }
  .confirm { position: fixed; left: 50%; top: 30%; transform: translateX(-50%); padding: 16px; z-index: 41; min-width: 280px; }
  .confirm .row { display: flex; gap: 8px; justify-content: flex-end; }
  .confirm .ok { border-color: var(--accent); }

  @media (max-width: 768px) {
    .shell, .shell.withpanel {
      grid-template-columns: minmax(0, 1fr); grid-template-rows: auto minmax(0, 1fr) auto;
      grid-template-areas: 'top' 'main' 'tabs'; align-content: stretch;
    }
    /* Brand and the foot controls share the top bar cell. */
    .brand, .foot { grid-area: top; background: transparent; border: 0; z-index: 6; }
    .shell::before { content: ''; grid-area: top; background: var(--side); border-bottom: 1px solid var(--border); }
    .brand { padding: calc(env(safe-area-inset-top) + 0px) 12px 0; justify-self: start; align-self: center; height: 40px; }
    .foot { justify-self: end; align-self: center; padding: calc(env(safe-area-inset-top) + 0px) 4px 0 0; height: 40px; }
    .tools .ghost.icon { width: 40px; height: 40px; }
    main { grid-area: main; }
    nav {
      grid-area: tabs; flex-direction: row; justify-content: space-around; gap: 0; padding: 0 4px env(safe-area-inset-bottom);
      border-right: 0; border-top: 1px solid var(--border); overflow: visible;
    }
    nav a {
      flex: 1 1 0; min-width: 0; height: var(--tabbar); flex-direction: column; justify-content: center; gap: 3px;
      padding: 0; border-radius: 8px; font-size: 10px;
    }
    nav a .lbl { flex: none; max-width: 100%; overflow: hidden; text-overflow: ellipsis; }
    nav a.active { background: transparent; color: var(--text); }
    .badge { display: none; }
    .tabdot { display: block; position: absolute; top: -2px; right: -4px; width: 7px; height: 7px; border-radius: 50%; background: var(--warn); }
  }
</style>
