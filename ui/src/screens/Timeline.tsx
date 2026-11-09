import { useEffect, useMemo, useState } from 'react';
import type { TimelineResult, TimelineRow, TimelineSegment } from '../../../packages/jobshop-client/rpc.js';
import { parseTitle } from '../../kit/titles.js';
import { StatusLetter } from '../components/StatusLetter.tsx';
import { useSize } from '../lib/useSize.ts';
import {
  arrowPath,
  arrowShape,
  durationLabel,
  markerText,
  parentIds,
  rowEnd,
  rowStart,
  segmentEnd,
  segmentText,
  stateLabels,
  timelineSpan,
  visibleRows,
} from '../lib/timeline.ts';
import { axisTicks, exactLabel, fullLabel, linkLabel, relativeLabel } from '../lib/trace.ts';

// Geometry in CSS pixels. The label column is also set in app.css (--tl-label).
const rowH = 30;
const barH = 14;
const axisH = 34;
const labelW = 220;
const minChart = 320;
const minZoom = 1;
const maxZoom = 64;

const legend = ['queued', 'working', 'waiting', 'human', 'deferred', 'retry'];

// One bar per ticket on a shared time axis, nested by parent, with
// arrows for blocking edges (#1103). Rows come from ticket.timeline in
// tree order; open bars grow to the "now" line.
export function Timeline({
  timeline,
  selected,
  onSelect,
  onTrace,
  onGraph,
}: {
  timeline: TimelineResult;
  selected: string;
  onSelect: (id: string) => void;
  onTrace: (ref: string) => void;
  onGraph: () => void;
}) {
  const [now, setNow] = useState(() => Date.now());
  const [zoom, setZoom] = useState(1);
  const [collapsed, setCollapsed] = useState<ReadonlySet<string>>(new Set());
  const [seg, setSeg] = useState<{ id: string; i: number } | null>(null);
  const [scrollRef, { width }] = useSize<HTMLDivElement>();

  // An active timeline ends at "now", which moves.
  useEffect(() => {
    const clock = setInterval(() => setNow(Date.now()), 15_000);
    return () => clearInterval(clock);
  }, []);

  const span = useMemo(() => timelineSpan(timeline, now), [timeline, now]);
  const parents = useMemo(() => parentIds(timeline.rows), [timeline.rows]);
  const rows = useMemo(() => visibleRows(timeline.rows, collapsed), [timeline.rows, collapsed]);
  const index = useMemo(() => new Map(rows.map((row, i) => [row.id, i])), [rows]);
  const byId = useMemo(() => new Map(timeline.rows.map((row) => [row.id, row])), [timeline.rows]);
  const conflicts = useMemo(
    () => new Set(timeline.arrows.filter((a) => a.state === 'conflict').map((a) => a.to)),
    [timeline.arrows],
  );

  const fit = Math.max(minChart, width - labelW - 24);
  const chartW = Math.round(fit * zoom);
  const scale = chartW / (span.end - span.start);
  const x = (t: number) => (t - span.start) * scale;
  // Relative ticks keep clear of the start label (a date) and the end label.
  const ticks = axisTicks(span.start, span.end, scale, span.start, span.end, 80).filter(
    (t) => x(t) > 150 && x(t) < chartW - 70,
  );
  const height = rows.length * rowH;
  const pick = byId.get(selected) || byId.get(timeline.root);

  function toggle(id: string) {
    setCollapsed((prev) => {
      const next = new Set(prev);
      if (!next.delete(id)) next.add(id);
      return next;
    });
  }

  const segTitle = (row: TimelineRow, s: TimelineSegment) =>
    `${segmentText(s)}\n${exactLabel(s.start)} to ${s.end ? exactLabel(s.end) : 'now'} (${durationLabel(segmentEnd(s, now) - Date.parse(s.start))})\n${row.id}`;

  return (
    <section aria-label="Ticket timeline" className="ticket-timeline">
      <div className="tag graph-heading">
        <h2>timeline</h2>
        <span className="muted">
          {timeline.rows.length} tickets · {timeline.arrows.length} arrows
        </span>
        <div className="graph-controls push-right" aria-label="Timeline zoom">
          <button
            aria-label="Zoom out"
            disabled={zoom <= minZoom}
            onClick={() => setZoom((z) => Math.max(minZoom, z / 2))}
          >
            −
          </button>
          <output aria-label="Timeline zoom level">{Math.round(zoom * 100)}%</output>
          <button
            aria-label="Zoom in"
            disabled={zoom >= maxZoom}
            onClick={() => setZoom((z) => Math.min(maxZoom, z * 2))}
          >
            +
          </button>
          <button disabled={zoom === 1} onClick={() => setZoom(1)}>
            fit
          </button>
          <button onClick={onGraph}>graph</button>
        </div>
      </div>
      {timeline.truncated && (
        <output className="block">
          This timeline exceeds the display limit. Showing a bounded portion. Open a related ticket as a new trace to
          explore further.
        </output>
      )}
      <ul className="tl-legend muted" aria-label="Segment states">
        {legend.map((state) => (
          <li key={state}>
            <span className="tl-swatch" data-state={state} aria-hidden="true" />
            {stateLabels[state]}
          </li>
        ))}
        <li>
          <span className="tl-arrow-sample" data-shape="blocking" aria-hidden="true" />
          blocked
        </li>
        <li>
          <span className="tl-arrow-sample" data-shape="tracks" aria-hidden="true" />
          tracks
        </li>
      </ul>
      <section className="tl-scroll" ref={scrollRef} aria-label="Timeline chart">
        <div className="tl-canvas" style={{ width: labelW + chartW + 16 + 'px' }}>
          <div className="tl-axis" style={{ height: axisH + 'px' }}>
            <div className="tl-corner" style={{ width: labelW + 'px' }} />
            <div className="tl-axis-track" style={{ width: chartW + 'px' }}>
              <span className="tl-tick tl-tick-edge" style={{ left: 0 }} title={exactLabel(span.start)}>
                <time dateTime={new Date(span.start).toISOString()}>{fullLabel(span.start)}</time>
              </span>
              {ticks.map((t) => (
                <span key={t} className="tl-tick" style={{ left: x(t) + 'px' }} title={exactLabel(t)}>
                  {relativeLabel(t - span.start)}
                </span>
              ))}
              <span
                className={'tl-tick tl-tick-edge tl-tick-end' + (span.active ? ' tl-now' : '')}
                style={{ left: chartW + 'px' }}
                title={exactLabel(span.end)}
              >
                {span.active ? 'now' : 'end'}
              </span>
            </div>
          </div>
          <div className="tl-body" style={{ height: height + 'px' }}>
            {rows.map((row) => {
              const end = rowEnd(row, now);
              return (
                <div
                  key={row.id}
                  className={'tl-row' + (row.id === pick?.id ? ' tl-picked' : '')}
                  data-conflict={conflicts.has(row.id) || undefined}
                  data-status={row.status}
                  style={{ height: rowH + 'px' }}
                >
                  <div className="tl-label" style={{ width: labelW + 'px', paddingLeft: row.depth * 14 + 4 + 'px' }}>
                    {parents.has(row.id) ? (
                      <button
                        className="tl-fold"
                        aria-expanded={!collapsed.has(row.id)}
                        aria-label={(collapsed.has(row.id) ? 'Expand ' : 'Collapse ') + row.id}
                        onClick={() => toggle(row.id)}
                      >
                        {collapsed.has(row.id) ? '▸' : '▾'}
                      </button>
                    ) : (
                      <span className="tl-fold" aria-hidden="true" />
                    )}
                    <StatusLetter ticket={row} />
                    <button
                      className="tl-name"
                      title={`${parseTitle(row.title).text || row.kind} (${row.id})`}
                      aria-pressed={row.id === pick?.id}
                      onClick={() => onSelect(row.id)}
                    >
                      {parseTitle(row.title).text || row.kind}
                    </button>
                  </div>
                  <div className="tl-track" style={{ width: chartW + 'px' }}>
                    {row.segments.map((s, i) => {
                      const left = x(Date.parse(s.start));
                      const w = Math.max(3, x(segmentEnd(s, now)) - left);
                      return (
                        <button
                          key={i}
                          className="tl-seg"
                          data-state={s.state}
                          data-open={!s.end || undefined}
                          aria-label={segmentText(s) + ', ' + row.id}
                          aria-pressed={seg?.id === row.id && seg.i === i}
                          title={segTitle(row, s)}
                          style={{ left: left + 'px', width: w + 'px', height: barH + 'px' }}
                          onClick={() => {
                            setSeg({ id: row.id, i });
                            onSelect(row.id);
                          }}
                        />
                      );
                    })}
                    {row.closed_at && (
                      <span
                        className="tl-end"
                        data-resolution={row.resolution || 'closed'}
                        title={`closed ${row.resolution || ''} ${exactLabel(row.closed_at)}`}
                        style={{ left: x(end) + 'px', height: barH + 'px' }}
                      />
                    )}
                    {row.markers.map((m, i) => (
                      <span
                        key={i}
                        className="tl-marker"
                        data-kind={m.kind}
                        title={markerText(m) + '\n' + exactLabel(m.at)}
                        style={{ left: x(Date.parse(m.at)) + 'px' }}
                      />
                    ))}
                  </div>
                </div>
              );
            })}
            {span.active && <span className="tl-nowline" style={{ left: labelW + chartW + 'px' }} aria-hidden="true" />}
            <svg
              className="tl-arrows"
              width={chartW}
              height={height}
              style={{ left: labelW + 'px' }}
              aria-hidden="true"
            >
              <defs>
                <marker
                  id="tl-head"
                  viewBox="0 0 10 10"
                  refX="9"
                  refY="5"
                  markerWidth="5"
                  markerHeight="5"
                  orient="auto"
                >
                  <path d="M 0 0 L 10 5 L 0 10 z" />
                </marker>
              </defs>
              {timeline.arrows.map((a, i) => {
                const from = index.get(a.from);
                const to = index.get(a.to);
                const blocker = byId.get(a.from);
                const blocked = byId.get(a.to);
                if (from === undefined || to === undefined || !blocker || !blocked) return null;
                const shape = arrowShape(a);
                const x1 = x(rowEnd(blocker, now));
                const x2 = x(rowStart(blocked));
                const y1 = from * rowH + rowH / 2;
                const y2 = to * rowH + rowH / 2;
                return (
                  <path
                    key={i}
                    className="tl-arrow"
                    data-shape={shape}
                    data-state={a.state}
                    d={arrowPath(x1, y1, x2, y2)}
                    markerEnd="url(#tl-head)"
                  >
                    <title>
                      {a.from}{' '}
                      {a.kind === 'tracks' ? 'is tracked by' : a.kind === 'wait' ? 'is waited on by' : 'blocks'} {a.to}
                      {a.state ? ' (' + a.state + ')' : ''}
                    </title>
                  </path>
                );
              })}
            </svg>
          </div>
        </div>
      </section>
      {pick && <RowDetail row={pick} now={now} seg={seg?.id === pick.id ? seg.i : -1} onTrace={onTrace} />}
    </section>
  );
}

