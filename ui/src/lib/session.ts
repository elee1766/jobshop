import type { SessionEvent } from '../../../packages/jobshop-client/rpc.js';

export interface PendingMessage {
  id: string;
  text: string;
  attachments: string[];
  state: 'sending' | 'queued';
}

export function sessionActivity(frames: SessionEvent[], fallback = 'idle') {
  for (let index = frames.length - 1; index >= 0; index--) {
    const frame = frames[index];
    if (frame.type === 'turn_end' || frame.error) return 'idle';
    if (['turn_start', 'text', 'tool', 'image', 'reasoning'].includes(frame.type)) return 'working';
  }
  return fallback;
}

// What the agent is doing now in a turn that is still running (JOB-192).
export type TurnPhase = { kind: 'thinking' | 'replying' } | { kind: 'tool'; title: string };

const running = (status: string) => status !== 'completed' && status !== 'failed' && status !== 'cancelled';

// turnPhase derives the phase from the frames the stream already delivers: the
// newest frame of a running turn says what it is doing, and a tool that has
// not finished takes precedence. It returns null when the last turn ended or
// the frames show no turn. `busy` is the host's own word that a turn is in
// flight, used when the frames loaded (the tail of a long turn) hold no clue.
export function turnPhase(frames: SessionEvent[], busy = false): TurnPhase | null {
  let last: TurnPhase | null = null;
  const settled = new Set<string>();
  for (let index = frames.length - 1; index >= 0; index--) {
    const frame = frames[index];
    if (frame.type === 'turn_end' || frame.error) return null;
    if (frame.type === 'turn_start') return last ?? { kind: 'thinking' };
    if (frame.type === 'tool' && frame.tool) {
      if (running(frame.tool.status) && !settled.has(frame.tool.id)) {
        return { kind: 'tool', title: frame.tool.intent || frame.tool.title || 'tool' };
      }
      settled.add(frame.tool.id);
      last ??= { kind: 'thinking' };
    } else if (frame.type === 'text' || frame.type === 'image') {
      last ??= { kind: 'replying' };
    } else if (frame.type === 'reasoning') {
      last ??= { kind: 'thinking' };
    }
  }
  return last ?? (busy ? { kind: 'thinking' } : null);
}

export function turnLabel(phase: TurnPhase) {
  return phase.kind === 'tool' ? 'running: ' + phase.title : phase.kind + '…';
}
