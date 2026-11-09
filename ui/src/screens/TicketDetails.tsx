import { useMemo } from 'react';
import type { GraphResult, ShowResult, Ticket } from '../../../packages/jobshop-client/rpc.js';
import { Markdown } from '../components/Markdown.tsx';
import { StatusLetter } from '../components/StatusLetter.tsx';
import { TitleAction } from '../components/TitleAction.tsx';
import { TitleText } from '../components/TitleText.tsx';
import { linkLabel, relationshipText, safeLinks, stateDetail } from '../lib/trace.ts';
import { TicketHistory } from './TicketHistory.tsx';

const priority = ['blocked by', 'blocks', 'waits for', 'waited on by', 'children'];
const rank = (kind: string) => (priority.indexOf(kind) < 0 ? 99 : priority.indexOf(kind));

interface Fields {
  wait?: { reason?: string };
  step?: string;
  state?: string;
  failures?: number;
}

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
  const node = graph.nodes.find((n) => n.id === ticket.id);
  const parent = graph.nodes.find((n) => n.id === ticket.parent_id);
  const groups = useMemo(() => {
    const out = new Map<string, Ticket[]>();
    for (const edge of graph.relations) {
      if (edge.from !== ticket.id && edge.to !== ticket.id) continue;
      const id = edge.from === ticket.id ? edge.to : edge.from;
      if (edge.kind === 'parent' && id === ticket.parent_id) continue;
      const other = graph.nodes.find((n) => n.id === id);
      if (!other) continue;
      const kind = edge.kind === 'parent' && edge.from === ticket.id ? 'children' : relationshipText(edge, ticket.id);
      const list = out.get(kind) || [];
      if (!list.some((n) => n.id === id)) list.push(other);
      out.set(kind, list);
    }
    return [...out].toSorted(([a], [b]) => rank(a) - rank(b));
  }, [graph, ticket.id, ticket.parent_id]);
  const fields = ticket.fields as Fields | undefined;
  const detailText = stateDetail(ticket);

  const link = (other: Ticket) => (
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
      <header className="ticket-overview">
        {ticket.parent_id && (
          <div className="ticket-parent">
            <span className="muted">child of</span>
            {parent ? (
              link(parent)
            ) : (
              <button className="mono" onClick={() => onSelect(ticket.parent_id!)}>
                {ticket.parent_id}
              </button>
            )}
          </div>
        )}
        <h2>
          <TitleText title={ticket.title} fallback={ticket.id} />
          {ticket.title && <small className="mono muted">{ticket.id}</small>}
        </h2>
        <p>
          <StatusLetter ticket={ticket} />{' '}
          {detailText && (
            <>
              <b>
                <Markdown inline text={detailText} />
              </b>{' '}
              ·{' '}
            </>
          )}
          {ticket.kind}
          {ticket.pool ? ' · ' + ticket.pool : ''}
        </p>
        <p className="muted">
          created {new Date(ticket.created_at).toLocaleString()} · updated{' '}
          {new Date(ticket.updated_at).toLocaleString()}
        </p>
        <div className="ticket-links">
          {safeLinks(node).map((href) => (
            <a key={href} href={href} target="_blank" rel="noopener noreferrer">
              {linkLabel(href)}
            </a>
          ))}
          <button onClick={() => onTrace(ticket.id)}>open as graph root</button>
          {ticket.holder && (
            <button className="watch-session" title={ticket.holder} onClick={() => onWatch(ticket.holder!)}>
              view running session ↗
            </button>
          )}
          {ticket.kind === 'session' && (
            <button className="watch-session" onClick={() => onWatch(ticket.id)}>
              view session ↗
            </button>
          )}
        </div>
        {!!ticket.labels?.length && <p className="muted">labels: {ticket.labels.join(', ')}</p>}
        {fields?.wait?.reason && (
          <p>
            <b>waiting:</b> <Markdown inline text={fields.wait.reason} />
          </p>
        )}
        {fields?.step && (
          <p>
            step: {fields.step}
            {fields.failures ? ' · ' + fields.failures + ' failures' : ''}
          </p>
        )}
      </header>
      {groups.length > 0 && (
        <div className="relationship-groups" aria-label="Ticket relationships">
          {groups.map(([kind, others]) => (
            <section className="relationship-group" aria-label={kind} key={kind}>
              <h3>
                {kind} <span className="muted">{others.length}</span>
              </h3>
              <ul className="rows">
                {others.map((other) => (
                  <li key={other.id}>{link(other)}</li>
                ))}
              </ul>
            </section>
          ))}
        </div>
      )}
      {!!detail.history?.length && (
        <TicketHistory entries={detail.history} ticket={ticket} graph={graph} onWatch={onWatch} />
      )}
      {ticket.body && (
        <details className="ticket-body" open>
          <summary>body</summary>
          <Markdown text={ticket.body} />
        </details>
      )}
      {!!detail.comments?.length && (
        <>
          <h3>comments</h3>
          {detail.comments.map((comment, i) => (
            <article className="ticket-comment" key={i}>
              <p className="muted">
                {comment.author} · {new Date(comment.created_at).toLocaleString()}
              </p>
              <Markdown text={comment.body} />
            </article>
          ))}
        </>
      )}
      <details>
        <summary>raw ticket data</summary>
        <pre>{JSON.stringify({ fields: ticket.fields, meta: ticket.meta }, null, 2)}</pre>
      </details>
    </div>
  );
}
