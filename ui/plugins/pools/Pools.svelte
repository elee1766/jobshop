<script>
  import { resource, params, errText } from '../../kit/plugin.svelte.js';
  import { ago } from '../../kit/format.js';
  import { poolRows, groupKeys, filterRows, labelCounts, poolGroups, triggerGroups } from './groups.js';

  /** @type {{ api: any, context?: any, config?: any }} */
  let { api } = $props();
  const url = params(() => api);
  const go = (/** @type {Record<string, string>} */ p) => api.go(p);
  const pools = resource(() => api, () => 'api/pools');
  let actErr = $state('');

  // State lives in the URL so a link reproduces the view: group (the
  // grouping key), q (name search), label (comma-joined selected labels)
  // and open (comma-joined groups expanded past their quiet pools).
  const group = $derived(url.group || 'instance');
  const q = $derived(url.q ?? '');
  const selected = $derived(url.label ? url.label.split(',') : []);
  const open = $derived(new Set(url.open ? url.open.split(',') : []));

  const rows = $derived(pools.data ? poolRows(pools.data) : []);
  const triggers = $derived(pools.data?.triggers ?? []);
  const keys = $derived(groupKeys([...rows, ...triggers]));
  const labels = $derived(labelCounts([...rows, ...triggers]));
  const shownRows = $derived(filterRows(rows, q, selected));
  const shownTriggers = $derived(filterRows(triggers, q, selected));
  const pg = $derived(poolGroups(shownRows, group));
  const tg = $derived(triggerGroups(shownTriggers, group));
  const idleTotal = $derived(rows.filter((p) => !p.desired && !p.open && !p.working && !p.quarantined && !p.paused && !p.drained && p.target == null).length);

  // A group is expanded when its key is in open. Ungrouped rows have the
  // one group named '', keyed '_'. Pool groups key as is, trigger groups
  // as t:<name>, so one URL holds both.
  /** @param {string} key */
  function toggle(key) {
    const next = new Set(open);
    if (!next.delete(key)) next.add(key);
    go({ open: [...next].join(',') });
  }
  /** @param {string} name @param {string} [pre] */
  const keyOf = (name, pre = '') => pre + (name || '_');

  /** @param {string} l */
  function toggleLabel(l) {
    const next = selected.includes(l) ? selected.filter((x) => x !== l) : [...selected, l];
    go({ label: next.join(',') });
  }

  /** @param {any} body */
  async function scale(body) {
    actErr = '';
    try { await api.post('api/pool/scale', body); pools.refresh(); } catch (e) { actErr = errText(e); }
  }

  /** @param {any} sp */
  function setTarget(sp) {
    const v = window.prompt('Target sessions for ' + sp.name + ' (' + sp.min + '..' + sp.max + ')', String(sp.target ?? sp.desired));
    if (v === null || v.trim() === '') return;
    const n = Number(v);
    if (!Number.isInteger(n) || n < 0) { actErr = 'target must be a whole number'; return; }
    scale({ pool: sp.name, target: n });
  }

  /** Returns every overridden pool of the group to auto. @param {any[]} list */
  async function autoGroup(list) {
    actErr = '';
    try {
      for (const sp of list) if (sp.target != null) await api.post('api/pool/scale', { pool: sp.name, auto: true });
      pools.refresh();
    } catch (e) { actErr = errText(e); }
  }
</script>

