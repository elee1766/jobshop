import { useAtomValue } from 'jotai';
import { useEffect, useMemo, useState, type KeyboardEvent, type ReactNode } from 'react';
import type {
  GraphNode,
  GraphResult,
  HolderInfo,
  ShowResult,
  Ticket,
  TimelineRow,
} from '../../../packages/jobshop-client/rpc.js';
import { Markdown } from '../components/Markdown.tsx';
import { StatusLetter } from '../components/StatusLetter.tsx';
import { ticketMenuRef, TicketMenuButton } from '../components/TicketMenu.tsx';
import { TitleAction } from '../components/TitleAction.tsx';
import { TitleText } from '../components/TitleText.tsx';
import {
  childProgress,
  detailTab,
  hasIO,
  ioOf,
  nodeDuration,
  nowOf,
  relationCount,
  relationGroups,
  summarize,
  type DetailTab,
  type KV,
} from '../lib/detail.ts';
import { setParams, useDeskRoute } from '../lib/route.ts';
import { durationLabel, holderState, liveText } from '../lib/timeline.ts';
import { exactLabel, fullLabel, linkLabel, safeLinks, stepName } from '../lib/trace.ts';
import { traceAtom } from '../state/live.ts';
import { segmentChoiceAtom } from '../state/ui.ts';
import { TicketEvents, WaitList } from './TicketEvents.tsx';

const pickTab = (id: DetailTab) => setParams({ tab: id === 'events' ? '' : id });

// What the ticket menu acts on: any ticket in this graph can be inspected.
const menuTarget = (t: { id: string; holder?: string; kind?: string }) => ({
  id: t.id,
  session: t.holder || (t.kind === 'session' ? t.id : ''),
  inspectable: true,
});

// A clock for "so far" durations: it moves while a ticket is open.
function useNow(open: boolean) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!open) return;
    const clock = setInterval(() => setNow(Date.now()), 15_000);
    return () => clearInterval(clock);
  }, [open]);
  return now;
}

