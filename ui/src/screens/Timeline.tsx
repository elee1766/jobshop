import { useSetAtom } from 'jotai';
import { useEffect, useMemo, useRef, useState } from 'react';
import type { HolderInfo, TimelineResult, TimelineRow, TimelineSegment } from '../../../packages/jobshop-client/rpc.js';
import { parseTitle } from '../../kit/titles.js';
import { StatusLetter } from '../components/StatusLetter.tsx';
import { setParams } from '../lib/route.ts';
import { segmentChoiceAtom } from '../state/ui.ts';
import { useSize } from '../lib/useSize.ts';
import {
  arrowPath,
  arrowShape,
  durationLabel,
  holderState,
  markerText,
  parentIds,
  rowEnd,
  rowStart,
  segmentEnd,
  segmentText,
  timelineSpan,
  visibleRows,
} from '../lib/timeline.ts';
import { axisTicks, exactLabel, fullLabel, relativeLabel } from '../lib/trace.ts';

// Geometry in CSS pixels. The label column is also set in app.css (--tl-label).
const rowH = 30;
const barH = 14;
const axisH = 34;
const labelW = 220;
const minChart = 320;
const minZoom = 1;
const maxZoom = 64;

// One bar per ticket on a shared time axis, nested by parent, with
// arrows for blocking edges (#1103). Rows come from ticket.timeline in
// tree order; open bars grow to the "now" line.
export function Timeline({
  timeline,
  sessions,
  selected,
  onSelect,
}: {
  timeline: TimelineResult;
  sessions?: Record<string, HolderInfo>;
  selected: string;
  onSelect: (id: string) => void;
}) {
  const [now, setNow] = useState(() => Date.now());
  const [zoom, setZoom] = useState(1);
  const [collapsed, setCollapsed] = useState<ReadonlySet<string>>(new Set());
  const [seg, setSeg] = useState<{ id: string; i: number } | null>(null);
  const choose = useSetAtom(segmentChoiceAtom);
  const [scrollRef, { width }] = useSize<HTMLElement>();

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

  // Opening a sub-item selects it where it sits in the whole graph, which may
  // be far down a large one: bring its row into the chart's own scroll (not
  // the page's) when the selection changes or the rows first arrive.
  const chart = useRef<HTMLElement | null>(null);
  const pickRow = index.get(pick?.id ?? '');
  const found = pick !== undefined && pickRow !== undefined;
  useEffect(() => {
    const el = chart.current;
    if (!el || !found) return;
    const top = axisH + pickRow * rowH;
    if (top < el.scrollTop + axisH) el.scrollTop = Math.max(0, top - axisH);
    else if (top + rowH > el.scrollTop + el.clientHeight) el.scrollTop = top + rowH - el.clientHeight;
  }, [pickRow, found]);

  function toggle(id: string) {
    setCollapsed((prev) => {
      const next = new Set(prev);
      if (!next.delete(id)) next.add(id);
      return next;
    });
  }

  const text = (s: TimelineSegment) => segmentText(s, { now, session: s.actor ? sessions?.[s.actor] : undefined });
  const segTitle = (row: TimelineRow, s: TimelineSegment) =>
    `${text(s)}\n${exactLabel(s.start)} to ${s.end ? exactLabel(s.end) : 'now'} (${durationLabel(segmentEnd(s, now) - Date.parse(s.start))})\n${row.id}`;

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
        </div>
      </div>
      {timeline.truncated && (
        <output className="block">
          This timeline exceeds the display limit. Showing a bounded portion. Open a related ticket as a new trace to
          explore further.
        </output>
      )}
      <section
        className="tl-scroll"
        ref={(node) => {
          chart.current = node;
          scrollRef(node);
        }}
        aria-label="Timeline chart"
      >
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
                      const live = holderState(s, s.actor ? sessions?.[s.actor] : undefined);
                      return (
                        <button
                          key={i}
                          className="tl-seg"
                          data-state={s.state}
                          data-open={!s.end || undefined}
                          data-turns={s.turns?.length || undefined}
                          data-live={live}
                          aria-label={text(s) + ', ' + row.id}
                          aria-pressed={seg?.id === row.id && seg.i === i}
                          title={segTitle(row, s)}
                          style={{ left: left + 'px', width: w + 'px', height: barH + 'px' }}
                          onClick={() => {
                            setSeg({ id: row.id, i });
                            choose((prev) => ({ id: row.id, index: i, n: (prev?.n ?? 0) + 1 }));
                            // The pane opens on its events tab, on this entry.
                            setParams({ tab: '' });
                            onSelect(row.id);
                          }}
                        >
                          {s.state === 'working' &&
                            s.turns?.map((turn, j) => {
                              const from = Math.max(0, x(Date.parse(turn.start)) - left);
                              const to = Math.min(w, x(turn.end ? Date.parse(turn.end) : segmentEnd(s, now)) - left);
                              return (
                                <span
                                  key={j}
                                  className="tl-turn"
                                  data-running={!turn.end || undefined}
                                  aria-hidden="true"
                                  style={{ left: from + 'px', width: Math.max(2, to - from) + 'px' }}
                                />
                              );
                            })}
                        </button>
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
    </section>
  );
}
