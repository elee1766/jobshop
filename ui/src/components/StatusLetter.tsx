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

// The same indicator for a Linear issue or a GitHub PR or issue (JOB-163): a
// letter, named in full for assistive technology and the tooltip. The tone
// reuses the ticket colors (open, in progress, closed).
export function ProviderStatus({ tone, letter, label }: { tone: string; letter: string; label: string }) {
  return (
    <span className="ticket-status mono" data-status={tone} role="img" title={label} aria-label={label}>
      {letter}
    </span>
  );
}

// A CI check's mark. Color is not the only carrier: each state has its own
// glyph and a name.
export function CheckMark({ state: result, mark, label }: { state: string; mark: string; label: string }) {
  return (
    <span role="img" aria-label={label} title={label} className={'check-' + result}>
      {mark}
    </span>
  );
}
