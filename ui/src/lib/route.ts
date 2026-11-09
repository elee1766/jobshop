// Desk's URL contract (JOB-150): the view and its references live in the
// query string, the home section in the hash. Desk serves one page at any
// proxy prefix, so there are no path routes. Wouter's browser location
// primitives provide the subscription (history, back and forward) and
// navigation. They read the raw search string: Wouter's useSearch decodes
// it first, which would turn an encoded "%41" in a filter into "A".
import { useEffect, useRef, useState } from 'react';
import { navigate, useLocationProperty } from 'wouter/use-browser-location';

export type View = 'home' | 'trace' | 'improvements' | 'system' | 'embed';
// The built-in navigation. 'embed' is a view too, but its tabs come from the
// service config (JOB-151), not from this list: ?view=embed&embed=<id>.
export const views: Exclude<View, 'embed'>[] = ['home', 'trace', 'improvements', 'system'];
const routable: View[] = [...views, 'embed'];
export const viewLabel: Record<Exclude<View, 'embed'>, string> = {
  home: 'home',
  trace: 'trace',
  improvements: 'learn',
  system: 'sys',
};
export type Section = '' | 'waiting' | 'inflight';

export interface Route {
  view: View;
  ref: string;
  inspect: string;
  audit: string;
  chat: string;
  embed: string;
  item: string;
  filter: string;
  since: boolean;
  section: Section;
}

export function parseRoute(search: string, hash: string): Route {
  const params = new URLSearchParams(search);
  const raw = params.get('view') || '';
  const view = (routable as string[]).includes(raw) ? (raw as View) : 'home';
  const h = hash.replace(/^#/, '');
  return {
    view,
    ref: params.get('ref') || '',
    inspect: params.get('inspect') || '',
    audit: params.get('audit') || '',
    chat: params.get('chat') || '',
    embed: view === 'embed' ? params.get('embed') || '' : '',
    item: params.get('item') || '',
    filter: params.get('filter') || '',
    since: params.get('since') === '1',
    section: view === 'home' && (h === 'waiting' || h === 'inflight') ? h : '',
  };
}

const currentSearch = () => location.search;
const currentHash = () => location.hash;

export function useDeskRoute(): Route {
  const search = useLocationProperty(currentSearch);
  const hash = useLocationProperty(currentHash);
  return parseRoute(search, hash);
}

export function currentRoute(): Route {
  return parseRoute(location.search, location.hash);
}

// Typed text (a filter) reaches the URL after a short pause, so typing does
// not write history on every key (Safari throttles replaceState). Until then
// it is pending: currentURL() folds it in, so any navigation in between
// carries it, and go() drops it because the new URL already has it.
let pending: Record<string, string> = {};
let timer: ReturnType<typeof setTimeout> | undefined;

function apply(url: URL, values: Record<string, string>) {
  for (const [key, value] of Object.entries(values)) {
    if (value) url.searchParams.set(key, value);
    else url.searchParams.delete(key);
  }
}

// currentURL is the URL to build the next one from.
export function currentURL(): URL {
  const url = new URL(location.href);
  apply(url, pending);
  return url;
}

// Writes a URL without leaving the page; Wouter's hooks see it.
export function go(url: URL, replace = false) {
  pending = {};
  clearTimeout(timer);
  const to = url.pathname + url.search + url.hash;
  if (to === location.pathname + location.search + location.hash) return;
  navigate(to, { replace });
}

// setParams merges query values into the current URL and replaces the
// history entry. The view is always written so the URL stays explicit.
export function setParams(values: Record<string, string>) {
  const url = currentURL();
  if (!url.searchParams.get('view')) url.searchParams.set('view', currentRoute().view);
  apply(url, values);
  go(url, true);
}

// queueParams is setParams after a pause, for typed text.
export function queueParams(values: Record<string, string>, delay = 250) {
  pending = { ...pending, ...values };
  clearTimeout(timer);
  timer = setTimeout(flushParams, delay);
}

export function flushParams() {
  if (!Object.keys(pending).length) return;
  const values = pending;
  setParams(values);
}

// useUrlText is a text field backed by a query parameter: the field shows
// what was typed at once, the URL follows after a pause, and a change that
// comes from the URL (Back, a link) replaces the text.
export function useUrlText(key: string, fromURL: string): [string, (value: string) => void] {
  const [text, setText] = useState(fromURL);
  const last = useRef(fromURL);
  useEffect(() => {
    if (fromURL !== last.current) {
      last.current = fromURL;
      setText(fromURL);
    }
  }, [fromURL]);
  useEffect(() => () => flushParams(), []);
  return [
    text,
    (value) => {
      last.current = value;
      setText(value);
      queueParams({ [key]: value });
    },
  ];
}
