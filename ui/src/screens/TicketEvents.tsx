import { useEffect, useMemo, useRef, useState } from 'react';
import type {
  GraphResult,
  HistoryEntry,
  HolderInfo,
  Ticket,
  TimelineRow,
  TimelineSegment,
  TimelineWait,
} from '../../../packages/jobshop-client/rpc.js';
import { eventItems, historyPartial, type EventItem } from '../lib/detail.ts';
import { durationLabel, holderState, isWorkflowSpan, segmentEnd, segmentText, turnTotals } from '../lib/timeline.ts';
import { exactLabel, fullLabel, linkLabel } from '../lib/trace.ts';
import { TicketHistory } from './TicketHistory.tsx';

const web = /^https?:\/\//;

// The conditions of a wait, each a link to what it waits on.
export function WaitList({ waits, onTrace }: { waits: TimelineWait[]; onTrace: (ref: string) => void }) {
  if (!waits.length) return null;
  return (
    <ul className="wait-list">
      {waits.map((w, j) => (
        <li key={j}>
          {w.target ? (
            <button className="mono" onClick={() => onTrace(w.target!)}>
              {w.label}
            </button>
          ) : w.url && web.test(w.url) ? (
            <a href={w.url} target="_blank" rel="noopener noreferrer">
              {w.label || linkLabel(w.url)}
            </a>
          ) : (
            w.label
          )}
          {w.state ? <span className="muted"> · {w.state}</span> : null}
        </li>
      ))}
    </ul>
  );
}

// The turns of one claim with exact times, for keyboard and touch.
export function TurnList({ seg, now }: { seg: TimelineSegment; now: number }) {
  const totals = turnTotals(seg, now);
  return (
    <details className="tl-turns">
      <summary>
        {totals?.count} turn{totals?.count === 1 ? '' : 's'}
      </summary>
      <ol>
        {seg.turns?.map((t, i) => (
          <li key={i}>
            <time dateTime={t.start} title={exactLabel(t.start)}>
              {fullLabel(t.start)}
            </time>
            {' · '}
            {durationLabel((t.end ? Date.parse(t.end) : segmentEnd(seg, now)) - Date.parse(t.start))}
            {t.end ? '' : ' so far'}
            {t.stop ? ' · ' + t.stop : ''}
          </li>
        ))}
      </ol>
    </details>
  );
}

// The events tab (JOB-185): one entry per span of the ticket's life and its
// close, newest first. Comments and other events are markers on the timeline
// (#1182). A toggle lists every history row instead, the raw rows no marker
// covers. A segment chosen on the timeline chart is highlighted and scrolled
// to.
export function TicketEvents({
  ticket,
  row,
  history,
  graph,
  sessions,
  now,
  chosen,
  onWatch,
  onTrace,
}: {
  ticket: Ticket;
  row?: TimelineRow;
  history: HistoryEntry[];
  graph: GraphResult;
  sessions?: Record<string, HolderInfo>;
  now: number;
  chosen: { index: number; n: number } | null;
  onWatch: (id: string) => void;
  onTrace: (ref: string) => void;
}) {
  // Without a timeline row (a session, a message, a cut-off trace) the raw
  // rows are all there is.
  const [raw, setRaw] = useState(false);
  const showRaw = raw || !row;
  const items = useMemo(() => (row ? eventItems(row, history) : []), [row, history]);
  const list = useRef<HTMLOListElement>(null);
  useEffect(() => {
    if (!chosen || showRaw) return;
    list.current?.querySelector('[data-chosen]')?.scrollIntoView({ block: 'nearest' });
  }, [chosen, showRaw]);

  return (
    <div className="ticket-events">
      <div className="events-bar">
        {row ? (
          <label>
            <input type="checkbox" checked={raw} onChange={(e) => setRaw(e.target.checked)} /> every history row
          </label>
        ) : (
          <span className="muted">This ticket is not on the timeline, so its history rows are listed.</span>
        )}
        {(row?.partial || historyPartial(history)) && (
          <span className="muted">
            Part of this history is not kept (an ephemeral or compacted ticket): earlier events are missing.
          </span>
        )}
      </div>
      {showRaw ? (
        history.length ? (
          <TicketHistory entries={history} ticket={ticket} graph={graph} onWatch={onWatch} />
        ) : (
          <p className="muted">No history rows.</p>
        )
      ) : (
        <ol className="events" ref={list}>
          {items.map((item) => (
            <Entry
              key={item.key}
              item={item}
              chosen={chosen !== null && item.type === 'segment' && item.index === chosen.index}
              sessions={sessions}
              now={now}
              onWatch={onWatch}
              onTrace={onTrace}
            />
          ))}
        </ol>
      )}
    </div>
  );
}

function Entry({
  item,
  chosen,
  sessions,
  now,
  onWatch,
  onTrace,
}: {
  item: EventItem;
  chosen: boolean;
  sessions?: Record<string, HolderInfo>;
  now: number;
  onWatch: (id: string) => void;
  onTrace: (ref: string) => void;
}) {
  switch (item.type) {
    case 'segment': {
      const s = item.seg;
      const session = s.actor ? sessions?.[s.actor] : undefined;
      const open = !s.end;
      const totals = turnTotals(s, now);
      return (
        <li className={chosen ? 'chosen' : undefined} data-chosen={chosen || undefined} data-kind="segment">
          <div className="event-head">
            <span
              className="tl-swatch"
              data-state={s.state}
              data-workflow={isWorkflowSpan(s) || undefined}
              aria-hidden="true"
            />
            <span data-live={holderState(s, session)}>{segmentText(s, { now, session })}</span>
          </div>
          <p className="muted event-when">
            <time dateTime={s.start} title={exactLabel(s.start)}>
              {fullLabel(s.start)}
            </time>
            {' · '}
            {durationLabel(segmentEnd(s, now) - Date.parse(s.start))}
            {open ? ' so far' : ''}
            {item.ended ? ' · ended: ' + item.ended : ''}
          </p>
          {s.actor && (
            <p className="event-links">
              <button title={s.actor} onClick={() => onWatch(s.actor!)}>
                view session ↗
              </button>
            </p>
          )}
          {totals && <TurnList seg={s} now={now} />}
          {s.wait && <WaitList waits={s.wait} onTrace={onTrace} />}
        </li>
      );
    }
    case 'close':
      return (
        <li data-kind="close">
          <div className="event-head">
            <span className="tl-end-sample" data-resolution={item.resolution || 'closed'} aria-hidden="true" />
            <b>closed{item.resolution ? ' ' + item.resolution : ''}</b>
            {item.note ? ' · ' + item.note : ''}
          </div>
          <p className="muted event-when">
            <time dateTime={new Date(item.at).toISOString()} title={exactLabel(item.at)}>
              {fullLabel(item.at)}
            </time>
            {item.by ? ' · by ' + item.by : ''}
          </p>
        </li>
      );
  }
}
