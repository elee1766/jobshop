import {
  browseKeys,
  browseViews,
  currentRoute,
  currentURL,
  go,
  setParams,
  type Section,
  type View,
} from '../lib/route.ts';
import { client } from '../lib/api.ts';
import { errorAtom, selectedAtom, traceAtom } from './live.ts';
import { store } from './store.ts';
import {
  busyAtom,
  chatHiddenAtom,
  chatOpenAtom,
  currentItemAtom,
  noteAtom,
  phoneTabAtom,
  watchedSessionAtom,
} from './ui.ts';
import { chatBridge } from './chatBridge.ts';

// Actions are plain functions over the store and the URL. They replace the
// closures that lived in App.svelte. Each write goes through the controller
// API; nothing here keeps its own copy of server data.

const setError = (message: unknown) => store.set(errorAtom, String(message));

// The trace the page last showed. A change of ref or inspected ticket from
// outside the actions (Back, Forward, a typed URL) clears the old graph
// before the new one arrives. The actions clear what they need themselves and
// record their key with markTrace, so they do not clear what they just set.
let shownTrace = '';
const traceKey = () => {
  const route = currentRoute();
  return route.ref + ':' + route.inspect;
};
export function markTrace() {
  shownTrace = traceKey();
}
export function syncTrace(ref: string, inspectRef: string) {
  const key = ref + ':' + inspectRef;
  if (key === shownTrace) return;
  shownTrace = key;
  store.set(traceAtom, null);
  store.set(selectedAtom, null);
  store.set(watchedSessionAtom, '');
}

export function showChat() {
  store.set(chatHiddenAtom, false);
  store.set(chatOpenAtom, true);
  store.set(phoneTabAtom, 'chats');
}

export function hideChat() {
  store.set(chatHiddenAtom, true);
  store.set(chatOpenAtom, false);
}

// leaveBrowse drops the query values the Linear and GitHub views own when the
// URL is going to another view or coming from one, so a group or filter of one
// provider never leaks into the other, and an issue key is not read as a
// waiting item on home.
function leaveBrowse(url: URL, to: View) {
  const from = currentRoute().view;
  if (from === to || !(browseViews.includes(from as never) || browseViews.includes(to as never))) return;
  for (const key of [...browseKeys, 'item']) url.searchParams.delete(key);
}

// navigate pushes a view. The other references stay in the URL, so a page
// and back keeps the in-flight filter, the trace and the audit.
export function navigate(view: View, section: Section = currentRoute().section, params: Record<string, string> = {}) {
  if (view !== 'home') store.set(phoneTabAtom, 'more');
  const url = currentURL();
  url.searchParams.set('view', view);
  // An embed id belongs to the embed view only.
  if (view !== 'embed') url.searchParams.delete('embed');
  // The Linear and GitHub views own their query values (JOB-163). Moving to
  // another view drops them, so a group or filter of one provider never leaks
  // into the other, and an issue key is not read as a waiting item on home.
  if (view !== currentRoute().view) leaveBrowse(url, view);
  for (const [key, value] of Object.entries(params)) {
    if (value) url.searchParams.set(key, value);
    else url.searchParams.delete(key);
  }
  url.hash = view === 'home' && section ? '#' + section : '';
  go(url);
}

// Catch-up shows what changed since the last visit; the toggle is the URL's
// since=1 and always lands on the top of home.
export function toggleSince() {
  store.set(phoneTabAtom, 'waiting');
  store.set(chatOpenAtom, false);
  navigate('home', '', { since: currentRoute().since ? '' : '1' });
  showSection('');
}

// Scroll to a home section; with no id, to the top of home. Scrolls .work
// itself: scrollIntoView could also move overflow:hidden ancestors.
export function showSection(id: string) {
  requestAnimationFrame(() => {
    const work = document.querySelector<HTMLElement>('.work');
    if (!work) return;
    const target = id ? document.getElementById(id) : null;
    work.scrollTop = target
      ? work.scrollTop + target.getBoundingClientRect().top - work.getBoundingClientRect().top
      : 0;
  });
}

export function goSection(id: Section) {
  store.set(phoneTabAtom, id || 'waiting');
  store.set(chatOpenAtom, false);
  navigate('home', id);
  showSection(id);
}

