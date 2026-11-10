import type {
  GraphNode,
  GraphResult,
  HistoryEntry,
  HolderInfo,
  Ticket,
  TimelineRow,
  TimelineSegment,
  TimelineWait,
} from '../../../packages/jobshop-client/rpc.js';
import { durationLabel, holderState, legendLabels, liveText, segmentEnd, turnTotals } from './timeline.ts';
import { relationshipText } from './trace.ts';

// The model behind the trace detail pane (JOB-185): what happened to a ticket
// and what it is doing now, derived only from its history and the timeline.
// Nothing here is estimated: a number that cannot be computed is left out.

const at = (t: string | undefined) => (t ? Date.parse(t) : NaN);

// The tabs of the lower area, with the value kept in the URL (?tab=). The
// marker and session tabs only exist while a tick or a session is open; the
// page falls back to events when the one in the URL has nothing to show.
export const detailTabs = [
  'events',
  'marker',
  'session',
  'details',
  'io',
  'relations',
  'body',
  'comments',
  'raw',
] as const;
export type DetailTab = (typeof detailTabs)[number];

export function detailTab(value: string): DetailTab {
  return (detailTabs as readonly string[]).includes(value) ? (value as DetailTab) : 'events';
}

// ---- time by state ----------------------------------------------------

export interface TimeShare {
  key: string;
  // The legend swatch (data-state) that draws this share.
  swatch: string;
  label: string;
  ms: number;
}

const shares: [string, string, string][] = [
  ['queued', 'queued', 'queued'],
  ['turn', 'working', legendLabels.working],
  ['held', 'held', legendLabels.held],
  ['working', 'working', 'working'],
  ['retry', 'retry', 'failed attempts'],
  ['waiting', 'waiting', legendLabels.waiting],
  ['human', 'human', legendLabels.human],
  ['deferred', 'deferred', legendLabels.deferred],
  ['unknown', 'unknown', legendLabels.unknown],
];

// Where a ticket's time went, from its timeline segments. A working span
// splits into time in a turn and time held with none when its claim has turn
// rows (JOB-184), else it is just working. A run or batch is working for its
// whole life, which says nothing: it has no split.
export function timeByState(row: TimelineRow, now: number): TimeShare[] {
  if (row.kind === 'run' || row.kind === 'batch') return [];
  const sums = new Map<string, number>();
  const add = (key: string, ms: number) => {
    if (ms > 0) sums.set(key, (sums.get(key) || 0) + ms);
  };
  for (const seg of row.segments) {
    const length = segmentEnd(seg, now) - at(seg.start);
    if (!(length > 0)) continue;
    if (seg.state === 'working') {
      const totals = turnTotals(seg, now);
      if (totals) {
        add('turn', totals.active);
        add('held', totals.held);
      } else add('working', length);
    } else add(seg.state, length);
  }
  return shares
    .filter(([key]) => sums.has(key))
    .map(([key, swatch, label]) => ({ key, swatch, label, ms: sums.get(key)! }));
}

// Claims taken and failed attempts, one per working or retry span with a
// session. Controller-run tickets (command, script) have none.
export function attempts(row: TimelineRow) {
  let claims = 0;
  let failed = 0;
  for (const seg of row.segments) {
    if (!seg.actor || (seg.state !== 'working' && seg.state !== 'retry')) continue;
    claims++;
    if (seg.state === 'retry') failed++;
  }
  return { claims, failed };
}

interface StepLike {
  retries?: number;
  failures?: number;
  step?: string;
  type?: string;
  state?: string;
  wait?: { reason?: string };
}
const fieldsOf = (ticket: Ticket) =>
  ticket.fields && typeof ticket.fields === 'object' ? (ticket.fields as StepLike) : {};

export interface Summary {
  created: number;
  ended: number | null;
  took: number;
  shares: TimeShare[];
  claims: number;
  failed: number;
  // A step's retry limit; 0 when the ticket has none.
  limit: number;
  // The timeline did not keep all of this ticket's history.
  partial: boolean;
  // The ticket is not in the timeline, so only its own fields are summarized.
  bare: boolean;
}

