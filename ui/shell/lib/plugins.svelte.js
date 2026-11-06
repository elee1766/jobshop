// Desk plugin host (#517). Plugins are custom elements; Desk imports a
// plugin's entry once, waits for its elements, then sets api, context
// and config on each element it creates. Built-in views use the same
// path, so this file is the whole contract on the browser side.

import { base, api as urlOf, get, post, live, onLive } from './api.svelte.js';
import { setBadge, onBadges, getTheme, setTheme } from './signals.js';
import { url, go, navigate, closeSession, onUrl } from './url.svelte.js';
import { push, pushStatus, enablePush, disablePush, testPush } from './push.svelte.js';

/**
 * Bumped when api gains a member a plugin may need; plugin.json "api": ">=N".
 * 2 (#535): session, push, openSession, closeSession, openTicket, params.
 * 3 (#541): go, onParams, query, has, mount; params keeps every URL key.
 * 4 (JOB-87): slot, plugin-defined slots named <plugin>:<name>.
 * 5 (JOB-85): the app slot; contributions, badge, badges, live, theme;
 *   has, mount and slot pick a contribution by plugin, element or id.
 */
export const API_VERSION = 5;

/**
 * @typedef {{ slot: string, element: string, id?: string, title?: string, match?: Record<string, string[]> }} Contribution
 * @typedef {{ name: string, version?: string, entry: string, api?: string, contributes: Contribution[], source: string, url: string, config?: any, overrides?: string }} PluginInfo
 */

export const plugins = $state({ list: /** @type {PluginInfo[]} */ ([]), loaded: false, error: '' });
export const toasts = $state({ items: /** @type {{ id: number, msg: string }[]} */ ([]) });
export const confirmBox = $state({ open: /** @type {{ msg: string, resolve: (v: boolean) => void } | null} */ (null) });

/** @type {Set<(ev: any) => void>} */
const listeners = new Set();

/** Called by the event connection for every raw event. @param {any} ev */
export function dispatch(ev) {
  for (const fn of listeners) {
    try { fn(ev); } catch (e) { console.error('desk plugin event handler', e); }
  }
}

let toastSeq = 0;

