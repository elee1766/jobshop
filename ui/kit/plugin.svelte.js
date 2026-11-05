// Helpers shared by the built-in plugins (#541). Not part of the shell:
// a built-in imports only from here and its own directory, and talks to
// Desk only through the `api` object (scripts/check-plugins.mjs). Every
// plugin bundles its own copy, so nothing here is shared at run time.

import { mount, unmount } from 'svelte';

/**
 * Defines <tag> as a custom element that mounts Component with
 * { api, context, config } props, which Desk sets before connecting it.
 * The plugin's CSS (style.css next to its bundle, emitted by the build)
 * loads through a <link>: the CSP has no inline styles.
 * @param {string} tag @param {any} Component
 */
export function defineElement(tag, Component) {
  if (customElements.get(tag)) return;
  const css = new URL(/* @vite-ignore */ './style.css', import.meta.url).href;
  customElements.define(tag, class extends HTMLElement {
    props = $state({ api: /** @type {any} */ (null), context: /** @type {any} */ ({}), config: /** @type {any} */ ({}) });
    /** @type {Record<string, any> | null} */
    app = null;
    set api(v) { this.props.api = v; }
    get api() { return this.props.api; }
    set context(v) { this.props.context = v ?? {}; }
    get context() { return this.props.context; }
    set config(v) { this.props.config = v ?? {}; }
    get config() { return this.props.config; }
    connectedCallback() {
      if (!document.querySelector(`link[href="${css}"]`)) {
        const l = document.createElement('link');
        l.rel = 'stylesheet';
        l.href = css;
        document.head.append(l);
      }
      this.style.display = 'contents';
      if (!this.app) this.app = mount(Component, { target: this, props: this.props });
    }
    disconnectedCallback() {
      if (this.app) unmount(this.app);
      this.app = null;
    }
  });
}

/**
 * A reactive GET: refetches when path() changes and, coalesced to once a
 * second, when any change event arrives.
 * @param {() => any} getApi @param {() => string | null} path
 */
export function resource(getApi, path) {
  const r = $state({ data: /** @type {any} */ (null), error: '', loading: true, refresh: () => { tick++; } });
  let tick = $state(0);
  $effect(() => {
    const api = getApi();
    if (!api) return;
    let t = 0;
    const off = api.subscribe(() => { if (!t) t = window.setTimeout(() => { t = 0; tick++; }, 1000); });
    return () => { off(); clearTimeout(t); };
  });
  // A slow endpoint (api/pools takes seconds) must not be starved by
  // change events: a tick while a fetch is in flight queues one more
  // fetch after it instead of discarding it. Only a new path discards.
  let inflight = false;
  let again = false;
  let cur = '';
  /** @param {any} api @param {string} p */
  const load = (api, p) => {
    inflight = true;
    again = false;
    const done = (/** @type {any} */ v) => {
      inflight = false;
      if (p !== cur) return;
      Object.assign(r, v);
      if (again) load(api, p);
    };
    api.get(p).then(
      (/** @type {any} */ d) => done({ data: d, error: '', loading: false }),
      (/** @type {any} */ e) => done({ error: String(e?.message || e), loading: false }),
    );
  };
  $effect(() => {
    const api = getApi();
    const p = path();
    void tick;
    if (!api || !p) return;
    if (p === cur && inflight) { again = true; return; }
    cur = p;
    load(api, p);
  });
  return r;
}

/**
 * api.query as Svelte state: { data, error, loading }.
 * @param {() => any} getApi @param {() => [string, any] | null} q method and params
 */
export function query(getApi, q) {
  const r = $state({ data: /** @type {any} */ (null), error: '', loading: true });
  $effect(() => {
    const api = getApi();
    const mp = q();
    if (!api || !mp) return;
    const h = api.query(mp[0], mp[1]);
    return h.subscribe((/** @type {any} */ s) => Object.assign(r, s));
  });
  return r;
}

/**
 * URL params as Svelte state, kept current through api.onParams.
 * @param {() => any} getApi
 */
export function params(getApi) {
  const p = $state(/** @type {Record<string, string>} */ ({}));
  $effect(() => {
    const api = getApi();
    if (!api) return;
    const set = (/** @type {Record<string, string>} */ v) => {
      for (const k of Object.keys(p)) if (!(k in v)) delete p[k];
      Object.assign(p, v);
    };
    set(api.params());
    return api.onParams(set);
  });
  return p;
}

/** @param {unknown} e */
export const errText = (e) => String(/** @type {any} */ (e)?.message || e);

/**
 * Svelte action hosting another plugin's contribution (api.mount):
 * <div use:slot={{ api, slot, target, context }}>.
 * @param {HTMLElement} node @param {{ api: any, slot: string, target: Record<string, any>, context?: any }} p
 */
export function slot(node, p) {
  let h = p.api.mount(node, p.slot, p.target, p.context);
  return {
    /** @param {{ api: any, slot: string, target: Record<string, any>, context?: any }} n */
    update(n) {
      if (h && JSON.stringify(n.target) === JSON.stringify(p.target)) h.context(n.context);
      else { h?.destroy(); h = n.api.mount(node, n.slot, n.target, n.context); }
      p = n;
    },
    destroy() { h?.destroy(); },
  };
}
