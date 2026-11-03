<script>
  import { url, go } from '../lib/url.svelte.js';
  import { resource } from '../lib/api.svelte.js';
  import { ago, fields, tone, links } from '../lib/format.js';

  const states = ['', 'running', 'needs_human', 'done', 'cancelled'];
  const work = resource(() => '/api/work' + (url.state ? '?state=' + encodeURIComponent(url.state) : ''));
</script>

<h1>Work</h1>
<div class="filters">
  {#each states as s (s)}
    <button class:on={url.state === s} onclick={() => go({ state: s })}>{s || 'all'}</button>
  {/each}
</div>

{#if work.error}
  <div class="card err">{work.error}</div>
{:else if work.data}
  {#if work.data.batches.length}
    <h2>Batches</h2>
    <div class="card scroll">
      <table>
        <tbody>
          {#each work.data.batches as b (b.id)}
            <tr class="click" onclick={() => go({ ticket: b.id })}>
              <td class="mono">{b.id}</td><td>{b.title}</td><td class="muted">{ago(b.updated_at)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
  <h2>Runs</h2>
  <div class="card scroll">
    <table data-testid="runs">
      <thead><tr><th>Run</th><th>State</th><th class="hide-sm">Workflow</th><th class="hide-sm">Links</th><th>Updated</th></tr></thead>
      <tbody>
        {#each work.data.runs as r (r.id)}
          {@const f = fields(r)}
          <tr class="click" onclick={() => go({ ticket: r.id })}>
            <td>{r.title}<div class="faint mono">{r.id} · {r.project}</div></td>
            <td><span class="pill {tone(f.state)}">{f.state || r.status}</span></td>
            <td class="hide-sm muted">{f.workflow || ''}</td>
            <td class="hide-sm">
              {#each links(r) as l (l.href)}<a href={l.href} target="_blank" rel="noopener noreferrer" onclick={(e) => e.stopPropagation()}>{l.label}</a> {/each}
            </td>
            <td class="muted">{ago(r.updated_at)}</td>
          </tr>
        {:else}
          <tr><td colspan="5" class="faint">No runs.</td></tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  h1 { font-size: 18px; font-weight: 600; margin: 0 0 12px; }
  h2 { font-size: 13px; font-weight: 500; color: var(--muted); margin: 18px 0 8px; }
  .filters { display: flex; gap: 6px; flex-wrap: wrap; }
  .filters .on { background: var(--accent-soft); border-color: var(--accent); }
  .err { padding: 12px; color: var(--bad); }
</style>
