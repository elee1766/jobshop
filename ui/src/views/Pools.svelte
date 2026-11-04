<script>
  import { resource, post, live } from '../lib/api.svelte.js';
  import { ago } from '../lib/format.js';

  const pools = resource(() => '/api/pools');
  let actErr = $state('');

  /** @param {any} body */
  async function scale(body) {
    actErr = '';
    try { await post('/api/pool/scale', body); live.tick++; } catch (e) { actErr = String(/** @type {any} */ (e).message || e); }
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

<header class="vh"><h1>Pools and triggers</h1></header>
<div class="pad">
  {#if pools.error}
    <div class="card err">{pools.error}</div>
  {:else if pools.data}
    {@const q = Object.entries(pools.data.quarantined ?? {})}
    {@const p = Object.entries(pools.data.paused ?? {})}
    {#if pools.data.pools?.length}
      <h2 class="st">Scale{#if pools.data.host_max} <span class="faint">(host max {pools.data.host_max})</span>{/if}</h2>
      {#if actErr}<div class="card err">{actErr}</div>{/if}
      <div class="card scroll">
        <table class="stack" data-testid="scale">
          <thead><tr><th>Pool</th><th>Desired</th><th>Open</th><th>Working</th><th>Idle</th><th>Cordoned</th><th>Min/max</th><th>Target</th><th></th></tr></thead>
          <tbody>
            {#each pools.data.pools as sp (sp.name)}
              <tr>
                <td class="mono c-id">{sp.name}{#if sp.drained}<span class="pill warn gap">drained</span>{/if}</td>
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
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
    <h2 class="st">Pool status</h2>
    <div class="card scroll">
      <table class="stack" data-testid="pools">
        <tbody>
          {#each q as [name, why] (name)}
            <tr><td class="mono c-id">{name}</td><td class="c-st"><span class="pill bad">quarantined</span></td><td class="muted full">{why}</td></tr>
          {/each}
          {#each p as [name, until] (name)}
            <tr><td class="mono c-id">{name}</td><td class="c-st"><span class="pill warn">paused</span></td><td class="muted full">until {ago(String(until))}</td></tr>
          {/each}
          {#if q.length === 0 && p.length === 0}
            <tr><td class="faint">All pools start sessions.</td></tr>
          {/if}
        </tbody>
      </table>
    </div>
    <h2 class="st">Triggers</h2>
    <div class="card scroll">
      <table class="stack" data-testid="triggers">
        <thead><tr><th>Trigger</th><th>Holds</th><th>Why not / stalls</th><th>Last fired</th><th>Next cron</th></tr></thead>
        <tbody>
          {#each pools.data.triggers as t (t.name)}
            <tr>
              <td class="mono c-id">{t.name}{#if t.open}<div class="faint">open {t.open}</div>{/if}</td>
              <td class="c-st"><span class="pill {t.holds ? 'accent' : ''}">{t.holds ? 'fires next' : 'waiting'}</span></td>
              <td class="muted full">{t.stalls?.length ? t.stalls.join(', ') : t.why ?? ''}</td>
              <td class="muted" data-label="last fired">{ago(t.last_fired)}</td>
              <td class="muted" data-label="next cron">{ago(t.next_cron)}</td>
            </tr>
          {:else}
            <tr><td class="faint">No triggers.</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .acts { white-space: nowrap; }
  .acts button { height: 24px; margin-right: 4px; }
  .gap { margin-left: 6px; }
  @media (max-width: 768px) {
    .acts button { height: 40px; padding: 0 14px; }
  }
</style>
