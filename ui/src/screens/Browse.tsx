import { useAtomValue } from 'jotai';
import { memo, useMemo, useRef, useState } from 'react';
import type { BrowseItem } from '../../../packages/jobshop-client/rpc.js';
import { TicketMenuButton, ticketMenuRef } from '../components/TicketMenu.tsx';
import {
  authorValue,
  checksText,
  groupItems,
  groupModes,
  listViews,
  providerName,
  statusLabel,
  statusLetter,
  statusTone,
  type Provider,
} from '../lib/browse.ts';
import { ago } from '../lib/model.ts';
import { currentURL, go, setParams, useDeskRoute } from '../lib/route.ts';
import { createSearcher } from '../lib/search.ts';
import { useShortcuts } from '../lib/shortcuts.ts';
import { useBrowseCatalog, useBrowseList, type ListParams } from '../state/browse.ts';
import { browseAtom, nowAtom } from '../state/live.ts';
import { StatusLine, WritesButton } from '../components/BrowseParts.tsx';
import { ProviderStatus } from '../components/StatusLetter.tsx';
import { Detail } from './BrowseDetail.tsx';

// The Linear and GitHub views (JOB-163): a list on the left, one item on the
// right, read-only. Both read through the controller (browse.*), which keeps
// the scope to the configured teams and repos and the provider limits. The
// view state is in the URL: ?view=linear&scope=JOB&list=active&item=JOB-150.

export function Linear() {
  return <Browse provider="linear" />;
}
export function GitHub() {
  return <Browse provider="github" />;
}

const kinds = ['pr', 'issue', 'all'];
const kindLabel: Record<string, string> = { pr: 'prs', issue: 'issues', all: 'prs and issues' };

// The team or repo an item reference belongs to.
export function scopeOf(provider: Provider, item: string): string {
  if (provider === 'github') return item.split('#')[0];
  const at = item.lastIndexOf('-');
  return at > 0 ? item.slice(0, at).toUpperCase() : '';
}

function Browse({ provider }: { provider: Provider }) {
  const route = useDeskRoute();
  const scope = useAtomValue(browseAtom);
  const scopes = scope
    ? provider === 'linear'
      ? scope.linear.map((s) => s.team)
      : scope.github.map((s) => s.repo)
    : null;
  if (!scopes) {
    return (
      <div className="browse-note">
        <output className="muted block">loading</output>
      </div>
    );
  }
  if (!scopes.length) return <NotConfigured provider={provider} />;
  // An item names its own scope, which a link from another repo or team sets.
  const named = route.item ? scopeOf(provider, route.item) : '';
  const current = scopes.includes(named) ? named : scopes.includes(route.scope) ? route.scope : scopes[0];
  return <Pane provider={provider} scopes={scopes} current={current} />;
}

function NotConfigured({ provider }: { provider: Provider }) {
  return (
    <div className="browse-note">
      <div className="tag">
        <span className="mono">/{provider}</span>
      </div>
      <div className="section-body">
        {provider === 'linear' ? (
          <p>
            No Linear team is configured for Desk. Desk reads the teams of the Linear identities in the service config (
            <code>identities</code> with <code>teams</code>), as the service identity.
          </p>
        ) : (
          <p>
            No GitHub repository is configured for Desk. Desk reads the repositories of the workflows and of{' '}
            <code>activity.repos</code> in the service config, as the service identity.
          </p>
        )}
      </div>
    </div>
  );
}

function Pane({ provider, scopes, current }: { provider: Provider; scopes: string[]; current: string }) {
  const route = useDeskRoute();
  const selected = route.item;
  return (
    <div className="browse" data-provider={provider} data-item={selected ? 'open' : 'none'}>
      <List provider={provider} scopes={scopes} current={current} />
      <section key={selected} className="browse-detail" aria-label={provider === 'linear' ? 'Issue' : 'Pull request'}>
        {selected ? (
          <Detail provider={provider} item={selected} />
        ) : (
          <>
            <div className="tag">
              <span className="mono">/{provider}</span>
            </div>
            <p className="section-body muted">
              {provider === 'linear' ? 'Select an issue to read it.' : 'Select a pull request or issue to read it.'}
            </p>
          </>
        )}
      </section>
    </div>
  );
}

const defaultGroup = (provider: Provider) => groupModes[provider][0];

