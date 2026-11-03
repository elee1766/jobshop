<script>
  import { resource } from '../lib/api.svelte.js';
  import { ago } from '../lib/format.js';

  const pools = resource(() => '/api/pools');
</script>

<h1>Pools and triggers</h1>
{#if pools.error}
  <div class="card err">{pools.error}</div>
{:else if pools.data}
  {@const q = Object.entries(pools.data.quarantined ?? {})}
  {@const p = Object.entries(pools.data.paused ?? {})}
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
</style>
