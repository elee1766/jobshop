import type { GraphNode, GraphRelation, GraphResult, Ticket } from '../../packages/jobshop-client/rpc.js';

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
  return ({ blocks: 'blocked by', waits_for: 'waits for', tracks: 'tracks', parent: 'parent of', created: 'led to' } as Record<string, string>)[kind] || kind;
}

export function linkLabel(href: string) {
  const url = new URL(href);
  if (url.hostname === 'github.com') return url.pathname.slice(1).replace('/pull/', ' PR #').replace('/issues/', ' #');
  if (url.hostname === 'linear.app') {
    const key = url.pathname.split('/issue/')[1]?.split('/')[0] || 'Linear';
    try { return decodeURIComponent(key); } catch { return key; }
  }
  return url.hostname + url.pathname;
}

export function safeLinks(node?: GraphNode) {
  return (node?.links || []).filter(link => {
    try { return ['http:', 'https:'].includes(new URL(link).protocol); } catch { return false; }
  });
}

// Creation time controls x; lanes prevent simultaneous tickets from overlapping.
export function graphLayout(graph: GraphResult) {
  const ordered = [...graph.nodes].sort((a, b) => Date.parse(a.created_at) - Date.parse(b.created_at) || a.id.localeCompare(b.id));
  const start = ordered.length ? Date.parse(ordered[0].created_at) : 0;
  const end = ordered.length ? Date.parse(ordered[ordered.length - 1].created_at) : start;
  const span = Math.max(600, Math.min(3600, graph.nodes.length * 160));
  const lanes: number[] = [];
  const positions = new Map<string, { x: number; y: number }>();
  for (const node of ordered) {
    const x = 16 + (end > start ? (Date.parse(node.created_at) - start) / (end - start) * span : 0);
    let lane = lanes.findIndex(right => right + 36 <= x);
    if (lane < 0) lane = lanes.length;
    lanes[lane] = x + 248;
    positions.set(node.id, { x, y: 64 + lane * 200 });
  }
  return {
    positions,
    ticks: Array.from({ length: end > start ? 5 : 1 }, (_, i) => ({ x: 16 + i / 4 * span, at: new Date(start + i / 4 * (end - start)).toISOString() })),
    width: span + 280,
    height: Math.max(264, 64 + lanes.length * 200),
  };
}

export function relationshipText(edge: GraphRelation, id: string) {
  if (edge.from === id) return relationLabel(edge.kind);
  return ({ blocks: 'blocks', waits_for: 'waited on by', tracks: 'tracked by', parent: 'child of', created: 'created from' } as Record<string, string>)[edge.kind] || edge.kind;
}
