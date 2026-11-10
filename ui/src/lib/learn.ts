import type { Learning, MemoryEdit } from '../../../packages/jobshop-client/rpc.js';

// Learn groups what the curator loop recorded by where each problem stands:
// a learning is the problem and a suggested fix, a memory edit is the change
// that was actually tried, and the gate's state on the edit is the outcome.
// A learning an edit cites is shown inside that edit, never twice.

export type Stage =
  | 'not-helped'
  | 'awaiting'
  | 'in-flight'
  | 'evaluating'
  | 'helped'
  | 'applied'
  | 'rejected'
  | 'expired';

export interface Section {
  stage: Stage;
  title: string;
  hint: string;
  // Open by default: recent or needing a look. The rest is history.
  open: boolean;
}

export const sections: Section[] = [
  { stage: 'not-helped', title: 'not helped', hint: 'the gate judged the change did not help', open: true },
  { stage: 'in-flight', title: 'fix proposed', hint: 'a curator pull request addresses it', open: true },
  { stage: 'evaluating', title: 'evaluating', hint: 'change applied, the gate has not decided', open: true },
  { stage: 'awaiting', title: 'awaiting curator', hint: 'open, no change proposed yet', open: true },
  { stage: 'helped', title: 'helped', hint: 'the gate found the change helped', open: false },
  { stage: 'applied', title: 'applied, not evaluated', hint: 'marked applied, no gate result recorded', open: false },
  { stage: 'rejected', title: 'rejected', hint: 'the proposed change was turned down', open: false },
  { stage: 'expired', title: 'expired', hint: 'old and never addressed, the curator no longer reads it', open: false },
];

export interface Where {
  text: string;
  // Empty for a local file or an id with nothing to open.
  href: string;
}

export interface Unit {
  key: string;
  stage: Stage;
  // When it was recorded (a learning) or merged (an edit).
  at: string;
  // Headline: the first cited problem, or empty when none is recorded.
  problem: string;
  more: number;
  // The suggested fix of the first cited learning that has one.
  fix: string;
  // "attempted fix" once a change landed (edit or applied), "proposed fix"
  // while a pull request is open or was turned down, else "suggested fix":
  // a proposal is not a change that was tried.
  fixLabel: string;
  where: Where | null;
  outcome: string;
  kind: string;
  scope: string;
  learnings: Learning[];
  // Cited learning ids that are not in the loaded list.
  missing: number[];
  edit?: MemoryEdit;
  search: string;
}

export const recentMs = 7 * 24 * 3600 * 1000;

export function isURL(value?: string) {
  return !!value && /^https?:\/\//i.test(value);
}

export function whereOf(pr?: string, path?: string): Where | null {
  if (path) return { text: path, href: '' };
  if (!pr) return null;
  if (isURL(pr)) return { text: prLabel(pr), href: pr };
  const memory = /^memory:(\d+)$/.exec(pr);
  return { text: memory ? 'memory edit #' + memory[1] : pr, href: '' };
}

