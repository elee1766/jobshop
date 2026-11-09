import type { Ticket } from '../../../packages/jobshop-client/rpc.js';
import { state, statusLetter } from '../lib/trace.ts';

// The leading O / R / C indicator, named by the ticket's state.
export function StatusLetter({ ticket, id }: { ticket: Ticket; id?: string }) {
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
