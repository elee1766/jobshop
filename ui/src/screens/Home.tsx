import { atom, useAtom, useAtomValue, useSetAtom } from 'jotai';
import { memo, useMemo } from 'react';
import type { AttentionItem } from '../../../packages/jobshop-client/rpc.js';
import { Markdown } from '../components/Markdown.tsx';
import { TitleAction } from '../components/TitleAction.tsx';
import { parseTitle } from '../../kit/titles.js';
import { ago, type Flight } from '../lib/model.ts';
import { useDeskRoute, useUrlText } from '../lib/route.ts';
import { useShortcuts } from '../lib/shortcuts.ts';
import { state as ticketState, statusLetter } from '../lib/trace.ts';
import { answer, attach, discuss, openTrace, toggleSince, triage } from '../state/actions.ts';
import { attention, lastFocusAtom, nowAtom, truncatedAtom } from '../state/live.ts';
import {
  busyAtom,
  collapsedAtom,
  currentItemAtom,
  currentKeyAtom,
  flightsAtom,
  focusedAtom,
  hideHoldsAtom,
  noteAtom,
  sortOldestAtom,
  watchedSessionAtom,
} from '../state/ui.ts';

const minuteOf = (t: number) => Math.floor(t / 60000) * 60000;

export function Home() {
  const route = useDeskRoute();
  return (
    <>
      {route.since && <SinceBar />}
      <Waiting />
      <NotMoving />
      <InFlight />
      <HomeShortcuts />
    </>
  );
}

// Home's keyboard shortcuts exist only while Home is mounted: a route change
// unregisters them, and the library skips text entry and IME composition.
function HomeShortcuts() {
  const ids = useAtomValue(attention.ids);
  const current = useAtomValue(currentItemAtom);
  const [, setFocused] = useAtom(focusedAtom);
  const move = (delta: number) => {
    const index = ids.findIndex((key) => key === (current ? current.kind + ':' + current.id : ''));
    const next = ids[Math.max(0, Math.min(ids.length - 1, index + delta))] || '';
    setFocused(next.slice(next.indexOf(':') + 1));
  };
  useShortcuts([
    { hotkey: 'J', callback: () => move(1) },
    { hotkey: 'K', callback: () => move(-1) },
    { hotkey: 'D', callback: () => void discuss() },
    ...(['1', '2', '3', '4', '5'] as const).map((key) => ({
      hotkey: key,
      callback: () => {
        if (current?.options?.some((o) => o.id === Number(key))) void answer(Number(key));
      },
    })),
  ]);
  return null;
}

function SinceBar() {
  const lastFocus = useAtomValue(lastFocusAtom);
  return (
    <div className="tag">
      Since {lastFocus && !lastFocus.startsWith('0001') ? new Date(lastFocus).toLocaleString() : 'first visit'}
      <button onClick={toggleSince}>clear</button>
    </div>
  );
}

function Collapse({ name, label }: { name: string; label: string }) {
  const [collapsed, set] = useAtom(collapsedAtom);
  return (
    <button aria-label={'Toggle ' + label} onClick={() => set({ ...collapsed, [name]: !collapsed[name] })}>
      {collapsed[name] ? '▸' : '▾'}
    </button>
  );
}

function Waiting() {
  const keys = useAtomValue(attention.ids);
  const collapsed = useAtomValue(collapsedAtom).waiting;
  return (
    <section className="waiting" id="waiting" aria-label="Waiting">
      <div className="tag">
        <Collapse name="waiting" label="waiting" />
        <span className="mono">/waiting</span>
        <button onClick={() => document.getElementById('answer-note')?.focus()}>answer</button>
        <button onClick={() => void discuss()}>discuss</button>
        <button className="wide" onClick={() => triage(keys.map((key) => key.slice(key.indexOf(':') + 1)))}>
          triage-in-chat
        </button>
        <span className="push-right muted">{keys.length}</span>
      </div>
      {!collapsed && (
        <div className="section-body rows">
          {!keys.length && <p className="muted">Nothing waiting on you.</p>}
          {keys.map((key) => (
            <AttentionRow key={key} itemKey={key} />
          ))}
        </div>
      )}
    </section>
  );
}