function prLabel(url: string) {
  const m = /^https?:\/\/github\.com\/([^/]+\/[^/]+)\/pull\/(\d+)/.exec(url);
  return m ? `${m[1]}#${m[2]}` : url.replace(/^https?:\/\//, '');
}

function editStage(state: string): Stage {
  return state === 'helped' ? 'helped' : state === 'not_helped' ? 'not-helped' : 'evaluating';
}

function outcomeOf(edit: MemoryEdit) {
  if (edit.state === 'helped') return 'helped: ' + (edit.reason || 'no reason recorded');
  if (edit.state === 'not_helped') return 'not helped: ' + (edit.reason || 'no reason recorded');
  return 'not evaluated yet';
}

export function learningOutcome(l: Learning) {
  if (l.gate) return l.gate.replace(/^(helped|not_helped):/, (_, s) => s.replace('_', ' ') + ':');
  if (l.status === 'rejected') return 'rejected' + (l.pr ? ': ' + (whereOf(l.pr)?.text || l.pr) : '');
  if (l.status === 'applied') return 'applied, no gate result recorded';
  if (l.status === 'expired')
    return 'expired, never addressed' + (l.pr ? ' (fix proposed: ' + (whereOf(l.pr)?.text || l.pr) + ')' : '');
  if (l.pr) return 'fix proposed, not merged yet';
  return 'not addressed yet';
}

export function stageOf(l: Learning): Stage {
  if (l.status === 'rejected') return 'rejected';
  if (l.status === 'applied') return 'applied';
  if (l.status === 'expired') return 'expired';
  return l.pr ? 'in-flight' : 'awaiting';
}

const stamp = (value?: string) => (value && !value.startsWith('0001') ? Date.parse(value) : 0);

export function build(learnings: Learning[], edits: MemoryEdit[]): Unit[] {
  const byID = new Map(learnings.map((l) => [l.id, l]));
  const cited = new Set<number>();
  const units: Unit[] = [];
  for (const edit of edits) {
    const ls: Learning[] = [];
    const missing: number[] = [];
    for (const id of edit.learnings || []) {
      cited.add(id);
      const l = byID.get(id);
      if (l) ls.push(l);
      else missing.push(id);
    }
    const first = ls[0];
    units.push({
      key: 'edit:' + edit.id,
      stage: editStage(edit.state),
      at: edit.merged_at || edit.created_at,
      problem: first?.summary || '',
      more: Math.max(0, ls.length - 1) + missing.length,
      fix: ls.find((l) => l.fix)?.fix || '',
      fixLabel: 'attempted fix',
      where: whereOf(edit.pr, edit.path),
      outcome: outcomeOf(edit),
      kind: first?.kind || '',
      scope: first?.scope || (edit.pools || []).map((p) => 'pool:' + p).join(' '),
      learnings: ls,
      missing,
      edit,
      search: '',
    });
  }
  for (const l of learnings) {
    if (cited.has(l.id)) continue;
    units.push({
      key: 'learning:' + l.id,
      stage: stageOf(l),
      at: l.created_at,
      problem: l.summary,
      more: 0,
      fix: l.fix || '',
      fixLabel: l.status === 'applied' ? 'attempted fix' : l.pr ? 'proposed fix' : 'suggested fix',
      where: whereOf(l.pr),
      outcome: learningOutcome(l),
      kind: l.kind,
      scope: l.scope,
      learnings: [l],
      missing: [],
      search: '',
    });
  }
  for (const u of units) {
    u.search = [
      u.problem,
      u.fix,
      u.kind,
      u.scope,
      u.where?.text,
      u.outcome,
      ...u.learnings.flatMap((l) => [l.summary, l.fix, l.ticket, l.session, '#' + l.id]),
      u.edit?.pr,
      u.edit?.path,
      ...(u.edit?.pools || []),
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase();
  }
  return units.toSorted((a, b) => stamp(b.at) - stamp(a.at) || (a.key < b.key ? -1 : 1));
}

export function matches(unit: Unit, filter: string) {
  const terms = filter.toLowerCase().split(/\s+/).filter(Boolean);
  return terms.every((term) => unit.search.includes(term));
}

export interface Summary {
  items: number;
  learnings: number;
  edits: number;
  fresh: number;
}

// Counts units, not records: an edit and the learnings it cites are one item.
export function summarize(units: Unit[], learnings: Learning[], edits: MemoryEdit[], now: number): Summary {
  return {
    items: units.length,
    learnings: learnings.length,
    edits: edits.length,
    fresh: units.filter((u) => isFresh(u.at, now)).length,
  };
}

export function isFresh(at: string, now: number) {
  const t = stamp(at);
  return t > 0 && now - t <= recentMs;
}

export const metricRows: {
  key: 'runs' | 'steps' | 'retries' | 'escalations' | 'findings' | 'recurrence';
  label: string;
}[] = [
  { key: 'runs', label: 'runs' },
  { key: 'steps', label: 'steps' },
  { key: 'retries', label: 'retries' },
  { key: 'escalations', label: 'escalations' },
  { key: 'findings', label: 'review findings' },
  { key: 'recurrence', label: 'same problem again' },
];