function List({ provider, scopes, current }: { provider: Provider; scopes: string[]; current: string }) {
  const route = useDeskRoute();
  const view = listViews[provider].includes(route.list) ? route.list : 'open';
  const kind = kinds.includes(route.kind) ? route.kind : 'pr';
  const group = groupModes[provider].includes(route.group) ? route.group : defaultGroup(provider);
  const params = useMemo<ListParams>(
    () =>
      provider === 'linear'
        ? { provider, scope: current, view, project: route.project, owner: route.owner, query: route.q }
        : { provider, scope: current, view, kind, author: route.author, query: route.q },
    [provider, current, view, kind, route.project, route.owner, route.author, route.q],
  );
  const list = useBrowseList(params);
  const catalog = useBrowseCatalog(provider === 'linear' ? current : '');
  const now = useAtomValue(nowAtom);

  // The box filters what is loaded as you type, with the shared fuzzy search.
  // Enter asks the provider, which finds what is not loaded.
  const [text, setText] = useState(route.q);
  const input = useRef<HTMLInputElement>(null);
  // A change that comes from the URL (Back, a link) replaces the text.
  const [typedFrom, setTypedFrom] = useState(route.q);
  if (typedFrom !== route.q) {
    setTypedFrom(route.q);
    setText(route.q);
  }
  const searcher = useMemo(
    () =>
      createSearcher(list.items, {
        fields: [
          { name: 'title', weight: 3, get: (i: BrowseItem) => i.title },
          { name: 'ref', weight: 2, get: (i: BrowseItem) => i.ref },
          { name: 'owner', weight: 1, get: (i: BrowseItem) => i.owner },
          { name: 'project', weight: 1, get: (i: BrowseItem) => i.project },
          { name: 'labels', weight: 1, get: (i: BrowseItem) => i.labels.join(' ') },
        ],
        ids: (i: BrowseItem) => [i.ref, i.id],
        label: (i: BrowseItem) => i.title,
      }),
    [list.items],
  );
  const local = text.trim() !== '' && text !== route.q;
  const shown = useMemo(() => (local ? searcher.search(text) : list.items), [local, searcher, text, list.items]);
  const groups = useMemo(() => groupItems(shown, group), [shown, group]);

  const rows = useRef<HTMLElement>(null);
  const move = (step: number) => {
    const at = shown.findIndex((i) => i.ref === route.item);
    const next = shown[Math.min(shown.length - 1, Math.max(0, at + step))];
    if (next) openItem(next.ref);
  };
  useShortcuts([
    { hotkey: 'J', callback: () => move(1) },
    { hotkey: 'K', callback: () => move(-1) },
    { hotkey: '/', callback: () => input.current?.focus() },
    { hotkey: 'R', callback: () => list.refresh() },
  ]);

  const authors = useMemo(() => [...new Set(list.items.map(authorValue).filter(Boolean))].toSorted(), [list.items]);
  const noun = provider === 'linear' ? 'issues' : kindLabel[kind];
  const fetched = list.status?.fetched_at ? ago(list.status.fetched_at, now) : '';
  const stamp = list.status?.state === 'cached' || list.status?.state === 'fresh' ? fetched : '';

  return (
    <section className="browse-list" aria-label={`${providerName[provider]} ${noun}`} ref={rows}>
      <div className="tag">
        <span className="mono">/{provider}</span>
        {scopes.length > 1 ? (
          <select
            aria-label={provider === 'linear' ? 'Team' : 'Repository'}
            value={current}
            onChange={(e) =>
              setParams({ scope: e.target.value, item: '', tab: '', project: '', owner: '', author: '' })
            }
          >
            {scopes.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        ) : (
          <span className="mono">{current}</span>
        )}
        <select
          aria-label="List"
          value={view}
          onChange={(e) => setParams({ list: e.target.value === 'open' ? '' : e.target.value })}
        >
          {listViews[provider].map((v) => (
            <option key={v} value={v}>
              {v}
            </option>
          ))}
        </select>
        {provider === 'github' && (
          <select
            aria-label="Kind"
            value={kind}
            onChange={(e) => setParams({ kind: e.target.value === 'pr' ? '' : e.target.value })}
          >
            {kinds.map((k) => (
              <option key={k} value={k}>
                {kindLabel[k]}
              </option>
            ))}
          </select>
        )}
        <button className="push-right" onClick={() => list.refresh()} disabled={list.busy} title="Read again (R)">
          {list.busy ? 'reading' : 'refresh'}
        </button>
        <WritesButton label={provider === 'linear' ? 'new issue' : 'new pr'} />
      </div>
      <search className="browse-search">
        <form
          className="browse-search-form"
          onSubmit={(e) => {
            e.preventDefault();
            setParams({ q: text.trim(), item: route.item });
          }}
        >
          <input
            ref={input}
            aria-label={`Search ${noun}`}
            placeholder={`search ${noun}`}
            value={text}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Escape') {
                setText(route.q);
                input.current?.blur();
              }
            }}
          />
          <kbd className="muted" aria-hidden="true">
            /
          </kbd>
        </form>
      </search>
      <div className="browse-filters">
        {provider === 'linear' ? (
          <>
            <label>
              project:{' '}
              <select value={route.project} onChange={(e) => setParams({ project: e.target.value })}>
                <option value="">all</option>
                <option value="none">none</option>
                {(catalog?.projects ?? []).map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.name}
                  </option>
                ))}
              </select>
            </label>
            <label>
              owner:{' '}
              <select value={route.owner} onChange={(e) => setParams({ owner: e.target.value })}>
                <option value="">all</option>
                <option value="none">nobody</option>
                {(catalog?.users ?? []).map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.name}
                  </option>
                ))}
              </select>
            </label>
          </>
        ) : (
          <label>
            author:{' '}
            <select value={route.author} onChange={(e) => setParams({ author: e.target.value })}>
              <option value="">all</option>
              {route.author && !authors.includes(route.author) && <option value={route.author}>{route.author}</option>}
              {authors.map((a) => (
                <option key={a} value={a}>
                  {a}
                </option>
              ))}
            </select>
          </label>
        )}
        <label>
          group:{' '}
          <select
            value={group}
            onChange={(e) => setParams({ group: e.target.value === defaultGroup(provider) ? '' : e.target.value })}
          >
            {groupModes[provider].map((g) => (
              <option key={g} value={g}>
                {g}
              </option>
            ))}
          </select>
        </label>
        {route.q && (
          <button onClick={() => setParams({ q: '' })} title="Back to the whole list">
            search: {route.q} ×
          </button>
        )}
      </div>
      <StatusLine status={list.status} error={list.error} onRetry={list.refresh} />
      {list.loading ? (
        <output className="section-body muted block">loading</output>
      ) : (
        <>
          <Rows provider={provider} groups={groups} selected={route.item} now={now} />
          {!shown.length && !list.error && list.status?.state !== 'error' && (
            <p className="section-body muted">
              {local ? 'Nothing loaded matches. ' : `No ${noun} match. `}
              {local && 'Press Enter to search the provider.'}
              {!local &&
                (route.q || route.project || route.owner || route.author || view !== 'open') &&
                'Try another list or clear the search.'}
            </p>
          )}
          {local && (
            <p className="section-body muted browse-hint">
              {shown.length} of {list.items.length} loaded match. Enter searches {providerName[provider]}.
            </p>
          )}
        </>
      )}
      <div className="browse-foot muted">
        <span>
          {list.items.length} loaded
          {list.total > list.items.length ? ` of ${list.total}` : list.hasMore ? ' so far' : ''}
          {stamp && ` · read ${stamp} ago`}
        </span>
        {list.hasMore && (
          <button onClick={list.loadMore} disabled={list.busy}>
            {list.busy ? 'loading' : 'load more'}
          </button>
        )}
      </div>
    </section>
  );
}

