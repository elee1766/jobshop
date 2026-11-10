import type { GraphResult, HistoryEntry, Ticket } from '../../../packages/jobshop-client/rpc.js';
import { Markdown } from '../components/Markdown.tsx';
import { foldTurns } from '../lib/history.ts';
import { durationLabel } from '../lib/timeline.ts';
import { stepName } from '../lib/trace.ts';

const titles: Record<string, string> = {
  create: 'created',
  claim: 'claimed',
  release: 'released',
  close: 'closed',
  reopen: 'reopened',
  update: 'updated',
  edge_add: 'relationship added',
  edge_remove: 'relationship removed',
  label_add: 'label added',
  label_remove: 'label removed',
  comment: 'commented',
  expire: 'claim expired',
};

const sessionID = /^[a-zA-Z0-9][a-zA-Z0-9._-]*-s-[a-zA-Z0-9]+$/;

export function TicketHistory({
  entries: all,
  ticket,
  graph,
  onWatch,
}: {
  entries: HistoryEntry[];
  ticket: Ticket;
  graph: GraphResult;
  onWatch: (id: string) => void;
}) {
  // Turn rows would add two lines per turn: they fold into their claim.
  const folded = foldTurns(all);
  const entries = folded.entries;
  const label = (id: string) => graph.nodes.find((node) => node.id === id)?.title || id;
  const isSession = (id: string) =>
    graph.nodes.some((node) => node.id === id && node.kind === 'session') || sessionID.test(id);
  function describe(entry: HistoryEntry) {
    const diff = (entry.diff || {}) as Record<string, any>;
    const fields = diff.fields || {};
    const turns = folded.claims.get(entry.seq);
    const title =
      entry.op === 'claim' && diff.resumed ? 'resumed work' : titles[entry.op] || entry.op.replaceAll('_', ' ');
    const step = stepName(fields) || stepName(ticket.fields);
    const summary = [
      diff.reason,
      diff.note,
      diff.resolution,
      diff.title,
      fields.wait?.reason,
      fields.state,
      diff.kind,
      diff.label,
    ]
      .filter((value) => typeof value === 'string' && value)
      .join(' · ');
    const refs = [
      ...new Set(
        [
          diff.holder,
          diff.parent_id,
          diff.to,
          diff.from,
          diff.session,
          diff.session_id,
          diff.ticket_id,
          diff.request_changes,
        ].filter((value) => typeof value === 'string' && value !== ticket.id),
      ),
    ] as string[];
    return { title, step, summary, refs, turns };
  }
  return (
    <div className="ticket-history">
      <ol className="rows">
        {entries.toReversed().map((entry) => {
          const event = describe(entry);
          return (
            <li key={entry.seq}>
              <div className="history-heading">
                <b>{event.title}</b>
                {event.step && <span className="history-step">step: {event.step}</span>}
                <time className="muted" dateTime={entry.at}>
                  {new Date(entry.at).toLocaleString()}
                </time>
              </div>
              {event.summary && <Markdown className="history-summary markdown" text={event.summary} inline />}
              {event.turns && (
                <p className="muted history-turns">
                  {event.turns.count} turn{event.turns.count === 1 ? '' : 's'}, {durationLabel(event.turns.active)} in
                  turns
                </p>
              )}
              {event.refs.length > 0 && (
                <div className="history-references">
                  {event.refs.map((id) =>
                    isSession(id) ? (
                      <button key={id} title={id} onClick={() => onWatch(id)}>
                        session: {label(id)} ↗
                      </button>
                    ) : (
                      <a key={id} title={id} href={'?view=trace&ref=' + encodeURIComponent(id)}>
                        {label(id)} <small className="mono">{id}</small>
                      </a>
                    ),
                  )}
                </div>
              )}
              <div className="history-footer">
                <span className="muted">by {entry.actor}</span>
                {isSession(entry.actor) && (
                  <button title={entry.actor} onClick={() => onWatch(entry.actor)}>
                    view session
                  </button>
                )}
                {entry.diff != null && Object.keys(entry.diff as object).length > 0 && (
                  <details className="history-raw">
                    <summary>details</summary>
                    <pre>{JSON.stringify(entry.diff, null, 2)}</pre>
                  </details>
                )}
              </div>
            </li>
          );
        })}
      </ol>
    </div>
  );
}
