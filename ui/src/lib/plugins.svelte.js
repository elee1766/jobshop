// Desk plugin host (#517). Plugins are custom elements; Desk imports a
// plugin's entry once, waits for its elements, then sets api, context
// and config on each element it creates. Built-in views use the same
// path, so this file is the whole contract on the browser side.

import { base, api as urlOf, get, post, live } from './api.svelte.js';
import { url, go } from './url.svelte.js';

/** Bumped when api gains a member a plugin may need; plugin.json "api": ">=N". */
export const API_VERSION = 1;

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
  navigate(view, params) { go({ view, ticket: '', session: '', artifact: '', ...(params || {}) }); },
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
});

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
