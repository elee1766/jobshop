import { useAtomValue } from 'jotai';
import { useEffect, useSyncExternalStore } from 'react';
import type {
  BrowseCatalogResult,
  BrowseDetail,
  BrowseItem,
  BrowseListParams,
  BrowseListResult,
  BrowseStatus,
  BrowseThread,
} from '../../../packages/jobshop-client/rpc.js';
import { client } from '../lib/api.ts';
import { browseAtom } from './live.ts';

// Reads for the Linear and GitHub views (JOB-163). Everything goes through
// the controller (browse.*), which owns the scope, the cache and the provider
// limits. Here each query or item is a small external store: components read
// it with useSyncExternalStore and an effect only starts or stops the read, so
// no component copies fetched data into its own state.
//
// A read is aborted when it is no longer wanted, the pages already loaded
// survive a refresh, and a new live revision (the controller saw the provider
// move) reads again in place without a loading flash. The last queries and
// items stay in memory (bounded), so going back to one shows it at once.

const message = (e: unknown) => (e instanceof Error ? e.message : String(e));
const aborted = (e: unknown) => e instanceof DOMException && e.name === 'AbortError';

class Resource<S> {
  private listeners = new Set<() => void>();
  state: S;
  constructor(initial: S) {
    this.state = initial;
  }
  subscribe = (fn: () => void) => {
    this.listeners.add(fn);
    return () => void this.listeners.delete(fn);
  };
  get = () => this.state;
  protected set(next: S | ((s: S) => S)) {
    this.state = typeof next === 'function' ? (next as (s: S) => S)(this.state) : next;
    for (const fn of this.listeners) fn();
  }
}

// A bounded most-recently-used map of resources.
function remember<T>(map: Map<string, T>, key: string, make: () => T, limit: number): T {
  let found = map.get(key);
  if (found) {
    map.delete(key);
  } else {
    found = make();
  }
  map.set(key, found);
  if (map.size > limit) map.delete(map.keys().next().value as string);
  return found;
}

// --- lists ----------------------------------------------------------------

export type ListParams = Omit<BrowseListParams, 'cursor' | 'limit' | 'refresh'>;

export interface ListState {
  items: BrowseItem[];
  next: string;
  hasMore: boolean;
  total: number;
  status?: BrowseStatus;
  // The first page is loading and nothing is shown yet.
  loading: boolean;
  // Another page, or a refresh of what is shown, is loading.
  busy: boolean;
  // The call itself failed (scope refused, Desk forbids browse, no network).
  error: string;
}

// mergeItems keeps the refreshed first page and the older pages that were
// loaded after it, without repeating an item that moved up.
export function mergeItems(first: BrowseItem[], loaded: BrowseItem[], pageSize: number): BrowseItem[] {
  const seen = new Set(first.map((item) => item.ref));
  return [...first, ...loaded.slice(pageSize).filter((item) => !seen.has(item.ref))];
}

class ListResource extends Resource<ListState> {
  private control: AbortController | null = null;
  private revision = -1;
  private started = false;
  constructor(
    private params: ListParams,
    private pageSize: number,
  ) {
    super({ items: [], next: '', hasMore: false, total: 0, loading: true, busy: false, error: '' });
  }

  private call(cursor: string, refresh: boolean, signal: AbortSignal) {
    return client.call(
      'browse.list',
      { ...this.params, cursor, limit: this.pageSize, refresh },
      { signal },
    ) as Promise<BrowseListResult>;
  }

  // sync starts the first read, and reads again when the revision moved.
  sync(revision: number) {
    const quiet = this.started && this.revision !== revision;
    if (this.started && !quiet) return;
    this.started = true;
    this.revision = revision;
    this.first(false, quiet);
  }