/** The one api object every plugin gets. */
export const deskApi = Object.freeze({
  version: API_VERSION,
  base,
  get user() { return who.user; },
  /** @param {string} path */
  get: (path) => get(path),
  /** @param {string} path @param {any} body */
  post: (path, body) => post(path, body),
  /** Admin RPC through Desk's policy. @param {string} method @param {any} [params] */
  call: (method, params) => post('api/rpc/' + encodeURIComponent(method), params ?? {}),
  /**
   * Runs a desk_commands entry (#542). Resolves to its stdout: parsed
   * JSON, else text. opts.ticket records the run in that ticket's history.
   * @param {string} name @param {any} [input] @param {{ ticket?: string }} [opts]
   */
  async exec(name, input, opts) {
    /** @type {{ output?: any, text?: string }} */
    const r = await post('api/rpc/desk.exec', { name, input: input ?? null, ticket: opts?.ticket || undefined });
    return r.output !== undefined ? r.output : (r.text ?? '');
  },
  /** @param {(ev: any) => void} fn */
  subscribe(fn) {
    listeners.add(fn);
    return () => { listeners.delete(fn); };
  },
  /**
   * Raw events filtered by any of kind, op, id, ticket_kind, label, pool.
   * @param {Record<string, string>} filter @param {(ev: any) => void} fn
   */
  events(filter, fn) {
    const keys = Object.entries(filter || {});
    return deskApi.subscribe((ev) => {
      if (keys.every(([k, v]) => ev[k] === v)) fn(ev);
    });
  },
  /** @param {string} view @param {Record<string, string>} [params] */
  navigate(view, params) { navigate(view, params); },
  /** @param {string} msg */
  toast(msg) {
    const id = ++toastSeq;
    toasts.items.push({ id, msg: String(msg) });
    setTimeout(() => { toasts.items = toasts.items.filter((t) => t.id !== id); }, 4000);
  },
  /** @param {string} msg @returns {Promise<boolean>} */
  confirm(msg) {
    return new Promise((resolve) => { confirmBox.open = { msg: String(msg), resolve }; });
  },
  /** The URL of a file under base. @param {string} path */
  url: (path) => urlOf(path),
  get lastSeq() { return live.lastSeq; },
  /**
   * A copy of Desk's URL state: view, ticket, session, item and every
   * other key, so a plugin view keeps its filters here and links and
   * back work (#541).
   */
  params: () => /** @type {Record<string, string>} */ ({ ...url }),
  /** Changes URL state and pushes a history entry. @param {Record<string, string>} patch */
  go(patch) { go(patch); },
  /** Called with params() after every URL change. @param {(params: Record<string, string>) => void} fn */
  onParams(fn) {
    paramListeners.add(fn);
    return () => { paramListeners.delete(fn); };
  },
  /**
   * Reactive admin call (#541): fetches, then refetches on a change
   * event for an id in the result or matching opts.watch, coalesced to
   * once a second. subscribe(fn) gets {data, error, loading}.
   * @param {string} method @param {any} [params] @param {{ watch?: Record<string, string> }} [opts]
   */
  query(method, params, opts) { return makeQuery(method, params, opts); },
  /**
   * Whether an enabled plugin contributes to slot for target (#541).
   * A target naming plugin, element or id picks by those instead of
   * match (api 5).
   * @param {string} slot @param {Record<string, any>} [target]
   */
  has(slot, target) { return !!firstFor(slot, target); },
  /**
   * Every enabled contribution to slot that target selects, in plugin
   * name order (JOB-85): what an app reads to build its nav, header and
   * drawer. Copies; the list is fixed once the page has loaded.
   * @param {string} slot @param {Record<string, any>} [target]
   */
  contributions(slot, target) {
    return contributions(slot)
      .filter((x) => selects(x, target))
      .map((x) => ({ ...$state.snapshot(x.c), plugin: x.plugin.name, source: x.plugin.source }));
  },
  /**
   * Mounts the first contribution to slot matching target inside host
   * (#541): how one plugin hosts another, e.g. artifact renderers, or an
   * app the built-in views ({plugin: 'work'} or {id: 'work'}).
   * Returns null when nothing matches, else { context(c), destroy() }.
   * @param {HTMLElement} host @param {string} slot @param {Record<string, any>} target @param {any} [context]
   */
  mount(host, slot, target, context) { return mountSlot(host, slot, target, context); },
  /**
   * Mounts every contribution to a plugin-defined slot (JOB-87), e.g.
   * 'work:toolbar', each in its own box inside host, in plugin order. A
   * contribution that fails shows an error card in its box only. target
   * is what match is checked against ({} when omitted). Returns null
   * when nothing matches, else { context(c), destroy() }.
   * @param {HTMLElement} host @param {string} slot @param {any} [context] @param {Record<string, any>} [target]
   */
  slot(host, slot, context, target) { return mountAll(host, slot, context, target); },
  /**
   * Opens a session in the side panel (full screen on a phone). item is
   * the attention item it resolves, "kind:id" (#535). readonly hides the
   * composer whatever access allows (JOB-54): a view mode for watching a
   * worker, not a permission.
   * @param {string} id @param {{ item?: string, readonly?: boolean }} [opts]
   */
  openSession(id, opts) { go({ session: id, item: opts?.item || '', ticket: '', ro: opts?.readonly ? '1' : '' }); },
  closeSession() { closeSession(); },
  /** @param {string} id */
  openTicket(id) { go({ ticket: id }); },
  /** Session chat (#535): the live stream, older history, sends. */
  session: Object.freeze({
    /**
     * Streams a session's events from tail journal entries before the
     * end. fn gets {type: 'open'|'event'|'end'|'error', event?}. Returns
     * a function that closes the stream.
     * @param {string} id @param {{ tail?: number }} opts @param {(m: { type: string, event?: any }) => void} fn
     */
    stream(id, opts, fn) {
      const es = new EventSource(urlOf('api/session/' + encodeURIComponent(id) + '/stream?tail=' + (opts?.tail || 200)));
      es.addEventListener('open', () => fn({ type: 'open' }));
      es.addEventListener('session', (e) => { try { fn({ type: 'event', event: JSON.parse(e.data) }); } catch { /* not JSON */ } });
      es.addEventListener('end', () => { es.close(); fn({ type: 'end' }); });
      es.addEventListener('error', () => fn({ type: 'error' }));
      return () => es.close();
    },
    /** Decoded events before seq, oldest first: {events, first}. @param {string} id @param {number} before @param {number} [limit] */
    older: (id, before, limit = 400) => get('api/session/' + encodeURIComponent(id) + '/events?before=' + before + '&limit=' + limit),
    /** {write, reason}: whether Desk may send to the session by id. @param {string} id */
    access: (id) => get('api/session/' + encodeURIComponent(id) + '/access'),
    /** Send by session id or key; queued while a turn runs unless interrupt. @param {{ session?: string, key?: string, text: string, interrupt?: boolean }} m */
    send: (m) => post('api/chat/send', m),
    /** Stop: cancel the running turn, send nothing. @param {string} id */
    cancel: (id) => post('api/chat/cancel', { session: id }),
  }),
  /**
   * Web Push opt-in (#535). status() is {state, error}; state is
   * unsupported, unavailable (Desk runs without --push), denied, off or on.
   * subscribe, unsubscribe and test resolve to the new status.
   */
  push: Object.freeze({
    status: () => pushStatus(),
    subscribe: async () => { await enablePush(); return { state: push.state, error: push.error }; },
    unsubscribe: async () => { await disablePush(); return { state: push.state, error: push.error }; },
    test: async () => { await testPush(); return { state: push.state, error: push.error }; },
  }),
  /**
   * Sets the count an app shows on a view's nav entry, 0 to clear
   * (JOB-85). The shell sets attention's from api/attention.
   * @param {string} view @param {number} n
   */
  badge(view, n) { setBadge(view, n); },
  /**
   * Called with {view: count} now and on every change. Returns an
   * unsubscribe function.
   * @param {(b: Record<string, number>) => void} fn
   */
  badges(fn) { return onBadges(fn); },
  /**
   * Called with whether the event stream is connected, now and on every
   * change. Returns an unsubscribe function.
   * @param {(connected: boolean) => void} fn
   */
  live(fn) { return onLive(fn); },
  /** 'dark' or 'light'; setting it applies and remembers it. */
  get theme() { return getTheme(); },
  set theme(v) { setTheme(v); },
});

