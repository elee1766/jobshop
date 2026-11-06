<script>
  // The drawer frame (#541): fetches the ticket and hosts action and
  // drawer plugins. The details are the built-in ticket plugin, ordered
  // first; every section gets the same context.
  import { fields } from '../../kit/format.js';
  import { resource, slot } from '../../kit/plugin.svelte.js';
  import X from '@lucide/svelte/icons/x';

  /** @type {{ api: any, id: string }} */
  let { api, id } = $props();
  const v = resource(() => api, () => 'api/ticket/' + encodeURIComponent(id));
  const close = () => api.go({ ticket: '' });
  /** Match target for drawer and action plugins (#517). @param {any} t @param {any} run */
  const target = (t, run) => ({ kind: t.kind, label: t.labels ?? [], pool: t.pool || '', workflow: fields(run?.ticket ?? t).workflow || '' });
  /**
   * Built-ins first so the details lead.
   * @param {string} name @param {any} t @param {any} run
   * @returns {{ plugin: string, element: string, title?: string, source: string }[]}
   */
  const slotted = (name, t, run) =>
    api.contributions(name, target(t, run))
      .sort((/** @type {any} */ a, /** @type {any} */ b) => Number(b.source === 'builtin') - Number(a.source === 'builtin'));
  /** A plain copy: elements of other plugins run their own Svelte. @param {any} data */
  const ctx = (data) => $state.snapshot({ ticket: data.ticket, run: data.run, trace: data.trace, comments: data.comments, sessions: data.sessions });
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
      {@const actions = slotted('action', t, v.data.run)}
      {#if actions.length}
        <div class="row actions" data-testid="drawer-actions">
          {#each actions as a (a.plugin + a.element)}
            <div class="plugin-host" data-plugin={a.plugin} use:slot={{ api, slot: 'action', target: { plugin: a.plugin, element: a.element }, context: c }}></div>
          {/each}
        </div>
      {/if}
      {#each slotted('drawer', t, v.data.run) as d (d.plugin + d.element)}
        {#if d.title}<h3>{d.title}</h3>{/if}
        <div data-testid="drawer-plugin" data-plugin={d.plugin}>
          <div class="plugin-host" data-plugin={d.plugin} use:slot={{ api, slot: 'drawer', target: { plugin: d.plugin, element: d.element }, context: c }}></div>
        </div>
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
