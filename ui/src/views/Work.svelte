<script>
  import { url, go } from '../lib/url.svelte.js';
  import { resource } from '../lib/api.svelte.js';
  import { ago, fields, tone, links } from '../lib/format.js';

  const states = ['', 'running', 'needs_human', 'done', 'cancelled'];
  const work = resource(() => 'api/work' + (url.state ? '?state=' + encodeURIComponent(url.state) : ''));
</script>

<header class="vh">
  <h1>Work</h1>
  <div class="chips actions" role="group" aria-label="state filter">
    {#each states as s (s)}
      <button class:on={url.state === s} onclick={() => go({ state: s })}>{s ? s.replace('_', ' ') : 'all'}</button>
    {/each}
  </div>
</header>
<div class="pad">
  {#if work.error}
    <div class="card err">{work.error}</div>
  {:else if work.data}
    {#if work.data.batches.length}
      <h2 class="st">Batches</h2>
      <div class="card scroll">
        <table class="stack">
          <tbody>
            {#each work.data.batches as b (b.id)}
              <tr class="click" onclick={() => go({ ticket: b.id })}>
                <td class="mono c-id">{b.id}</td><td class="c-title">{b.title}</td><td class="muted c-time">{ago(b.updated_at)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
    <h2 class="st">Runs</h2>
    <div class="card scroll">
      <table class="stack" data-testid="runs">
        <thead><tr><th>Run</th><th>State</th><th>Workflow</th><th>Links</th><th>Updated</th></tr></thead>
        <tbody>
          {#each work.data.runs as r (r.id)}
            {@const f = fields(r)}
            <tr class="click" onclick={() => go({ ticket: r.id })}>
              <td class="c-title">{r.title}<div class="faint mono">{r.id} · {r.project}</div></td>
              <td class="c-id"><span class="pill {tone(f.state)}">{f.state || r.status}</span></td>
              <td class="muted">{f.workflow || ''}</td>
              <td class="links">
                {#each links(r) as l (l.href)}<a href={l.href} target="_blank" rel="noopener noreferrer" onclick={(e) => e.stopPropagation()}>{l.label}</a> {/each}
              </td>
              <td class="muted c-time">{ago(r.updated_at)}</td>
            </tr>
          {:else}
            <tr><td colspan="5" class="faint">No runs.</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .vh .chips { min-width: 0; }
  .links a { display: inline-block; padding: 2px 0; margin-right: 8px; }
  @media (max-width: 768px) {
    .links a { padding: 8px 0; }
  }
</style>
