<script>
  import { url, go } from '../lib/url.svelte.js';
  import { resource, post } from '../lib/api.svelte.js';
  import { chatLabel } from '../lib/format.js';
  import Plus from '@lucide/svelte/icons/plus';
  import Escalations from './Escalations.svelte';

  // Operator chat: the sessions of the chat pool and a form for a fresh one.
  // A session opens in the side panel (SessionPanel), so the list stays put.
  const chats = resource(() => 'api/chat');
  const escalations = resource(() => 'api/escalations');
  // Escalation sessions have their own list below.
  const others = $derived((chats.data?.sessions ?? []).filter((/** @type {any} */ s) => !s.key.startsWith('escalation:')));
  let first = $state('');
  let starting = $state(false);
  let error = $state('');

  async function start() {
    if (!first.trim()) return;
    starting = true;
    error = '';
    try {
      const r = await post('api/chat/new', { text: first });
      first = '';
      go({ session: r.session });
    } catch (e) {
      error = String(/** @type {Error} */ (e).message || e);
    } finally {
      starting = false;
    }
  }
</script>

<header class="vh">
  <h1>Chat</h1>
  <span class="sub">pool <span class="mono">{chats.data?.pool ?? '…'}</span></span>
</header>

<div class="list">
  <form class="new" onsubmit={(e) => { e.preventDefault(); start(); }}>
    <textarea bind:value={first} rows="3" placeholder="What should the operator look at?" data-testid="chat-new-input"></textarea>
    <div class="row">
      <button type="submit" class="primary" disabled={starting || !first.trim()} data-testid="chat-new"><Plus size={13} /> New session</button>
    </div>
    {#if error}<div class="err">{error}</div>{/if}
  </form>
  {#if chats.error}<div class="err">{chats.error}</div>{/if}
  <div data-testid="chat-list">
    {#each others as s (s.id)}
      <a
        href={'?view=chat&session=' + s.id}
        class="item"
        class:active={url.session === s.id}
        onclick={(e) => { e.preventDefault(); go({ session: s.id }); }}
      >
        <span class="dot" class:accent={s.alive}></span>
        <span class="name">{chatLabel(s.key)}</span>
        <span class="mono faint id">{s.id}</span>
      </a>
    {:else}
      <div class="faint pad0">No open operator sessions.</div>
    {/each}
  </div>
  <Escalations items={escalations} />
</div>

<style>
  .list { flex: 1; min-height: 0; overflow-y: auto; padding: 12px 20px 24px; max-width: 960px; width: 100%; }
  .item { display: flex; align-items: center; gap: 8px; height: 34px; padding: 0 8px; border-radius: var(--radius); color: var(--muted); min-width: 0; }
  .item:hover { background: var(--hover); text-decoration: none; color: var(--text); }
  .item.active { background: var(--raised); color: var(--text); }
  .name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .id { flex: none; font-size: 11px; }
  .pad0 { padding: 8px; }
  .new { display: flex; flex-direction: column; gap: 6px; padding: 0 0 12px; }
  .new textarea { resize: vertical; }
  .new .row { display: flex; justify-content: flex-end; }
  .err { color: var(--bad); padding: 4px 2px; font-size: 12px; }

  @media (max-width: 768px) {
    .list { padding: 8px 8px 16px; }
    .item { height: 48px; padding: 0 10px; font-size: 14px; border-bottom: 1px solid var(--border); border-radius: 0; }
    .vh .sub { display: none; }
  }
</style>
