import { useAtom, useAtomValue } from 'jotai';
import { useEffect, useRef, useState } from 'react';
import { useMedia } from '../lib/useMedia.ts';
import { setParams, useDeskRoute } from '../lib/route.ts';
import { inspect, openTrace } from '../state/actions.ts';
import { selectedAtom, traceAtom } from '../state/live.ts';
import { markAtom, watchedSessionAtom } from '../state/ui.ts';
import { TicketDetails } from './TicketDetails.tsx';
import { Timeline } from './Timeline.tsx';
import { TraceSearch } from './TraceSearch.tsx';

// The TL page (#1183): the timeline on top, always the same size, and under it
// one area for whatever is selected. A press on the chart picks a ticket, a
// bar or a tick and the area shows it: the events of the ticket, the tick's
// text, the read-only session of its holder, or the full details, each behind
// a tab that sits right under the chart. Nothing is stacked, so no panel pushes
// another out of reach.
// On a phone the area does not share the screen with the chart. A press opens
// it as a full-screen popup over the page. The chart stays mounted under the
// popup, so closing it returns to the same selection, zoom and scroll.
export function Trace() {
  const route = useDeskRoute();
  const trace = useAtomValue(traceAtom);
  const selected = useAtomValue(selectedAtom);
  const [watched, setWatched] = useAtom(watchedSessionAtom);
  const [, setMark] = useAtom(markAtom);
  const phone = useMedia('(max-width: 599px)');
  // A link to a tab or to an inspected ticket opens the popup at once; a link
  // to a reference alone shows the chart. A new search starts on the chart.
  const [popup, setPopup] = useState(route.tab !== '' || route.inspect !== '');
  const id = selected?.id || trace?.graph?.root || '';
  // A chart with no rows has nothing to press, so its area stays inline.
  const pressable = !!trace?.timeline?.rows.length;
  const popped = phone && pressable && popup;
  const close = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (!popped) return;
    close.current?.focus({ preventScroll: true });
    // The session viewer inside closes itself on Escape first (its listener is
    // added first and marks the key handled); the popup takes the next one.
    const key = (event: KeyboardEvent) => {
      if (event.key !== 'Escape' || event.defaultPrevented || event.isComposing) return;
      setPopup(false);
    };
    window.addEventListener('keydown', key);
    return () => window.removeEventListener('keydown', key);
  }, [popped]);
  // The session tab is the way to a session, so showing one asks for it.
  const watch = (session: string) => {
    setWatched(session);
    if (session) setParams({ tab: 'session' });
  };
  const lower = trace?.graph && selected && (
    <TicketDetails
      detail={trace.inspection || trace}
      graph={trace.graph}
      watched={watched}
      onSelect={(next) => {
        setMark(null);
        inspect(next);
      }}
      onTrace={openTrace}
      onWatch={watch}
    />
  );
  return (
    <>
      <TraceSearch
        key={route.ref}
        reference={route.ref}
        onOpen={(ref) => {
          setPopup(false);
          setMark(null);
          openTrace(ref);
        }}
      />
      {trace?.graph && (
        <div className="trace-layout">
          {trace.timeline ? (
            <Timeline
              timeline={trace.timeline}
              sessions={trace.sessions}
              selected={id}
              onSelect={inspect}
              onPick={() => setPopup(true)}
            />
          ) : (
            <section aria-label="Ticket timeline" className="ticket-timeline">
              <div className="tag graph-heading">
                <h2>timeline</h2>
              </div>
              <p className="muted section-body">loading timeline…</p>
            </section>
          )}
          {popped ? (
            <dialog open className="lower-popup" aria-label="Selected item">
              <div className="tag">
                <span className="mono muted">{id}</span>
                <button
                  ref={close}
                  className="push-right"
                  aria-label="Close and return to the timeline"
                  onClick={() => setPopup(false)}
                >
                  close ×
                </button>
              </div>
              <div className="lower-popup-body">{lower}</div>
            </dialog>
          ) : (
            // A phone shows the area inline only when the chart has nothing to press.
            (!phone || !pressable) && lower
          )}
        </div>
      )}
    </>
  );
}
