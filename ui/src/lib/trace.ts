import type { GraphNode, GraphRelation, GraphResult, Ticket } from '../../../packages/jobshop-client/rpc.js';

export function stepName(fields: unknown): string {
  if (!fields || typeof fields !== 'object') return '';
  const step = (fields as { step?: unknown }).step;
  return typeof step === 'string' ? step : '';
}

export function state(ticket: Ticket) {
  const fields = ticket.fields as { state?: string } | undefined;
  return [fields?.state || ticket.status, ticket.resolution].filter(Boolean).join(' · ');
}

export function statusLetter(ticket: Ticket): string {
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

// The tooltip and text of a wait edge: what it waits on, its state and why.
export function waitText(edge: GraphRelation) {
  return [edge.state || 'waiting', edge.reason].filter(Boolean).join(': ');
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

// Graph card geometry in canvas pixels, shared by the layout, edges and CSS.
export const card = { width: 248, height: 148, lane: 164, left: 16, top: 16 };

// The trace's time span: start is the first creation, end the latest close
// (completed) or now (anything still open).
export function traceEnd(graph: GraphResult, now: number) {
  const active = graph.nodes.some((node) => node.status !== 'closed');
  let last = 0;
  for (const node of graph.nodes)
    last = Math.max(last, Date.parse(node.created_at), node.closed_at ? Date.parse(node.closed_at) : 0);
  return { active, at: active ? Math.max(now, last) : last };
}

// Creation time controls x through one scale (x = left + (t - start) * scale),
// which the time axis reuses; lanes keep simultaneous tickets apart. The
// scale spreads creations over the canvas; the trace end (now or the last
// close) may lie right of the last card.
// External nodes (what a wait names outside the store) are placed by the
// time the earliest wait on them began.
export function graphLayout(graph: GraphResult) {
  const items = [
    ...graph.nodes.map((node) => ({ id: node.id, at: node.created_at })),
    ...(graph.external || []).map((ext) => ({ id: ext.id, at: ext.at })),
  ];
  const ordered = items.toSorted((a, b) => Date.parse(a.at) - Date.parse(b.at) || a.id.localeCompare(b.id));
  const start = ordered.length ? Date.parse(ordered[0].at) : 0;
  const last = ordered.length ? Date.parse(ordered[ordered.length - 1].at) : start;
  const span = Math.max(600, Math.min(3600, items.length * 160));
  const scale = span / (last > start ? last - start : 60_000);
  const lanes: number[] = [];
  const positions = new Map<string, { x: number; y: number }>();
  for (const node of ordered) {
    const x = card.left + (Date.parse(node.at) - start) * scale;
    let lane = lanes.findIndex((right) => right + 36 <= x);
    if (lane < 0) lane = lanes.length;
    lanes[lane] = x + card.width;
    positions.set(node.id, { x, y: card.top + lane * card.lane });
  }
  return {
    positions,
    start,
    scale,
    width: card.left + span + card.width + 16,
    height: Math.max(card.top + card.lane + 16, card.top + lanes.length * card.lane),
  };
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
const shortTime = new Intl.DateTimeFormat(undefined, {
  month: 'short',
  day: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
});

// Endpoint labels: date and minute.
export function fullLabel(at: number | string) {
  return fullTime.format(new Date(at));
}
// Tooltips: date, seconds and time zone.
export function exactLabel(at: number | string) {
  return exactTime.format(new Date(at));
}
// Compact card times.
export function shortLabel(at: number | string) {
  return shortTime.format(new Date(at));
}

// The parts of state() a status letter does not already say: a run's
// fields.state and the resolution, minus the status itself.
export function stateDetail(ticket: Ticket) {
  const same = new Set([ticket.status, ticket.status === 'in_progress' ? 'running' : '']);
  const fields = ticket.fields as { state?: string } | undefined;
  return [fields?.state, ticket.resolution].filter((part): part is string => !!part && !same.has(part)).join(' · ');
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
