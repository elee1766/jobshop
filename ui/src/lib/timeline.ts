import type {
  HolderInfo,
  TimelineArrow,
  TimelineMarker,
  TimelineResult,
  TimelineRow,
  TimelineSegment,
} from '../../../packages/jobshop-client/rpc.js';
import { relativeLabel } from './trace.ts';

// The timeline's time span: the first creation to the latest close, or now
// while anything is open. A span is never shorter than a minute, so a lone
// new ticket still has a scale.
export function timelineSpan(tl: TimelineResult, now: number) {
  let start = Infinity;
  let last = 0;
  let active = false;
  for (const row of tl.rows) {
    start = Math.min(start, Date.parse(row.created_at));
    last = Math.max(last, Date.parse(row.created_at));
    if (row.closed_at) last = Math.max(last, Date.parse(row.closed_at));
    else active = true;
  }
  if (!tl.rows.length) return { start: now, end: now + 60_000, active: false };
  const end = active ? Math.max(now, last) : last;
  return { start, end: Math.max(end, start + 60_000), active };
}

// A segment's end: its own, or now while it is still going.
export function segmentEnd(seg: TimelineSegment, now: number) {
  return seg.end ? Date.parse(seg.end) : now;
}

// A row's bar end: its close, or now.
export function rowEnd(row: TimelineRow, now: number) {
  return row.closed_at ? Date.parse(row.closed_at) : now;
}

// Where a row's work begins, for arrows into it: its first claim, else its
// creation.
export function rowStart(row: TimelineRow) {
  return Date.parse(row.started || row.created_at);
}

// The rows left after collapsing: a collapsed row hides everything nested
// under it. Rows come in tree order, so a row is hidden while its depth is
// deeper than a collapsed ancestor's.
export function visibleRows(rows: TimelineRow[], collapsed: ReadonlySet<string>) {
  const out: TimelineRow[] = [];
  let hideBelow = Infinity;
  for (const row of rows) {
    if (row.depth > hideBelow) continue;
    hideBelow = Infinity;
    out.push(row);
    if (collapsed.has(row.id)) hideBelow = row.depth;
  }
  return out;
}

// Rows that have rows nested under them.
export function parentIds(rows: TimelineRow[]) {
  const ids = new Set<string>();
  for (const row of rows) if (row.parent) ids.add(row.parent);
  return ids;
}

// "1h 30m": a duration, at most two units.
export function durationLabel(ms: number) {
  return relativeLabel(Math.max(0, ms)).slice(1);
}

// What a state is called in details and accessible names.
export const stateLabels: Record<string, string> = {
  queued: 'queued',
  working: 'working',
  held: 'held, idle',
  lost: 'host not alive',
  waiting: 'waiting on an event',
  human: 'waiting on a human',
  deferred: 'deferred',
  retry: 'failed attempt',
  unknown: 'no history kept',
};

// A working bar splits into time in a turn and time held with none
// (JOB-184). The swatch state and its words.
export const legendLabels: Record<string, string> = { ...stateLabels, working: 'in a turn' };

// One line saying what a segment is, without its times.
export function segmentText(seg: TimelineSegment, opts: { now?: number; session?: HolderInfo } = {}) {
  const parts = [stateLabels[seg.state] || seg.state];
  if (seg.actor) parts.push(seg.actor);
  if (seg.detail) parts.push(seg.detail);
  if (seg.state === 'deferred' && seg.until) parts.push('until ' + new Date(seg.until).toLocaleString());
  if (seg.ended) parts.push('ended ' + seg.ended);
  const totals = turnTotals(seg, opts.now ?? Date.now());
  if (totals) parts.push(turnsText(totals));
  const live = liveText(holderState(seg, opts.session), opts.session);
  if (live) parts.push(live);
  return parts.join(' · ');
}

// A holder's state now, for the open working span only (JOB-184). Nothing
// is claimed about a holder the controller cannot describe (an external or
// finished session): 'unknown', and the bar is drawn as held.
export type HolderState = 'running' | 'idle' | 'lost' | 'unknown';

