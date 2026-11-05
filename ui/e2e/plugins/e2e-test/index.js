// Test plugin for e2e/plugins.spec.js (#517): a view that calls
// api.get, api.call and api.post, a drawer section and a CSV renderer.
// Plain JS, no build step, as a user plugin would be.

class View extends HTMLElement {
  async connectedCallback() {
    this.innerHTML = '<h1>E2E plugin</h1><div data-testid="e2e-get">…</div><div data-testid="e2e-call">…</div><div data-testid="e2e-post">…</div>';
    const q = (/** @type {string} */ id) => /** @type {HTMLElement} */ (this.querySelector(`[data-testid=${id}]`));
    try {
      const w = await this.api.get('api/whoami');
      q('e2e-get').textContent = 'get ok ' + typeof w.user;
    } catch (e) { q('e2e-get').textContent = 'get failed ' + e; }
    try {
      const v = await this.api.call('version', {});
      q('e2e-call').textContent = 'call ok ' + (v.version ? 'version' : JSON.stringify(v));
    } catch (e) { q('e2e-call').textContent = 'call failed ' + e; }
    try {
      await this.api.call('claim', {});
      q('e2e-post').textContent = 'claim allowed';
    } catch (e) { q('e2e-post').textContent = 'claim refused ' + e.message; }
  }
}

class Drawer extends HTMLElement {
  set context(c) { this.textContent = 'drawer for ' + (c?.ticket?.id || '?'); this.dataset.testid = 'e2e-drawer'; }
}

class Csv extends HTMLElement {
  set context(c) {
    this.dataset.testid = 'e2e-csv';
    fetch(c.file.url).then((r) => r.text()).then((t) => {
      const rows = t.trim().split('\n').map((l) => l.split(','));
      const table = document.createElement('table');
      for (const r of rows) {
        const tr = table.insertRow();
        for (const cell of r) tr.insertCell().textContent = cell;
      }
      this.replaceChildren(table);
    });
  }
}

customElements.define('e2e-test-view', View);
customElements.define('e2e-test-drawer', Drawer);
customElements.define('e2e-test-csv', Csv);
