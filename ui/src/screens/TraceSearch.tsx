import { useEffect, useRef, useState, type FormEvent } from 'react';
import type { Ticket } from '../../../packages/jobshop-client/rpc.js';
import { TitleAction } from '../components/TitleAction.tsx';
import { ticketMenuRef, TicketMenuButton } from '../components/TicketMenu.tsx';
import { client } from '../lib/api.ts';
import { state as ticketState, statusLetter } from '../lib/trace.ts';

const pageSize = 50;
type Status = 'open' | 'in_progress' | 'closed' | '';

// The server search is authoritative: this lists what it returned, a page at
// a time, and never filters a partial list as if it were the whole dataset.
// The parent keys it by `reference`, so a new reference (a link, Back) starts
// the field from that reference.
export function TraceSearch({ reference, onOpen }: { reference: string; onOpen: (id: string) => void }) {
  const [query, setQuery] = useState(reference);
  const [status, setStatus] = useState<Status>('');
  const [results, setResults] = useState<Ticket[]>([]);
  const [searched, setSearched] = useState<string | null>(null);
  const [more, setMore] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const request = useRef<AbortController | undefined>(undefined);
  const last = useRef<{ q: string; status: Status }>({ q: '', status: '' });
  const shown = useRef<Ticket[]>([]);

  useEffect(() => () => request.current?.abort(), []);

  function clear() {
    request.current?.abort();
    request.current = undefined;
    setBusy(false);
    setSearched(null);
    setResults([]);
    shown.current = [];
    setError('');
    setMore(false);
  }

  async function search(append = false) {
    request.current?.abort();
    const controller = new AbortController();
    request.current = controller;
    if (!append) {
      last.current = { q: query.trim(), status };
      setSearched(last.current.q);
      setResults([]);
      shown.current = [];
      setMore(false);
    }
    setBusy(true);
    setError('');
    try {
      const tickets = await client.call(
        'ticket.list',
        {
          query: last.current.q,
          status: last.current.status,
          order: 'newest',
          limit: pageSize + 1,
          after: append ? shown.current.at(-1)?.id : undefined,
        },
        { signal: controller.signal },
      );
      if (request.current !== controller) return;
      setMore(tickets.length > pageSize);
      shown.current = append ? [...shown.current, ...tickets.slice(0, pageSize)] : tickets.slice(0, pageSize);
      setResults(shown.current);
    } catch (e) {
      if (request.current === controller && !controller.signal.aborted) setError(String(e));
    } finally {
      if (request.current === controller) setBusy(false);
    }
  }

  const submit = (event: FormEvent) => {
    event.preventDefault();
    void search();
  };

  return (
    <>
      <div className="tag trace-search-bar">
        <span className="mono">/tl</span>
        <search aria-label="Search tickets">
          <form onSubmit={submit}>
            <input
              aria-label="TL reference"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              maxLength={512}
              placeholder="title, ticket, session, repo#PR"
            />
            <select aria-label="Ticket status" value={status} onChange={(e) => setStatus(e.target.value as Status)}>
              <option value="">all</option>
              <option value="open">open</option>
              <option value="in_progress">running</option>
              <option value="closed">closed</option>
            </select>
            <button type="submit">search</button>
            <button
              type="button"
              disabled={!query.trim()}
              title="Open an exact ticket, session or external reference"
              onClick={() => onOpen(query.trim())}
            >
              open
            </button>
          </form>
        </search>
      </div>
      {searched !== null ? (
        <section className="trace-search-results" aria-label="Ticket search results" aria-busy={busy}>
          <div className="tag">
            <output>
              {busy ? 'searching…' : `${results.length}${more ? '+' : ''} matches`} · {searched || 'recent tickets'}
            </output>
            <button className="push-right" onClick={clear}>
              dismiss
            </button>
          </div>
          {error ? (
            <p className="error" role="alert">
              {error}
            </p>
          ) : (
            !busy &&
            !results.length && (
              <p className="muted">no matching tickets. try a title, ticket id, or related reference.</p>
            )
          )}
          {results.length > 0 && (
            <div className="table-scroll">
              <table className="mono rows">
                <thead>
                  <tr>
                    <th title="Status: O open, R running, C closed">s</th>
                    <th>ticket</th>
                    <th>kind</th>
                    <th className="wide">pool</th>
                  </tr>
                </thead>
                <tbody>
                  {results.map((ticket) => (
                    <tr
                      key={ticket.id}
                      ref={ticketMenuRef({
                        id: ticket.id,
                        session: ticket.holder || (ticket.kind === 'session' ? ticket.id : ''),
                      })}
                    >
                      <td
                        className="ticket-status"
                        data-status={ticket.status}
                        title={ticketState(ticket)}
                        aria-label={ticketState(ticket)}
                      >
                        {statusLetter(ticket)}
                      </td>
                      <td>
                        <TitleAction
                          className="ticket-label"
                          title={ticket.title}
                          fallback={ticket.id}
                          tip={ticket.id}
                          href={'?view=trace&ref=' + encodeURIComponent(ticket.id)}
                        />
                        <TicketMenuButton
                          target={{
                            id: ticket.id,
                            session: ticket.holder || (ticket.kind === 'session' ? ticket.id : ''),
                          }}
                        />
                      </td>
                      <td>{ticket.kind}</td>
                      <td className="wide">{ticket.pool || '–'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
          {more && (
            <button disabled={busy} onClick={() => void search(true)}>
              more
            </button>
          )}
        </section>
      ) : (
        !reference && (
          <p className="muted trace-search-hint">
            search all tickets, including completed work, or open an exact reference.
          </p>
        )
      )}
    </>
  );
}
