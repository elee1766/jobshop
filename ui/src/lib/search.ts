// Shared client-side fuzzy search over a local collection (JOB-150).
//
// Only what is loaded is searched: a caller says so when the server holds
// more, and server-side search stays authoritative. Ranking, best first:
//   1. an exact identifier,
//   2. an identifier prefix,
//   3. Fuse's weighted fuzzy match over the named fields,
//   4. a subsequence of the label (the older picker behavior: the first and
//      last letters of a title still find it).
import Fuse, { type IFuseOptions } from 'fuse.js';

export interface SearchField<T> {
  name: string;
  weight: number;
  get: (item: T) => string | undefined;
}

export interface Searcher<T> {
  search(query: string): T[];
}

export interface SearchSpec<T> {
  fields: SearchField<T>[];
  // Identifiers that rank first on an exact or prefix match.
  ids: (item: T) => string[];
  // The text a subsequence query is matched against.
  label?: (item: T) => string;
  fuse?: Partial<IFuseOptions<unknown>>;
}

const norm = (value: string | undefined) => (value || '').toLocaleLowerCase();

function subsequence(text: string, query: string) {
  let index = 0;
  for (const char of norm(text)) if (char === query[index]) index++;
  return index === query.length;
}

// createSearcher indexes items once. Build a new one when the collection
// changes: at Desk's collection sizes a rebuild is cheap and keeps the index
// in step with live additions, updates and removals.
export function createSearcher<T>(items: T[], spec: SearchSpec<T>): Searcher<T> {
  const docs = items.map((item) => {
    const doc: Record<string, string> = {};
    for (const field of spec.fields) doc[field.name] = field.get(item) || '';
    return { item, doc };
  });
  const fuse = new Fuse(docs, {
    keys: spec.fields.map((field) => ({ name: ['doc', field.name], weight: field.weight })),
    threshold: 0.34,
    ignoreLocation: true,
    ...spec.fuse,
  });
  return {
    search(query) {
      const q = norm(query.trim());
      if (!q) return items;
      const exact: T[] = [];
      const prefix: T[] = [];
      for (const { item } of docs) {
        const found = spec.ids(item).map(norm);
        if (found.includes(q)) exact.push(item);
        else if (found.some((id) => id.startsWith(q))) prefix.push(item);
      }
      const taken = new Set([...exact, ...prefix]);
      const fuzzy = fuse
        .search(q.replace(/\s+/g, ' '))
        .map((hit) => hit.item.item)
        .filter((item) => !taken.has(item));
      fuzzy.forEach((item) => taken.add(item));
      const squashed = q.replace(/\s+/g, '');
      const loose = spec.label
        ? items.filter((item) => !taken.has(item) && subsequence(spec.label!(item), squashed))
        : [];
      return [...exact, ...prefix, ...fuzzy, ...loose];
    },
  };
}
