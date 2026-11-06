<script>
  // The session holding a step (JOB-54): display name, pool and model,
  // a busy dot and the running tool. A click opens it read-only.

  /** @type {{ h: any, onopen: (id: string) => void, label?: string }} */
  let { h, onopen, label = '' } = $props();
</script>

<button type="button" class="holder" onclick={(e) => { e.stopPropagation(); onopen(h.id); }} title={'Open ' + h.id + ' read-only'} data-testid="holder" data-session={h.id}>
  <span class="dot" class:busy={h.busy} class:dead={h.alive === false}></span>
  {#if label}<span>{label}</span>{/if}
  <span class="name" data-testid="holder-name">{h.display_name || h.id}</span>
  {#if h.pool || h.model}<span class="faint" data-testid="holder-pool">{[h.pool, h.model].filter(Boolean).join(' · ')}</span>{/if}
  {#if h.busy}<span class="tool muted" data-testid="holder-tool">{h.tool || 'working'}</span>{/if}
</button>

<style>
  .holder {
    display: inline-flex; align-items: baseline; gap: 6px; min-width: 0; max-width: 100%; height: auto; padding: 2px 8px;
    font-size: 12px; text-align: left; white-space: nowrap;
  }
  .name { overflow: hidden; text-overflow: ellipsis; min-width: 0; }
  .tool { overflow: hidden; text-overflow: ellipsis; min-width: 0; font-style: italic; }
  .dot { flex: none; width: 7px; height: 7px; border-radius: 50%; background: var(--faint); align-self: center; }
  .dot.busy { background: var(--accent); animation: pulse 1.2s ease-in-out infinite; }
  .dot.dead { background: var(--bad, #c44); }
  @keyframes pulse { 50% { opacity: 0.35; } }
  @media (max-width: 768px) { .holder { min-height: 36px; } }
</style>
