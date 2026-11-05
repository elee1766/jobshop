<script>
  import { url, go } from '../lib/url.svelte.js';
  import { resource, post, api } from '../lib/api.svelte.js';
  import { ago, bytes, markdown } from '../lib/format.js';
  import PluginSlot from './PluginSlot.svelte';
  import { contributions, matches } from '../lib/plugins.svelte.js';

  // Artifacts (#508): deliverables from agents. Unread first; `a` acks
  // and moves to the next, j/k move, o opens the source ticket.

  const list = resource(() => {
    const q = new URLSearchParams();
    if (!url.read) q.set('unread', '1');
    if (url.trigger) q.set('trigger', url.trigger);
    if (url.pool) q.set('pool', url.pool);
    return 'api/artifacts?' + q.toString();
  });
  const items = $derived(/** @type {any[]} */ (list.data ?? []));
  const current = $derived(url.artifact || items[0]?.ticket.id || '');
  // With nothing selected this re-reads the list; `shown` ignores it.
  const one = resource(() => (current ? 'api/artifacts/' + encodeURIComponent(current) : 'api/artifacts?unread=1'));
  const shown = $derived(current && one.data?.ticket?.id === current ? one.data : null);
  let ackError = $state('');

  /** @param {number} d */
  function move(d) {
    const i = items.findIndex((a) => a.ticket.id === current);
    const next = items[Math.min(Math.max(i + d, 0), items.length - 1)];
    if (next) go({ artifact: next.ticket.id });
  }

  async function ack() {
    if (!shown || shown.read) return;
    const i = items.findIndex((a) => a.ticket.id === current);
    const next = items[i + 1] ?? items[i - 1];
    ackError = '';
    try {
      await post('api/artifacts/' + encodeURIComponent(current) + '/ack', {});
      go({ artifact: url.read ? current : next && next.ticket.id !== current ? next.ticket.id : '' });
    } catch (e) {
      ackError = String(/** @type {any} */ (e).message || e);
    }
  }

  function openSource() {
    if (shown?.artifact.ticket) go({ ticket: shown.artifact.ticket });
  }

  /** @param {KeyboardEvent} e */
  function key(e) {
    const t = /** @type {HTMLElement} */ (e.target);
    if (url.ticket || e.metaKey || e.ctrlKey || e.altKey || t.closest('input, textarea, select')) return;
    if (e.key === 'j') move(1);
    else if (e.key === 'k') move(-1);
    else if (e.key === 'a') ack();
    else if (e.key === 'o') openSource();
  }

  /** @param {any} f */
  const isImage = (f) => !f.pruned && /^image\/(png|jpeg|gif|webp)$/.test(f.type || '');
  /** @param {any} f */
  const isText = (f) => !f.pruned && f.size <= 256 * 1024 && (/^text\//.test(f.type || '') || f.type === 'application/json' || !f.type);
  /** @param {string} name */
  const fileURL = (name) => api('api/artifacts/' + encodeURIComponent(current) + '/files/' + encodeURIComponent(name));

  /** @param {string} name */
  async function text(name) {
    const r = await fetch(fileURL(name));
    if (!r.ok) throw new Error(r.status + ' ' + r.statusText);
    return r.text();
  }

  /** @param {any} a */
  const source = (a) => a.artifact.trigger || a.artifact.pool || '';

  /** An artifact plugin (#517) matched by MIME type or extension. @param {any} f */
  const renderer = (f) => {
    if (f.pruned) return null;
    const ext = (f.name.match(/\.([^.]+)$/)?.[1] || '').toLowerCase();
    return contributions('artifact').find((x) => matches(x.c.match, { type: (f.type || '').split(';')[0], ext })) ?? null;
  };
</script>

<svelte:window onkeydown={key} />

<h1>Artifacts</h1>
<p class="muted sub">Reports and files agents delivered. <span class="faint">j/k move · a ack and next · o open source</span></p>

<div class="filters">
  <input placeholder="trigger" value={url.trigger} onchange={(e) => go({ trigger: e.currentTarget.value, artifact: '' })} />
  <input placeholder="pool" value={url.pool} onchange={(e) => go({ pool: e.currentTarget.value, artifact: '' })} />
  <label><input type="checkbox" checked={!!url.read} onchange={(e) => go({ read: e.currentTarget.checked ? '1' : '', artifact: '' })} /> show read</label>
</div>

{#if list.error}
  <div class="card err">{list.error}</div>
{:else if list.loading && !list.data}
  <p class="faint">Loading…</p>
{:else if items.length === 0}
  <div class="card empty" data-testid="artifacts-empty">
    <div class="big">Nothing to read.</div>
    <div class="muted">No {url.read ? '' : 'unread '}artifacts.</div>
  </div>
{:else}
  <div class="split">
    <div class="card scroll list" data-testid="artifact-list">
      {#each items as a (a.ticket.id)}
        <button class="item" class:active={a.ticket.id === current} class:read={a.read} onclick={() => go({ artifact: a.ticket.id })}>
          <div class="title">{a.ticket.title}</div>
          <div class="faint">
            {#if source(a)}{source(a)} · {/if}{ago(a.ticket.created_at)}
            {#if a.artifact.files?.length} · {a.artifact.files.length} file{a.artifact.files.length > 1 ? 's' : ''}{/if}
            {#if a.read} · read{/if}
          </div>
        </button>
      {/each}
    </div>
    <article class="card pane" data-testid="artifact-pane">
      {#if one.error && current}
        <div class="err">{one.error}</div>
      {:else if shown}
        {@const m = shown.artifact}
        <header>
          <h2>{shown.ticket.title}</h2>
          {#if shown.read}
            <span class="pill ok">read</span>
          {:else}
            <button data-testid="artifact-ack" onclick={ack}>Ack</button>
          {/if}
        </header>
        <div class="row faint">
          <span class="mono">{shown.ticket.id}</span>
          <span>{ago(shown.ticket.created_at)}</span>
          {#if m.ticket}<a href={'?ticket=' + m.ticket} onclick={(e) => { e.preventDefault(); go({ ticket: m.ticket }); }}>source {m.ticket}</a>{/if}
          {#if m.run}<a href={'?ticket=' + m.run} onclick={(e) => { e.preventDefault(); go({ ticket: m.run }); }}>run {m.run}</a>{/if}
          {#if m.trigger}<a href={'?view=artifacts&read=1&trigger=' + encodeURIComponent(m.trigger)} onclick={(e) => { e.preventDefault(); go({ trigger: m.trigger, read: '1', artifact: '' }); }}>trigger {m.trigger}</a>{/if}
          {#if m.pool}<span>pool {m.pool}</span>{/if}
        </div>
        {#if ackError}<div class="err">{ackError}</div>{/if}
        <!-- Sanitized by markdown() in format.js; see the comment there. -->
        <div class="md" data-testid="artifact-body">{@html markdown(shown.ticket.body)}</div>
        {#if m.files?.length}
          <h3>Attachments</h3>
          {#each m.files as f (f.name)}
            <div class="file">
              <div class="row">
                {#if f.pruned}
                  <span class="mono">{f.name}</span><span class="faint">pruned after retention</span>
                {:else}
                  <a href={fileURL(f.name)} download={f.name} class="mono">{f.name}</a>
                {/if}
                <span class="faint">{bytes(f.size)}</span>
              </div>
              {#if renderer(f)}
                {@const r = renderer(f)}
                {#if r}<PluginSlot plugin={r.plugin} element={r.c.element} context={{ artifact: shown, file: { ...f, url: fileURL(f.name) } }} />{/if}
              {:else if isImage(f)}
                <img src={fileURL(f.name)} alt={f.name} />
              {:else if isText(f)}
                {#await text(f.name) then body}<pre>{body}</pre>{:catch e}<div class="err">{e.message}</div>{/await}
              {/if}
            </div>
          {/each}
        {/if}
        {#if shown.comments?.length}
          <h3>Comments</h3>
          {#each shown.comments as c (c.id)}
            <div class="comment"><div class="faint"><span class="mono">{c.author}</span> · {ago(c.created_at)}</div><pre>{c.body}</pre></div>
          {/each}
        {/if}
      {:else}
        <p class="faint">Loading…</p>
      {/if}
    </article>
  </div>
{/if}

<style>
  h1 { font-size: 18px; font-weight: 600; margin: 0; }
  .sub { margin: 2px 0 12px; }
  .filters { display: flex; gap: 8px; align-items: center; margin-bottom: 12px; flex-wrap: wrap; }
  .filters input:not([type]) { width: 140px; }
  .split { display: grid; grid-template-columns: minmax(220px, 320px) 1fr; gap: 14px; align-items: start; }
  .list { padding: 4px; max-height: calc(100vh - 160px); overflow-y: auto; }
  .item { display: block; width: 100%; text-align: left; border: 0; background: none; padding: 8px 10px; border-radius: 6px; }
  .item:hover { background: var(--raised); }
  .item.active { background: var(--accent-soft); }
  .item .title { font-weight: 500; }
  .item.read .title { color: var(--muted); font-weight: 400; }
  .pane { padding: 14px 18px; min-width: 0; }
  header { display: flex; justify-content: space-between; align-items: center; gap: 10px; }
  h2 { font-size: 16px; font-weight: 600; margin: 0; }
  h3 { font-size: 12px; font-weight: 500; color: var(--muted); margin: 18px 0 6px; text-transform: uppercase; letter-spacing: 0.04em; }
  .row { display: flex; flex-wrap: wrap; gap: 6px 12px; align-items: center; margin: 6px 0; }
  .md { overflow-wrap: anywhere; line-height: 1.5; }
  .md :global(pre) { background: var(--bg); border: 1px solid var(--border); border-radius: 6px; padding: 10px; overflow: auto; }
  .md :global(img) { max-width: 100%; }
  .md :global(table) { border-collapse: collapse; }
  .md :global(th), .md :global(td) { border: 1px solid var(--border); padding: 4px 8px; }
  .file { margin: 10px 0; }
  .file img { max-width: 100%; border: 1px solid var(--border); border-radius: 6px; }
  .file pre, .comment pre { background: var(--bg); border: 1px solid var(--border); border-radius: 6px; padding: 10px; max-height: 360px; overflow: auto; }
  .comment { border-left: 2px solid var(--border); padding: 4px 10px; margin: 8px 0; }
  .empty { padding: 40px 20px; text-align: center; }
  .big { font-size: 15px; font-weight: 500; margin-bottom: 4px; color: var(--ok); }
  .err { padding: 12px; color: var(--bad); }
  @media (max-width: 720px) { .split { grid-template-columns: 1fr; } }
</style>
