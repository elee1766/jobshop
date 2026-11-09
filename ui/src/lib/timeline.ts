import type {
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

// What a state is called in the legend, details and accessible names.
export const stateLabels: Record<string, string> = {
  queued: 'queued',
  working: 'working',
  waiting: 'waiting on an event',
  human: 'waiting on a human',
  deferred: 'deferred',
  retry: 'failed attempt',
  unknown: 'no history kept',
};

// One line saying what a segment is, without its times.
export function segmentText(seg: TimelineSegment) {
  const parts = [stateLabels[seg.state] || seg.state];
  if (seg.actor) parts.push(seg.actor);
  if (seg.detail) parts.push(seg.detail);
  if (seg.state === 'deferred' && seg.until) parts.push('until ' + new Date(seg.until).toLocaleString());
  if (seg.ended) parts.push('ended ' + seg.ended);
  return parts.join(' · ');
}

export function markerText(marker: TimelineMarker) {
  return [marker.kind, marker.label].filter(Boolean).join(': ');
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
