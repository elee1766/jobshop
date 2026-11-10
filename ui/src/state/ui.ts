import { atom } from 'jotai';
import type { ChatSession } from '../../../packages/jobshop-client/index.js';
import { flightsOf } from '../lib/model.ts';
import { attention, chats, nowAtom, runs, systemAtom } from './live.ts';

// Transient UI state (JOB-150). Navigation state (view, ref, filter, ...)
// is the URL, read through lib/route.ts, not mirrored here.

export type Chat = ChatSession;

export const focusedAtom = atom('');
export const phoneTabAtom = atom('waiting');
export const chatOpenAtom = atom(false);
export const watchedSessionAtom = atom('');
export const collapsedAtom = atom<Record<string, boolean>>({});
export const sortOldestAtom = atom(true);
export const hideHoldsAtom = atom(false);
export const keyboardAtom = atom(false);

// Whether the chat rail is hidden is remembered in this browser.
function readHidden() {
  try {
    return localStorage.getItem('desk.chat.hidden') === 'true';
  } catch {
    return false;
  }
}
const hiddenBase = atom(readHidden());
export const chatHiddenAtom = atom(
  (get) => get(hiddenBase),
  (_get, set, value: boolean) => {
    set(hiddenBase, value);
    try {
      localStorage.setItem('desk.chat.hidden', String(value));
    } catch {
      /* Storage may be disabled. */
    }
  },
);

// The attention item the keyboard and answer buttons act on: the focused
// one, else the first. Rows read the key, so moving focus renders two rows.
export const currentKeyAtom = atom((get) => {
  const keys = get(attention.ids);
  const focused = get(focusedAtom);
  return keys.find((key) => key.slice(key.indexOf(':') + 1) === focused) || keys[0] || '';
});
export const currentItemAtom = atom((get) => {
  const key = get(currentKeyAtom);
  return key ? get(attention.entity(key)) : undefined;
});

// Flights are rebuilt per run only when that run changes or the minute turns.
const minute = (t: number) => Math.floor(t / 60000) * 60000;
export const flightsAtom = atom((get) =>
  flightsOf(
    get(runs.ids).flatMap((id) => {
      const run = get(runs.entity(id));
      return run ? [run] : [];
    }),
    minute(get(nowAtom)),
  ),
);

// The live state of one session, by id. System snapshots are reconciled, so
// a row re-renders only when its own session changes.
const liveSessions = new Map<string, ReturnType<typeof makeLiveSession>>();
const makeLiveSession = (id: string) => atom((get) => get(systemAtom)?.sessions?.find((s) => s.id === id));
export function liveSessionAtom(id: string) {
  let a = liveSessions.get(id);
  if (!a) liveSessions.set(id, (a = makeLiveSession(id)));
  return a;
}

export const busyAtom = atom(false);

// The note sent with an answer to the focused waiting item.
export const noteAtom = atom('');

// The chat list in the controller's order, as entities.
export const chatListAtom = atom((get) =>
  get(chats.ids).flatMap((id) => {
    const chat = get(chats.entity(id));
    return chat ? [chat] : [];
  }),
);

// Writes in flight, by "pool:<name>" or "session:<id>", so a control shows
// pending and cannot be submitted twice (JOB-178).
export const pendingAtom = atom<Record<string, boolean>>({});

// The admin methods this Desk allows (GET api/rpc), or null until read. A
// control for a method that is not listed stays hidden: desk.rpc is the grant
// Desk can see, and the controller still checks its own.
export const rpcMethodsAtom = atom<string[] | null>(null);

// The segment the timeline chart last chose (JOB-185): the detail pane opens
// its events tab on that entry. n counts choices, so choosing the same bar
// again scrolls to it again.
export const segmentChoiceAtom = atom<{ id: string; index: number; n: number } | null>(null);

// The comment or history tick open in the lower panel of the TL page (#1183):
// the row and the marker's key (see markerKey in lib/timeline.ts). It is
// here, not in the timeline, because the chart presses the tick and the panel
// under it shows the text.
export const markAtom = atom<{ id: string; key: string } | null>(null);