/**
 * Whether target selects a contribution (api 5). plugin, element and id
 * are names, not match keys: each one given must equal the
 * contribution's. The other keys go to match, which is skipped when only
 * names are given, since the caller chose the contribution already.
 * @param {{ plugin: PluginInfo, c: Contribution }} x @param {Record<string, any>} [target]
 */
function selects(x, target) {
  const rest = { ...(target ?? {}) };
  let named = false;
  for (const [k, v] of /** @type {[string, string][]} */ ([['plugin', x.plugin.name], ['element', x.c.element], ['id', x.c.id ?? '']])) {
    if (typeof rest[k] !== 'string') continue;
    if (rest[k] !== v) return false;
    named = true;
    delete rest[k];
  }
  if (named && Object.keys(rest).length === 0) return true;
  return matches(x.c.match, rest);
}

/** @param {string} slot @param {Record<string, any>} [target] */
function firstFor(slot, target) {
  return contributions(slot).find((x) => selects(x, target)) ?? null;
}

/** @param {HTMLElement} host @param {string} slot @param {Record<string, any>} target @param {any} context */
function mountSlot(host, slot, target, context) {
  const hit = firstFor(slot, target);
  if (!hit) return null;
  return mountOne(host, hit.plugin, hit.c, context);
}

/** @param {HTMLElement} host @param {string} slot @param {any} context @param {Record<string, any>} [target] */
function mountAll(host, slot, context, target) {
  const hits = contributions(slot).filter((x) => selects(x, target));
  if (!hits.length) return null;
  const boxes = hits.map((hit) => {
    const box = document.createElement('div');
    box.className = 'plugin-host';
    box.dataset.plugin = hit.plugin.name;
    box.dataset.slot = slot;
    return box;
  });
  host.replaceChildren(...boxes);
  const handles = hits.map((hit, i) => mountOne(boxes[i], hit.plugin, hit.c, context));
  return {
    /** @param {any} c */
    context(c) { for (const h of handles) h.context(c); },
    destroy() { for (const h of handles) h.destroy(); host.replaceChildren(); },
  };
}

/**
 * Creates one contribution's element in host. An import, define or
 * module error, or an error thrown from the plugin's own files while it
 * is mounted, replaces the element with an error card in host only.
 * @param {HTMLElement} host @param {PluginInfo} plugin @param {Contribution} c @param {any} context
 */
function mountOne(host, plugin, c, context) {
  /** @type {any} */
  let el = null;
  let gone = false;
  let ctx = context ?? {};
  /** @param {unknown} e */
  const fail = (e) => {
    if (gone) return;
    const d = document.createElement('div');
    d.className = 'card plugin-error';
    d.dataset.testid = 'plugin-error';
    d.dataset.plugin = plugin.name;
    d.textContent = `Plugin ${plugin.name} failed: ${/** @type {any} */ (e)?.message || e}`;
    host.replaceChildren(d);
    el = null;
  };
  /** @param {ErrorEvent} e */
  const onErr = (e) => {
    if (e.filename && e.filename.includes('/' + plugin.url)) fail(e.message || e.error);
  };
  window.addEventListener('error', onErr);
  loadPlugin(plugin)
    .then(() => whenDefined(c.element))
    .then(() => {
      if (gone) return;
      el = document.createElement(c.element);
      el.api = deskApi;
      el.config = plugin.config ?? {};
      el.context = ctx;
      host.replaceChildren(el);
    })
    .catch(fail);
  return {
    /** @param {any} next */
    context(next) {
      ctx = next ?? {};
      try { if (el) el.context = ctx; } catch (e) { fail(e); }
    },
    destroy() { gone = true; window.removeEventListener('error', onErr); host.replaceChildren(); el = null; },
  };
}

