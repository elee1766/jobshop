<script>
  import { resource } from '../lib/api.svelte.js';
  import { ago } from '../lib/format.js';

  const d = resource(() => '/api/deploys');
  /** @param {string} [r] */
  const commit = (r) => (r ? 'https://github.com/elee1766/jobshop/commit/' + encodeURIComponent(r) : '');
  const short = (/** @type {string} */ r) => (r || '').slice(0, 7);
</script>

<header class="vh"><h1>Deploys</h1></header>
<div class="pad">
  {#if d.error}
    <div class="card err">{d.error}</div>
  {:else if d.data}
    {@const x = d.data}
    {@const behind = x.deployed && x.server.commit && !x.deployed.startsWith(x.server.commit)}
    <div class="grid" data-testid="deploys">
      <div class="card cell">
        <div class="faint lab">Controller running</div>
        <div class="v mono">{x.server.version}</div>
        {#if x.server.commit}<a href={commit(x.server.commit)} target="_blank" rel="noopener noreferrer" class="mono">{short(x.server.commit)}</a>{/if}
      </div>
      <div class="card cell">
        <div class="faint lab">Deployed rev</div>
        {#if x.deployed}
          <a class="v mono" href={commit(x.deployed)} target="_blank" rel="noopener noreferrer">{short(x.deployed)}</a>
          <div class="muted">{ago(x.deployed_at)}</div>
          {#if behind}<span class="pill warn">controller not restarted on it yet</span>{/if}
        {:else}<div class="muted">no self-deploy state</div>{/if}
      </div>
      <div class="card cell">
        <div class="faint lab">Last failed rev (rolled back)</div>
        {#if x.failed}
          <a class="v mono" href={commit(x.failed)} target="_blank" rel="noopener noreferrer">{short(x.failed)}</a>
          <div class="muted">{ago(x.failed_at)}</div>
          {#if x.failed === x.deployed}<span class="pill">superseded</span>{:else}<span class="pill bad">rolled back</span>{/if}
        {:else}<div class="muted">none</div>{/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(220px, 100%), 1fr)); gap: 10px; }
  .cell { padding: 12px 14px; display: flex; flex-direction: column; gap: 4px; align-items: flex-start; min-width: 0; }
  .lab { font-size: 11px; text-transform: uppercase; letter-spacing: 0.05em; }
  .v { font-size: 16px; overflow-wrap: anywhere; }
</style>
