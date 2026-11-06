<script>
  // Built-in ticket details (#541): the first drawer section. The shell
  // owns the drawer frame and hands every drawer plugin the same
  // context, { ticket, run, trace, comments } from api/ticket/{id}.
  import { resource } from '../../kit/plugin.svelte.js';
  import { ago, fields, tone, links, meta, safe, isSessionId } from '../../kit/format.js';
  import Trace from '../../kit/Trace.svelte';
  import Holder from '../../kit/Holder.svelte';

  /** @type {{ api: any, context?: any, config?: any }} */
  let { api, context = {} } = $props();
  const go = (/** @type {Record<string, string>} */ p) => api.go(p);
  const t = $derived(context.ticket);
  // External activity linked to this run or ticket (#255).
  const ext = resource(() => api, () => (t ? 'api/activity?limit=20&ticket=' + encodeURIComponent(t.id) : null));
  /** @param {string} id */
  const openSession = (id) => go({ session: id, ticket: '' });
  // Workers open read-only (JOB-54); writing is the explicit "Message".
  /** @param {string} id */
  const watch = (id) => api.openSession(id, { readonly: true });
  /** Holder info from the view's sessions map. @param {string} id */
  const holderInfo = (id) => context.sessions?.[id] ?? { id };
  /** @param {any} x */
  const report = (x) => meta(x).report || null;
</script>

{#snippet sessionLink(/** @type {string} */ sid)}
  <a href={'?session=' + sid} class="mono sl" onclick={(e) => { e.preventDefault(); e.stopPropagation(); openSession(sid); }}>{sid}</a>
{/snippet}

{#if t}
  {@const f = fields(t)}
  {@const rp = report(t)}
  <div class="ticket">
    <h2>{t.title}</h2>
    <div class="row">
      <span class="pill">{t.kind}</span>
      <span class="pill {tone(f.state || t.status)}">{f.state || t.status}{t.resolution ? ' · ' + t.resolution : ''}</span>
      {#if t.pool}<span class="pill">{t.pool}</span>{/if}
      {#each t.labels ?? [] as l (l)}<span class="pill accent">{l}</span>{/each}
    </div>
    <div class="row">
      {#each links(t) as l (l.href)}<a href={l.href} target="_blank" rel="noopener noreferrer">{l.label}</a>{/each}
      {#if t.holder}
        {#if isSessionId(t.holder) && t.status === 'in_progress'}
          <span class="muted held" data-testid="held-by">held by <Holder h={holderInfo(t.holder)} onopen={watch} label="View session" />
            <button type="button" class="ghost" onclick={() => openSession(t.holder)} data-testid="message-session">Message this session</button></span>
        {:else}<span class="muted">held by {#if isSessionId(t.holder)}{@render sessionLink(t.holder)}{:else}<span class="mono">{t.holder}</span>{/if}</span>{/if}
      {/if}
      <span class="faint">updated {ago(t.updated_at)}</span>
    </div>
    {#if t.parent_id}
      <div class="row"><span class="muted">parent</span> <a href={'?ticket=' + t.parent_id} class="mono" onclick={(e) => { e.preventDefault(); go({ ticket: t.parent_id }); }}>{t.parent_id}</a></div>
    {/if}
    {#if rp}
      <div class="report"><span class="pill accent">{rp.state}</span> {#if rp.note}<span>{rp.note}</span>{/if} <span class="faint">{ago(rp.at)}</span></div>
    {/if}

    {#if context.run}
      <h3>Steps</h3>
      <ol class="steps">
        {#each context.run.steps as s (s.id)}
          {@const sf = fields(s)}
          {@const sr = report(s)}
          <li>
            <div class="stepline">
              <a href={'?ticket=' + s.id} onclick={(e) => { e.preventDefault(); go({ ticket: s.id }); }}>{sf.step || s.title}</a>
              <span class="pill {tone(s.status)}">{s.status}{s.resolution ? ' · ' + s.resolution : ''}</span>
              {#if sf.failures}<span class="pill bad">{sf.failures} failures</span>{/if}
              {#if s.status === 'in_progress' && s.holder}<Holder h={holderInfo(s.holder)} onopen={watch} />{/if}
            </div>
            {#if sr?.note}<div class="muted note">{sr.state}: {sr.note} <span class="faint">{ago(sr.at)}</span></div>{/if}
          </li>
        {/each}
      </ol>
    {/if}

    {#if t.body}
      <h3>Body</h3>
      <pre class="body">{t.body}</pre>
    {/if}

    {#if context.trace}
      <h3>Trace</h3>
      <Trace node={context.trace} {go} {watch} />
    {/if}

    {#if ext.data?.items?.length}
      <h3>External activity</h3>
      <ul class="ext" data-testid="drawer-activity">
        {#each ext.data.items as it (it.id)}
          <li>
            <span class="mono">{it.actor}</span> {it.title}
            <span class="faint">· {ago(it.at)}</span>
            {#if it.url && safe(it.url)}<a href={it.url} target="_blank" rel="noopener noreferrer">open</a>{/if}
          </li>
        {/each}
      </ul>
    {/if}

    {#if context.comments?.length}
      <h3>Comments</h3>
      {#each context.comments as c (c.id)}
        <div class="comment">
          <div class="faint">{#if isSessionId(c.author)}{@render sessionLink(c.author)}{:else}<span class="mono">{c.author}</span>{/if} · {ago(c.created_at)}</div>
          <pre>{c.body}</pre>
        </div>
      {/each}
    {/if}
  </div>
{/if}

<style>
  h2 { font-size: 16px; font-weight: 600; margin: 12px 0 8px; overflow-wrap: anywhere; }
  h3 { font-size: 11px; font-weight: 500; color: var(--faint); margin: 20px 0 6px; text-transform: uppercase; letter-spacing: 0.05em; }
  .row { display: flex; flex-wrap: wrap; gap: 6px 10px; align-items: center; margin: 6px 0; }
  .held { display: inline-flex; flex-wrap: wrap; gap: 6px; align-items: baseline; }
  .report { margin: 8px 0; padding: 8px 10px; background: var(--raised); border-radius: var(--radius); overflow-wrap: anywhere; }
  .steps { margin: 0; padding-left: 20px; }
  .steps li { margin: 6px 0; }
  .stepline { display: flex; flex-wrap: wrap; gap: 4px 8px; align-items: center; }
  .note { font-size: 12px; margin-top: 2px; overflow-wrap: anywhere; }
  .body { background: var(--bg); border: 1px solid var(--border); border-radius: var(--radius); padding: 10px; max-height: 320px; overflow: auto; }
  .comment { border-left: 2px solid var(--border); padding: 4px 10px; margin: 8px 0; }
  .ext { margin: 0; padding-left: 18px; }
  .ext li { margin: 4px 0; overflow-wrap: anywhere; }
  @media (max-width: 768px) {
    .steps li a, .ext li a, .sl { display: inline-block; padding: 8px 0; }
  }
</style>