// The trace's detail pane (JOB-185): what happened to the selected ticket and
// what it is doing now first, the detail behind tabs. Both layouts use it.
export function TicketDetails({
  detail,
  graph,
  onSelect,
  onTrace,
  onWatch,
}: {
  detail: ShowResult;
  graph: GraphResult;
  onSelect: (id: string) => void;
  onTrace: (id: string) => void;
  onWatch: (id: string) => void;
}) {
  const ticket = detail.ticket;
  const trace = useAtomValue(traceAtom);
  const route = useDeskRoute();
  const choice = useAtomValue(segmentChoiceAtom);
  const node = graph.nodes.find((n) => n.id === ticket.id);
  // A wait target in this graph is selected, not opened again (JOB-190).
  const target = (id: string) => (graph.nodes.some((n) => n.id === id) ? onSelect(id) : onTrace(id));
  const parent = graph.nodes.find((n) => n.id === ticket.parent_id);
  const timeline = trace?.timeline;
  const row = useMemo(() => timeline?.rows.find((r) => r.id === ticket.id), [timeline, ticket.id]);
  const open = ticket.status !== 'closed';
  const now = useNow(open);
  const history = useMemo(() => detail.history || [], [detail.history]);
  const comments = useMemo(() => detail.comments || [], [detail.comments]);
  const groups = useMemo(() => relationGroups(graph, ticket), [graph, ticket]);
  const io = useMemo(() => ioOf(ticket), [ticket]);
  const sessions = trace?.sessions;
  const holder: HolderInfo | undefined = ticket.holder ? sessions?.[ticket.holder] : undefined;
  const kids = useMemo(() => childProgress(timeline?.rows, ticket.id), [timeline, ticket.id]);
  const sum = useMemo(() => summarize(ticket, row, now), [ticket, row, now]);
  const current = nowOf(ticket, row, kids, holder, history);
  const step = stepName(ticket.fields);

  // Tabs that have nothing to show are left out, not shown empty.
  const tabs: [DetailTab, string][] = [['events', 'events']];
  if (hasIO(ticket)) tabs.push(['io', 'input / result']);
  const related = relationCount(groups);
  tabs.push(['relations', 'relations' + (related ? ' ' + related : '')]);
  if (ticket.body) tabs.push(['body', 'body']);
  if (comments.length) tabs.push(['comments', 'comments ' + comments.length]);
  tabs.push(['raw', 'raw']);
  const wanted = detailTab(route.tab);
  const tab = tabs.some(([id]) => id === wanted) ? wanted : 'events';

  // A segment chosen on the chart (it also resets the tab to events).
  const chosen = choice && choice.id === ticket.id ? choice : null;

  function keys(event: KeyboardEvent<HTMLButtonElement>) {
    const at = tabs.findIndex(([id]) => id === tab);
    const to =
      event.key === 'ArrowRight'
        ? at + 1
        : event.key === 'ArrowLeft'
          ? at - 1
          : event.key === 'Home'
            ? 0
            : event.key === 'End'
              ? tabs.length - 1
              : -1;
    if (to < 0) return;
    event.preventDefault();
    const next = tabs[(to + tabs.length) % tabs.length][0];
    pickTab(next);
    requestAnimationFrame(() => document.getElementById('detail-tab-' + next)?.focus());
  }

  const link = (other: Pick<GraphNode, 'id' | 'title' | 'status' | 'resolution' | 'fields'>) => (
    <>
      <StatusLetter ticket={other} />
      <TitleAction
        className="relation-ticket ticket-label"
        title={other.title}
        fallback={other.id}
        tip={other.id}
        onActivate={() => onSelect(other.id)}
      >
        {other.title && <small className="mono muted">{other.id}</small>}
      </TitleAction>
    </>
  );

  return (
    <div className="inspector" aria-label="Ticket details">
      <header className="ticket-overview" ref={ticketMenuRef(menuTarget(ticket))}>
        {ticket.parent_id && (
          <div
            className="ticket-parent"
            ref={ticketMenuRef(parent ? menuTarget(parent) : menuTarget({ id: ticket.parent_id }))}
          >
            <span className="muted">child of</span>
            {parent ? (
              <>
                {link(parent)}
                <TicketMenuButton target={menuTarget(parent)} />
              </>
            ) : (
              <button className="mono" onClick={() => onSelect(ticket.parent_id!)}>
                {ticket.parent_id}
              </button>
            )}
          </div>
        )}
        <h2>
          <StatusLetter ticket={ticket} /> <TitleText title={ticket.title} fallback={ticket.id} />
          {ticket.title && <small className="mono muted">{ticket.id}</small>}
        </h2>
        <p className="summary-kind">
          {ticket.kind}
          {ticket.pool ? ' · ' + ticket.pool : ''}
          {step ? ' · step ' + step : ''}
          {!!ticket.labels?.length && <span className="muted"> · {ticket.labels.join(', ')}</span>}
        </p>
      </header>

      <dl className="facts" aria-label="Summary">
        <Fact name="started">
          <time dateTime={ticket.created_at} title={exactLabel(ticket.created_at)}>
            {fullLabel(ticket.created_at)}
          </time>
        </Fact>
        <Fact name={sum.ended ? 'ended' : 'to'}>
          {sum.ended ? (
            <time dateTime={ticket.closed_at} title={exactLabel(sum.ended)}>
              {fullLabel(sum.ended)}
            </time>
          ) : (
            'now'
          )}
        </Fact>
        <Fact name={sum.ended ? 'took' : 'open for'}>{durationLabel(sum.took)}</Fact>
        {sum.claims > 0 && (
          <Fact name="attempts">
            {sum.claims}
            {sum.failed > 0 || sum.limit > 0 ? (
              <>
                {' · '}
                <span className={sum.failed > 0 ? 'warn' : undefined}>failed {sum.failed}</span>
                {sum.limit > 0 ? ' of ' + sum.limit + ' allowed' : ''}
              </>
            ) : null}
          </Fact>
        )}
        {kids.total > 0 && (
          <Fact name={ticket.kind === 'run' ? 'steps' : 'children'}>
            {kids.closed} of {kids.total} closed
          </Fact>
        )}
        {ticket.status === 'closed' && (
          <Fact name="outcome">{[ticket.resolution || 'closed', runState(ticket)].filter(Boolean).join(' · ')}</Fact>
        )}
        {ticket.holder && (
          <Fact name="holder">
            <button className="watch-session mono" title={ticket.holder} onClick={() => onWatch(ticket.holder!)}>
              {ticket.holder}
            </button>
            {liveSummary(row, ticket.holder, holder)}
          </Fact>
        )}
        {safeLinks(node).map((href) => (
          <Fact key={href} name={linkKind(href)}>
            <a href={href} target="_blank" rel="noopener noreferrer">
              {linkLabel(href)}
            </a>
          </Fact>
        ))}
      </dl>
      {sum.shares.length > 0 && sum.took > 0 && (
        <div className="time-by-state" aria-label="Time by state">
          <div className="time-bar" aria-hidden="true">
            {sum.shares.map((s) => (
              <span key={s.key} className="tl-swatch" data-state={s.swatch} style={{ flexGrow: s.ms }} />
            ))}
          </div>
          <ul className="muted">
            {sum.shares.map((s) => (
              <li key={s.key}>
                <span className="tl-swatch" data-state={s.swatch} aria-hidden="true" /> {s.label} {durationLabel(s.ms)}
              </li>
            ))}
          </ul>
        </div>
      )}
      {sum.bare ? (
        <p className="muted summary-note">
          This ticket is not on the timeline, so only its own fields are summarized. Time by state and attempts are left
          out.
        </p>
      ) : (
        sum.partial && (
          <p className="muted summary-note">
            Part of this ticket's history is not kept, so its time by state and attempts cover only what is.
          </p>
        )
      )}

      <section className="now-block" data-kind={current.kind} aria-label={open ? 'Now' : 'How it ended'}>
        <span className="now-label">{open ? 'now' : 'ended'}</span>
        <div>
          <p>
            <b data-live={current.kind === 'lost' ? 'lost' : undefined}>{current.text}</b>
            {current.since !== undefined && open && (
              <span className="muted">
                {' '}
                · for{' '}
                <time dateTime={new Date(current.since).toISOString()} title={exactLabel(current.since)}>
                  {durationLabel(now - current.since)}
                </time>
              </span>
            )}
            {current.until !== undefined && (
              <span>
                {' '}
                until{' '}
                <time dateTime={new Date(current.until).toISOString()} title={exactLabel(current.until)}>
                  {new Date(current.until).toLocaleString()}
                </time>
              </span>
            )}
          </p>
          {current.waits.length > 0 && <WaitList waits={current.waits} onTrace={target} />}
          {current.kind === 'human' && ticket.body && (
            <p>
              <button onClick={() => pickTab('body')}>read the question</button>
            </p>
          )}
        </div>
      </section>

      <div className="ticket-links">
        <TicketMenuButton target={menuTarget(ticket)} />
        {ticket.kind === 'session' && (
          <button className="watch-session" onClick={() => onWatch(ticket.id)}>
            view session ↗
          </button>
        )}
      </div>

      <div className="tag detail-tabs" role="tablist" aria-label="Ticket detail">
        {tabs.map(([id, label]) => (
          <button
            key={id}
            id={'detail-tab-' + id}
            role="tab"
            aria-selected={tab === id}
            aria-controls="detail-panel"
            className={tab === id ? 'chosen' : undefined}
            onClick={() => pickTab(id)}
            onKeyDown={keys}
          >
            {label}
          </button>
        ))}
      </div>
      <div id="detail-panel" role="tabpanel" aria-labelledby={'detail-tab-' + tab} className="detail-panel">
        {tab === 'events' && (
          <TicketEvents
            ticket={ticket}
            row={row}
            history={history}
            comments={comments}
            graph={graph}
            sessions={sessions}
            now={now}
            chosen={chosen}
            onWatch={onWatch}
            onTrace={target}
          />
        )}
        {tab === 'io' && <InputResult io={io} />}
        {tab === 'relations' && (
          <div className="relationship-groups" aria-label="Ticket relationships">
            {groups.length === 0 && <p className="muted">No related tickets in this trace.</p>}
            {groups.map(({ kind, nodes }) => (
              <section className="relationship-group" aria-label={kind} key={kind}>
                <h3>
                  {kind} <span className="muted">{nodes.length}</span>
                </h3>
                <ul className="rows">
                  {nodes.map((other) => (
                    <li key={other.id} ref={ticketMenuRef(menuTarget(other))}>
                      {link(other)}
                      <span className="muted relation-took">{nodeDuration(other, now)}</span>
                      <TicketMenuButton target={menuTarget(other)} />
                    </li>
                  ))}
                </ul>
              </section>
            ))}
          </div>
        )}
        {tab === 'body' && <Markdown text={ticket.body || ''} />}
        {tab === 'comments' &&
          comments.map((comment) => (
            <article className="ticket-comment" key={comment.id}>
              <p className="muted">
                {comment.author} ·{' '}
                <time dateTime={comment.created_at}>{new Date(comment.created_at).toLocaleString()}</time>
              </p>
              <Markdown text={comment.body} />
            </article>
          ))}
        {tab === 'raw' && (
          <>
            <p className="muted">
              fields and meta of {ticket.id}, revision {ticket.revision}. created{' '}
              {new Date(ticket.created_at).toLocaleString()}, updated {new Date(ticket.updated_at).toLocaleString()}.
            </p>
            <pre className="raw-json">{JSON.stringify({ fields: ticket.fields, meta: ticket.meta }, null, 2)}</pre>
          </>
        )}
      </div>
    </div>
  );
}

