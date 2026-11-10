import type { ShowResult, TimelineRow } from '../../../packages/jobshop-client/rpc.js';
import { parseTitle } from '../../kit/titles.js';
import { Markdown } from '../components/Markdown.tsx';
import { markerKinds } from '../lib/timeline.ts';
import { exactLabel, fullLabel, linkLabel } from '../lib/trace.ts';

const web = /^https?:\/\//;

// What one marker on the timeline says, opened on demand under the chart
// (#1182). The chart only draws a tick, so the text stays out of it. A comment
// shows its author, time and full text; a history event its label and raw row.
// The ticket's own thread is only loaded for the selected row, so the panel
// shows the marker's one line until it arrives.
export function MarkerPanel({
  row,
  index,
  detail,
  onStep,
  onClose,
}: {
  row: TimelineRow;
  index: number;
  detail?: ShowResult;
  onStep: (to: number) => void;
  onClose: () => void;
}) {
  const marker = row.markers[index];
  const thread = detail?.ticket.id === row.id ? detail : undefined;
  const comment = marker.comment ? thread?.comments?.find((c) => c.id === marker.comment) : undefined;
  const entry = marker.seq ? thread?.history?.find((h) => h.seq === marker.seq) : undefined;
  const diff = entry?.diff != null && Object.keys(entry.diff as object).length > 0 ? entry.diff : undefined;
  const author = comment?.author || marker.actor;
  const last = row.markers.length - 1;
  return (
    <section className="tl-marker-panel" aria-label="Marker detail">
      <div className="tag">
        <h3>{markerKinds[marker.kind] || marker.kind}</h3>
        <span className="muted" title={row.id}>
          {parseTitle(row.title).text || row.kind}
        </span>
        <span className="muted">
          {index + 1} of {row.markers.length}
        </span>
        <div className="push-right tl-marker-steps">
          <button aria-label="Previous marker" disabled={index <= 0} onClick={() => onStep(index - 1)}>
            prev
          </button>
          <button aria-label="Next marker" disabled={index >= last} onClick={() => onStep(index + 1)}>
            next
          </button>
          <button aria-label="Close marker detail" onClick={onClose}>
            close
          </button>
        </div>
      </div>
      <div className="tl-marker-body">
        <p className="muted">
          {author && <b>{author}</b>}
          {author && ' · '}
          <time dateTime={marker.at} title={exactLabel(marker.at)}>
            {fullLabel(marker.at)}
          </time>
        </p>
        {comment ? (
          <Markdown text={comment.body} />
        ) : (
          <p>
            {marker.url && web.test(marker.url) ? (
              <a href={marker.url} target="_blank" rel="noopener noreferrer">
                {marker.label || linkLabel(marker.url)}
              </a>
            ) : (
              marker.label || marker.kind
            )}
          </p>
        )}
        {comment && marker.url && web.test(marker.url) && (
          <p>
            <a href={marker.url} target="_blank" rel="noopener noreferrer">
              {linkLabel(marker.url)}
            </a>
          </p>
        )}
        {marker.comment && !comment && (
          <p className="muted">{thread ? 'The full text is no longer available.' : 'loading the full text…'}</p>
        )}
        {diff !== undefined && <pre className="tl-marker-raw">{JSON.stringify(diff, null, 2)}</pre>}
      </div>
    </section>
  );
}
