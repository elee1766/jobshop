import type { RunResult, SessionInfo, Ticket } from '../../../packages/jobshop-client/rpc.js';

// What a run's current step is doing now, apart from the run's lifecycle
// (#1185). The run stays "running" while its step waits. Only a session in a
// turn is executing, and that comes from the live session state, never from
// the step having a holder or a retained session.
export type ActivityKind = 'turn' | 'held' | 'lost' | 'timer' | 'event' | 'human' | 'resource' | 'gated' | 'queued';

export interface Activity {
  kind: ActivityKind;
  label: string;
  // The reason, resume time or condition, when known.
  detail: string;
}

// The live state of a session as the system view reports it.
export type LiveSession = Pick<SessionInfo, 'alive' | 'busy' | 'tool' | 'host_pid' | 'exit' | 'suspended'>;

const humanLabels = ['needs-human', 'question', 'proposed'];
// A step with a when condition is held until its needs finish, with a
// defer_until in 2200. It is not a timer.
const gatedFrom = Date.UTC(2100, 0, 1);

const clock = (value: string) => new Date(value).toISOString().slice(0, 16).replace('T', ' ') + 'Z';

function inTime(value: string, now: number) {
  const minutes = Math.max(0, Math.floor((Date.parse(value) - now) / 60000));
  return minutes < 60
    ? minutes + 'm'
    : minutes < 1440
      ? Math.floor(minutes / 60) + 'h'
      : Math.floor(minutes / 1440) + 'd';
}

const until = (value: string, now: number) => `until ${clock(value)} (in ${inTime(value, now)})`;
const join = (...parts: string[]) => parts.filter(Boolean).join(' · ');

export function activityOf(
  run: RunResult,
  step: Ticket | undefined,
  session: LiveSession | undefined,
  now: number,
): Activity {
  const rf = (run.run.fields as Record<string, any>) || {};
  const sf = (step?.fields as Record<string, any>) || {};
  const wait = sf.wait;
  if (rf.state === 'needs_human' || step?.labels?.some((l) => humanLabels.includes(l))) {
    return { kind: 'human', label: 'waiting on a person', detail: wait?.reason || '' };
  }
  if (step?.holder) {
    if (!session) return { kind: 'held', label: 'held', detail: 'session state unknown' };
    if (session.suspended) return { kind: 'held', label: 'held, suspended', detail: '' };
    if (!session.alive && (session.host_pid || session.exit)) {
      return { kind: 'lost', label: 'host lost', detail: 'no live session is working this step' };
    }
    if (session.busy) return { kind: 'turn', label: 'in a turn', detail: session.tool || '' };
    return session.alive
      ? { kind: 'held', label: 'held, between turns', detail: '' }
      : { kind: 'held', label: 'held', detail: 'session state unknown' };
  }
  if (wait?.conds?.length) {
    const open = wait.conds.filter((c: any) => !c.met);
    const times = open.filter((c: any) => c.type === 'time' && c.at).map((c: any) => Date.parse(c.at));
    // A wait on times alone is a deferral until that time; with mode any it
    // ends at the earliest, else at the latest.
    if (open.length && times.length === open.length) {
      const end = new Date(wait.mode === 'any' ? Math.min(...times) : Math.max(...times)).toISOString();
      return { kind: 'timer', label: 'deferred', detail: join(wait.reason, until(end, now)) };
    }
    const names = open.map((c: any) => c.note || c.type).join(', ');
    return { kind: 'event', label: 'waiting on an event', detail: join(wait.reason, names) };
  }
  const slot = run.waits?.find((w) => w.ticket === step?.id);
  if (slot) return { kind: 'resource', label: 'waiting for a slot', detail: slot.key };
  const defer = step?.defer_until;
  if (defer && !defer.startsWith('0001') && Date.parse(defer) > now) {
    if (Date.parse(defer) >= gatedFrom) return { kind: 'gated', label: 'gated', detail: 'waits for earlier steps' };
    // The reason belongs to the release that set this time.
    const why = sf.defer && Date.parse(sf.defer.until) === Date.parse(defer) ? sf.defer.reason : '';
    return { kind: 'timer', label: 'deferred', detail: join(why, until(defer, now)) };
  }
  return { kind: 'queued', label: 'queued', detail: step?.pool || '' };
}