  private first(refresh: boolean, quiet: boolean) {
    this.control?.abort();
    const ctl = new AbortController();
    this.control = ctl;
    if (!quiet) this.set((s) => ({ ...s, busy: true, error: '' }));
    this.call('', refresh, ctl.signal)
      .then((res) => {
        if (ctl.signal.aborted) return;
        this.set((s) => {
          if (res.status.state === 'error') {
            // A failed read keeps what was shown and says why.
            return { ...s, status: res.status, loading: false, busy: false, error: '' };
          }
          const paged = s.items.length > this.pageSize;
          return {
            items: paged ? mergeItems(res.items, s.items, this.pageSize) : res.items,
            next: paged ? s.next : res.next || '',
            hasMore: paged ? s.hasMore : res.has_more,
            total: res.total || 0,
            status: res.status,
            loading: false,
            busy: false,
            error: '',
          };
        });
      })
      .catch((e) => {
        if (aborted(e) || ctl.signal.aborted) return;
        this.set((s) => ({ ...s, loading: false, busy: false, error: message(e) }));
      });
  }

  refresh = () => this.first(true, false);

  more = () => {
    const s = this.state;
    if (!s.next || s.busy) return;
    const ctl = new AbortController();
    this.set((x) => ({ ...x, busy: true }));
    this.call(s.next, false, ctl.signal)
      .then((res) => {
        this.set((x) => {
          if (res.status.state === 'error') {
            return { ...x, busy: false, error: res.status.message || 'The next page could not be read.' };
          }
          const seen = new Set(x.items.map((item) => item.ref));
          return {
            ...x,
            items: [...x.items, ...res.items.filter((item) => !seen.has(item.ref))],
            next: res.next || '',
            hasMore: res.has_more,
            status: res.status,
            busy: false,
            error: '',
          };
        });
      })
      .catch((e) => {
        if (!aborted(e)) this.set((x) => ({ ...x, busy: false, error: message(e) }));
      });
  };

  // release drops a read nobody is waiting for. The next sync reads again.
  release() {
    this.control?.abort();
    this.control = null;
    this.started = false;
    this.set((s) => ({ ...s, busy: false }));
  }
}

const lists = new Map<string, ListResource>();

export function useBrowseList(params: ListParams, pageSize = 50) {
  const key = JSON.stringify([params, pageSize]);
  const revision = useAtomValue(browseAtom)?.revision ?? 0;
  const res = remember(lists, key, () => new ListResource(params, pageSize), 24);
  const state = useSyncExternalStore(res.subscribe, res.get);
  useEffect(() => {
    res.sync(revision);
  }, [res, revision]);
  useEffect(() => () => res.release(), [res]);
  return { ...state, loadMore: res.more, refresh: res.refresh };
}

// --- details --------------------------------------------------------------

export interface DetailState {
  detail: BrowseDetail | null;
  loading: boolean;
  older: boolean;
  error: string;
  // Comment ids the reader has already had in front of them. A comment that
  // arrives later is new.
  seen: Set<string>;
}

export const commentIds = (threads: BrowseThread[]) => threads.flatMap((t) => t.comments.map((c) => c.id));

// mergeThreads folds more threads into a list by id: a thread that is in both
// gets the comments of both, oldest first.
export function mergeThreads(base: BrowseThread[], more: BrowseThread[]): BrowseThread[] {
  const out = base.map((t) => ({ ...t, comments: [...t.comments] }));
  for (const thread of more) {
    const at = out.findIndex((t) => t.id === thread.id);
    if (at < 0) {
      out.push(thread);
      continue;
    }
    const have = new Set(out[at].comments.map((c) => c.id));
    const merged = [...out[at].comments, ...thread.comments.filter((c) => !have.has(c.id))].toSorted(
      (a, b) => Date.parse(a.at) - Date.parse(b.at),
    );
    out[at] = { ...out[at], comments: merged, hidden: Math.max(out[at].hidden ?? 0, thread.hidden ?? 0) || undefined };
  }
  return out;
}

class DetailResource extends Resource<DetailState> {
  private control: AbortController | null = null;
  private revision = -1;
  private started = false;
  private pagedOlder = false;
  constructor(
    private provider: string,
    private ref: string,
    private kind: string,
  ) {
    super({ detail: null, loading: true, older: false, error: '', seen: new Set() });
  }

