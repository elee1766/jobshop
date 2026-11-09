// Structural sharing for live snapshots: a new snapshot keeps the previous
// reference for every part that did not change, so memoized components and
// narrow atoms skip work on an unchanged ticket (JOB-150).
export function reconcile<T>(prev: T, next: T): T {
  if (prev === next) return prev;
  if (Array.isArray(prev) && Array.isArray(next)) {
    const items = next.map((item, i) => reconcile(prev[i], item));
    return (items.length === prev.length && items.every((item, i) => item === prev[i]) ? prev : items) as T;
  }
  if (
    prev &&
    next &&
    typeof prev === 'object' &&
    typeof next === 'object' &&
    !Array.isArray(prev) &&
    !Array.isArray(next)
  ) {
    const before = prev as Record<string, unknown>;
    const after = next as Record<string, unknown>;
    const keys = Object.keys(after);
    const merged: Record<string, unknown> = {};
    let same = keys.length === Object.keys(before).length;
    for (const key of keys) {
      merged[key] = reconcile(before[key], after[key]);
      if (merged[key] !== before[key]) same = false;
    }
    return (same ? prev : merged) as T;
  }
  return next;
}