<header class="vh">
  <h1>Pools and triggers</h1>
  {#if pools.data}<span class="sub">{rows.length} pools, {idleTotal} idle{#if pools.data.host_max} · host max {pools.data.host_max}{/if}</span>{/if}
  <div class="actions">
    <label class="faint" for="pool-group">group by</label>
    <select id="pool-group" data-testid="group-by" value={group} onchange={(e) => go({ group: e.currentTarget.value === 'instance' ? '' : e.currentTarget.value, open: '' })}>
      {#each keys as k (k)}<option value={k}>{k}</option>{/each}
    </select>
  </div>
</header>
<div class="pad">
  {#if pools.error}
    <div class="card err">{pools.error}</div>
  {:else if pools.data}
    <div class="bar">
      <input type="text" data-testid="pool-search" placeholder="search pools and triggers" aria-label="search by name" value={q}
        onchange={(e) => go({ q: e.currentTarget.value.trim() })} />
      {#if q || selected.length}<button data-testid="clear-filters" onclick={() => go({ q: '', label: '' })}>Clear</button>{/if}
    </div>
    {#if labels.length}
      <div class="chips lbl" role="group" aria-label="label filter" data-testid="label-chips">
        {#each labels as [l, n] (l)}
          <button class:on={selected.includes(l)} onclick={() => toggleLabel(l)}>{l} <span class="faint">{n}</span></button>
        {/each}
      </div>
    {/if}
    {#if actErr}<div class="card err">{actErr}</div>{/if}

    <h2 class="st">Pools</h2>
    <div class="card scroll">
      <table class="stack" data-testid="scale">
        <thead><tr><th>{group === 'none' ? 'Pool' : group + ' / pool'}</th><th>Desired</th><th>Open</th><th>Working</th><th>Idle</th><th>Cordoned</th><th>Min/max</th><th>Target</th><th></th></tr></thead>
        {#each pg as g (g.name)}
          {@const key = keyOf(g.name)}
          {@const expanded = open.has(key)}
          <tbody data-testid="pool-group" data-group={g.name}>
            {#if group !== 'none'}
              <tr class="grp click" onclick={() => toggle(key)} aria-expanded={expanded}>
                <td class="mono c-id"><span class="caret">{expanded || !g.folded.length ? '▾' : '▸'}</span> {g.name}
                  <span class="faint gap">{g.pools.length} pool{g.pools.length === 1 ? '' : 's'}</span>
                  {#if g.quiet}<span class="faint gap" data-testid="idle-note">{g.pools.length} idle</span>{/if}
                  {#if g.quarantined}<span class="pill bad gap">{g.quarantined} quarantined</span>{/if}
                  {#if g.paused}<span class="pill warn gap">{g.paused} paused</span>{/if}
                  {#if g.drained}<span class="pill warn gap">{g.drained} drained</span>{/if}
                  {#if g.targeted}<span class="pill accent gap">{g.targeted} targeted</span>{/if}
                </td>
                <td data-label="desired">{g.desired}</td>
                <td data-label="open">{g.open}</td>
                <td data-label="working">{g.working}</td>
                <td data-label="idle">{g.idle}</td>
                <td></td><td></td><td></td>
                <td class="c-act acts">
                  {#if g.targeted}<button onclick={(e) => { e.stopPropagation(); autoGroup(g.pools); }}>Auto all</button>{/if}
                </td>
              </tr>
            {/if}
            {#each expanded ? g.pools : g.shown as sp (sp.name)}
              {@render poolRow(sp)}
            {/each}
            <!-- A quiet group is already one summary row; its header expands it. -->
            {#if !expanded && g.folded.length && (!g.quiet || group === 'none')}
              <tr class="click fold" onclick={() => toggle(key)} data-testid="fold"><td class="faint" colspan="9">{g.folded.length} idle pool{g.folded.length === 1 ? '' : 's'}</td></tr>
            {/if}
          </tbody>
        {:else}
          <tbody><tr><td class="faint" colspan="9">{rows.length ? 'No pools match.' : 'No ticket pools.'}</td></tr></tbody>
        {/each}
      </table>
    </div>

    <h2 class="st">Triggers</h2>
    <div class="card scroll">
      <table class="stack" data-testid="triggers">
        <thead><tr><th>Trigger</th><th>Holds</th><th>Why not / stalls</th><th>Last fired</th><th>Next cron</th></tr></thead>
        {#each tg as g (g.name)}
          {@const key = keyOf(g.name, 't:')}
          {@const expanded = open.has(key)}
          <tbody data-testid="trigger-group" data-group={g.name}>
            {#if group !== 'none'}
              <tr class="grp click" onclick={() => toggle(key)} aria-expanded={expanded}>
                <td class="mono c-id" colspan="5"><span class="caret">{expanded || !g.folded.length ? '▾' : '▸'}</span> {g.name}
                  <span class="faint gap">{g.triggers.length} trigger{g.triggers.length === 1 ? '' : 's'}</span>
                  {#if g.stalled}<span class="pill bad gap">{g.stalled} stalled</span>{/if}
                  {#if g.firing}<span class="pill accent gap">{g.firing} firing</span>{/if}
                  {#if g.quiet}<span class="faint gap">all waiting</span>{/if}
                </td>
              </tr>
            {/if}
            {#each expanded ? g.triggers : g.shown as t (t.name)}{@render triggerRow(t)}{/each}
            {#if !expanded && g.folded.length && (!g.quiet || group === 'none')}
              <tr class="click fold" onclick={() => toggle(key)}><td class="faint" colspan="5">{g.folded.length} waiting</td></tr>
            {/if}
          </tbody>
        {:else}
          <tbody><tr><td class="faint" colspan="5">{triggers.length ? 'No triggers match.' : 'No triggers.'}</td></tr></tbody>
        {/each}
      </table>
    </div>
  {/if}
</div>

{#snippet poolRow(/** @type {any} */ sp)}
  <tr data-testid="pool-row" data-pool={sp.name}>
    <td class="mono c-id ind">{sp.name}
      {#if sp.quarantined}<span class="pill bad gap">quarantined</span>{/if}
      {#if sp.paused}<span class="pill warn gap">paused</span>{/if}
      {#if sp.drained}<span class="pill warn gap">drained</span>{/if}
      {#if sp.why}<div class="muted why">{sp.why}</div>{/if}
      {#if sp.until}<div class="muted why">until {ago(String(sp.until))}</div>{/if}
    </td>
    <td data-label="desired">{sp.desired}</td>
    <td data-label="open">{sp.open}</td>
    <td data-label="working">{sp.working}</td>
    <td data-label="idle">{sp.idle}</td>
    <td data-label="cordoned">{sp.cordoned}</td>
    <td class="muted" data-label="min/max">{sp.min}/{sp.max}</td>
    <td class="c-st">{#if sp.target != null}<span class="pill accent">target {sp.target}</span>{:else}<span class="muted">auto</span>{/if}</td>
    <td class="c-act acts">
      <button onclick={() => setTarget(sp)}>Scale</button>
      {#if sp.target != null}<button onclick={() => scale({ pool: sp.name, auto: true })}>Auto</button>{/if}
    </td>
  </tr>
{/snippet}

{#snippet triggerRow(/** @type {any} */ t)}
  <tr data-testid="trigger-row" data-trigger={t.name}>
    <td class="mono c-id ind">{t.name}{#if t.open}<div class="faint">open {t.open}</div>{/if}</td>
    <td class="c-st"><span class="pill {t.stalls?.length ? 'bad' : t.holds ? 'accent' : ''}">{t.stalls?.length ? 'stalled' : t.holds ? 'fires next' : 'waiting'}</span></td>
    <td class="muted full">{t.stalls?.length ? t.stalls.join(', ') : t.why ?? ''}</td>
    <td class="muted" data-label="last fired">{ago(t.last_fired)}</td>
    <td class="muted" data-label="next cron">{ago(t.next_cron)}</td>
  </tr>
{/snippet}

<style>
  .bar { display: flex; gap: 6px; margin-bottom: 8px; }
  .bar input { flex: 1; max-width: 360px; }
  .lbl { margin-bottom: 8px; flex-wrap: wrap; overflow: visible; }
  .acts { white-space: nowrap; }
  .acts button { height: 24px; margin-right: 4px; }
  .gap { margin-left: 6px; }
  .grp { background: var(--hover); }
  .grp td { font-weight: 500; }
  .caret { display: inline-block; width: 1em; color: var(--faint); }
  .ind { padding-left: 28px; }
  .why { font-family: var(--sans, inherit); max-width: 60ch; }
  .fold td { cursor: pointer; padding-left: 28px; }
  @media (max-width: 768px) {
    .acts button { height: 40px; padding: 0 14px; }
    .ind { padding-left: 0; }
  }
</style>
