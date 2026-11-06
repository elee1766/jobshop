<script>
  // The built-in Desk app (JOB-85): brand, nav, main view, session panel,
  // ticket drawer and the phone tab bar. It was the shell's App.svelte;
  // as an `app` plugin it uses only the api, so a replacement app can do
  // everything it does.
  import { onMount } from 'svelte';
  import Inbox from '@lucide/svelte/icons/inbox';
  import MessageSquare from '@lucide/svelte/icons/message-square';
  import ListChecks from '@lucide/svelte/icons/list-checks';
  import ActivityIcon from '@lucide/svelte/icons/activity';
  import Terminal from '@lucide/svelte/icons/terminal';
  import Boxes from '@lucide/svelte/icons/boxes';
  import Rocket from '@lucide/svelte/icons/rocket';
  import Sun from '@lucide/svelte/icons/sun';
  import Moon from '@lucide/svelte/icons/moon';
  import Puzzle from '@lucide/svelte/icons/puzzle';
  import { params, slot } from '../../kit/plugin.svelte.js';
  import { panel } from './panel.svelte.js';
  import Drawer from './Drawer.svelte';
  import SidePanel from './SidePanel.svelte';

  /** @type {{ api: any, context?: any, config?: any }} */
  let { api } = $props();
  const url = params(() => api);

  let user = $state('');
  let connected = $state(false);
  let badges = $state(/** @type {Record<string, number>} */ ({}));
  let light = $state(false);
  onMount(() => {
    light = api.theme === 'light';
    user = api.user;
    // api.user fills in once the shell's whoami returns.
    if (!user) api.get('api/whoami').then((/** @type {any} */ w) => (user = w.user || '')).catch(() => {});
    const offLive = api.live((/** @type {boolean} */ v) => (connected = v));
    const offBadges = api.badges((/** @type {Record<string, number>} */ b) => (badges = b));
    return () => { offLive(); offBadges(); };
  });
  function toggleTheme() {
    light = !light;
    api.theme = light ? 'light' : 'dark';
  }

  // Every view is a plugin (#541). Known ids keep their icon, key and
  // nav order; other plugin views follow in plugin order.
  /** @type {Record<string, { icon: any, key: string }>} */
  const known = {
    attention: { icon: Inbox, key: 'a' },
    chat: { icon: MessageSquare, key: 'c' },
    work: { icon: ListChecks, key: 'w' },
    activity: { icon: ActivityIcon, key: 'y' },
    sessions: { icon: Terminal, key: 's' },
    pools: { icon: Boxes, key: 'p' },
    deploys: { icon: Rocket, key: 'd' },
  };
  const order = Object.keys(known);
  const rank = (/** @type {string} */ id) => { const i = order.indexOf(id); return i < 0 ? order.length : i; };

  // Contributions are fixed once the shell has loaded plugins, which it
  // does before mounting the app.
  /** @type {{ plugin: string, element: string, id?: string, title?: string }[]} */
  const pviews = $derived(api.contributions('view'));
  const nav = $derived(
    pviews
      .filter((p, i) => p.id && pviews.findIndex((q) => q.id === p.id) === i)
      .map((p, i) => ({ id: p.id ?? '', title: p.title || p.id || '', i }))
      .sort((a, b) => rank(a.id) - rank(b.id) || a.i - b.i)
      .map((v) => ({ ...v, icon: known[v.id]?.icon || Puzzle, key: known[v.id]?.key || '' })),
  );
  /** @type {{ plugin: string, element: string }[]} */
  const headers = $derived(api.contributions('header'));
  const view = $derived(url.view || 'attention');
  const pview = $derived(pviews.find((v) => v.id === view));

  const go = (/** @type {Record<string, string>} */ p) => api.go(p);

  // "g" then a letter jumps to a view, as in Linear.
  let armed = 0;
  /** @param {KeyboardEvent} e */
  function onkey(e) {
    // The ticket drawer sits on top and closes first; it marks its Escape handled.
    if (e.key === 'Escape' && url.session && !url.ticket && !e.defaultPrevented) { api.closeSession(); return; }
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

<div class="shell" class:withpanel={!!url.session} style:--pw={panel.w + 'px'} data-testid="desk-app">
  <div class="brand"><span class="mark"></span>Jobshop <span class="faint">Desk</span></div>
  <nav aria-label="views">
    {#if user}<div class="user faint" title="signed in through the proxy">{user}</div>{/if}
    {#each nav as v (v.id)}
      {@const Icon = v.icon}
      <a
        href={'?view=' + v.id}
        class:active={view === v.id}
        data-view={v.id}
        title={'Go to ' + v.title + (v.key ? ' (g ' + v.key + ')' : '')}
        onclick={(e) => { e.preventDefault(); go({ view: v.id, ticket: '' }); }}
      >
        <span class="ico"><Icon size={14} />{#if (badges[v.id] ?? 0) > 0}<span class="tabdot"></span>{/if}</span>
        <span class="lbl">{v.title}</span>
        {#if (badges[v.id] ?? 0) > 0}<span class="badge">{badges[v.id]}</span>{/if}
      </a>
    {/each}
    {#each headers as h (h.plugin + h.element)}
      <div class="hdr plugin-host" data-plugin={h.plugin} use:slot={{ api, slot: 'header', target: { plugin: h.plugin, element: h.element } }}></div>
    {/each}
  </nav>
  <div class="foot">
    <span class="live" title={connected ? 'live' : 'disconnected'}>
      <span class="dot" class:ok={connected} class:bad={!connected}></span>
      <span class="faint">{connected ? 'live' : 'offline'}</span>
    </span>
    <span class="tools">
      <button class="ghost icon" onclick={toggleTheme} aria-label="toggle theme" title="toggle theme">
        {#if light}<Moon size={14} />{:else}<Sun size={14} />{/if}
      </button>
    </span>
  </div>
  <main>
    {#if pview}
      {#key pview.plugin + pview.element}
        <div class="plugin-host" data-plugin={pview.plugin} use:slot={{ api, slot: 'view', target: { plugin: pview.plugin, element: pview.element } }}></div>
      {/key}
    {:else}<p class="faint pad" data-testid="view-missing">No plugin shows <span class="mono">{view}</span>.</p>
    {/if}
  </main>
  {#if url.session}{#key url.session}<SidePanel {api} id={url.session} item={url.item || ''} readonly={url.ro === '1'} />{/key}{/if}
  {#if url.ticket}<Drawer {api} id={url.ticket} />{/if}
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
  .hdr { padding: 4px 8px; }

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