// The text form of one bar: every segment and marker, with exact times, and
// the waits it was parked on. Keyboard and touch reach everything the bar
// shows here.
function RowDetail({
  row,
  now,
  seg,
  onTrace,
}: {
  row: TimelineRow;
  now: number;
  seg: number;
  onTrace: (ref: string) => void;
}) {
  return (
    <div className="tl-detail" aria-label={'Timeline of ' + row.id}>
      <h3>
        <span className="mono">{row.id}</span> {parseTitle(row.title).text}
      </h3>
      {row.partial && (
        <p className="muted">
          Part of this ticket's history is not kept (an ephemeral or compacted ticket), so some of its time is shown as
          unknown.
        </p>
      )}
      <ol className="tl-segments">
        {row.segments.map((s, i) => (
          <li key={i} className={i === seg ? 'chosen' : undefined}>
            <span className="tl-swatch" data-state={s.state} aria-hidden="true" />
            <span>{segmentText(s)}</span>{' '}
            <span className="muted">
              <time dateTime={s.start} title={exactLabel(s.start)}>
                {fullLabel(s.start)}
              </time>
              {' · '}
              {durationLabel(segmentEnd(s, now) - Date.parse(s.start))}
              {s.end ? '' : ' so far'}
            </span>
            {s.wait && s.wait.length > 0 && (
              <ul>
                {s.wait.map((w, j) => (
                  <li key={j}>
                    {w.target ? (
                      <button className="mono" onClick={() => onTrace(w.target!)}>
                        {w.label}
                      </button>
                    ) : w.url && /^https?:\/\//.test(w.url) ? (
                      <a href={w.url} target="_blank" rel="noopener noreferrer">
                        {w.label || linkLabel(w.url)}
                      </a>
                    ) : (
                      w.label
                    )}
                    {w.state ? <span className="muted"> · {w.state}</span> : null}
                  </li>
                ))}
              </ul>
            )}
          </li>
        ))}
      </ol>
      {row.markers.length > 0 && (
        <details>
          <summary>{row.markers.length} markers</summary>
          <ul className="tl-markers">
            {row.markers.map((m, i) => (
              <li key={i}>
                <span className="tl-marker tl-marker-sample" data-kind={m.kind} aria-hidden="true" />
                <time dateTime={m.at} title={exactLabel(m.at)}>
                  {fullLabel(m.at)}
                </time>{' '}
                {m.url && /^https?:\/\//.test(m.url) ? (
                  <a href={m.url} target="_blank" rel="noopener noreferrer">
                    {markerText(m)}
                  </a>
                ) : (
                  markerText(m)
                )}
              </li>
            ))}
          </ul>
        </details>
      )}
    </div>
  );
}