export function summarize(ticket: Ticket, row: TimelineRow | undefined, now: number): Summary {
  const created = at(ticket.created_at);
  const ended = ticket.closed_at ? at(ticket.closed_at) : null;
  const f = fieldsOf(ticket);
  const counts = row ? attempts(row) : { claims: 0, failed: 0 };
  return {
    created,
    ended,
    took: Math.max(0, (ended ?? now) - created),
    shares: row ? timeByState(row, now) : [],
    claims: counts.claims,
    failed: counts.failed,
    limit: ticket.kind === 'step' && f.retries && f.retries > 0 ? f.retries : 0,
    partial: !!row?.partial,
    bare: !row,
  };
}

// Steps of a run or batch: how many direct children are closed.
export function childProgress(rows: TimelineRow[] | undefined, id: string) {
  const kids = (rows || []).filter((row) => row.parent === id);
  return { total: kids.length, closed: kids.filter((row) => row.status === 'closed').length };
}

// ---- now ---------------------------------------------------------------

export interface Now {
  kind: 'closed' | 'turn' | 'idle' | 'lost' | 'held' | 'waiting' | 'human' | 'deferred' | 'queued' | 'running' | 'open';
  text: string;
  // The start of the span the ticket is in, when known.
  since?: number;
  until?: number;
  holder?: string;
  waits: TimelineWait[];
}

// What an open ticket is doing or waiting for, or how a closed one ended.
// The open span of its timeline row says it; without a row only the ticket's
// own fields can.
export function nowOf(
  ticket: Ticket,
  row: TimelineRow | undefined,
  kids: { total: number; closed: number },
  session: HolderInfo | undefined,
  history: HistoryEntry[] = [],
): Now {
  if (ticket.status === 'closed') {
    const last = history.findLast((e) => e.op === 'close');
    const diff = (last?.diff || {}) as { note?: string; reason?: string };
    const why = diff.note || diff.reason;
    // The outcome is its own fact; here is what the close row says about why.
    return {
      kind: 'closed',
      text: why || ['closed', ticket.resolution].filter(Boolean).join(' '),
      since: ticket.closed_at ? at(ticket.closed_at) : undefined,
      waits: [],
    };
  }
  const seg = row?.segments.findLast((s) => !s.end);
  if (seg) return nowOfSegment(ticket, seg, kids, session);
  if (ticket.status === 'in_progress' && ticket.holder) {
    return { kind: 'held', text: 'held by ' + ticket.holder, holder: ticket.holder, waits: [] };
  }
  const f = fieldsOf(ticket);
  if (f.wait?.reason) return { kind: 'waiting', text: f.wait.reason, waits: [] };
  if (ticket.defer_until && at(ticket.defer_until) > Date.now()) {
    return { kind: 'deferred', text: 'deferred', until: at(ticket.defer_until), waits: [] };
  }
  return { kind: 'open', text: ticket.status === 'open' ? 'open, not claimed' : ticket.status, waits: [] };
}

function nowOfSegment(
  ticket: Ticket,
  seg: TimelineSegment,
  kids: { total: number; closed: number },
  session: HolderInfo | undefined,
): Now {
  const since = at(seg.start);
  switch (seg.state) {
    case 'working': {
      if (!seg.actor) {
        const steps = kids.total ? `${kids.closed} of ${kids.total} steps closed` : '';
        return { kind: 'running', text: ['running', steps].filter(Boolean).join(' · '), since, waits: [] };
      }
      const live = holderState(seg, session);
      const words = liveText(live, session);
      const kind = live === 'running' ? 'turn' : live === 'idle' ? 'idle' : live === 'lost' ? 'lost' : 'held';
      return { kind, text: words || 'held by ' + seg.actor, since, holder: seg.actor, waits: [] };
    }
    case 'waiting':
      return { kind: 'waiting', text: seg.detail || 'waiting on an event', since, waits: seg.wait || [] };
    case 'human':
      return { kind: 'human', text: seg.detail || 'waiting on a human', since, waits: [] };
    case 'deferred':
      return { kind: 'deferred', text: 'deferred', since, until: seg.until ? at(seg.until) : undefined, waits: [] };
    case 'queued':
      return { kind: 'queued', text: 'queued' + (ticket.pool ? ' for ' + ticket.pool : ''), since, waits: [] };
    default:
      return { kind: 'open', text: legendLabels[seg.state] || seg.state, since, waits: [] };
  }
}

