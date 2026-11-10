import { useAtom, useAtomValue } from 'jotai';
import { useState } from 'react';
import { useDeskRoute } from '../lib/route.ts';
import { inspect, openTrace } from '../state/actions.ts';
import { selectedAtom, traceAtom } from '../state/live.ts';
import { watchedSessionAtom } from '../state/ui.ts';
import { SessionViewer } from './SessionViewer.tsx';
import { TicketDetails } from './TicketDetails.tsx';
import { Timeline } from './Timeline.tsx';
import { TraceSearch } from './TraceSearch.tsx';

export function Trace() {
  const route = useDeskRoute();
  const trace = useAtomValue(traceAtom);
  const selected = useAtomValue(selectedAtom);
  const [watched, setWatched] = useAtom(watchedSessionAtom);
  // The timeline is the default (#1181). The details replace it on request and
  // the timeline stays mounted, so its zoom, folds and scroll survive the swap.
  // A deep link to a detail tab asks for the details. Following a link from
  // the details to another ticket stays on them, like a tab would. A new search
  // starts on the timeline.
  const [details, show] = useState(route.tab !== '');
  const id = selected?.id || trace?.graph?.root || '';
  // Without a timeline there is nothing to swap with: the details stand alone.
  const swapped = details || !trace?.timeline;
  const viewer = watched && <SessionViewer key={watched} id={watched} onClose={() => setWatched('')} />;
  return (
    <>
      <TraceSearch
        key={route.ref}
        reference={route.ref}
        onOpen={(ref) => {
          show(false);
          openTrace(ref);
        }}
      />
      {trace?.graph && (
        <div className={'trace-layout' + (swapped ? ' swapped' : '')}>
          {trace.timeline ? (
            <Timeline
              timeline={trace.timeline}
              sessions={trace.sessions}
              selected={id}
              detail={trace.inspection || trace}
              hidden={swapped}
              onSelect={inspect}
              onDetails={() => show(true)}
            />
          ) : (
            <section aria-label="Ticket timeline" className="ticket-timeline">
              <div className="tag graph-heading">
                <h2>timeline</h2>
              </div>
              <p className="muted section-body">loading timeline…</p>
            </section>
          )}
          {swapped ? (
            <section aria-label="Details view" className="trace-details">
              {trace.timeline && (
                <div className="tag graph-heading">
                  <h2>details</h2>
                  <span className="muted mono">{id}</span>
                  <button
                    className="push-right"
                    aria-label="Show the timeline"
                    title="Back to the timeline at the same selection"
                    onClick={() => show(false)}
                  >
                    timeline
                  </button>
                </div>
              )}
              <div className={'trace-bottom' + (watched ? ' watching' : '')}>
                {selected && (
                  <TicketDetails
                    detail={trace.inspection || trace}
                    graph={trace.graph}
                    onSelect={inspect}
                    onTrace={openTrace}
                    onWatch={setWatched}
                  />
                )}
                {viewer}
              </div>
            </section>
          ) : (
            watched && <div className="trace-bottom">{viewer}</div>
          )}
        </div>
      )}
    </>
  );
}
