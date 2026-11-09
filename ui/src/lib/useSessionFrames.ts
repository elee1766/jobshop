import { useEffect, useReducer, useRef, useState } from 'react';
import type { SessionEvent } from '../../../packages/jobshop-client/rpc.js';
import { client } from './api.ts';

export interface SessionStream {
  id: string | null;
  frames: SessionEvent[];
  // connecting | connected | reconnecting | ended | error | closed
  connection: string;
  error: string;
}

const empty = (id: string | null): SessionStream => ({
  id,
  frames: [],
  connection: id ? 'connecting' : 'closed',
  error: '',
});

type Change = { id: string | null; apply: (s: SessionStream) => SessionStream };

// The stored stream is keyed by session id. A change for another id starts
// from an empty stream, so a switch never shows the old session's frames and
// no effect has to reset state.
function reduce(stored: SessionStream, change: Change): SessionStream {
  return change.apply(stored.id === change.id ? stored : empty(change.id));
}

// useSessionFrames follows one session's stream. Frames that arrive in a
// burst are applied once per animation frame, so a busy session costs one
// render per frame, not one per event. Going offline closes the stream and
// coming online reopens it from the last sequence number seen.
export function useSessionFrames(id: string | null) {
  const [stored, dispatch] = useReducer(reduce, id, empty);
  const [online, setOnline] = useState(() => navigator.onLine);
  const last = useRef<{ id: string | null; seq: number }>({ id, seq: 0 });
  const seen = useRef(new Set<number>());

  useEffect(() => {
    const up = () => setOnline(true);
    const down = () => setOnline(false);
    window.addEventListener('online', up);
    window.addEventListener('offline', down);
    return () => {
      window.removeEventListener('online', up);
      window.removeEventListener('offline', down);
    };
  }, []);

  useEffect(() => {
    if (last.current.id !== id) {
      last.current = { id, seq: 0 };
      seen.current = new Set();
    }
    if (!id) return;
    if (!online) {
      dispatch({ id, apply: (s) => (s.connection === 'ended' ? s : { ...s, connection: 'reconnecting' }) });
      return;
    }
    let live = true;
    let buffer: SessionEvent[] = [];
    let scheduled = 0;
    const flush = () => {
      scheduled = 0;
      if (!live || !buffer.length) return;
      const batch = buffer;
      buffer = [];
      dispatch({ id, apply: (s) => ({ ...s, frames: [...s.frames, ...batch] }) });
    };
    const close = client.session(
      id,
      (event, kind) => {
        if (!live) return;
        if (kind === 'end') {
          flush();
          dispatch({ id, apply: (s) => ({ ...s, connection: 'ended' }) });
          return;
        }
        if (kind !== 'session') return;
        const frame = event as SessionEvent;
        if (frame.seq) {
          last.current.seq = Math.max(last.current.seq, frame.seq);
          if (seen.current.has(frame.seq)) return;
          seen.current.add(frame.seq);
        }
        buffer.push(frame);
        if (!scheduled) scheduled = requestAnimationFrame(flush);
      },
      {
        since: last.current.seq,
        onState: (state, detail) => {
          if (!live) return;
          dispatch({
            id,
            apply: (s) =>
              s.connection === 'ended'
                ? s
                : {
                    ...s,
                    connection: state,
                    error: state === 'error' ? detail?.error || 'Unable to read this session.' : s.error,
                  },
          });
        },
      },
    );
    return () => {
      live = false;
      if (scheduled) cancelAnimationFrame(scheduled);
      close();
    };
  }, [id, online]);

  // Older pages of history are prepended by the transcript through this.
  const prepend = (older: SessionEvent[]) => {
    for (const frame of older) if (frame.seq) seen.current.add(frame.seq);
    dispatch({
      id,
      apply: (s) => {
        const have = new Set(s.frames.map((f) => f.seq));
        return { ...s, frames: [...older.filter((f) => !have.has(f.seq)), ...s.frames] };
      },
    });
  };

  const current = stored.id === id ? stored : empty(id);
  return { ...current, prepend };
}