// ---- events ------------------------------------------------------------

export type EventItem =
  | { type: 'segment'; key: string; at: number; index: number; seg: TimelineSegment; ended: string }
  | { type: 'close'; key: string; at: number; resolution: string; by: string; note: string };

// Ops that never end a span the way a reader cares about.
const quiet = new Set(['turn_start', 'turn_end', 'comment', 'edge_add', 'edge_remove', 'label_add', 'label_remove']);

// How a span ended, from the history row written at its end. A deferral that
// simply ran out has none.
export function endedBy(seg: TimelineSegment, history: HistoryEntry[]): string {
  if (!seg.end) return '';
  // A wait already says how it ended ("ended met") in its own text.
  if (seg.ended) return '';
  const end = at(seg.end);
  const entry = history.find((e) => at(e.at) === end && !quiet.has(e.op));
  if (!entry) return '';
  const diff = (entry.diff || {}) as { reason?: string; resolution?: string; note?: string; to?: string };
  const why = diff.reason || diff.note;
  switch (entry.op) {
    case 'claim':
      return 'claimed by ' + entry.actor;
    case 'release':
      return ['released', why && '(' + why + ')'].filter(Boolean).join(' ');
    case 'lease_expired':
      return 'claim expired';
    case 'handoff':
      return 'handed off' + (diff.to ? ' to ' + diff.to : '');
    case 'close':
      return ['closed', diff.resolution, why && '(' + why + ')'].filter(Boolean).join(' ');
    case 'reopen':
      return 'reopened';
    default:
      return '';
  }
}

// One entry per span of the ticket's life (a claim, a wait, a human hold,
// time queued or deferred) and the close, by time. Newest first. Comments and
// the history rows no span shows are markers on the timeline (#1182).
export function eventItems(row: TimelineRow, history: HistoryEntry[]): EventItem[] {
  const items: EventItem[] = row.segments.map((seg, index) => ({
    type: 'segment',
    key: 'seg:' + index,
    at: at(seg.start),
    index,
    seg,
    ended: endedBy(seg, history),
  }));
  if (row.closed_at) {
    const last = history.findLast((e) => e.op === 'close');
    const diff = (last?.diff || {}) as { note?: string; reason?: string };
    items.push({
      type: 'close',
      key: 'close',
      at: at(row.closed_at),
      resolution: row.resolution || '',
      by: last?.actor || '',
      note: diff.note || diff.reason || '',
    });
  }
  // Ties keep timeline order; the stable sort then reversed puts the later one first.
  return items.toSorted((a, b) => a.at - b.at).toReversed();
}

// A history that does not start at the ticket's creation is partial.
export function historyPartial(history: HistoryEntry[]) {
  return history.length > 0 && history[0].op !== 'create';
}

// ---- input and result --------------------------------------------------

export interface KV {
  key: string;
  value: string;
  // Long values are collapsed behind their first line.
  long: boolean;
}

const longValue = 140;

export function kv(key: string, value: unknown): KV | null {
  if (value === undefined || value === null || value === '') return null;
  let text: string;
  if (typeof value === 'string') text = value;
  else if (typeof value === 'number' || typeof value === 'boolean') text = String(value);
  else if (Array.isArray(value) && value.every((v) => typeof v === 'string' || typeof v === 'number'))
    text = value.join(', ');
  else text = JSON.stringify(value, null, 2);
  return { key, value: text, long: text.length > longValue || text.includes('\n') };
}

function pairs(object: unknown): KV[] {
  if (!object || typeof object !== 'object') return [];
  return Object.entries(object as Record<string, unknown>)
    .map(([key, value]) => kv(key, value))
    .filter((row): row is KV => !!row);
}

const some = (...list: (KV | null)[]) => list.filter((row): row is KV => !!row);

