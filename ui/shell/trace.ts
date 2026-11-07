import type { GraphNode, GraphRelation, GraphResult, Ticket } from '../../packages/jobshop-client/rpc.js';

export function state(ticket: Ticket) {
  const fields = ticket.fields as { state?: string } | undefined;
  return [fields?.state || ticket.status, ticket.resolution].filter(Boolean).join(' · ');
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

// Breadth-first columns keep every relationship visible, including cycles and shared dependencies.
export function graphLayout(graph: GraphResult) {
  const depth = new Map([[graph.root, 0]]);
  const queue = [graph.root];
  for (let i = 0; i < queue.length; i++) {
    for (const edge of graph.relations) {
      const next = edge.from === queue[i] ? edge.to : edge.to === queue[i] ? edge.from : '';
      if (next && !depth.has(next)) {
        depth.set(next, depth.get(queue[i])! + 1);
        queue.push(next);
      }
    }
  }
  const rows = new Map<number, number>();
  const positions = new Map<string, { x: number; y: number }>();
  for (const node of graph.nodes) {
    const column = depth.get(node.id) || 0;
    const row = rows.get(column) || 0;
    positions.set(node.id, { x: 16 + column * 300, y: 16 + row * 180 });
    rows.set(column, row + 1);
  }
  return {
    positions,
    width: Math.max(280, ...[...positions.values()].map(p => p.x + 264)),
    height: Math.max(176, ...[...positions.values()].map(p => p.y + 160)),
  };
}

export function relationshipText(edge: GraphRelation, id: string) {
  if (edge.from === id) return relationLabel(edge.kind);
  return ({ blocks: 'blocks', waits_for: 'waited on by', tracks: 'tracked by', parent: 'child of', created: 'created from' } as Record<string, string>)[edge.kind] || edge.kind;
}
