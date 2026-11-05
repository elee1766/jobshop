<script>
  import { resource, post, live } from '../lib/api.svelte.js';
  import { ago } from '../lib/format.js';

  const pools = resource(() => 'api/pools');
  let actErr = $state('');

  /** @param {any} body */
  async function scale(body) {
    actErr = '';
    try { await post('api/pool/scale', body); live.tick++; } catch (e) { actErr = String(/** @type {any} */ (e).message || e); }
  }

  /** @param {any} sp */
  function setTarget(sp) {
    const v = window.prompt('Target sessions for ' + sp.name + ' (' + sp.min + '..' + sp.max + ')', String(sp.target ?? sp.desired));
    if (v === null || v.trim() === '') return;
    const n = Number(v);
    if (!Number.isInteger(n) || n < 0) { actErr = 'target must be a whole number'; return; }
    scale({ pool: sp.name, target: n });
  }
</script>

<h1>Pools and triggers</h1>
{#if pools.error}
  <div class="card err">{pools.error}</div>
{:else if pools.data}
  {@const q = Object.entries(pools.data.quarantined ?? {})}
  {@const p = Object.entries(pools.data.paused ?? {})}
  {#if pools.data.pools?.length}
    <h2>Scale{#if pools.data.host_max} <span class="faint">(host max {pools.data.host_max})</span>{/if}</h2>
    {#if actErr}<div class="card err">{actErr}</div>{/if}
    <div class="card scroll">
      <table data-testid="scale">
        <thead><tr><th>Pool</th><th>Desired</th><th>Open</th><th>Working</th><th>Idle</th><th class="hide-sm">Cordoned</th><th class="hide-sm">Min/max</th><th>Target</th><th></th></tr></thead>
        <tbody>
          {#each pools.data.pools as sp (sp.name)}
            <tr>
              <td class="mono">{sp.name}{#if sp.drained}<span class="pill warn">drained</span>{/if}</td>
              <td>{sp.desired}</td>
              <td>{sp.open}</td>
              <td>{sp.working}</td>
              <td>{sp.idle}</td>
              <td class="hide-sm">{sp.cordoned}</td>
              <td class="hide-sm muted">{sp.min}/{sp.max}</td>
              <td>{#if sp.target != null}<span class="pill accent">{sp.target}</span>{:else}<span class="muted">auto</span>{/if}</td>
              <td class="acts">
                <button onclick={() => setTarget(sp)}>Scale</button>
                {#if sp.target != null}<button onclick={() => scale({ pool: sp.name, auto: true })}>Auto</button>{/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
  <h2>Pool status</h2>
  <div class="card scroll">
    <table data-testid="pools">
      <tbody>
        {#each q as [name, why] (name)}
          <tr><td class="mono">{name}</td><td><span class="pill bad">quarantined</span></td><td class="muted">{why}</td></tr>
        {/each}
        {#each p as [name, until] (name)}
          <tr><td class="mono">{name}</td><td><span class="pill warn">paused</span></td><td class="muted">until {ago(String(until))}</td></tr>
        {/each}
        {#if q.length === 0 && p.length === 0}
          <tr><td class="faint">All pools start sessions.</td></tr>
        {/if}
      </tbody>
    </table>
  </div>
  <h2>Triggers</h2>
  <div class="card scroll">
    <table data-testid="triggers">
      <thead><tr><th>Trigger</th><th>Holds</th><th class="hide-sm">Why not / stalls</th><th>Last fired</th><th class="hide-sm">Next cron</th></tr></thead>
      <tbody>
        {#each pools.data.triggers as t (t.name)}
          <tr>
            <td class="mono">{t.name}{#if t.open}<div class="faint">open {t.open}</div>{/if}</td>
            <td><span class="pill {t.holds ? 'accent' : ''}">{t.holds ? 'fires next' : 'waiting'}</span></td>
            <td class="hide-sm muted">{t.stalls?.length ? t.stalls.join(', ') : t.why ?? ''}</td>
            <td class="muted">{ago(t.last_fired)}</td>
            <td class="hide-sm muted">{ago(t.next_cron)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  h1 { font-size: 18px; font-weight: 600; margin: 0 0 4px; }
  h2 { font-size: 13px; font-weight: 500; color: var(--muted); margin: 18px 0 8px; }
  .err { padding: 12px; color: var(--bad); }
  .acts { white-space: nowrap; }
  .acts button { font-size: 12px; margin-right: 4px; }
  .pill.warn { margin-left: 6px; }
</style>
