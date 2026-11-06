// Session console layout and tabs. Desktop: a resizable strip along the
// bottom of the page column (height persisted). Phone: full screen with a
// tab bar to dismiss it. The URL's session is the active tab; the tab list
// is kept in localStorage so it survives reloads.

const FLOOR = 160;
const PAGE_MIN = 120; // the page above keeps a strip visible
const DEFAULT = 360;
const MAX_TABS = 12;

/** @param {number} h */
function clamp(h) {
  return Math.round(Math.max(FLOOR, Math.min(window.innerHeight - PAGE_MIN, h)));
}

const saved = Number(localStorage.getItem('desk-console-h'));
const mq = window.matchMedia('(max-width: 768px)');

export const panel = $state({ h: clamp(saved > 0 ? saved : DEFAULT), mobile: mq.matches });

mq.addEventListener('change', (e) => { panel.mobile = e.matches; });
window.addEventListener('resize', () => { panel.h = clamp(panel.h); });

/** @param {number} h */
export function setHeight(h) {
  panel.h = clamp(h);
  localStorage.setItem('desk-console-h', String(panel.h));
}

export function resetHeight() {
  setHeight(DEFAULT);
}

/** @typedef {{ id: string, item: string, ro: string }} Tab */

/** @returns {Tab[]} */
function load() {
  try {
    const v = JSON.parse(localStorage.getItem('desk-console-tabs') || '[]');
    return Array.isArray(v) ? v.filter((t) => t && typeof t.id === 'string').slice(0, MAX_TABS) : [];
  } catch {
    return [];
  }
}

export const tabs = $state({ list: load() });

function save() {
  localStorage.setItem('desk-console-tabs', JSON.stringify(tabs.list));
}

/** Adds or updates the tab for the URL's session. @param {Record<string, string>} url */
export function track(url) {
  if (!url.session) return;
  const t = { id: url.session, item: url.item || '', ro: url.ro || '' };
  const i = tabs.list.findIndex((x) => x.id === t.id);
  if (i < 0) tabs.list = [...tabs.list, t].slice(-MAX_TABS);
  else if (tabs.list[i].item !== t.item || tabs.list[i].ro !== t.ro) tabs.list[i] = t;
  else return;
  save();
}

/** @param {any} api @param {Tab} t */
export function show(api, t) {
  api.openSession(t.id, { item: t.item, readonly: t.ro === '1' });
}

/** Closes one tab; when it was active, the neighbor takes over, or the console hides. @param {any} api @param {string} active @param {string} id */
export function closeTab(api, active, id) {
  const i = tabs.list.findIndex((x) => x.id === id);
  if (i < 0) return;
  tabs.list = tabs.list.filter((x) => x.id !== id);
  save();
  if (active !== id) return;
  const next = tabs.list[Math.min(i, tabs.list.length - 1)];
  if (next) show(api, next);
  else api.go({ session: '', item: '', ro: '' });
}
