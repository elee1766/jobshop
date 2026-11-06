// Shell signals any app can show (JOB-85): view badges and the theme.
// The built-in frame used to compute these itself; they live in the
// shell now so a replacement app gets them through the api.

/** @type {Record<string, number>} */
let badges = {};
/** @type {Set<(b: Record<string, number>) => void>} */
const badgeFns = new Set();

/**
 * Sets the attention count shown on a view's nav entry; 0 clears it.
 * @param {string} view @param {number} n
 */
export function setBadge(view, n) {
  const v = Math.max(0, Math.floor(Number(n) || 0));
  if ((badges[view] ?? 0) === v) return;
  const next = { ...badges };
  if (v) next[view] = v;
  else delete next[view];
  badges = next;
  for (const fn of badgeFns) {
    try { fn({ ...badges }); } catch (e) { console.error('desk badges handler', e); }
  }
}

/**
 * Called with every view's badge now and on each change. Returns an
 * unsubscribe function.
 * @param {(b: Record<string, number>) => void} fn
 */
export function onBadges(fn) {
  badgeFns.add(fn);
  fn({ ...badges });
  return () => { badgeFns.delete(fn); };
}

/** @typedef {'dark' | 'light'} Theme */

/** @type {Theme} */
let theme = localStorage.getItem('desk-theme') === 'light' ? 'light' : 'dark';
// Applied before any app mounts, so a light theme does not flash dark.
document.documentElement.dataset.theme = theme;

export function getTheme() { return theme; }

/** @param {string} t */
export function setTheme(t) {
  theme = t === 'light' ? 'light' : 'dark';
  document.documentElement.dataset.theme = theme;
  localStorage.setItem('desk-theme', theme);
}
