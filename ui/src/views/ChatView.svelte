<script>
  import { url, go } from '../lib/url.svelte.js';
  import { resource, post } from '../lib/api.svelte.js';
  import Chat from './Chat.svelte';

  // Operator chat: sessions of the chat pool (escalation threads from the
  // gateway and desk:<id> sessions), and a button for a fresh one.
  const chats = resource(() => '/api/chat');
  let first = $state('');
  let starting = $state(false);
  let error = $state('');

  const current = $derived(chats.data?.sessions?.find((/** @type {any} */ s) => s.id === url.session));

  async function start() {
    if (!first.trim()) return;
    starting = true;
    error = '';
    try {
      const r = await post('/api/chat/new', { text: first });
      first = '';
      go({ session: r.session });
    } catch (e) {
      error = String(/** @type {Error} */ (e).message || e);
    } finally {
      starting = false;
    }
  }

  /** @param {string} k */
  const label = (k) => (k.startsWith('desk:') ? 'Desk' : k.startsWith('matrix:') ? 'Matrix' : k.split(':')[0] || 'session');
</script>

<h1>Chat</h1>
<p class="muted sub">Operator sessions in pool <span class="mono">{chats.data?.pool ?? '…'}</span>. Desk and Matrix share sessions by key.</p>

<div class="card new">
  <form onsubmit={(e) => { e.preventDefault(); start(); }}>
    <textarea bind:value={first} rows="2" placeholder="Start a new operator session: what should it look at?" data-testid="chat-new-input"></textarea>
    <button type="submit" disabled={starting || !first.trim()} data-testid="chat-new">New session</button>
  </form>
  {#if error}<div class="err">{error}</div>{/if}
</div>

<div class="split">
  <div class="card list">
    {#if chats.error}<div class="err">{chats.error}</div>{/if}
    {#each chats.data?.sessions ?? [] as s (s.id)}
      <a
        href={'?view=chat&session=' + s.id}
        class:active={url.session === s.id}
        onclick={(e) => { e.preventDefault(); go({ session: s.id }); }}
      >
        <span class="pill accent">{label(s.key)}</span>
        <span class="mono">{s.id}</span>
        <div class="faint key">{s.key}</div>
      </a>
    {:else}
      <div class="faint pad">No open operator sessions.</div>
    {/each}
  </div>
  <div class="card pane">
    {#if url.session}
      {#key url.session}
        <Chat id={url.session} key={current?.key ?? ''} />
      {/key}
    {:else}
      <div class="faint pad">Pick a session or start one.</div>
    {/if}
  </div>
</div>

<style>
  h1 { font-size: 18px; font-weight: 600; margin: 0; }
  .sub { margin: 2px 0 12px; }
  .new { padding: 10px; margin-bottom: 12px; }
  .new form { display: flex; gap: 8px; align-items: flex-end; }
  textarea {
    flex: 1; font: inherit; color: inherit; background: var(--raised); border: 1px solid var(--border);
    border-radius: 6px; padding: 6px 8px; resize: vertical;
  }
  .split { display: grid; grid-template-columns: 260px 1fr; gap: 12px; }
  .list { padding: 4px; max-height: 70vh; overflow: auto; }
  .list a { display: block; padding: 6px 8px; border-radius: 6px; color: var(--text); }
  .list a:hover { background: var(--raised); text-decoration: none; }
  .list a.active { background: var(--accent-soft); }
  .key { font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .pane { min-width: 0; }
  .pad { padding: 12px; }
  .err { color: var(--bad); padding: 6px 2px; }
  button:disabled { opacity: 0.5; cursor: default; }
  @media (max-width: 720px) { .split { grid-template-columns: 1fr; } .list { max-height: 30vh; } }
</style>