const Rows = memo(function Rows({
  provider,
  groups,
  selected,
  now,
}: {
  provider: Provider;
  groups: ReturnType<typeof groupItems>;
  selected: string;
  now: number;
}) {
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});
  if (!groups.some((g) => g.items.length)) return null;
  return (
    <div className="table-scroll">
      <table className="mono browse-rows" aria-label={provider === 'linear' ? 'Issues' : 'Pull requests and issues'}>
        <thead>
          <tr>
            <th title="Status">s</th>
            <th>id</th>
            <th>{provider === 'linear' ? 'issue' : 'title'}</th>
            <th>{provider === 'linear' ? 'owner' : 'ci'}</th>
          </tr>
        </thead>
        {groups.map((g) => (
          <tbody key={g.key || 'all'}>
            {g.label && (
              <tr className="browse-group">
                <th colSpan={4}>
                  <button
                    aria-expanded={!collapsed[g.key]}
                    onClick={() => setCollapsed({ ...collapsed, [g.key]: !collapsed[g.key] })}
                  >
                    {collapsed[g.key] ? '▸' : '▾'} {g.label} <span className="muted">{g.items.length}</span>
                  </button>
                </th>
              </tr>
            )}
            {!collapsed[g.key] &&
              g.items.map((item) => <Row key={item.ref} item={item} on={item.ref === selected} now={now} />)}
          </tbody>
        ))}
      </table>
    </div>
  );
});

function openItem(ref: string) {
  const url = currentURL();
  url.searchParams.set('item', ref);
  url.searchParams.delete('tab');
  go(url);
}

const Row = memo(function Row({ item, on, now }: { item: BrowseItem; on: boolean; now: number }) {
  const target = useMemo(
    () => ({
      id: item.ref,
      ref: item.provider + ':' + item.ref,
      provider: { name: providerName[item.provider as Provider], url: item.url },
    }),
    [item],
  );
  const owner = item.provider === 'linear' ? item.owner : checksText(item.checks);
  return (
    <tr ref={ticketMenuRef(target)} className={on ? 'chosen' : undefined} aria-selected={on}>
      <td>
        <ProviderStatus tone={statusTone(item)} letter={statusLetter(item)} label={statusLabel(item)} />
      </td>
      <td className="browse-id muted">{item.id}</td>
      <td className="browse-title">
        <button
          className="browse-open"
          aria-current={on ? 'true' : undefined}
          title={`${item.ref} · ${item.title}${item.updated_at ? ' · updated ' + ago(item.updated_at, now) + ' ago' : ''}`}
          onClick={() => openItem(item.ref)}
        >
          {item.title}
        </button>
        {item.provider === 'github' && item.kind === 'issue' && <span className="muted"> issue</span>}
      </td>
      <td className={'browse-end' + (item.checks ? ' checks-' + item.checks : '')}>
        {owner || <span className="muted">-</span>}
        <TicketMenuButton target={target} />
      </td>
    </tr>
  );
});