/** @type {Set<(p: Record<string, string>) => void>} */
const paramListeners = new Set();

onUrl(() => {
  const p = { ...url };
  for (const fn of paramListeners) {
    try { fn(p); } catch (e) { console.error('desk plugin params handler', e); }
  }
});

/** Collects ticket ids in a result so a change to one refetches it. @param {any} v @param {Set<string>} out */
function idsOf(v, out, depth = 0) {
  if (!v || typeof v !== 'object' || depth > 4) return out;
  if (Array.isArray(v)) { for (const x of v) idsOf(x, out, depth + 1); return out; }
  if (typeof v.id === 'string') out.add(v.id);
  for (const k of ['ticket', 'run', 'batch', 'steps', 'tickets', 'items']) if (k in v) idsOf(v[k], out, depth + 1);
  return out;
}

/** @param {string} method @param {any} params @param {{ watch?: Record<string, string> }} [opts] */
function makeQuery(method, params, opts) {
  let state = { data: /** @type {any} */ (null), error: '', loading: true };
  /** @type {Set<(s: any) => void>} */
  const subs = new Set();
  let ids = new Set();
  let timer = 0;
  let closed = false;
  const emit = () => { for (const fn of subs) fn({ ...state }); };
  const load = () => {
    deskApi.call(method, params).then(
      (d) => { if (closed) return; ids = idsOf(d, new Set()); state = { data: d, error: '', loading: false }; emit(); },
      (e) => { if (closed) return; state = { ...state, error: String(e?.message || e), loading: false }; emit(); },
    );
  };
  const watch = Object.entries(opts?.watch || {});
  const off = deskApi.subscribe((ev) => {
    const hit = (ev.id && ids.has(ev.id)) || (watch.length > 0 && watch.every(([k, v]) => ev[k] === v)) || ev.kind === 'lagged';
    if (hit && !timer) timer = window.setTimeout(() => { timer = 0; load(); }, 1000);
  });
  load();
  return {
    get data() { return state.data; },
    get error() { return state.error; },
    get loading() { return state.loading; },
    refresh: load,
    /** @param {(s: any) => void} fn */
    subscribe(fn) {
      subs.add(fn);
      fn({ ...state });
      return () => {
        subs.delete(fn);
        if (subs.size === 0) { closed = true; off(); clearTimeout(timer); }
      };
    },
  };
}

export const who = $state({ user: '' });

/** @type {Map<string, Promise<void>>} */
const imports = new Map();

/**
 * Imports a plugin's entry once. Rejects with the import error, or with
 * an API version error before importing.
 * @param {PluginInfo} p
 */
export function loadPlugin(p) {
  let pr = imports.get(p.name);
  if (!pr) {
    const need = Number(String(p.api || '').replace(/^>=\s*/, '')) || 0;
    pr = need > API_VERSION
      ? Promise.reject(new Error(`needs Desk plugin API ${need}, this Desk has ${API_VERSION}`))
      : import(/* @vite-ignore */ urlOf(p.url + p.entry)).then(() => {});
    imports.set(p.name, pr);
  }
  return pr;
}

/**
 * Waits for a custom element, failing after ms so a plugin that never
 * defines it shows an error card instead of an empty slot.
 * @param {string} name @param {number} [ms]
 */
export function whenDefined(name, ms = 10000) {
  return Promise.race([
    customElements.whenDefined(name),
    new Promise((_, rej) => setTimeout(() => rej(new Error(`element <${name}> was not defined`)), ms)),
  ]);
}

export async function loadPlugins() {
  try {
    plugins.list = await get('api/plugins');
    plugins.error = '';
  } catch (e) {
    plugins.error = String(/** @type {any} */ (e).message || e);
  }
  plugins.loaded = true;
}

/**
 * Contributions for a slot, with their plugin.
 * @param {string} slot
 * @returns {{ plugin: PluginInfo, c: Contribution }[]}
 */
export function contributions(slot) {
  const out = [];
  for (const p of plugins.list) for (const c of p.contributes) if (c.slot === slot) out.push({ plugin: p, c });
  return out;
}

/**
 * Whether a contribution's match accepts a target. Every listed key must
 * accept one of its values; no match accepts everything.
 * @param {Record<string, string[]> | undefined} match
 * @param {Record<string, string | string[] | undefined>} target
 */
export function matches(match, target) {
  if (!match) return true;
  return Object.entries(match).every(([k, want]) => {
    const have = target[k];
    const vals = Array.isArray(have) ? have : have ? [have] : [];
    return vals.some((v) => want.some((w) => w === v || (w.endsWith('/*') && v.startsWith(w.slice(0, -1)))));
  });
}

export { url };
