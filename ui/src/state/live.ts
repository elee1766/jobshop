import { atom } from 'jotai';
import type {
  AttentionItem,
  BrowseScopeResult,
  Event,
  Learning,
  MemoryEdit,
  RunResult,
} from '../../../packages/jobshop-client/rpc.js';
import type {
  AutomationAudit,
  ChatSession,
  EmbedView,
  LiveEvent,
  SystemView,
  TraceView,
} from '../../../packages/jobshop-client/index.js';
import { client } from '../lib/api.ts';
import { reconcile } from '../lib/reconcile.ts';
import { parseRoute } from '../lib/route.ts';
import { collection } from './entities.ts';
import { store } from './store.ts';

// Live application data (JOB-150). One owner (LiveOwner) opens the single
// api/live stream for the tab and writes here; screens only read atoms.
// Collections keep membership apart from contents, so a changed ticket
// renders its own row and nothing else.

export const attentionKinds = ['question', 'decision', 'ticket', 'run', 'wait', 'guard', 'pool'];
export const attentionKey = (item: AttentionItem) => item.kind + ':' + item.id;

export const attention = collection<AttentionItem>(attentionKey);
export const runs = collection<RunResult>((run) => run.run.id);
export const chats = collection<ChatSession>((chat) => chat.id);

export const systemAtom = atom<SystemView | null>(null);
export const improvementsAtom = atom<{ learnings: Learning[]; edits: MemoryEdit[] } | null>(null);
// The configured iframe tabs (desk.embeds). Null until the first snapshot, so
// a direct link does not claim an embed is missing before Desk has asked.
export const embedsAtom = atom<EmbedView[] | null>(null);
// What the Linear and GitHub views may read, and the revision that moves when
// the controller learns the providers changed (JOB-163). Null until the first
// snapshot. Lists and details are read on demand, not streamed.
export const browseAtom = atom<BrowseScopeResult | null>(null);
export const traceAtom = atom<TraceView | null>(null);
// The ticket the trace page details: the inspected one, else the root.
export const selectedAtom = atom<TraceView['ticket'] | null>(null);
export const auditAtom = atom<AutomationAudit | { id: string; error: string } | null>(null);
export const truncatedAtom = atom(false);
export const connectedAtom = atom(false);
export const errorAtom = atom('');
export const nowAtom = atom(Date.now());
export const lastFocusAtom = atom('');

// The chat panel handles commands the controller sends to the open chat.
type DeskListener = (event: Event) => void;
const deskListeners = new Set<DeskListener>();
export function onDeskCommand(fn: DeskListener) {
  deskListeners.add(fn);
  return () => void deskListeners.delete(fn);
}

function receive(event: LiveEvent) {
  switch (event.kind) {
    case 'attention':
      attention.apply(
        store,
        (event.data.items || []).filter((item) => attentionKinds.includes(item.kind)),
      );
      break;
    case 'overview':
      runs.apply(store, event.data.runs || []);
      store.set(truncatedAtom, !!event.data.truncated);
      break;
    case 'system':
      store.set(systemAtom, (prev) => reconcile(prev, event.data));
      store.set(nowAtom, Date.now());
      break;
    case 'automation': {
      const route = parseRoute(location.search, location.hash);
      if (event.data.id === route.audit) store.set(auditAtom, (prev) => reconcile(prev, event.data));
      break;
    }
    case 'scheduler': {
      store.set(nowAtom, Date.now());
      const status = event.data.system;
      if (status) store.set(systemAtom, (prev) => (prev ? { ...prev, status } : prev));
      break;
    }
    case 'chats':
      chats.apply(store, event.data.sessions || []);
      break;
    case 'improvements':
      store.set(improvementsAtom, (prev) => reconcile(prev, event.data));
      break;
    case 'embeds':
      store.set(embedsAtom, (prev) => reconcile(prev, event.data.embeds));
      break;
    case 'browse':
      store.set(browseAtom, (prev) => reconcile(prev, event.data));
      break;
    case 'trace': {
      if (parseRoute(location.search, location.hash).view !== 'trace') break;
      if ('error' in event.data) {
        store.set(errorAtom, event.data.error);
        break;
      }
      const data = event.data;
      store.set(traceAtom, (prev) => reconcile(prev, data));
      store.set(selectedAtom, data.inspected || data.ticket);
      break;
    }
    case 'desk':
      for (const fn of deskListeners) fn(event.data);
      break;
  }
}

export interface StreamParams {
  trace?: string;
  inspect?: string;
  timeline?: boolean;
  audit?: string;
}

// openStream connects the live stream and returns its close function. The
// caller owns the connection: there is one per tab.
export function openStream(params: StreamParams) {
  return client.live(receive, {
    ...params,
    onState: (state, detail) => {
      store.set(connectedAtom, state === 'connected');
      if (state === 'error') store.set(errorAtom, detail?.error || 'Live updates unavailable');
      if (state === 'connected') store.set(errorAtom, '');
    },
  });
}

// resetLive drops every scoped atom, so nothing from one controller shows
// under another (an endpoint change or a test).
export function resetLive() {
  attention.clear(store);
  runs.clear(store);
  chats.clear(store);
  store.set(systemAtom, null);
  store.set(improvementsAtom, null);
  store.set(embedsAtom, null);
  store.set(browseAtom, null);
  store.set(traceAtom, null);
  store.set(selectedAtom, null);
  store.set(auditAtom, null);
  store.set(truncatedAtom, false);
  store.set(connectedAtom, false);
  store.set(errorAtom, '');
}