const AttentionRow = memo(function AttentionRow({ itemKey }: { itemKey: string }) {
  const item = useAtomValue(useMemo(() => attention.entity(itemKey), [itemKey]));
  const focused = useAtomValue(useMemo(() => atom((get) => get(currentKeyAtom) === itemKey), [itemKey]));
  const setFocused = useSetAtom(focusedAtom);
  const now = minuteOf(useAtomValue(nowAtom));
  if (!item) return null;
  return (
    <article className={focused ? 'focused' : undefined}>
      <Reference item={item} />
      {item.kind !== 'pool' && (
        <a href={'?view=trace&ref=' + encodeURIComponent(item.id)} aria-label={'Trace ' + (item.ref || item.id)}>
          trace
        </a>
      )}
      <TitleAction className="item-title" title={item.title} onActivate={() => setFocused(item.id)}>
        {' '}
        <span className="muted">{ago(item.since, now)}</span>
      </TitleAction>
      {focused ? (
        <Focused item={item} />
      ) : (
        <span className="compact-options muted">{(item.options || []).map((o) => o.title).join(' · ')}</span>
      )}
    </article>
  );
});

function Reference({ item }: { item: AttentionItem }) {
  return item.url && /^https?:\/\//i.test(item.url) ? (
    <a href={item.url} target="_blank" rel="noopener noreferrer">
      <b>{item.ref || item.id}</b>
    </a>
  ) : (
    <b>{item.ref || item.id}</b>
  );
}

function Focused({ item }: { item: AttentionItem }) {
  const busy = useAtomValue(busyAtom);
  const [note, setNote] = useAtom(noteAtom);
  return (
    <>
      <Markdown className="attention-context markdown" text={item.context || item.summary || ''} />
      <div className="options">
        {(item.options || []).map((option) => (
          <button
            key={option.id}
            className={option.recommended ? 'recommended' : undefined}
            disabled={busy}
            onClick={() => void answer(option.id)}
          >
            {option.id} {option.title}
            {option.recommended ? ' (recommended)' : ''}
            {option.description ? ': ' + option.description : ''}
          </button>
        ))}
      </div>
      <details className="answer-note">
        <summary>add answer note</summary>
        <label htmlFor="answer-note">note sent with your choice</label>
        <textarea id="answer-note" value={note} onChange={(e) => setNote(e.target.value)} rows={2} />
      </details>
    </>
  );
}

function NotMoving() {
  const flights = useAtomValue(flightsAtom);
  const [hideHolds, setHideHolds] = useAtom(hideHoldsAtom);
  const collapsed = useAtomValue(collapsedAtom)['not-moving'];
  const now = minuteOf(useAtomValue(nowAtom));
  const stalled = useMemo(() => flights.filter((r) => r.stalled && (!hideHolds || !r.hold)), [flights, hideHolds]);
  return (
    <section className="not-moving" aria-label="Not moving">
      <div className="tag">
        <Collapse name="not-moving" label="not moving" />
        <span className="mono">/not-moving</span>
        <button onClick={() => attach(stalled.map((r) => r.ref))}>diagnose</button>
        <button onClick={() => setHideHolds(!hideHolds)}>{hideHolds ? 'show-holds' : 'hide-holds'}</button>
        <span className="push-right muted">{stalled.length}</span>
      </div>
      {!collapsed && (
        <div className="section-body rows">
          {stalled.map((row) => (
            <p key={row.id} className={row.hold ? 'muted' : undefined}>
              <button className="mono" onClick={() => openTrace(row.id)}>
                {row.ref}
              </button>{' '}
              {ago(row.since, now)} <Markdown inline text={row.diagnosis} />
              {row.hold ? ' (external hold?)' : ''}
            </p>
          ))}
          {!stalled.length && <p className="muted">No stalled work. Threshold: 3h since last push or run activity.</p>}
        </div>
      )}
    </section>
  );
}