export interface IO {
  inputs: KV[];
  result: KV[];
}

// A run's inputs and outcome, a step's inputs, outcome, output and last
// command, a batch's key and issues. Other tickets have inputs only if their
// fields hold some.
export function ioOf(ticket: Ticket): IO {
  const f = (ticket.fields && typeof ticket.fields === 'object' ? ticket.fields : {}) as Record<string, any>;
  const resolution = kv('resolution', ticket.resolution);
  switch (ticket.kind) {
    case 'run':
      return {
        inputs: [...some(kv('workflow', f.workflow), kv('key', f.key)), ...pairs(f.vars), ...pairs(f.inputs)],
        result: [...some(kv('state', f.state), resolution), ...pairs(f.automation?.output)],
      };
    case 'step':
      return {
        inputs: [
          ...some(
            kv('step', f.step),
            kv('run', f.run),
            kv('type', f.type),
            kv('needs', f.needs),
            kv('script', f.script),
            kv('timeout', f.timeout),
          ),
          ...pairs(f.inputs),
        ],
        result: [
          ...some(
            resolution,
            kv('result', f.result),
            kv('outcome', f.outcome),
            kv('failures', f.failures),
            kv('retries', f.retries),
            kv('command', f.command),
          ),
          ...pairs(f.structured_output),
          ...some(kv('output', f.output)),
        ],
      };
    case 'batch':
      return {
        inputs: some(kv('key', f.key), kv('issues', f.issues), kv('order', f.order)),
        result: some(kv('state', f.state), resolution),
      };
    default:
      return { inputs: pairs(f.inputs), result: some(resolution) };
  }
}

export function hasIO(ticket: Ticket) {
  if (['run', 'step', 'batch'].includes(ticket.kind)) return true;
  const io = ioOf(ticket);
  return io.inputs.length > 0;
}

// ---- relations -----------------------------------------------------------

const groupOrder = ['parent', 'children', 'blocked by', 'blocks', 'waits for', 'waited on by'];
const groupRank = (kind: string) => {
  const i = groupOrder.indexOf(kind);
  return i < 0 ? 99 : i;
};

export interface RelationGroup {
  kind: string;
  nodes: GraphNode[];
}

// The tickets related to one ticket, by how: parent, children, blockers and
// blocked, waits and the rest.
export function relationGroups(graph: GraphResult, ticket: Pick<Ticket, 'id' | 'parent_id'>): RelationGroup[] {
  const out = new Map<string, GraphNode[]>();
  const add = (kind: string, node: GraphNode | undefined) => {
    if (!node) return;
    const list = out.get(kind) || [];
    if (!list.some((n) => n.id === node.id)) list.push(node);
    out.set(kind, list);
  };
  const find = (id: string) => graph.nodes.find((n) => n.id === id);
  for (const edge of graph.relations) {
    if (edge.from !== ticket.id && edge.to !== ticket.id) continue;
    const id = edge.from === ticket.id ? edge.to : edge.from;
    let kind = edge.kind === 'parent' && edge.from === ticket.id ? 'children' : relationshipText(edge, ticket.id);
    if (kind === 'child of') kind = 'parent';
    add(kind, find(id));
  }
  // The parent link may be missing from the graph's edges.
  if (ticket.parent_id && !out.get('parent')?.some((n) => n.id === ticket.parent_id))
    add('parent', find(ticket.parent_id));
  return [...out].map(([kind, nodes]) => ({ kind, nodes })).toSorted((a, b) => groupRank(a.kind) - groupRank(b.kind));
}

export const relationCount = (groups: RelationGroup[]) => new Set(groups.flatMap((g) => g.nodes.map((n) => n.id))).size;

// How long a related ticket has existed or took: creation to close, or so
// far while open.
export function nodeDuration(node: Pick<GraphNode, 'created_at' | 'closed_at' | 'status'>, now: number) {
  const end = node.closed_at ? at(node.closed_at) : now;
  const ms = end - at(node.created_at);
  if (!(ms >= 0)) return '';
  return durationLabel(ms) + (node.status === 'closed' ? '' : ' so far');
}