export function openTrace(ref: string) {
  const route = currentRoute();
  if (ref !== route.ref || route.inspect) {
    store.set(selectedAtom, null);
    store.set(traceAtom, null);
    store.set(watchedSessionAtom, '');
  }
  store.set(chatOpenAtom, false);
  const url = currentURL();
  leaveBrowse(url, 'trace');
  url.searchParams.set('ref', ref);
  url.searchParams.delete('inspect');
  url.searchParams.set('view', 'trace');
  url.hash = '';
  store.set(phoneTabAtom, 'more');
  go(url);
  markTrace();
}

export function inspect(id: string) {
  const trace = store.get(traceAtom);
  const node = trace?.graph?.nodes.find((n) => n.id === id);
  store.set(watchedSessionAtom, node?.holder || (node?.kind === 'session' ? id : ''));
  if (currentRoute().inspect === id) return;
  store.set(selectedAtom, null);
  const url = currentURL();
  leaveBrowse(url, 'trace');
  url.searchParams.set('inspect', id);
  url.searchParams.set('view', 'trace');
  url.hash = '';
  store.set(phoneTabAtom, 'more');
  go(url);
  markTrace();
}

// setLayout switches the trace between the timeline ('', the default, so no
// query value) and the graph. Both keep the ref and the inspected ticket.
export function setLayout(layout: '' | 'graph') {
  setParams({ layout });
}

export function openAudit(id: string) {
  const url = currentURL();
  leaveBrowse(url, 'system');
  if (id) url.searchParams.set('audit', id);
  else url.searchParams.delete('audit');
  url.searchParams.set('view', 'system');
  url.hash = '';
  store.set(phoneTabAtom, 'more');
  go(url);
}

export function setFilter(value: string) {
  setParams({ filter: value });
}

// setChatParam records the open chat in the URL without a history step.
export function setChatParam(id: string) {
  const url = currentURL();
  if (id) url.searchParams.set('chat', id);
  else url.searchParams.delete('chat');
  go(url, true);
}

export async function discuss(item = store.get(currentItemAtom)) {
  if (!item) return;
  try {
    const result = await client.post('api/attention/chat', { kind: item.kind, id: item.id });
    openChat({ id: result.session, key: result.key, title: item.ref || item.id });
    showChat();
  } catch (e) {
    setError(e);
  }
}

// openChat asks the chat panel to open a chat the controller just made.
function openChat(chat: { id: string; key: string; title: string }) {
  chatBridge.open?.({ ...chat, state: 'idle', alive: true, updated: '' });
}

export async function answer(option: number): Promise<boolean> {
  const item = store.get(currentItemAtom);
  if (!item || store.get(busyAtom)) return false;
  store.set(busyAtom, true);
  try {
    const result = await client.post('api/attention/answer', {
      kind: item.kind,
      id: item.id,
      option,
      expected_title: item.options?.find((choice) => choice.id === option)?.title,
      note: store.get(noteAtom),
    });
    store.set(noteAtom, '');
    if (result.session) {
      openChat({ id: result.session, key: result.key, title: item.ref || item.id });
      showChat();
    }
    return true;
  } catch (e) {
    setError(e);
    return false;
  } finally {
    store.set(busyAtom, false);
  }
}

export function triage(ids: string[]) {
  showChat();
  chatBridge.start?.(
    'Summarize these waiting items, recommend an order, and identify decisions I need to make. Do not change work until I choose an action.',
    ids,
    'triage ' + new Date().toISOString().slice(0, 19),
  );
}

// discussProvider starts a named chat with a Linear issue or GitHub PR or issue
// attached (JOB-163). The reference is provider:ref, which Desk resolves
// through browse.show when the message is sent.
export function discussProvider(provider: string, ref: string) {
  showChat();
  chatBridge.start?.('', [provider + ':' + ref], ref);
}

export function attach(refs: string[]) {
  showChat();
  chatBridge.attach?.(refs);
}

export async function closeSession(id: string) {
  try {
    const dry = await client.call('session.close', { id, dry_run: true });
    if (
      !window.confirm(
        `Close ${id}? Its turn stops and its slot is freed. ${dry.held?.length || 0} held tickets will be released. The transcript remains.`,
      )
    )
      return;
    await client.call('session.close', { id, reason: 'closed in Desk' });
  } catch (e) {
    setError(e);
  }
}