export function holderState(seg: TimelineSegment, session?: HolderInfo): HolderState | undefined {
  if (seg.state !== 'working' || seg.end || !seg.actor) return undefined;
  if (!session) return 'unknown';
  if (session.host_lost) return 'lost';
  if (session.busy) return 'running';
  return session.alive ? 'idle' : 'unknown';
}

// What the live state means in words, with the tool the session is running.
export function liveText(state: HolderState | undefined, session?: HolderInfo) {
  switch (state) {
    case 'running':
      return 'turn running' + (session?.tool ? ': ' + session.tool : '');
    case 'idle':
      return 'held, idle between turns';
    case 'lost':
      return 'host not alive, nobody is working this ticket';
    default:
      return '';
  }
}

export interface TurnTotals {
  count: number;
  active: number;
  held: number;
}

// Time in turns and time held with none, for one claim. Null when the claim
// has no turn rows: it is not idle, it is unknown.
export function turnTotals(seg: TimelineSegment, now: number): TurnTotals | null {
  if (!seg.turns?.length) return null;
  const end = segmentEnd(seg, now);
  const active = seg.turns.reduce(
    (sum, t) => sum + Math.max(0, (t.end ? Date.parse(t.end) : end) - Date.parse(t.start)),
    0,
  );
  return { count: seg.turns.length, active, held: Math.max(0, end - Date.parse(seg.start) - active) };
}

// "active 14m in 6 turns · held 1h 20m": held is the time with no turn.
export function turnsText(t: TurnTotals) {
  return `active ${durationLabel(t.active)} in ${t.count} turn${t.count === 1 ? '' : 's'} · held ${durationLabel(t.held)}`;
}

// What a marker kind is called in the marker panel and accessible names.
export const markerKinds: Record<string, string> = {
  comment: 'comment',
  event: 'event',
  pr: 'pr opened',
  merged: 'pr merged',
  session: 'taken over',
};

// A marker's accessible name: what it is, its one line and its time.
export function markerText(marker: TimelineMarker) {
  return [marker.kind, marker.label].filter(Boolean).join(': ');
}

// Ticks closer than gap pixels would cover each other and one could not be
// pressed. They are drawn as one tick at the first of them; the marker panel
// steps through the rest (#1182). markers are in time order, x maps a time to
// a pixel. Each group lists the indexes of its markers in the row.
export function clusterMarkers(markers: TimelineMarker[], x: (time: number) => number, gap = 10) {
  const groups: { at: number; indexes: number[] }[] = [];
  markers.forEach((marker, i) => {
    const at = x(Date.parse(marker.at));
    const last = groups[groups.length - 1];
    if (last && at - last.at < gap) last.indexes.push(i);
    else groups.push({ at, indexes: [i] });
  });
  return groups;
}

export type ArrowShape = 'met' | 'blocking' | 'conflict' | 'tracks' | 'wait';

// How an arrow is drawn: a blocking edge by its state, a wait by its own.
export function arrowShape(arrow: TimelineArrow): ArrowShape {
  if (arrow.kind === 'tracks') return 'tracks';
  if (arrow.kind === 'wait') return 'wait';
  return (arrow.state as ArrowShape) || 'blocking';
}

// An elbow from the prerequisite's end (x1, y1) to the dependent's start
// (x2, y2). When the dependent starts before the prerequisite ends, the
// path turns back through the gap between the two rows.
export function arrowPath(x1: number, y1: number, x2: number, y2: number, gap = 6) {
  if (x2 >= x1 + gap * 2) return `M ${x1} ${y1} H ${x1 + gap} V ${y2} H ${x2}`;
  const mid = (y1 + y2) / 2;
  return `M ${x1} ${y1} H ${x1 + gap} V ${mid} H ${x2 - gap} V ${y2} H ${x2}`;
}