function Fact({ name, children }: { name: string; children: ReactNode }) {
  return (
    <div>
      <dt className="muted">{name}</dt>
      <dd>{children}</dd>
    </div>
  );
}

function linkKind(href: string) {
  const host = new URL(href).hostname;
  if (host === 'github.com') return href.includes('/pull/') ? 'pr' : 'issue';
  if (host === 'linear.app') return 'linear';
  return 'link';
}

// A run's or batch's own state when it is not just closed.
function runState(ticket: Ticket) {
  const state = (ticket.fields as { state?: string } | undefined)?.state;
  return state && state !== ticket.status && state !== ticket.resolution ? state : '';
}

// The live state of the session holding a ticket, from its open working bar.
function liveSummary(row: TimelineRow | undefined, id: string, session?: HolderInfo) {
  const seg = row?.segments.findLast((s) => !s.end && s.state === 'working' && s.actor === id);
  const words = seg ? liveText(holderState(seg, session), session) : '';
  return words ? <span className="muted"> · {words}</span> : null;
}

function InputResult({ io }: { io: { inputs: KV[]; result: KV[] } }) {
  return (
    <div className="io">
      <Pairs title="input" rows={io.inputs} />
      <Pairs title="result" rows={io.result} />
    </div>
  );
}

function Pairs({ title, rows }: { title: string; rows: KV[] }) {
  return (
    <section aria-label={title}>
      <h3>{title}</h3>
      {rows.length === 0 ? (
        <p className="muted">{title === 'result' ? 'No result yet.' : 'No input recorded.'}</p>
      ) : (
        <dl className="pairs">
          {rows.map((row) => (
            <div key={row.key}>
              <dt className="mono muted">{row.key}</dt>
              <dd>
                {row.long ? (
                  <details>
                    <summary>{preview(row.value)}</summary>
                    <pre>{row.value}</pre>
                  </details>
                ) : (
                  row.value
                )}
              </dd>
            </div>
          ))}
        </dl>
      )}
    </section>
  );
}

function preview(text: string) {
  const line = text.split('\n')[0];
  return line.length > 90 ? line.slice(0, 89) + '…' : line + (text.includes('\n') ? ' …' : '');
}
