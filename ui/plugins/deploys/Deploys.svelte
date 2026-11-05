<script>
  // Built-in Deploys view as a plugin (#517). Talks to Desk only through
  // `api`; must not import from ui/src (scripts/check-plugins.mjs).

  /** @type {{ api: any, context?: any, config?: any }} */
  let { api } = $props();

  /** @type {any} */
  let data = $state(null);
  let error = $state('');

  async function load() {
    try {
      data = await api.get('api/deploys');
      error = '';
    } catch (e) {
      error = String(/** @type {any} */ (e).message || e);
    }
  }

  $effect(() => {
    if (!api) return;
    load();
    // Deploy state lives in files, not tickets; refetch on any change
    // burst at most once a second, as the shell's views do.
    let t = 0;
    const off = api.subscribe(() => {
      if (!t) t = window.setTimeout(() => { t = 0; load(); }, 1000);
    });
    return () => { off(); clearTimeout(t); };
  });

  /** @param {string} [r] */
  const commit = (r) => (r ? 'https://github.com/elee1766/jobshop/commit/' + encodeURIComponent(r) : '');
  const short = (/** @type {string} */ r) => (r || '').slice(0, 7);

  /** @param {string} t */
  function ago(t) {
    if (!t || t.startsWith('0001')) return '';
    const s = Math.round((Date.now() - new Date(t).getTime()) / 1000);
    if (s < 60) return s + 's ago';
    if (s < 3600) return Math.round(s / 60) + 'm ago';
    if (s < 86400) return Math.round(s / 3600) + 'h ago';
    return Math.round(s / 86400) + 'd ago';
  }
</script>

<div class="desk-deploys">
  <h1>Deploys</h1>
  {#if error}
    <div class="card err">{error}</div>
  {:else if data}
    {@const x = data}
    {@const behind = x.deployed && x.server.commit && !x.deployed.startsWith(x.server.commit)}
    <div class="grid" data-testid="deploys">
      <div class="card cell">
        <div class="faint">Controller running</div>
        <div class="v mono">{x.server.version}</div>
        {#if x.server.commit}<a href={commit(x.server.commit)} target="_blank" rel="noopener noreferrer" class="mono">{short(x.server.commit)}</a>{/if}
      </div>
      <div class="card cell">
        <div class="faint">Deployed rev</div>
        {#if x.deployed}
          <a class="v mono" href={commit(x.deployed)} target="_blank" rel="noopener noreferrer">{short(x.deployed)}</a>
          <div class="muted">{ago(x.deployed_at)}</div>
          {#if behind}<span class="pill warn">controller not restarted on it yet</span>{/if}
        {:else}<div class="muted">no self-deploy state</div>{/if}
      </div>
      <div class="card cell">
        <div class="faint">Last failed rev (rolled back)</div>
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
  .desk-deploys h1 { font-size: 18px; font-weight: 600; margin: 0 0 12px; }
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 12px; }
  .cell { padding: 14px; display: flex; flex-direction: column; gap: 4px; align-items: flex-start; }
  .v { font-size: 16px; }
  .err { padding: 12px; color: var(--bad); }
</style>
