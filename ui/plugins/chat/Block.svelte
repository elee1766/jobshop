<script>
  // One transcript block, after jcode-web's Block.svelte (#535).
  import { md } from './format.js';

  /** @type {{ block: import('./transcript.svelte.js').Block }} */
  let { block } = $props();
  let open = $state(false);

  /** @param {string} s @param {number} [n] */
  function brief(s, n = 140) {
    const one = (s || '').replace(/\s+/g, ' ').trim();
    return one.length > n ? one.slice(0, n) + '…' : one;
  }
  /** @param {string} s */
  function pretty(s) {
    try { return JSON.stringify(JSON.parse(s), null, 2); } catch { return s; }
  }
  /** The intent field of a tool's JSON input, as jcode tools send it. @param {string} s */
  function intent(s) {
    const m = /"intent"\s*:\s*"((?:[^"\\]|\\.)*)"/.exec(s || '');
    return m ? m[1] : '';
  }
  /** @type {Record<string, string>} */
  const sources = { operator: 'operator', desk: 'Desk user', controller: 'controller prompt', message: 'message' };
</script>

{#if block.kind === 'user'}
  <div class="user" data-source={block.source} data-testid="block-user">
    <span class="tag">{sources[block.source] || block.source}</span>
    {#if block.source === 'controller'}
      <details class="prompt"><summary>{brief(block.text, 120)}</summary><pre>{block.text}</pre></details>
    {:else}
      <div class="md">{@html md(block.text)}</div>
    {/if}
  </div>
{:else if block.kind === 'assistant'}
  <div class="assistant md" class:live={block.live} data-testid="block-assistant">{@html md(block.text)}</div>
{:else if block.kind === 'tool'}
  <div class="tool" data-state={block.state} data-testid="block-tool" data-tool={block.id}>
    <button class="tool-head" onclick={() => (open = !open)} aria-expanded={open}>
      <span class="tname">{block.name}</span>
      <span class="tintent">{intent(block.input) || brief(block.input, 80)}</span>
      <span class="tstate" data-testid="tool-state">{block.state}</span>
    </button>
    {#if open}
      {#if block.input}<pre class="io">{pretty(block.input)}</pre>{/if}
      {#if block.output}<pre class="io out">{block.output}</pre>{/if}
    {/if}
  </div>
{:else if block.kind === 'error'}
  <p class="error" data-testid="block-error">{block.text}</p>
{:else}
  <details class="system" data-testid="block-system">
    <summary>{brief(block.text, 160)}</summary>
    {#if block.text.length > 160 || block.text.includes('\n')}<pre>{block.text}</pre>{/if}
  </details>
{/if}

<style>
  .user { margin: 14px 0 8px; padding: 8px 12px; border-radius: 8px; background: var(--desk-accent-soft); min-width: 0; overflow-wrap: anywhere; }
  .user[data-source='controller'] { background: transparent; border: 1px dashed var(--desk-border-strong); }
  .tag { display: block; font-size: 11px; color: var(--desk-faint); margin-bottom: 2px; }
  .prompt summary { cursor: pointer; color: var(--desk-muted); font-size: 12px; overflow-wrap: anywhere; }
  .prompt pre, .system pre { white-space: pre-wrap; overflow-wrap: anywhere; font-family: var(--desk-font); font-size: 12px; margin: 4px 0 0; }
  .assistant { margin: 6px 0; min-width: 0; overflow-wrap: anywhere; }
  .md :global(p) { margin: 6px 0; }
  .md :global(pre) { background: var(--desk-raised); padding: 8px 10px; border-radius: 6px; overflow-x: auto; white-space: pre; max-width: 100%; }
  .md :global(code) { font-family: var(--desk-mono); font-size: 12px; background: var(--desk-raised); padding: 1px 4px; border-radius: 3px; }
  .md :global(pre code) { background: none; padding: 0; }
  .md :global(table) { border-collapse: collapse; font-size: 12px; display: block; overflow-x: auto; max-width: 100%; }
  .md :global(td), .md :global(th) { border: 1px solid var(--desk-border); padding: 3px 8px; }
  .md :global(ul), .md :global(ol) { padding-left: 20px; }
  .md :global(a) { color: var(--desk-accent); }
  .live > :global(:last-child)::after { content: '▍'; color: var(--desk-accent); }
  .tool { margin: 2px 0; font-size: 12px; min-width: 0; }
  .tool-head {
    display: flex; gap: 8px; width: 100%; text-align: left; padding: 2px 0; height: auto; min-height: 22px;
    background: none; border: 0; color: var(--desk-muted); justify-content: flex-start; white-space: nowrap; min-width: 0; overflow: hidden;
  }
  .tool-head:hover { color: var(--desk-text); background: none; }
  .tname { font-family: var(--desk-mono); color: var(--desk-faint); flex: none; max-width: 40%; overflow: hidden; text-overflow: ellipsis; }
  .tool[data-state='running'] .tname, .tool[data-state='input'] .tname { color: var(--desk-accent); }
  .tool[data-state='failed'] .tname, .tool[data-state='failed'] .tstate { color: var(--desk-bad); }
  .tintent { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; }
  .tstate { flex: none; font-size: 11px; color: var(--desk-faint); }
  .io {
    margin: 4px 0 8px; padding: 8px 10px; background: var(--desk-raised); border-radius: 6px;
    max-height: 24rem; overflow: auto; white-space: pre; font-family: var(--desk-mono); font-size: 11.5px; max-width: 100%;
  }
  .out { color: var(--desk-muted); }
  .system { color: var(--desk-faint); font-size: 12px; margin: 6px 0; }
  .system summary { cursor: pointer; overflow-wrap: anywhere; }
  .error { color: var(--desk-bad); font-size: 12px; overflow-wrap: anywhere; }
  @media (max-width: 768px) {
    .tool-head { min-height: 32px; align-items: center; }
  }
</style>
