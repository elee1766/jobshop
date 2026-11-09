// Pure helpers of the Linear and GitHub views (JOB-163): what a row shows,
// how a list is grouped, and where a linked item opens. No React and no I/O,
// so node's test runner covers them.
import type {
  BrowseItem,
  BrowseLinked,
  BrowseScopeResult,
  BrowseStatus,
} from '../../../packages/jobshop-client/rpc.js';

export type Provider = 'linear' | 'github';

export const providerName: Record<Provider, string> = { linear: 'Linear', github: 'GitHub' };

// The views of a list, the provider's own words. The first is the default.
export const listViews: Record<Provider, string[]> = {
  linear: ['open', 'active', 'backlog', 'closed', 'all'],
  github: ['open', 'merged', 'closed', 'all'],
};

export const groupModes: Record<Provider, string[]> = {
  linear: ['status', 'owner', 'project', 'none'],
  github: ['state', 'review', 'checks', 'author', 'none'],
};

// The one-letter status in the first column. Its full name is the label, so
// the letter is never the only carrier of the state.
export function statusLetter(item: BrowseItem): string {
  if (item.provider === 'github') return { open: 'O', draft: 'D', merged: 'M', closed: 'C' }[item.state] ?? '?';
  const types: Record<string, string> = {
    triage: 'T',
    backlog: 'B',
    unstarted: 'O',
    started: 'I',
    completed: 'C',
    canceled: 'X',
  };
  return types[item.state_type] ?? '?';
}

// The ticket-status colors of the rest of Desk: open, in progress, closed.
export function statusTone(item: BrowseItem): 'open' | 'in_progress' | 'closed' {
  const type = item.provider === 'github' ? item.state : item.state_type;
  if (type === 'started' || type === 'draft') return 'in_progress';
  if (type === 'completed' || type === 'canceled' || type === 'merged' || type === 'closed') return 'closed';
  return 'open';
}

export function statusLabel(item: BrowseItem): string {
  return item.provider === 'github' ? item.state : `${item.state} (${item.state_type})`;
}

const checkWord: Record<string, string> = { pass: 'pass', fail: 'fail', run: 'run' };
export const checksText = (value?: string) => checkWord[value ?? ''] ?? '-';

export interface Group {
  key: string;
  label: string;
  items: BrowseItem[];
}

const linearOrder = ['triage', 'started', 'unstarted', 'backlog', 'completed', 'canceled'];
const linearLabel: Record<string, string> = {
  triage: 'triage',
  started: 'active',
  unstarted: 'todo',
  backlog: 'backlog',
  completed: 'done',
  canceled: 'canceled',
};
const githubOrder = ['open', 'draft', 'merged', 'closed'];
const reviewLabel: Record<string, string> = {
  required: 'review required',
  changes: 'changes requested',
  approved: 'approved',
  '': 'no review state',
};
const reviewOrder = ['required', 'changes', 'approved', ''];
const checksLabel: Record<string, string> = { fail: 'failing', run: 'running', pass: 'passing', '': 'no checks' };
const checksOrder = ['fail', 'run', 'pass', ''];

// groupItems keeps the list's order inside each group and orders the groups by
// the provider's own sequence, or by first appearance when it has none.
export function groupItems(items: BrowseItem[], mode: string): Group[] {
  const keyOf = (item: BrowseItem): [string, string] => {
    switch (mode) {
      case 'status':
        return [item.state_type, linearLabel[item.state_type] ?? item.state_type];
      case 'state':
        return [item.state, item.state];
      case 'owner':
      case 'author':
        return [item.owner ?? '', item.owner || 'nobody'];
      case 'project':
        return [item.project ?? '', item.project || 'no project'];
      case 'review':
        return [item.review ?? '', reviewLabel[item.review ?? ''] ?? item.review ?? ''];
      case 'checks':
        return [item.checks ?? '', checksLabel[item.checks ?? ''] ?? item.checks ?? ''];
    }
    return ['', ''];
  };
  if (mode === 'none' || !mode) return [{ key: '', label: '', items }];
  const groups = new Map<string, Group>();
  for (const item of items) {
    const [key, label] = keyOf(item);
    const group = groups.get(key) ?? { key, label, items: [] };
    group.items.push(item);
    groups.set(key, group);
  }
  const order = { status: linearOrder, state: githubOrder, review: reviewOrder, checks: checksOrder }[mode];
  const out = [...groups.values()];
  if (order) out.sort((a, b) => rank(order, a.key) - rank(order, b.key));
  return out;
}
const rank = (order: string[], key: string) => {
  const at = order.indexOf(key);
  return at < 0 ? order.length : at;
};

