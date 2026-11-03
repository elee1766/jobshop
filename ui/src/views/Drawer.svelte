<script>
  import { go } from '../lib/url.svelte.js';
  import { resource } from '../lib/api.svelte.js';
  import { ago, fields, tone, links } from '../lib/format.js';
  import Trace from './Trace.svelte';

  /** @type {{ id: string }} */
  let { id } = $props();
  const v = resource(() => '/api/ticket/' + encodeURIComponent(id));
  const close = () => go({ ticket: '' });
</script>

<svelte:window onkeydown={(e) => { if (e.key === 'Escape') close(); }} />

<div class="backdrop" onclick={close} role="presentation"></div>
<aside class="drawer" data-testid="drawer" aria-label="ticket">
  <header>
    <span class="mono muted">{id}</span>
    <button onclick={close} aria-label="close">✕</button>
  </header>
  {#if v.error}
    <div class="err">{v.error}</div>
  {:else if v.data}
    {@const t = v.data.ticket}
    {@const f = fields(t)}
    <h2>{t.title}</h2>
    <div class="row">
      <span class="pill">{t.kind}</span>
      <span class="pill {tone(f.state || t.status)}">{f.state || t.status}{t.resolution ? ' · ' + t.resolution : ''}</span>
      {#if t.pool}<span class="pill">{t.pool}</span>{/if}
      {#each t.labels ?? [] as l (l)}<span class="pill accent">{l}</span>{/each}
    </div>
    <div class="row">
      {#each links(t) as l (l.href)}<a href={l.href} target="_blank" rel="noopener noreferrer">{l.label}</a>{/each}
      {#if t.holder}<span class="muted">held by <span class="mono">{t.holder}</span></span>{/if}
      <span class="faint">updated {ago(t.updated_at)}</span>
    </div>
    {#if t.parent_id}
      <div class="row"><span class="muted">parent</span> <a href={'?ticket=' + t.parent_id} class="mono" onclick={(e) => { e.preventDefault(); go({ ticket: t.parent_id }); }}>{t.parent_id}</a></div>
    {/if}

    {#if v.data.run}
      <h3>Steps</h3>
      <ol class="steps">
        {#each v.data.run.steps as s (s.id)}
          {@const sf = fields(s)}
          <li>
            <a href={'?ticket=' + s.id} onclick={(e) => { e.preventDefault(); go({ ticket: s.id }); }}>{sf.step || s.title}</a>
            <span class="pill {tone(s.status)}">{s.status}{s.resolution ? ' · ' + s.resolution : ''}</span>
            {#if sf.failures}<span class="pill bad">{sf.failures} failures</span>{/if}
          </li>
        {/each}
      </ol>
    {/if}

    {#if t.body}
      <h3>Body</h3>
      <pre class="body">{t.body}</pre>
    {/if}

    {#if v.data.trace}
      <h3>Trace</h3>
      <Trace node={v.data.trace} />
    {/if}

    {#if v.data.comments?.length}
      <h3>Comments</h3>
      {#each v.data.comments as c (c.id)}
        <div class="comment">
          <div class="faint"><span class="mono">{c.author}</span> · {ago(c.created_at)}</div>
          <pre>{c.body}</pre>
        </div>
      {/each}
    {/if}
  {:else}
    <p class="faint">Loading…</p>
  {/if}
</aside>

<style>
  .backdrop { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.35); z-index: 10; }
  .drawer {
    position: fixed; top: 0; right: 0; bottom: 0; width: min(620px, 100vw); z-index: 11;
    background: var(--panel); border-left: 1px solid var(--border); overflow-y: auto; padding: 14px 18px 40px;
  }
  header { display: flex; justify-content: space-between; align-items: center; }
  h2 { font-size: 16px; font-weight: 600; margin: 10px 0 8px; }
  h3 { font-size: 12px; font-weight: 500; color: var(--muted); margin: 18px 0 6px; text-transform: uppercase; letter-spacing: 0.04em; }
  .row { display: flex; flex-wrap: wrap; gap: 6px 10px; align-items: center; margin: 6px 0; }
  .steps { margin: 0; padding-left: 20px; }
  .steps li { margin: 4px 0; }
  .body { background: var(--bg); border: 1px solid var(--border); border-radius: 6px; padding: 10px; max-height: 320px; overflow: auto; }
  .comment { border-left: 2px solid var(--border); padding: 4px 10px; margin: 8px 0; }
  .err { color: var(--bad); padding: 12px 0; }
</style>
