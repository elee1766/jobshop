<script>
  // The built-in nav (JOB-86): the view links, the signed-in user, header
  // widgets, the live dot and theme toggle, and "g" then a letter to jump
  // to a view. It was the desk app's own markup; as a plugin it reads the
  // views from api.views() and the app tells it where it is docked.
  //
  // context.dock is left, right, top, bottom or none; context.narrow is
  // true on a phone. A bottom bar on a phone is a tab bar, as before; its
  // live dot and theme toggle sit in the top bar the app reserves, with
  // position: fixed.
  import { onMount } from 'svelte';
  import Sun from '@lucide/svelte/icons/sun';
  import Moon from '@lucide/svelte/icons/moon';
  import { slot } from '../../kit/plugin.svelte.js';
  import { params } from '../../kit/plugin.svelte.js';
  import { iconFor } from './icons.js';

  /** @type {{ api: any, context?: { dock?: string, narrow?: boolean }, config?: any }} */
  let { api, context = {} } = $props();
  const url = params(() => api);
  const dock = $derived(context.dock || 'left');
  const bar = $derived(dock === 'top' || dock === 'bottom');
  const tabs = $derived(bar && !!context.narrow);

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

  // The manifests are fixed after the shell loaded plugins, which it does
  // before the app (and so this nav) mounts.
  /** @type {{ id: string, title: string, icon: string, key: string }[]} */
  const views = $derived(api.views());
  /** @type {{ plugin: string, element: string }[]} */
  const headers = $derived(api.contributions('header'));
  const view = $derived(url.view || 'attention');

  const go = (/** @type {Record<string, string>} */ p) => api.go(p);

  // "g" then a letter jumps to a view, as in Linear.
  let armed = 0;
  /** @param {KeyboardEvent} e */
  function onkey(e) {
    const el = /** @type {HTMLElement} */ (e.target);
    if (e.metaKey || e.ctrlKey || e.altKey || el.closest('input, textarea, select, [contenteditable]')) return;
    if (armed && Date.now() - armed < 1200) {
      armed = 0;
      const v = views.find((x) => x.key && x.key === e.key);
      if (v) { e.preventDefault(); go({ view: v.id, ticket: '' }); }
    } else if (e.key === 'g') {
      armed = Date.now();
    }
  }
</script>

<svelte:window onkeydown={onkey} />

<nav aria-label="views" class="nav {dock}" class:bar class:tabs data-dock={dock}>
  <div class="links">
    {#if user}<div class="user faint" title="signed in through the proxy">{user}</div>{/if}
    {#each views as v (v.id)}
      {@const Icon = iconFor(v.icon)}
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
  </div>
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
</nav>

<style>
  .nav {
    flex: 1; min-width: 0; min-height: 0; display: flex; flex-direction: column;
    background: var(--desk-side); color: var(--desk-text);
  }
  .nav.left { width: 212px; flex: 1 1 auto; border-right: 1px solid var(--desk-border); }
  .nav.right { width: 212px; flex: 1 1 auto; border-left: 1px solid var(--desk-border); }
  .links { flex: 1; min-height: 0; padding: 4px 8px; display: flex; flex-direction: column; gap: 1px; overflow-y: auto; }
  .user { padding: 0 8px 6px; font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  a {
    display: flex; align-items: center; gap: 9px; height: 28px; padding: 0 8px; border-radius: var(--desk-radius);
    color: var(--desk-muted); font-size: 13px; text-decoration: none;
  }
  a:hover { background: var(--desk-hover); text-decoration: none; color: var(--desk-text); }
  a.active { background: var(--desk-raised); color: var(--desk-text); }
  a.active .ico { color: var(--desk-accent); }
  .ico { position: relative; display: inline-flex; flex: none; }
  .lbl { flex: 1; min-width: 0; }
  .badge {
    min-width: 18px; height: 16px; padding: 0 5px; border-radius: 999px; font-size: 11px; line-height: 16px; text-align: center;
    background: var(--desk-accent-soft); color: var(--desk-accent);
  }
  .tabdot { display: none; }
  .hdr { padding: 4px 8px; }
  .foot { display: flex; align-items: center; gap: 4px; padding: 8px 12px; border-top: 1px solid var(--desk-border); }
  .live { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; }
  .tools { margin-left: auto; display: inline-flex; align-items: center; gap: 2px; }

  /* A bar above or below the page: links in a row, status at the end. */
  .nav.bar { flex-direction: row; align-items: center; width: 100%; flex: 1 1 auto; }
  .nav.top { border-bottom: 1px solid var(--desk-border); }
  .nav.bottom { border-top: 1px solid var(--desk-border); }
  .bar .links { flex-direction: row; align-items: center; gap: 2px; padding: 0 8px; overflow-x: auto; overflow-y: hidden; scrollbar-width: none; }
  .bar .user { padding: 0 8px; }
  .bar a { height: 36px; flex: none; }
  .bar .foot { border-top: 0; padding: 0 8px; flex: none; }

  /* Phone tab bar: icon over label, equal widths, a dot instead of a count. */
  .tabs .links { flex: 1; justify-content: space-around; gap: 0; padding: 0 4px; overflow: visible; }
  .tabs.bottom { padding-bottom: env(safe-area-inset-bottom); }
  .tabs .user { display: none; }
  .tabs a {
    flex: 1 1 0; min-width: 0; height: var(--desk-tabbar); flex-direction: column; justify-content: center; gap: 3px;
    padding: 0; border-radius: 8px; font-size: 10px;
  }
  .tabs .lbl { flex: none; max-width: 100%; overflow: hidden; text-overflow: ellipsis; }
  .tabs a.active { background: transparent; color: var(--desk-text); }
  .tabs .badge { display: none; }
  .tabs .tabdot { display: block; position: absolute; top: -2px; right: -4px; width: 7px; height: 7px; border-radius: 50%; background: var(--desk-warn); }
  .tabs .tools .ghost.icon { width: 40px; height: 40px; }
  /* The bottom tab bar has no room for status: it goes in the top bar the app reserves. */
  .tabs.bottom .foot {
    position: fixed; top: var(--vvtop, 0px); right: 0; z-index: 6; height: calc(40px + env(safe-area-inset-top));
    padding: env(safe-area-inset-top) 4px 0 0; border: 0; background: transparent;
  }
</style>