  private call(cursor: string, refresh: boolean, signal: AbortSignal) {
    return client.call(
      'browse.show',
      { provider: this.provider, ref: this.ref, kind: this.kind, cursor, refresh },
      { signal },
    ) as Promise<BrowseDetail>;
  }

  sync(revision: number) {
    const quiet = this.started && this.revision !== revision;
    if (this.started && !quiet) return;
    this.started = true;
    this.revision = revision;
    this.first(false);
  }

  private first(refresh: boolean) {
    this.control?.abort();
    const ctl = new AbortController();
    this.control = ctl;
    this.call('', refresh, ctl.signal)
      .then((res) => {
        if (ctl.signal.aborted) return;
        this.set((s) => {
          if (res.status.state === 'error' && s.detail) {
            // Keep what is shown and say why it is not newer.
            return { ...s, loading: false, error: '', detail: { ...s.detail, status: res.status } };
          }
          const detail =
            this.pagedOlder && s.detail
              ? { ...res, threads: mergeThreads(res.threads, s.detail.threads), older: s.detail.older }
              : res;
          // The first read is what the reader has seen; later ones add new.
          const seen = s.detail ? s.seen : new Set(commentIds(res.threads));
          return { detail, loading: false, older: false, error: '', seen };
        });
      })
      .catch((e) => {
        if (aborted(e) || ctl.signal.aborted) return;
        this.set((s) => ({ ...s, loading: false, error: message(e) }));
      });
  }

  refresh = () => this.first(true);

  loadOlder = () => {
    const s = this.state;
    if (!s.detail?.older || s.older) return;
    const ctl = new AbortController();
    this.set((x) => ({ ...x, older: true }));
    this.call(s.detail.older, false, ctl.signal)
      .then((res) => {
        this.pagedOlder = true;
        this.set((x) => {
          if (!x.detail) return { ...x, older: false };
          if (res.status.state === 'error') {
            return { ...x, older: false, error: res.status.message || 'Older activity could not be read.' };
          }
          const threads = mergeThreads(x.detail.threads, res.threads);
          // What was paged in is read, not new.
          const seen = new Set([...x.seen, ...commentIds(res.threads)]);
          return { ...x, older: false, error: '', seen, detail: { ...x.detail, threads, older: res.older } };
        });
      })
      .catch((e) => {
        if (!aborted(e)) this.set((x) => ({ ...x, older: false, error: message(e) }));
      });
  };

  acknowledge = () => {
    this.set((s) => (s.detail ? { ...s, seen: new Set(commentIds(s.detail.threads)) } : s));
  };

  release() {
    this.control?.abort();
    this.control = null;
    this.started = false;
  }
}

const details = new Map<string, DetailResource>();

export function useBrowseDetail(provider: string, ref: string, kind: string) {
  const revision = useAtomValue(browseAtom)?.revision ?? 0;
  const res = remember(details, `${provider}|${ref}|${kind}`, () => new DetailResource(provider, ref, kind), 12);
  const state = useSyncExternalStore(res.subscribe, res.get);
  useEffect(() => {
    res.sync(revision);
  }, [res, revision]);
  useEffect(() => () => res.release(), [res]);
  return { ...state, loadOlder: res.loadOlder, acknowledge: res.acknowledge, refresh: res.refresh };
}

// --- catalog --------------------------------------------------------------

class CatalogResource extends Resource<BrowseCatalogResult | null> {
  private started = false;
  constructor(private team: string) {
    super(null);
  }
  sync() {
    if (this.started) return;
    this.started = true;
    client
      .call('browse.catalog', { team: this.team })
      .then((res) => {
        // A failed read stays unstarted so the next visit asks again.
        if (res.status.state === 'error') this.started = false;
        else this.set(res);
      })
      .catch(() => {
        this.started = false;
      });
  }
}

const catalogs = new Map<string, CatalogResource>();
const none = new CatalogResource('');
none.sync = () => {};

export function useBrowseCatalog(team: string) {
  const res = team ? remember(catalogs, team, () => new CatalogResource(team), 8) : none;
  const state = useSyncExternalStore(res.subscribe, res.get);
  useEffect(() => {
    res.sync();
  }, [res]);
  return state;
}
