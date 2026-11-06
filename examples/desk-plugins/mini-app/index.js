// A whole-page Desk app (JOB-85): a top bar and one view, in plain JS
// with no build step. Desk mounts the last `app` contribution instead of
// its own frame, so this replaces the built-in desk app. It shows the
// api an app has: contributions for the nav, mount for a view, badges,
// live, theme and URL state. See docs/desk-plugins.md.

class MiniApp extends HTMLElement {
  connectedCallback() {
    const api = this.api;
    this.offs = [];
    const bar = document.createElement('header');
    bar.className = 'mini-bar';
    bar.style.cssText = 'display:flex;gap:8px;align-items:center;padding:8px 12px;border-bottom:1px solid var(--desk-border);background:var(--desk-side);flex-wrap:wrap';
    const title = document.createElement('b');
    title.textContent = this.config.title || 'Mini app';
    title.dataset.testid = 'mini-title';
    bar.append(title);

    // One button per view any enabled plugin contributes.
    this.links = new Map();
    for (const v of api.contributions('view')) {
      if (!v.id || this.links.has(v.id)) continue;
      const b = document.createElement('button');
      b.dataset.view = v.id;
      b.textContent = v.title || v.id;
      b.onclick = () => api.navigate(v.id);
      this.links.set(v.id, b);
      bar.append(b);
    }

    const live = document.createElement('span');
    live.dataset.testid = 'mini-live';
    live.style.marginLeft = 'auto';
    this.offs.push(api.live((on) => { live.textContent = on ? 'live' : 'offline'; }));
    const theme = document.createElement('button');
    theme.dataset.testid = 'mini-theme';
    theme.textContent = 'theme';
    theme.onclick = () => { api.theme = api.theme === 'light' ? 'dark' : 'light'; };
    bar.append(live, theme);

    this.main = document.createElement('main');
    this.main.dataset.testid = 'mini-main';
    this.main.style.cssText = 'flex:1;min-height:0;overflow:auto;display:flex;flex-direction:column';
    const page = document.createElement('div');
    page.style.cssText = 'position:fixed;inset:0;display:flex;flex-direction:column;background:var(--desk-bg)';
    page.append(bar, this.main);
    this.replaceChildren(page);

    this.offs.push(api.badges((b) => {
      for (const [id, el] of this.links) el.dataset.badge = String(b[id] || 0);
    }));
    this.offs.push(api.onParams((p) => this.show(p.view)));
    this.show(api.params().view);
  }

  /** Mounts the built-in (or replaced) view with that id. */
  show(view) {
    if (view === this.view) return;
    this.view = view;
    for (const [id, el] of this.links) el.classList.toggle('on', id === view);
    this.mounted?.destroy();
    this.mounted = this.api.mount(this.main, 'view', { id: view });
    if (!this.mounted) this.main.textContent = 'No view ' + view;
  }

  disconnectedCallback() {
    for (const off of this.offs) off();
    this.mounted?.destroy();
  }
}

customElements.define('mini-app', MiniApp);
