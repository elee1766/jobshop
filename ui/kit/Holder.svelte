<script>
  // The session holding a step (JOB-54), as a one-line chip: busy dot,
  // pool, model and the running tool. The display name is often a long
  // task title that repeats the run, so it only shows when there is no
  // pool, and always in the tooltip. A click opens it read-only.

  /** @type {{ h: any, onopen: (id: string) => void, label?: string }} */
  let { h, onopen, label = '' } = $props();
  const name = $derived(h.display_name || h.id);
  const tip = $derived([name, [h.pool, h.model].filter(Boolean).join(' · '), h.busy ? h.tool || 'working' : '', h.id]
    .filter(Boolean).join('\n') + '\nClick to open read-only');
</script>

<button type="button" class="holder" onclick={(e) => { e.stopPropagation(); onopen(h.id); }} title={tip} data-testid="holder" data-session={h.id} data-name={name}>
  <span class="dot" class:busy={h.busy} class:dead={h.alive === false}></span>
  {#if label}<span>{label}</span>{/if}
  {#if h.pool}<span class="name" data-testid="holder-pool">{h.pool}</span>{:else}<span class="name" data-testid="holder-name">{name}</span>{/if}
  {#if h.model}<span class="faint model">{h.model}</span>{/if}
  {#if h.busy}<span class="tool muted" data-testid="holder-tool">{h.tool || 'working'}</span>{/if}
</button>

<style>
  .holder {
    display: inline-flex; align-items: baseline; gap: 6px; min-width: 0; max-width: min(100%, 420px); height: auto; padding: 2px 8px;
    font-size: 12px; text-align: left; white-space: nowrap; overflow: hidden;
  }
  .name { flex: none; max-width: 16em; overflow: hidden; text-overflow: ellipsis; }
  .model { flex: 0 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis; }
  .tool { flex: 0 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis; font-style: italic; }
  .dot { flex: none; width: 7px; height: 7px; border-radius: 50%; background: var(--faint); align-self: center; }
  .dot.busy { background: var(--accent); animation: pulse 1.2s ease-in-out infinite; }
  .dot.dead { background: var(--bad, #c44); }
  @keyframes pulse { 50% { opacity: 0.35; } }
  @media (max-width: 768px) { .holder { min-height: 36px; } .model { display: none; } }
</style>
