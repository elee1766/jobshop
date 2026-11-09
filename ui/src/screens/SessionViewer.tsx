import { useEffect, useRef, useState } from 'react';
import { sessionActivity } from '../lib/session.ts';
import { useSessionFrames } from '../lib/useSessionFrames.ts';
import { Transcript } from './Transcript.tsx';

// A read-only view of one session. Escape and the dismiss button close it;
// focus starts on the dismiss button so the keyboard can reach it at once.
export function SessionViewer({ id, onClose }: { id: string; onClose: () => void }) {
  const { frames, connection, error, prepend } = useSessionFrames(id);
  const [viewport, setViewport] = useState<HTMLDivElement | null>(null);
  const closeButton = useRef<HTMLButtonElement>(null);
  const activity = sessionActivity(frames);

  useEffect(() => closeButton.current?.focus({ preventScroll: true }), []);
  useEffect(() => {
    const key = (event: KeyboardEvent) => {
      if (event.key !== 'Escape' || event.isComposing) return;
      event.preventDefault();
      onClose();
    };
    window.addEventListener('keydown', key);
    return () => window.removeEventListener('keydown', key);
  }, [onClose]);

  return (
    <section className="session-viewer" aria-label="Read-only session">
      <div className="tag">
        <span>session</span>
        <output className={'chat-state' + (connection === 'connected' && activity === 'working' ? ' working' : '')}>
          read only · {connection === 'connected' ? activity : connection}
        </output>
        <button ref={closeButton} className="push-right" onClick={onClose}>
          dismiss ×
        </button>
      </div>
      <div className="session-identity mono">{id}</div>
      <div className="session-output" ref={setViewport} aria-live="polite" aria-label="Session transcript">
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        {!frames.length && !error && (
          <p className="muted">
            {connection === 'ended' ? 'No journal messages recorded.' : 'Waiting for session output…'}
          </p>
        )}
        <Transcript frames={frames} viewport={viewport} onOlder={prepend} />
      </div>
    </section>
  );
}
