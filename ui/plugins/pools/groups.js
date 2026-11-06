// Grouping and collapse rules for the Pools and triggers view (JOB-81).
// Pure functions over the rows api/pools returns, so the view stays thin.

/**
 * @typedef {{ name: string, project?: string, labels?: string[], desired: number, open: number,
 *   working: number, idle: number, cordoned: number, min: number, max: number, target?: number | null,
 *   drained?: boolean, why?: string, until?: string, quarantined?: boolean, paused?: boolean }} PoolRow
 * @typedef {{ name: string, holds?: boolean, why?: string, open?: string, stalls?: string[], labels?: string[],
 *   last_fired?: string, next_cron?: string }} TriggerRow
 */

/** Group key shown for rows that lack the grouping label. */
export const NONE = '(none)';

/**
 * One row per pool: the scale rows plus quarantined and paused pools the
 * scale list does not carry, with the status folded in as fields.
 * @param {any} data api/pools
 * @returns {PoolRow[]}
 */
export function poolRows(data) {
  const quarantined = data?.quarantined ?? {};
  const paused = data?.paused ?? {};
  /** @type {Map<string, PoolRow>} */
  const rows = new Map();
  for (const sp of data?.pools ?? []) rows.set(sp.name, { ...sp });
  for (const name of new Set([...Object.keys(quarantined), ...Object.keys(paused)])) {
    if (!rows.has(name)) rows.set(name, { name, desired: 0, open: 0, working: 0, idle: 0, cordoned: 0, min: 0, max: 0 });
  }
  for (const [name, why] of Object.entries(quarantined)) Object.assign(/** @type {PoolRow} */ (rows.get(name)), { quarantined: true, why });
  for (const [name, until] of Object.entries(paused)) Object.assign(/** @type {PoolRow} */ (rows.get(name)), { paused: true, until });
  return [...rows.values()].sort((a, b) => a.name.localeCompare(b.name));
}

/**
 * A pool with no sessions, nothing queued and no override. Quiet pools
 * fold away; quarantined, paused, drained, target-overridden and active
 * ones never do.
 * @param {PoolRow} p
 */
export function poolQuiet(p) {
  return !p.quarantined && !p.paused && !p.drained && p.target == null && !p.desired && !p.open && !p.working && !p.cordoned;
}

/**
 * A trigger that is only waiting: it does not fire next and has no stalls.
 * @param {TriggerRow} t
 */
export function triggerQuiet(t) {
  return !t.holds && !t.stalls?.length;
}

/**
 * The group a row belongs to for a grouping key: "none", "project", or
 * a label key k (the value of its k:v label).
 * @param {{ project?: string, labels?: string[] }} row @param {string} key
 */
export function groupOf(row, key) {
  if (key === 'none') return '';
  if (key === 'project') return row.project || NONE;
  const prefix = key + ':';
  const l = (row.labels ?? []).find((x) => x.startsWith(prefix));
  return l ? l.slice(prefix.length) : NONE;
}

/**
 * The grouping keys on offer: the fixed ones, then every label key (the
 * part before the first colon) in use.
 * @param {{ labels?: string[] }[]} rows
 */
export function groupKeys(rows) {
  const keys = new Set(['instance', 'pack']);
  for (const r of rows) for (const l of r.labels ?? []) {
    const i = l.indexOf(':');
    if (i > 0) keys.add(l.slice(0, i));
  }
  return ['none', 'project', ...keys];
}

/**
 * Rows matching the search (a name substring) and every selected label.
 * @template {{ name: string, labels?: string[] }} T
 * @param {T[]} rows @param {string} q @param {string[]} labels
 * @returns {T[]}
 */
export function filterRows(rows, q, labels) {
  const needle = q.trim().toLowerCase();
  return rows.filter((r) => (!needle || r.name.toLowerCase().includes(needle)) && labels.every((l) => (r.labels ?? []).includes(l)));
}

/**
 * Every label on the rows with how many rows carry it, most common first.
 * @param {{ labels?: string[] }[]} rows
 */
export function labelCounts(rows) {
  /** @type {Map<string, number>} */
  const m = new Map();
  for (const r of rows) for (const l of r.labels ?? []) m.set(l, (m.get(l) ?? 0) + 1);
  return [...m.entries()].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]));
}

/**
 * Pool groups, busiest first: groups with anything needing attention
 * (quarantined, paused, working) come before idle ones, then by name.
 * @param {PoolRow[]} rows @param {string} key
 */
export function poolGroups(rows, key) {
  /** @type {Map<string, PoolRow[]>} */
  const by = new Map();
  for (const r of rows) {
    const g = groupOf(r, key);
    by.set(g, [...(by.get(g) ?? []), r]);
  }
  const out = [...by.entries()].map(([name, pools]) => {
    const sum = (/** @type {(p: PoolRow) => number} */ f) => pools.reduce((n, p) => n + f(p), 0);
    return {
      name,
      pools,
      desired: sum((p) => p.desired), open: sum((p) => p.open), working: sum((p) => p.working), idle: sum((p) => p.idle),
      quarantined: sum((p) => (p.quarantined ? 1 : 0)),
      paused: sum((p) => (p.paused ? 1 : 0)),
      drained: sum((p) => (p.drained ? 1 : 0)),
      targeted: sum((p) => (p.target != null ? 1 : 0)),
      quiet: pools.every(poolQuiet),
      // Pools that always show, and the idle ones folded behind a row.
      shown: pools.filter((p) => !poolQuiet(p)),
      folded: pools.filter(poolQuiet),
    };
  });
  const rank = (/** @type {typeof out[number]} */ g) => (g.quarantined || g.paused ? 0 : g.working || g.open ? 1 : g.quiet ? 3 : 2);
  return out.sort((a, b) => rank(a) - rank(b) || a.name.localeCompare(b.name));
}

/**
 * Trigger groups: stalled or firing ones first, in the groups too.
 * @param {TriggerRow[]} rows @param {string} key
 */
export function triggerGroups(rows, key) {
  const k = key === 'project' ? 'instance' : key;
  /** @type {Map<string, TriggerRow[]>} */
  const by = new Map();
  for (const r of rows) {
    const g = groupOf(r, k);
    by.set(g, [...(by.get(g) ?? []), r]);
  }
  const rank = (/** @type {TriggerRow} */ t) => (t.stalls?.length ? 0 : t.holds ? 1 : 2);
  const out = [...by.entries()].map(([name, triggers]) => {
    const sorted = [...triggers].sort((a, b) => rank(a) - rank(b) || a.name.localeCompare(b.name));
    return {
      name,
      triggers: sorted,
      firing: triggers.filter((t) => t.holds).length,
      stalled: triggers.filter((t) => t.stalls?.length).length,
      quiet: triggers.every(triggerQuiet),
      shown: sorted.filter((t) => !triggerQuiet(t)),
      folded: sorted.filter(triggerQuiet),
    };
  });
  return out.sort((a, b) => (a.quiet === b.quiet ? a.name.localeCompare(b.name) : a.quiet ? 1 : -1) || b.stalled - a.stalled);
}