function InFlight() {
  const flights = useAtomValue(flightsAtom);
  const route = useDeskRoute();
  const [filter, setFilter] = useUrlText('filter', route.filter);
  const lastFocus = useAtomValue(lastFocusAtom);
  const truncated = useAtomValue(truncatedAtom);
  const [sortOldest, setSortOldest] = useAtom(sortOldestAtom);
  const collapsed = useAtomValue(collapsedAtom).inflight;
  const now = minuteOf(useAtomValue(nowAtom));
  const filtered = useMemo(() => {
    const terms = filter.toLowerCase().split(/\s+/).filter(Boolean);
    const since = route.since && lastFocus ? Date.parse(lastFocus) : 0;
    return flights
      .filter(
        (r) =>
          terms.every((term) =>
            `${parseTitle(r.title).text} ${r.id} ${r.step} ${r.repo} ${r.ref} ${r.waiting} ${r.prs.map((p) => p.label).join(' ')}`
              .toLowerCase()
              .includes(term),
          ) &&
          (!route.since || !lastFocus || Date.parse(r.run.run.updated_at) > since),
      )
      .toSorted((a, b) => (sortOldest ? 1 : -1) * (Date.parse(a.since) - Date.parse(b.since)));
  }, [flights, filter, route.since, lastFocus, sortOldest]);
  const steps = useMemo(() => [...new Set(flights.map((r) => r.step))], [flights]);
  return (
    <section className="inflight" id="inflight" aria-label="In flight">
      <div className="tag">
        <Collapse name="inflight" label="inflight" />
        <span className="mono">/inflight</span>
        <button onClick={() => setSortOldest(!sortOldest)}>sort</button>
        <button
          onClick={() => {
            if (filtered[0]) openTrace(filtered[0].id);
          }}
        >
          trace
        </button>
        <input
          aria-label="Filter inflight"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder="step, repo or ref"
        />
        <button onClick={() => setFilter('')}>clear</button>
        <span className="push-right muted">{filtered.length}</span>
      </div>
      {!collapsed && (
        <div className="section-body">
          <div className="step-strip mono">
            {steps.map((step) => (
              <button
                key={step}
                className={filter === step ? 'chosen' : undefined}
                onClick={() => setFilter(filter === step ? '' : step)}
              >
                {step} {flights.filter((r) => r.step === step).length}
              </button>
            ))}
          </div>
          {truncated && <p className="warn">Showing the first 100 runs. Narrow the controller query for more.</p>}
          <div className="table-scroll refs">
            <table className="mono rows">
              <thead>
                <tr>
                  <th title="Status: O open, R running, C closed">s</th>
                  <th>ticket</th>
                  <th className="ref-col">issue</th>
                  <th className="ref-col">pr</th>
                  <th>step</th>
                  <th>waiting</th>
                  <th>
                    <button onClick={() => setSortOldest(!sortOldest)}>age</button>
                  </th>
                  <th className="wide">session</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((row) => (
                  <FlightRow key={row.id} row={row} now={now} />
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </section>
  );
}

const FlightRow = memo(function FlightRow({ row, now }: { row: Flight; now: number }) {
  const setWatched = useSetAtom(watchedSessionAtom);
  const ext = { target: '_blank', rel: 'noopener noreferrer' } as const;
  return (
    <tr className={row.stalled ? 'stalled' : undefined}>
      <td className="ticket-status" title={ticketState(row.run.run)} aria-label={ticketState(row.run.run)}>
        {statusLetter(row.run.run)}
      </td>
      <td>
        <TitleAction
          className="ticket-label"
          title={row.title}
          fallback={row.ref}
          tip={row.ref === row.id ? row.id : `${row.id} · ${row.ref}`}
          onActivate={(e) => (e.shiftKey ? attach([row.ref]) : openTrace(row.id))}
        />
        {row.issueURL && (
          <a className="ticket-issue ref-stack" href={row.issueURL} {...ext}>
            {row.ref}
          </a>
        )}
        {row.prs.map((pr) => (
          <a key={pr.url} className="ticket-issue ref-stack" href={pr.url} {...ext}>
            {pr.label}
          </a>
        ))}
      </td>
      <td className="ref-col">
        {row.issueURL ? (
          <a href={row.issueURL} {...ext}>
            {row.ref}
          </a>
        ) : (
          '–'
        )}
      </td>
      <td className="ref-col">
        {row.prs.length
          ? row.prs.map((pr, i) => (
              <span key={pr.url}>
                {i ? ', ' : ''}
                <a href={pr.url} {...ext}>
                  {pr.label}
                </a>
              </span>
            ))
          : '–'}
      </td>
      <td>{row.step}</td>
      <td>
        <Markdown inline text={row.waiting} />
      </td>
      <td title={row.source}>{ago(row.since, now)}</td>
      <td className="wide">
        <button
          disabled={!row.holder}
          onClick={() => {
            openTrace(row.id);
            setWatched(row.holder);
          }}
        >
          {row.state}
        </button>
      </td>
    </tr>
  );
});
