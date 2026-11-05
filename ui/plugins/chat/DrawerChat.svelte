<script>
  // Drawer section on session and step tickets (#535): the session's
  // chat inline, with a link that opens it in the side panel. A step
  // shows its holder's chat; an escalation its escalation chat.
  import Pane from './Pane.svelte';

  /** @type {{ api: any, context?: any, config?: any }} */
  let { api, context = {} } = $props();

  const t = $derived(context.ticket || {});
  /** @param {any} x */
  const meta = (x) => (typeof x?.meta === 'string' ? JSON.parse(x.meta) : x?.meta || {});
  const sid = $derived.by(() => {
    if (t.kind === 'session') return t.id;
    const e = meta(t).escalation || meta(t).gateway;
    if (e?.session) return e.session;
    return t.holder && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(t.holder) ? t.holder : '';
  });
</script>

{#if sid}
  <div class="wrap" data-testid="escalation-chat">
    <div class="bar"><a href={'?session=' + sid} onclick={(e) => { e.preventDefault(); api.openSession(sid); }}>open in side panel</a></div>
    <div class="frame"><Pane {api} context={{ session: sid }} /></div>
  </div>
{/if}

<style>
  .wrap { border: 1px solid var(--desk-border); border-radius: var(--desk-radius); overflow: hidden; }
  .bar { padding: 4px 10px; font-size: 12px; border-bottom: 1px solid var(--desk-border); background: var(--desk-panel); }
  .bar a { color: var(--desk-accent); }
  .frame { height: min(480px, 60dvh); }
  @media (max-width: 768px) { .frame { height: 70dvh; } }
</style>
