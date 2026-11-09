import { state, statusLetter, type Stateful } from '../lib/trace.ts';

// The leading O / R / C indicator, named by the ticket's state.
export function StatusLetter({ ticket, id }: { ticket: Stateful; id?: string }) {
  return (
    <span
      id={id}
      className="ticket-status mono"
      data-status={ticket.status}
      role="img"
      title={state(ticket)}
      aria-label={state(ticket)}
    >
      {statusLetter(ticket)}
    </span>
  );
}