// The login to filter a GitHub list by: a GitHub App's search form carries
// [bot], which the controller turns into app/<login>.
export const authorValue = (item: BrowseItem) => (item.owner ?? '') + (item.owner_bot ? '[bot]' : '');

export interface Place {
  // An internal link: the view and the query that open it.
  view?: Provider;
  params?: Record<string, string>;
  // The provider's own page, when there is one.
  url?: string;
}

const repoRef = /^([\w.-]+\/[\w.-]+)#(\d+)$/;
const linearKey = /^([A-Za-z][A-Za-z0-9]*)-\d+$/;

// whereIs resolves a linked item. Inside the configured scope it opens in
// Desk; outside it only links to the provider, so Desk never offers a read it
// would refuse. A run opens its trace, which is not decided here.
export function whereIs(link: BrowseLinked, scope: BrowseScopeResult | null): Place {
  const url = link.url?.startsWith('https://') ? link.url : undefined;
  const repo = repoRef.exec(link.ref);
  if (repo && scope?.github.some((s) => s.repo === repo[1])) {
    return { view: 'github', params: { scope: repo[1], item: link.ref, tab: '' }, url };
  }
  const team = linearKey.exec(link.ref);
  if (team && scope?.linear.some((s) => s.team === team[1].toUpperCase())) {
    return { view: 'linear', params: { scope: team[1].toUpperCase(), item: link.ref.toUpperCase(), tab: '' }, url };
  }
  return { url };
}

// Which of a status's messages the reader sees first. A stale list still
// shows its rows; an error with nothing to show replaces them.
export function statusNotice(status: BrowseStatus | undefined, now = Date.now()): string {
  if (!status || status.state === 'fresh' || status.state === 'cached') return '';
  const wait = status.retry_at ? Math.max(0, Math.ceil((Date.parse(status.retry_at) - now) / 1000)) : 0;
  const retry = wait > 0 ? ` Try again in ${wait < 90 ? wait + 's' : Math.ceil(wait / 60) + 'm'}.` : '';
  const prefix = status.state === 'stale' ? 'Showing an older copy. ' : '';
  return prefix + (status.message || 'The provider could not be read.') + retry;
}

// A short label for an error code, for the tag bar.
export const codeLabel: Record<string, string> = {
  auth: 'auth',
  rate_limited: 'rate limited',
  unavailable: 'unavailable',
  not_configured: 'not configured',
  not_found: 'not found',
  error: 'error',
};

// The text a new-comment divider counts: the ids in a detail now that were
// not in the first read of it.
export function newComments(seen: Set<string>, ids: string[]): string[] {
  return seen.size === 0 ? [] : ids.filter((id) => !seen.has(id));
}

// How long ago, in the compact form of the rest of Desk (10m, 3h, 2d).
export function age(value?: string, now = Date.now()): string {
  if (!value || value.startsWith('0001')) return '';
  const minutes = Math.max(0, Math.floor((now - Date.parse(value)) / 60000));
  if (minutes < 1) return 'just now';
  if (minutes < 60) return minutes + 'm ago';
  if (minutes < 1440) return Math.floor(minutes / 60) + 'h ago';
  return Math.floor(minutes / 1440) + 'd ago';
}

// A seconds count as 1m42s, for a check's duration.
export function duration(seconds?: number): string {
  if (!seconds || seconds < 0) return '';
  if (seconds < 60) return seconds + 's';
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return m < 60 ? `${m}m${String(s).padStart(2, '0')}s` : `${Math.floor(m / 60)}h${String(m % 60).padStart(2, '0')}m`;
}
