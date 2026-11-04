<script>
  import { go } from '../lib/url.svelte.js';
  import { resource } from '../lib/api.svelte.js';
  import { ago, fields, tone, links, meta, safe } from '../lib/format.js';
  import Trace from './Trace.svelte';
  import Chat from './Chat.svelte';
  import X from '@lucide/svelte/icons/x';

  /** @type {{ id: string }} */
  let { id } = $props();
  const v = resource(() => '/api/ticket/' + encodeURIComponent(id));
  // External activity linked to this run or ticket (#255).
  const ext = resource(() => '/api/activity?limit=20&ticket=' + encodeURIComponent(id));
  const close = () => go({ ticket: '' });
  /**
   * The escalation chat: meta.escalation (controller-owned, #379), else the
   * older meta.gateway session and key.
   * @param {any} t
   */
  const thread = (t) => {
    const m = meta(t);
    const e = m.escalation;
    if (e?.session) return { session: e.session, key: e.key || '' };
    const g = m.gateway;
    if (g?.session) return { session: g.session, key: g.key || '' };
    return null;
  };
  /** @param {any} t */
  const report = (t) => meta(t).report || null;
</script>

<svelte:window onkeydown={(e) => { if (e.key === 'Escape') close(); }} />

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
      {@const f = fields(t)}
      {@const th = thread(t)}
      {@const rp = report(t)}
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
      {#if rp}
        <div class="report"><span class="pill accent">{rp.state}</span> {#if rp.note}<span>{rp.note}</span>{/if} <span class="faint">{ago(rp.at)}</span></div>
      {/if}

      {#if th}
        <h3>Escalation chat</h3>
        <div class="panel chatpanel" data-testid="escalation-chat"><Chat id={th.session} key={th.key} /></div>
      {/if}

      {#if v.data.run}
        <h3>Steps</h3>
        <ol class="steps">
          {#each v.data.run.steps as s (s.id)}
            {@const sf = fields(s)}
            {@const sr = report(s)}
            <li>
              <div class="stepline">
                <a href={'?ticket=' + s.id} onclick={(e) => { e.preventDefault(); go({ ticket: s.id }); }}>{sf.step || s.title}</a>
                <span class="pill {tone(s.status)}">{s.status}{s.resolution ? ' · ' + s.resolution : ''}</span>
                {#if sf.failures}<span class="pill bad">{sf.failures} failures</span>{/if}
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

      {#if v.data.trace}
        <h3>Trace</h3>
        <Trace node={v.data.trace} />
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
  </div>
</aside>

<style>
  .backdrop { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.45); z-index: 10; }
  .drawer {
    position: fixed; top: 0; right: 0; bottom: 0; width: min(620px, 100vw); z-index: 11; display: flex; flex-direction: column;
    background: var(--panel); border-left: 1px solid var(--border); box-shadow: var(--shadow);
  }
  header {
    flex: none; display: flex; justify-content: space-between; align-items: center; min-height: 44px; padding: 0 10px 0 18px;
    border-bottom: 1px solid var(--border);
  }
  .id { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .scrollbody { flex: 1; min-height: 0; overflow-y: auto; overscroll-behavior: contain; padding: 4px 18px 40px; }
  h2 { font-size: 16px; font-weight: 600; margin: 12px 0 8px; overflow-wrap: anywhere; }
  h3 { font-size: 11px; font-weight: 500; color: var(--faint); margin: 20px 0 6px; text-transform: uppercase; letter-spacing: 0.05em; }
  .row { display: flex; flex-wrap: wrap; gap: 6px 10px; align-items: center; margin: 6px 0; }
  .report { margin: 8px 0; padding: 8px 10px; background: var(--raised); border-radius: var(--radius); overflow-wrap: anywhere; }
  .steps { margin: 0; padding-left: 20px; }
  .steps li { margin: 6px 0; }
  .stepline { display: flex; flex-wrap: wrap; gap: 4px 8px; align-items: center; }
  .note { font-size: 12px; margin-top: 2px; overflow-wrap: anywhere; }
  .body { background: var(--bg); border: 1px solid var(--border); border-radius: var(--radius); padding: 10px; max-height: 320px; overflow: auto; }
  .comment { border-left: 2px solid var(--border); padding: 4px 10px; margin: 8px 0; }
  .panel { border: 1px solid var(--border); border-radius: var(--radius); margin-top: 6px; overflow: hidden; }
  .chatpanel { height: min(480px, 60dvh); }
  .ext { margin: 0; padding-left: 18px; }
  .ext li { margin: 4px 0; overflow-wrap: anywhere; }
  .err { color: var(--bad); padding: 12px 0; }

  @media (max-width: 720px) {
    .drawer { width: 100vw; border-left: 0; padding-top: env(safe-area-inset-top); }
    .backdrop { display: none; }
    header { min-height: 52px; padding: 0 6px 0 14px; }
    .scrollbody { padding: 2px 14px calc(32px + env(safe-area-inset-bottom)); }
    .steps li a, .ext li a { display: inline-block; padding: 8px 0; }
    .chatpanel { height: 70dvh; }
  }
</style>
