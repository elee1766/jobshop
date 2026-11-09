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
