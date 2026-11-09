import type { AttentionItem } from '../../../packages/jobshop-client/rpc.js';
import { atom } from 'jotai';
import { store } from './store.ts';

// The ticket context menu (JOB-136). One menu serves the whole page. A row
// opens it with a snapshot of the ticket it was opened on, so an action acts
// on that ticket and never on whatever is selected by the time it runs.

export interface TicketMenuTarget {
  // The exact ticket id: what "copy ticket id" copies and the footer shows.
  id: string;
  // What "attach to chat" attaches (an issue key, a repo#n); the id when empty.
  ref?: string;
  // The session "view session" opens: the ticket's holder, or the ticket
  // itself when it is a session.
  session?: string;
  // The trace page inspects tickets in its graph in place ("inspect").
  inspectable?: boolean;
  // Set on waiting items, which can also be discussed in chat.
  item?: AttentionItem;
  // Set on a Linear issue or GitHub PR or issue (JOB-163). It is not a
  // Jobshop ticket, so the menu offers only what applies to it: attach to
  // chat, copy the reference and open it at the provider.
  provider?: { name: string; url: string };
}

// Anything with a rectangle to sit next to: an element or the pointer.
export interface MenuAnchor {
  getBoundingClientRect(): DOMRect;
}

export interface TicketMenuState {
  target: TicketMenuTarget;
  anchor: MenuAnchor;
  // Focus goes back here when the menu closes without navigating.
  returnTo: HTMLElement | null;
  // A closed menu stays in the atom so the popup can finish closing and
  // return focus; the next open replaces it.
  open: boolean;
}

export const ticketMenuAtom = atom<TicketMenuState | null>(null);

export function openTicketMenu(state: Omit<TicketMenuState, 'open'>) {
  store.set(ticketMenuAtom, { ...state, open: true });
}

export function closeTicketMenu() {
  const state = store.get(ticketMenuAtom);
  if (state?.open) store.set(ticketMenuAtom, { ...state, open: false });
}

// A short confirmation after an action whose result is otherwise invisible
// (a copy). Failures go to the error banner or stay in the menu.
export const noticeAtom = atom('');
