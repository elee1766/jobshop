// Deploys view (#577), moved out of Desk into packs/deploy. Plain JS, no
// build step. Reads the self-deploy state through the pack's
// deploy-state desk command and the controller version through the api;
// rollbacks themselves show in Attention as deploy:rollback tickets.

const repo = 'https://github.com/elee1766/jobshop/commit/';

function ago(sec) {
  if (!sec) return '';
  const s = Math.round(Date.now() / 1000 - sec);
  if (s < 60) return s + 's ago';
  if (s < 3600) return Math.round(s / 60) + 'm ago';
  if (s < 86400) return Math.round(s / 3600) + 'h ago';
  return Math.round(s / 86400) + 'd ago';
}

function el(tag, props = {}, ...kids) {
  const e = document.createElement(tag);
  Object.assign(e, props);
  e.append(...kids.filter((k) => k != null));
  return e;
}

function rev(r) {
  // textContent only: the values come from files on the host.
  return el('a', { className: 'mono', href: repo + encodeURIComponent(r), target: '_blank', rel: 'noopener noreferrer', textContent: r.slice(0, 7) });
}

function cell(label, ...body) {
  const c = el('div', { className: 'card' }, el('div', { className: 'faint', textContent: label }), ...body);
  Object.assign(c.style, { padding: '14px', display: 'flex', flexDirection: 'column', gap: '4px', alignItems: 'flex-start' });
  return c;
}

class DeploysView extends HTMLElement {
  set api(v) {
    this._api = v;
    if (this.isConnected) this.start();
  }

  connectedCallback() {
    if (this._api) this.start();
  }

  disconnectedCallback() {
    if (this.off) this.off();
    this.off = null;
    clearTimeout(this.t);
  }

  start() {
    if (this.off) return;
    this.load();
    // State lives in files; refetch at most once a second on any change
    // burst, as Desk's views do.
    this.off = this._api.subscribe(() => {
      if (!this.t) this.t = setTimeout(() => { this.t = 0; this.load(); }, 1000);
    });
  }

  async load() {
    try {
      const [state, ver] = await Promise.all([this._api.exec('deploy-state'), this._api.call('version')]);
      this.render(state || {}, ver || {});
    } catch (e) {
      this.replaceChildren(el('h1', { textContent: 'Deploys' }), el('div', { className: 'card', textContent: String(e.message || e) }));
    }
  }

  render(s, ver) {
    const behind = s.deployed && ver.commit && !s.deployed.startsWith(ver.commit);
    const grid = el('div', {},
      cell('Controller running', el('div', { className: 'mono', textContent: ver.version || '' }), ver.commit ? rev(ver.commit) : null),
      cell('Deployed rev', ...(s.deployed
        ? [rev(s.deployed), el('div', { className: 'muted', textContent: ago(s.deployed_at) }), behind ? el('span', { className: 'pill warn', textContent: 'controller not restarted on it yet' }) : null]
        : [el('div', { className: 'muted', textContent: 'no self-deploy state' })])),
      cell('Last failed rev', ...(s.failed
        ? [rev(s.failed), el('div', { className: 'muted', textContent: ago(s.failed_at) }),
          s.failed === s.deployed || s.failed_at <= s.deployed_at
            ? el('span', { className: 'pill', textContent: 'superseded' })
            : el('span', { className: 'pill bad', textContent: 'rolled back' })]
        : [el('div', { className: 'muted', textContent: 'none' })])),
    );
    grid.dataset.testid = 'deploys';
    Object.assign(grid.style, { display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '12px' });
    const h = el('h1', { textContent: 'Deploys' });
    Object.assign(h.style, { fontSize: '18px', fontWeight: '600', margin: '0 0 12px' });
    this.replaceChildren(h, grid);
  }
}

if (!customElements.get('desk-deploys-view')) customElements.define('desk-deploys-view', DeploysView);
