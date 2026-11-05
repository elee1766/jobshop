<script>
  // One plugin element in a slot (#517). Imports the plugin, creates its
  // element, sets api/context/config, and shows an error card in place
  // of a plugin that fails to import, define or render.
  import { deskApi, loadPlugin, whenDefined } from '../lib/plugins.svelte.js';

  /** @type {{ plugin: import('../lib/plugins.svelte.js').PluginInfo, element: string, context?: any }} */
  let { plugin, element, context = {} } = $props();

  /** @type {HTMLElement | undefined} */
  let host = $state();
  let error = $state('');
  /** @type {any} */
  let el = null;

  $effect(() => {
    const target = host;
    const p = plugin;
    const name = element;
    if (!target) return;
    let gone = false;
    error = '';
    /** @param {ErrorEvent} e */
    const onErr = (e) => {
      // Errors thrown from the plugin's own module or element.
      if (e.filename && e.filename.includes('/' + p.url)) error = String(e.message || e.error);
    };
    window.addEventListener('error', onErr);
    loadPlugin(p)
      .then(() => whenDefined(name))
      .then(() => {
        if (gone) return;
        el = document.createElement(name);
        el.api = deskApi;
        el.config = p.config ?? {};
        el.context = $state.snapshot(context);
        target.replaceChildren(el);
      })
      .catch((e) => { if (!gone) error = String(e?.message || e); });
    return () => {
      gone = true;
      window.removeEventListener('error', onErr);
      target.replaceChildren();
      el = null;
    };
  });

  // A new context (another ticket in the drawer) is set on the live element.
  $effect(() => {
    const c = $state.snapshot(context);
    if (el) el.context = c;
  });
</script>

{#if error}
  <div class="card plugin-error" data-testid="plugin-error" data-plugin={plugin.name}>
    <b>Plugin {plugin.name}</b> failed: <span class="mono">{error}</span>
    <div class="faint">from {plugin.source}</div>
  </div>
{/if}
<div bind:this={host} class="plugin-host" class:hidden={!!error} data-plugin={plugin.name}></div>

<style>
  .plugin-error { padding: 12px; color: var(--bad); margin: 6px 0; }
  .hidden { display: none; }
</style>
