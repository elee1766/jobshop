<script>
  // The shell (JOB-85): only the runtime. It loads plugins, keeps the
  // event stream, URL state, toasts and confirm, and mounts one `app`
  // contribution, which owns the whole page. The built-in app is the
  // desk plugin (ui/plugins/desk), today's frame.
  import { onMount } from 'svelte';
  import { connect, resource, get } from './lib/api.svelte.js';
  import { initPush } from './lib/push.svelte.js';
  import { setBadge } from './lib/signals.js';
  import { plugins, loadPlugins, contributions, dispatch, who, toasts, confirmBox } from './lib/plugins.svelte.js';
  import PluginSlot from './views/PluginSlot.svelte';

  onMount(() => connect(dispatch));
  onMount(loadPlugins);
  // Registers the service worker, so the PWA installs and pushes arrive
  // even before a plugin asks for api.push.
  onMount(() => { initPush(); });
  onMount(() => { get('api/whoami').then((w) => (who.user = w.user || '')).catch(() => {}); });

  // Keep --vvh as tall as the visible viewport so an app can keep its
  // composer above the on-screen keyboard (iOS does not resize the
  // layout viewport).
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

  // The attention count any app can badge (api.badges).
  const attention = resource(() => 'api/attention');
  $effect(() => { if (attention.data) setBadge('attention', attention.data.length ?? 0); });

  // One app, the last by plugin name as for session (JOB-85). Another
  // app replaces the built-in desk by taking its name or with desk
  // disabled (desk.plugins.desk.enabled: false). A failing app shows
  // PluginSlot's error card, not a blank page.
  const app = $derived(contributions('app').at(-1));
</script>

{#if !plugins.loaded}
  <p class="faint boot">Loading…</p>
{:else if app}
  {#key app.plugin.name + app.c.element}<PluginSlot plugin={app.plugin} element={app.c.element} />{/key}
{:else}
  <div class="card boot-error" data-testid="app-missing" role="alert">
    <b>Desk has no app.</b>
    {#if plugins.error}Plugins did not load: <span class="mono">{plugins.error}</span>{:else}No enabled plugin contributes the <span class="mono">app</span> slot; enable the built-in <span class="mono">desk</span> plugin or install one.{/if}
  </div>
{/if}
{#if toasts.items.length}
  <div class="toasts" data-testid="toasts">{#each toasts.items as t (t.id)}<div class="card toast">{t.msg}</div>{/each}</div>
{/if}
<!-- done takes resolve before clearing: c follows confirmBox.open and
     reads null after it, so OK did nothing (#603). -->
{#if confirmBox.open}
  {@const c = confirmBox.open}
  {@const done = (/** @type {boolean} */ v) => { const r = c.resolve; confirmBox.open = null; r(v); }}
  <div class="backdrop" role="presentation" onclick={() => done(false)}></div>
  <div class="card confirm" role="dialog" aria-modal="true" data-testid="confirm">
    <p>{c.msg}</p>
    <div class="row"><button onclick={() => done(false)}>Cancel</button><button class="ok" onclick={() => done(true)}>OK</button></div>
  </div>
{/if}

<style>
  .boot { padding: 16px 20px; }
  .boot-error { margin: 24px auto; max-width: 560px; padding: 16px; color: var(--bad); }
  .toasts { position: fixed; right: 16px; bottom: 16px; display: flex; flex-direction: column; gap: 6px; z-index: 50; }
  .toast { padding: 8px 12px; }
  .backdrop { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.4); z-index: 60; }
  .confirm { position: fixed; left: 50%; top: 30%; transform: translateX(-50%); padding: 16px; z-index: 61; min-width: 280px; }
  .confirm .row { display: flex; gap: 8px; justify-content: flex-end; }
  .confirm .ok { border-color: var(--accent); }
</style>
