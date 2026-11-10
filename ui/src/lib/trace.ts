import type { GraphNode, GraphRelation, Ticket } from '../../../packages/jobshop-client/rpc.js';

export function stepName(fields: unknown): string {
  if (!fields || typeof fields !== 'object') return '';
  const step = (fields as { step?: unknown }).step;
  return typeof step === 'string' ? step : '';
}

// What the status letter and state text read of a ticket (a timeline row has these too).
export type Stateful = Pick<Ticket, 'status' | 'resolution' | 'fields'>;

export function state(ticket: Stateful) {
  const fields = ticket.fields as { state?: string } | undefined;
  return [fields?.state || ticket.status, ticket.resolution].filter(Boolean).join(' · ');
}

export function statusLetter(ticket: Stateful): string {
  return ({ open: 'O', in_progress: 'R', closed: 'C' } as Record<string, string>)[ticket.status] || '?';
}

export function relationLabel(kind: string) {
  return (
    (
      {
        blocks: 'blocked by',
        waits_for: 'waits for',
        tracks: 'tracks',
        parent: 'parent of',
        created: 'led to',
        wait: 'is parked on',
      } as Record<string, string>
    )[kind] || kind
  );
}

export function linkLabel(href: string) {
  const url = new URL(href);
  if (url.hostname === 'github.com') return url.pathname.slice(1).replace('/pull/', ' PR #').replace('/issues/', ' #');
  if (url.hostname === 'linear.app') {
    const key = url.pathname.split('/issue/')[1]?.split('/')[0] || 'Linear';
    try {
      return decodeURIComponent(key);
    } catch {
      return key;
    }
  }
  return url.hostname + url.pathname;
}

export function safeLinks(node?: GraphNode) {
  return (node?.links || []).filter((link) => {
    try {
      return ['http:', 'https:'].includes(new URL(link).protocol);
    } catch {
      return false;
    }
  });
}

const second = 1000,
  minute = 60 * second,
  hour = 60 * minute,
  day = 24 * hour;
const tickSteps = [
  second,
  5 * second,
  15 * second,
  30 * second,
  minute,
  2 * minute,
  5 * minute,
  10 * minute,
  15 * minute,
  30 * minute,
  hour,
  2 * hour,
  3 * hour,
  6 * hour,
  12 * hour,
  day,
  2 * day,
  7 * day,
  14 * day,
  30 * day,
  90 * day,
  365 * day,
];

// Relative axis ticks: start + k*step strictly between start and end, inside
// the visible window [from, to], with step the smallest that puts ticks at
// least gap pixels apart.
export function axisTicks(start: number, end: number, pxPerMs: number, from: number, to: number, gap = 88) {
  if (!(pxPerMs > 0) || end <= start) return [];
  const step = tickSteps.find((s) => s * pxPerMs >= gap) ?? tickSteps[tickSteps.length - 1];
  const ticks: number[] = [];
  for (let k = Math.max(1, Math.ceil((from - start) / step)); ticks.length < 200; k++) {
    const t = start + k * step;
    if (t >= end || t > to) break;
    ticks.push(t);
  }
  return ticks;
}

// "+1h 30m": an offset from the trace start, at most two units.
export function relativeLabel(ms: number) {
  const units: [number, string][] = [
    [day, 'd'],
    [hour, 'h'],
    [minute, 'm'],
    [second, 's'],
  ];
  const parts: string[] = [];
  let rest = Math.round(ms / second) * second;
  for (const [size, name] of units) {
    if (parts.length === 2) break;
    const n = Math.floor(rest / size);
    if (n || (parts.length && size !== second)) parts.push(n + name);
    rest -= n * size;
  }
  while (parts.length > 1 && parts[parts.length - 1].startsWith('0')) parts.pop();
  return '+' + (parts.join(' ') || '0s');
}

const fullTime = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' });
const exactTime = new Intl.DateTimeFormat(undefined, { dateStyle: 'full', timeStyle: 'long' });

// Endpoint labels: date and minute.
export function fullLabel(at: number | string) {
  return fullTime.format(new Date(at));
}
// Tooltips: date, seconds and time zone.
export function exactLabel(at: number | string) {
  return exactTime.format(new Date(at));
}
export function relationshipText(edge: GraphRelation, id: string) {
  if (edge.from === id) return relationLabel(edge.kind);
  return (
    (
      {
        blocks: 'blocks',
        waits_for: 'waited on by',
        tracks: 'tracked by',
        parent: 'child of',
        created: 'created from',
        wait: 'waited on by',
      } as Record<string, string>
    )[edge.kind] || edge.kind
  );
}
