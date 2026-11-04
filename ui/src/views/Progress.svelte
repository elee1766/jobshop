<script>
  import { go } from '../lib/url.svelte.js';
  import { get, live } from '../lib/api.svelte.js';
  import { ago, fields, meta, tone, isSessionId } from '../lib/format.js';
  import SessionLink from './SessionLink.svelte';

  /** Running runs with each step's latest worker report (meta.report) and the newest comment. */
  /** @type {{ limit?: number }} */
  let { limit = 6 } = $props();

  /** @type {{ run: any, steps: any[], comment: any }[]} */
  let rows = $state([]);
  let loaded = $state(false);

  $effect(() => {
    void live.tick;
    let stale = false;
    (async () => {
      try {
        const w = await get('/api/work?state=running&limit=' + limit);
        const runs = (w.runs ?? []).slice(0, limit);
        const out = await Promise.all(runs.map(async (/** @type {any} */ run) => {
          try {
            const v = await get('/api/ticket/' + encodeURIComponent(run.id));
            const cs = v.comments ?? [];
            return { run, steps: v.run?.steps ?? [], comment: cs.length ? cs[cs.length - 1] : null };
          } catch {
            return { run, steps: [], comment: null };
          }
        }));
        if (!stale) { rows = out; loaded = true; }
      } catch {
        if (!stale) loaded = true;
      }
    })();
    return () => { stale = true; };
  });

  /** Steps that have reported, newest first; else the step that is in progress. @param {any[]} steps */
  function reported(steps) {
    return steps
      .map((s) => ({ s, r: meta(s).report }))
      .filter((x) => x.r && (x.r.note || x.r.state))
      .sort((a, b) => String(b.r.at).localeCompare(String(a.r.at)));
  }

  /** @param {string} s */
  const rtone = (s) => (s === 'failed' || s === 'blocked' ? 'bad' : s === 'waiting' || s === 'dropped' ? 'warn' : s === 'ready' ? 'ok' : 'accent');
</script>

{#if rows.length}
  <div class="progress" data-testid="progress">
    {#each rows as { run, steps, comment } (run.id)}
      {@const rep = reported(steps)}
      {@const active = steps.find((s) => s.status === 'in_progress')}
      <div class="item card">
        <button class="head" onclick={() => go({ ticket: run.id })}>
          <span class="dot accent"></span>
          <span class="title">{run.title}</span>
          <span class="mono faint id">{run.id}</span>
          <span class="faint when">{ago(run.updated_at)}</span>
        </button>
        <ul>
          {#each rep.slice(0, 3) as { s, r } (s.id)}
            <li>
              <span class="pill {rtone(r.state)}">{r.state}</span>
              <span class="step muted">{fields(s).step || s.title}</span>
              {#if isSessionId(r.by)}<SessionLink id={r.by} />{/if}
              {#if r.note}<span class="note">{r.note}</span>{/if}
              <span class="faint when">{ago(r.at)}</span>
            </li>
          {:else}
            {#if active}
              <li><span class="pill {tone(active.status)}">{active.status}</span><span class="step muted">{fields(active).step || active.title}</span><span class="faint">no report yet</span></li>
            {:else}
              <li><span class="faint">waiting for a step to start</span></li>
            {/if}
          {/each}
          {#if comment}
            <li><span class="pill">comment</span><span class="step muted">{#if isSessionId(comment.author)}<SessionLink id={comment.author} />{:else}<span class="mono">{comment.author}</span>{/if}</span><span class="note">{comment.body}</span><span class="faint when">{ago(comment.created_at)}</span></li>
          {/if}
        </ul>
      </div>
    {/each}
  </div>
{:else if !loaded}
  <div class="faint">Loading…</div>
{:else}
  <div class="faint none" data-testid="progress-empty">No runs in progress.</div>
{/if}

<style>
  .progress { display: flex; flex-direction: column; gap: 8px; }
  .item { overflow: hidden; }
  .head {
    width: 100%; height: auto; min-height: 32px; justify-content: flex-start; gap: 8px; padding: 6px 12px; border: 0; border-radius: 0;
    background: transparent; text-align: left; white-space: normal;
  }
  .head:hover:not(:disabled) { background: var(--hover); }
  .title { flex: 1; min-width: 0; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .id { flex: none; }
  ul { list-style: none; margin: 0; padding: 0 12px 8px 28px; display: flex; flex-direction: column; gap: 4px; }
  li { display: flex; gap: 8px; align-items: baseline; min-width: 0; flex-wrap: wrap; }
  .step { flex: none; }
  .note { flex: 1 1 200px; min-width: 0; overflow-wrap: anywhere; white-space: pre-wrap; display: -webkit-box; -webkit-line-clamp: 3; line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; }
  .when { flex: none; margin-left: auto; font-size: 11px; }
  .none { padding: 4px 0; }
  @media (max-width: 768px) {
    .head { min-height: 44px; }
    .id { display: none; }
  }
</style>
