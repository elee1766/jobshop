import { useAtom, useAtomValue } from 'jotai';
import { useDeskRoute } from '../lib/route.ts';
import { inspect, openAudit, openTrace } from '../state/actions.ts';
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
  const invocation = (selected?.meta as { automation?: { invocation?: string } } | undefined)?.automation?.invocation;
  return (
    <>
      <TraceSearch key={route.ref} reference={route.ref} onOpen={openTrace} />
      {trace?.graph && (
        <div className={'trace-layout' + (watched ? ' watching' : '')}>
          {trace.timeline ? (
            <Timeline
              timeline={trace.timeline}
              sessions={trace.sessions}
              selected={selected?.id || trace.graph.root}
              onSelect={inspect}
            />
          ) : (
            <section aria-label="Ticket timeline" className="ticket-timeline">
              <div className="tag graph-heading">
                <h2>timeline</h2>
              </div>
              <p className="muted section-body">loading timeline…</p>
            </section>
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
            {watched && <SessionViewer key={watched} id={watched} onClose={() => setWatched('')} />}
          </div>
        </div>
      )}
      {invocation && (
        <p className="section-body">
          automation invocation{' '}
          <button className="mono" onClick={() => openAudit(invocation)}>
            {invocation}
          </button>
        </p>
      )}
    </>
  );
}
