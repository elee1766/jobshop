import { atom, type Atom, type PrimitiveAtom, type Store } from 'jotai';
import { reconcile } from '../lib/reconcile.ts';

// A collection keeps list membership and order apart from entity contents:
// `ids` changes only when the set or order changes, and each entity atom
// changes only when that entity's contents do. A row subscribes to its own
// atom, so an update to one ticket does not render the others (JOB-150).
export interface Collection<T> {
  ids: PrimitiveAtom<string[]>;
  entity(id: string): Atom<T | undefined>;
  // Applies a full snapshot (initial snapshot, patch or reconnect alike).
  apply(store: Store, items: T[]): void;
  // Drops everything, for an endpoint change.
  clear(store: Store): void;
}

export function collection<T>(keyOf: (item: T) => string): Collection<T> {
  const ids = atom<string[]>([]);
  // Atoms are kept once made, so a subscriber never holds an orphan: a
  // removed entity reads undefined and a returning one updates the same atom.
  const atoms = new Map<string, PrimitiveAtom<T | undefined>>();
  const slot = (id: string) => {
    let a = atoms.get(id);
    if (!a) {
      a = atom<T | undefined>(undefined);
      atoms.set(id, a);
    }
    return a;
  };
  return {
    ids,
    entity: slot,
    apply(store, items) {
      const next: string[] = [];
      const seen = new Set<string>();
      for (const item of items) {
        const id = keyOf(item);
        if (seen.has(id)) continue;
        seen.add(id);
        next.push(id);
        const a = slot(id);
        const prev = store.get(a);
        const merged = reconcile(prev, item);
        if (merged !== prev) store.set(a, merged);
      }
      for (const [id, a] of atoms) if (!seen.has(id) && store.get(a) !== undefined) store.set(a, undefined);
      const prevIds = store.get(ids);
      if (prevIds.length !== next.length || prevIds.some((id, i) => id !== next[i])) store.set(ids, next);
    },
    clear(store) {
      for (const a of atoms.values()) if (store.get(a) !== undefined) store.set(a, undefined);
      store.set(ids, []);
    },
  };
}
