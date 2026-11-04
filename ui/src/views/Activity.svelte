<script>
  import { url, go } from '../lib/url.svelte.js';
  import { resource } from '../lib/api.svelte.js';
  import { ago, safe } from '../lib/format.js';
  import GitCommit from '@lucide/svelte/icons/git-commit-horizontal';
  import GitPullRequest from '@lucide/svelte/icons/git-pull-request';
  import GitMerge from '@lucide/svelte/icons/git-merge';
  import GitBranch from '@lucide/svelte/icons/git-branch';
  import MessageSquare from '@lucide/svelte/icons/message-square';
  import MessageSquareCode from '@lucide/svelte/icons/message-square-code';
  import Eye from '@lucide/svelte/icons/eye';
  import CircleCheck from '@lucide/svelte/icons/circle-check';
  import CircleX from '@lucide/svelte/icons/circle-x';
  import CircleDot from '@lucide/svelte/icons/circle-dot';
  import Tag from '@lucide/svelte/icons/tag';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import UserRound from '@lucide/svelte/icons/user-round';
  import ActivityIcon from '@lucide/svelte/icons/activity';

  // Filters live in the URL: source, repo, kind, who (humans default).
  const kinds = ['', 'push', 'pr', 'merge', 'review', 'review_comment', 'comment', 'issue', 'ci', 'branch', 'release', 'state', 'assignee'];
  const whos = ['humans', 'bots', 'all'];
  const who = $derived(url.who || 'humans');

  const feed = resource(() => {
    const q = new URLSearchParams({ since: '168h', limit: '300' });
    if (url.source) q.set('source', url.source);
    if (url.repo) q.set('repo', url.repo);
    if (url.kind) q.set('kind', url.kind);
    if (who !== 'all') q.set('bots', who);
    return '/api/activity?' + q.toString();
  });

  /** "github:owner/name" -> { source, repo } @param {string} s */
  const split = (s) => {
    const i = s.indexOf(':');
    return { source: s.slice(0, i), repo: s.slice(i + 1) };
  };

  /** @param {any} it */
  function icon(it) {
    switch (it.kind) {
      case 'push': return GitCommit;
      case 'pr': return GitPullRequest;
      case 'merge': return GitMerge;
      case 'branch': return GitBranch;
      case 'comment': return MessageSquare;
      case 'review_comment': return MessageSquareCode;
      case 'review': return Eye;
      case 'ci': return / success /.test(it.title) ? CircleCheck : CircleX;
      case 'issue': return CircleDot;
      case 'release': return Tag;
      case 'state': return RefreshCw;
      case 'assignee': return UserRound;
      default: return ActivityIcon;
    }
  }

  /** @param {any} it */
  const ciTone = (it) => (it.kind !== 'ci' ? '' : / success /.test(it.title) ? 'ok' : / (failure|timed_out|startup_failure) /.test(it.title) ? 'bad' : '');

  /** Short plain-text body: one paragraph, at most 280 chars. @param {string} b */
  function short(b) {
    if (!b) return '';
    const s = b.replace(/\r/g, '').trim();
    return s.length > 280 ? s.slice(0, 280) + '…' : s;
  }

  /** @param {string} t */
  function day(t) {
    const d = new Date(t);
    const today = new Date();
    const y = new Date(today);
    y.setDate(today.getDate() - 1);
    if (d.toDateString() === today.toDateString()) return 'Today';
    if (d.toDateString() === y.toDateString()) return 'Yesterday';
    return d.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' });
  }

  /** Items grouped by local day, newest first. */
  const days = $derived.by(() => {
    /** @type {{ day: string, items: any[] }[]} */
    const out = [];
    for (const it of feed.data?.items ?? []) {
      const d = day(it.at);
      if (!out.length || out[out.length - 1].day !== d) out.push({ day: d, items: [] });
      out[out.length - 1].items.push(it);
    }
    return out;
  });
</script>

<header class="vh">
  <h1>Activity</h1>
  <div class="chips actions" data-testid="activity-filters">
    {#each whos as w (w)}
      <button class:on={who === w} onclick={() => go({ who: w === 'humans' ? '' : w })}>{w}</button>
    {/each}
    <span class="sep"></span>
    <button class:on={!url.source && !url.repo} onclick={() => go({ source: '', repo: '' })}>all sources</button>
    {#each feed.data?.sources ?? [] as s (s)}
      {@const p = split(s)}
      <button class:on={url.repo === p.repo} onclick={() => go({ source: p.source, repo: p.repo })}>{p.repo}</button>
    {/each}
    <span class="sep"></span>
    <select aria-label="kind" value={url.kind} onchange={(e) => go({ kind: e.currentTarget.value })}>
      {#each kinds as k (k)}<option value={k}>{k || 'all kinds'}</option>{/each}
    </select>
  </div>
</header>

<div class="pad">
  {#if feed.error}
    <div class="card err">{feed.error}</div>
  {:else if feed.data}
    {#if !feed.data.sources.length}
      <div class="card empty faint">Activity is off. Set <span class="mono">activity: {'{}'}</span> in jobshop.yaml.</div>
    {/if}
    {#each days as g (g.day)}
      <h2 class="st">{g.day}</h2>
      <ul class="card feed" data-testid="activity">
        {#each g.items as it (it.id)}
          {@const Icon = icon(it)}
          <li>
            <span class="icon {ciTone(it)}" title={it.kind}><Icon size={14} /></span>
            <div class="main">
              <div class="line">
                <span class="actor">{it.actor || 'unknown'}</span>
                {#if it.bot}<span class="pill">bot</span>{/if}
                <span class="title">{it.title}</span>
              </div>
              {#if it.body}<div class="body muted">{short(it.body)}</div>{/if}
              <div class="meta faint">
                <span class="mono">{it.repo}</span>
                <span title={it.at}>{ago(it.at)}</span>
                {#if it.url && safe(it.url)}<a href={it.url} target="_blank" rel="noopener noreferrer">open</a>{/if}
                {#each it.tickets ?? [] as id (id)}
                  <a href={'?ticket=' + id} class="mono" onclick={(e) => { e.preventDefault(); go({ ticket: id }); }}>{id}</a>
                {/each}
              </div>
            </div>
          </li>
        {/each}
      </ul>
    {:else}
      {#if feed.data.sources.length}<div class="card empty faint">Nothing in the last 7 days for these filters.</div>{/if}
    {/each}
  {/if}
</div>

<style>
  .vh .chips { min-width: 0; }
  .empty { margin-top: 12px; }
  .feed { list-style: none; margin: 0; padding: 0; }
  .feed li { display: flex; gap: 10px; padding: 9px 12px; border-top: 1px solid var(--border); }
  .feed li:first-child { border-top: none; }
  .icon { color: var(--muted); padding-top: 2px; flex: none; }
  .icon.ok { color: var(--ok); }
  .icon.bad { color: var(--bad); }
  .main { min-width: 0; flex: 1; }
  .line { display: flex; gap: 6px; flex-wrap: wrap; align-items: baseline; }
  .actor { font-weight: 600; }
  .title { overflow-wrap: anywhere; }
  .body { white-space: pre-wrap; overflow-wrap: anywhere; margin-top: 3px; }
  .meta { display: flex; gap: 4px 12px; flex-wrap: wrap; margin-top: 3px; font-size: 12px; }
  .meta a { padding: 2px 0; }
  @media (max-width: 720px) {
    .vh { flex-wrap: wrap; padding-top: 6px; padding-bottom: 6px; }
    .vh .chips { flex: 1 0 100%; margin-left: 0; order: 3; }
    .meta a { padding: 8px 0; }
  }
</style>
