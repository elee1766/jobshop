// A top-bar nav for the built-in Desk app (JOB-86), plain JS with no
// build step. It contributes the `nav` slot, so it replaces the built-in
// nav (the last nav by plugin name wins; "top-nav" sorts after "nav").
// The desk app docks it by `placement` and sets `context` to
// { dock, narrow }. It lists api.views(), marks the current view, shows
// the attention badge, and shrinks to icons-only through a container
// query in style.css when the bar is narrow. See docs/desk-plugins.md.

const css = new URL('./style.css', import.meta.url).href;

class TopNav extends HTMLElement {
  connectedCallback() {
    if (!document.querySelector(`link[href="${css}"]`)) {
      const l = document.createElement('link');
      l.rel = 'stylesheet';
      l.href = css;
      document.head.append(l);
    }
    const api = this.api;
    this.offs = [];
    const nav = document.createElement('nav');
    nav.className = 'top-nav';
    nav.setAttribute('aria-label', 'views');
    nav.dataset.testid = 'top-nav';

    this.links = new Map();
    for (const v of api.views()) {
      const a = document.createElement('a');
      a.href = '?view=' + v.id;
      a.dataset.view = v.id;
      a.title = v.title;
      const label = document.createElement('span');
      label.className = 'lbl';
      label.textContent = v.title;
      const badge = document.createElement('span');
      badge.className = 'n';
      a.append(label, badge);
      a.onclick = (e) => { e.preventDefault(); api.navigate(v.id); };
      this.links.set(v.id, { a, badge });
      nav.append(a);
    }
    this.replaceChildren(nav);

    this.offs.push(api.badges((b) => {
      for (const [id, { badge }] of this.links) badge.textContent = b[id] ? String(b[id]) : '';
    }));
    const mark = (p) => {
      for (const [id, { a }] of this.links) a.classList.toggle('on', id === (p.view || 'attention'));
    };
    this.offs.push(api.onParams(mark));
    mark(api.params());
  }

  disconnectedCallback() {
    for (const off of this.offs || []) off();
  }
}

customElements.define('top-nav-bar', TopNav);
