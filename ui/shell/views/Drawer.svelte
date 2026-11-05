<script>
  // The drawer frame (#541): fetches the ticket and hosts action and
  // drawer plugins. The details are the built-in ticket plugin, ordered
  // first; every section gets the same context.
  import { go } from '../lib/url.svelte.js';
  import { resource } from '../lib/api.svelte.js';
  import { fields } from '../../kit/format.js';
  import X from '@lucide/svelte/icons/x';
  import PluginSlot from './PluginSlot.svelte';
  import { contributions, matches } from '../lib/plugins.svelte.js';

  /** @type {{ id: string }} */
  let { id } = $props();
  const v = resource(() => 'api/ticket/' + encodeURIComponent(id));
  const close = () => go({ ticket: '' });
  /** Match target for drawer and action plugins (#517). @param {any} t @param {any} run */
  const target = (t, run) => ({ kind: t.kind, label: t.labels ?? [], pool: t.pool || '', workflow: fields(run?.ticket ?? t).workflow || '' });
  /** Built-ins first so the details lead. @param {string} slot @param {any} t @param {any} run */
  const slotted = (slot, t, run) =>
    contributions(slot)
      .filter((x) => matches(x.c.match, target(t, run)))
      .sort((a, b) => Number(b.plugin.source === 'builtin') - Number(a.plugin.source === 'builtin'));
  /** @param {any} data */
  const ctx = (data) => ({ ticket: data.ticket, run: data.run, trace: data.trace, comments: data.comments });
</script>

<svelte:window onkeydown={(e) => { if (e.key === 'Escape') { e.preventDefault(); close(); } }} />

<div class="backdrop" onclick={close} role="presentation"></div>
<aside class="drawer" data-testid="drawer" aria-label="ticket">
  <header>
    <span class="mono muted id">{id}</span>
    <button class="ghost icon" onclick={close} aria-label="close"><X size={14} /></button>
  </header>
  <div class="scrollbody">
    {#if v.error}
      <div class="err">{v.error}</div>
    {:else if v.data}
      {@const t = v.data.ticket}
      {@const c = ctx(v.data)}
      {#if slotted('action', t, v.data.run).length}
        <div class="row actions" data-testid="drawer-actions">
          {#each slotted('action', t, v.data.run) as a (a.plugin.name + a.c.element)}
            <PluginSlot plugin={a.plugin} element={a.c.element} context={c} />
          {/each}
        </div>
      {/if}
      {#each slotted('drawer', t, v.data.run) as d (d.plugin.name + d.c.element)}
        {#if d.c.title}<h3>{d.c.title}</h3>{/if}
        <div data-testid="drawer-plugin" data-plugin={d.plugin.name}><PluginSlot plugin={d.plugin} element={d.c.element} context={c} /></div>
      {:else}
        <p class="faint">No drawer plugins are enabled.</p>
      {/each}
    {:else}
      <p class="faint">Loading…</p>
    {/if}
  </div>
</aside>

<style>
  .backdrop { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.45); z-index: 30; }
  .drawer {
    position: fixed; top: 0; right: 0; bottom: 0; width: min(620px, 100vw); z-index: 31; display: flex; flex-direction: column;
    background: var(--panel); border-left: 1px solid var(--border); box-shadow: var(--shadow);
  }
  header {
    flex: none; display: flex; justify-content: space-between; align-items: center; min-height: 44px; padding: 0 10px 0 18px;
    border-bottom: 1px solid var(--border);
  }
  .id { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .scrollbody { flex: 1; min-height: 0; overflow-y: auto; overscroll-behavior: contain; padding: 4px 18px 40px; }
  h3 { font-size: 11px; font-weight: 500; color: var(--faint); margin: 20px 0 6px; text-transform: uppercase; letter-spacing: 0.05em; }
  .row { display: flex; flex-wrap: wrap; gap: 6px 10px; align-items: center; margin: 6px 0; }
  .err { color: var(--bad); padding: 12px 0; }

  @media (max-width: 768px) {
    .drawer { width: 100vw; border-left: 0; padding-top: env(safe-area-inset-top); }
    .backdrop { display: none; }
    header { min-height: 52px; padding: 0 6px 0 14px; }
    .scrollbody { padding: 2px 14px calc(32px + env(safe-area-inset-bottom)); }
  }
</style>
