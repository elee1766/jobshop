import {
  useEffect,
  useEffectEvent,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
  type PointerEvent as ReactPointerEvent,
} from 'react';
import type { GraphResult } from '../../../packages/jobshop-client/rpc.js';
import { parseTitle } from '../../kit/titles.js';
import { Markdown } from '../components/Markdown.tsx';
import { StatusLetter } from '../components/StatusLetter.tsx';
import { TitleText } from '../components/TitleText.tsx';
import { useSize } from '../lib/useSize.ts';
import {
  axisTicks,
  card,
  exactLabel,
  fullLabel,
  graphLayout,
  linkLabel,
  relationLabel,
  relativeLabel,
  safeLinks,
  shortLabel,
  stateDetail,
  stepName,
  traceEnd,
} from '../lib/trace.ts';

// Height of the time axis pinned to the top of the viewport.
const axis = 48;
const pad = 4;
const minZoom = 0.05;
const maxZoom = 3;

interface Camera {
  zoom: number;
  x: number;
  y: number;
}

// A callback ref that reports the element's width.
const measureWidth = (set: (w: number) => void) => (node: HTMLElement | null) => {
  if (node) set(node.clientWidth);
};

// The graph's camera is `null` while it fits the canvas to the viewport and
// becomes a value once the reader pans or zooms. Reset returns to null.
export function Graph({
  graph,
  selected,
  onSelect,
  onWatch,
}: {
  graph: GraphResult;
  selected: string;
  onSelect: (id: string) => void;
  onWatch: (id: string) => void;
}) {
  const layout = useMemo(() => graphLayout(graph), [graph]);
  const [viewportRef, { width, height }] = useSize<HTMLDivElement>();
  const viewport = useRef<HTMLDivElement | null>(null);
  const [touched, setTouched] = useState<Camera | null>(null);
  const [dragging, setDragging] = useState(false);
  const [now, setNow] = useState(() => Date.now());
  const pointers = useRef(new Map<number, { x: number; y: number }>());
  const [tip, setTip] = useState<{ text: string; left: number } | null>(null);
  const [pinned, setPinned] = useState(false);

  // An active trace ends at "now", which moves.
  useEffect(() => {
    const clock = setInterval(() => setNow(Date.now()), 30_000);
    return () => clearInterval(clock);
  }, []);

  const fit: Camera = useMemo(
    () => ({
      zoom: width && height ? Math.min(1, (width - 16) / layout.width, (height - axis - 16) / layout.height) : 1,
      x: 8,
      y: axis + 8,
    }),
    [width, height, layout],
  );
  const camera = touched ?? fit;
  const { zoom, x, y } = camera;

  // Canvas x of a time, then its viewport x through the camera: the axis and
  // the cards share this one transform, so ticks stay on their cards while
  // panning and zooming.
  const screenX = (t: number) => x + (card.left + (t - layout.start) * layout.scale) * zoom;
  const timeAt = (px: number) => layout.start + ((px - x) / zoom - card.left) / layout.scale;
  const end = traceEnd(graph, now);
  const startX = screenX(layout.start);
  const endX = screenX(end.at);

  // Endpoint labels stay in view (clamped to the edges) while their marks
  // sit at the exact position. Relative tick labels take the second row;
  // when the endpoint labels would collide the end label moves there too.
  const [startBox, setStartBox] = useState(0);
  const [endBox, setEndBox] = useState(0);
  const [tipBox, setTipBox] = useState(0);
  const startLeft = Math.max(pad, Math.min(width - pad - startBox, startX));
  const endLeft = Math.max(pad, Math.min(width - pad - endBox, endX - endBox));
  const stacked = endLeft < startLeft + startBox + 8;
  const ticks = axisTicks(layout.start, end.at, layout.scale * zoom, timeAt(0), timeAt(width)).filter((t) => {
    const px = screenX(t);
    return (
      px > 32 &&
      px < width - 32 &&
      px - startX > 40 &&
      Math.abs(endX - px) > 40 &&
      (!stacked || px + 32 < endLeft - 4 || px - 32 > endLeft + endBox + 4)
    );
  });

  function zoomAt(next: number, clientX?: number, clientY?: number) {
    const el = viewport.current;
    if (!el) return;
    const box = el.getBoundingClientRect();
    const px = clientX == null ? width / 2 : clientX - box.left;
    const py = clientY == null ? height / 2 : clientY - box.top;
    next = Math.max(minZoom, Math.min(maxZoom, next));
    const ratio = next / camera.zoom;
    setTouched({ zoom: next, x: px - (px - camera.x) * ratio, y: py - (py - camera.y) * ratio });
  }

  // The wheel handler must be non-passive to stop page scroll while zooming.
  // An effect event reads the newest camera without re-attaching the listener.
  const wheel = useEffectEvent((event: WheelEvent) => {
    event.preventDefault();
    const delta = event.deltaY * (event.deltaMode === 1 ? 16 : event.deltaMode === 2 ? height : 1);
    zoomAt(zoom * Math.exp(-Math.max(-400, Math.min(400, delta)) * 0.002), event.clientX, event.clientY);
  });
  useEffect(() => {
    const el = viewport.current;
    if (!el) return;
    const handler = (event: WheelEvent) => wheel(event);
    el.addEventListener('wheel', handler, { passive: false });
    return () => el.removeEventListener('wheel', handler);
  }, []);

  function pointerDown(event: ReactPointerEvent<HTMLDivElement>) {
    if (event.button !== 0 || (event.target as Element).closest('button, a')) return;
    event.currentTarget.setPointerCapture(event.pointerId);
    pointers.current.set(event.pointerId, { x: event.clientX, y: event.clientY });
    setDragging(true);
    setTouched((c) => c ?? fit);
    setTip(null);
    setPinned(false);
    event.currentTarget.focus();
  }
  function pointerMove(event: ReactPointerEvent<HTMLDivElement>) {
    const map = pointers.current;
    const previous = map.get(event.pointerId);
    if (!previous) return;
    const before = [...map.values()];
    map.set(event.pointerId, { x: event.clientX, y: event.clientY });
    if (map.size === 2) {
      const after = [...map.values()];
      const distance = (points: typeof before) => Math.hypot(points[1].x - points[0].x, points[1].y - points[0].y);
      const dx = (after[0].x + after[1].x - before[0].x - before[1].x) / 2;
      const dy = (after[0].y + after[1].y - before[0].y - before[1].y) / 2;
      const box = event.currentTarget.getBoundingClientRect();
      const cx = (before[0].x + before[1].x) / 2 - box.left;
      const cy = (before[0].y + before[1].y) / 2 - box.top;
      setTouched((c) => {
        const cur = c ?? fit;
        const next =
          distance(before) > 0
            ? Math.max(minZoom, Math.min(maxZoom, (cur.zoom * distance(after)) / distance(before)))
            : cur.zoom;
        const ratio = next / cur.zoom;
        return { zoom: next, x: cx - (cx - cur.x) * ratio + dx, y: cy - (cy - cur.y) * ratio + dy };
      });
    } else if (map.size === 1) {
      const dx = event.clientX - previous.x;
      const dy = event.clientY - previous.y;
      setTouched((c) => {
        const cur = c ?? fit;
        return { ...cur, x: cur.x + dx, y: cur.y + dy };
      });
    }
  }
  function pointerEnd(event: ReactPointerEvent<HTMLDivElement>) {
    pointers.current.delete(event.pointerId);
    setDragging(pointers.current.size > 0);
  }
  function keydown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape' && tip) {
      setTip(null);
      setPinned(false);
      return;
    }
    if (event.target !== event.currentTarget) return;
    const delta: Record<string, [number, number]> = {
      ArrowLeft: [40, 0],
      ArrowRight: [-40, 0],
      ArrowUp: [0, 40],
      ArrowDown: [0, -40],
    };
    if (event.key in delta) {
      event.preventDefault();
      const [dx, dy] = delta[event.key];
      setTouched({ ...camera, x: camera.x + dx, y: camera.y + dy });
    } else if (event.key === '+' || event.key === '=') {
      event.preventDefault();
      zoomAt(zoom * 1.25);
    } else if (event.key === '-') {
      event.preventDefault();
      zoomAt(zoom / 1.25);
    } else if (event.key === '0') {
      event.preventDefault();
      setTouched(null);
    }
  }

  // Exact time on hover, focus or tap (a tap pins it until the next tap).
  const show = (t: number, left: number) => setTip({ text: exactLabel(t), left });
  const hide = () => {
    if (!pinned) setTip(null);
  };
  const toggle = (t: number, left: number) => {
    const next = !(pinned && tip?.text === exactLabel(t));
    setPinned(next);
    if (next) show(t, left);
    else setTip(null);
  };

  const stamp = (t: number, label: string, kind: string, left: number, center: number) => (
    <button
      className={'graph-stamp graph-stamp-' + kind}
      style={{ left: left + 'px' }}
      aria-label={label + ', ' + exactLabel(t)}
      onPointerEnter={() => show(t, center)}
      onPointerLeave={hide}
      onFocus={() => show(t, center)}
      onBlur={() => {
        setPinned(false);
        setTip(null);
      }}
      onClick={() => toggle(t, center)}
    >
      <time dateTime={new Date(t).toISOString()}>{label}</time>
    </button>
  );

  const measure = measureWidth;

  return (
    <section aria-label="Ticket dependency graph" className="ticket-graph">
      <div className="tag graph-heading">
        <h2>graph</h2>
        <span className="muted">
          {graph.nodes.length} tickets · {graph.relations.length} relationships
        </span>
        <div className="graph-controls push-right" aria-label="Graph zoom">
          <button aria-label="Zoom out" disabled={zoom <= minZoom} onClick={() => zoomAt(zoom / 1.25)}>
            −
          </button>
          <output aria-label="Graph zoom level">{Math.round(zoom * 100)}%</output>
          <button aria-label="Zoom in" disabled={zoom >= maxZoom} onClick={() => zoomAt(zoom * 1.25)}>
            +
          </button>
          <button onClick={() => setTouched(null)}>reset</button>
        </div>
      </div>
      {graph.truncated && (
        <output className="block">
          This graph exceeds the display limit. Showing a bounded portion. Open a related ticket as a new graph to
          explore further.
        </output>
      )}
      <p className="muted graph-help">
        Creation time → left to right: a card's left edge is when it was created. Arrows show relationships, not
        execution order. Drag to pan; wheel or pinch to zoom. Keyboard: arrows, +/−, 0 to reset. Focus or tap a time for
        the exact date.
      </p>
      {/* The graph camera has application keyboard controls; ticket buttons remain native. */}
      <div
        className={'graph-scroll' + (dragging ? ' dragging' : '')}
        ref={(node) => {
          viewport.current = node;
          viewportRef(node);
        }}
        style={{ '--graph-axis': axis + 'px' } as React.CSSProperties}
        role="application"
        aria-label="Ticket graph navigation"
        tabIndex={0}
        onPointerDown={pointerDown}
        onPointerMove={pointerMove}
        onPointerUp={pointerEnd}
        onPointerCancel={pointerEnd}
        onLostPointerCapture={pointerEnd}
        onKeyDown={keydown}
      >
        <div className="graph-grid" aria-hidden="true">
          {ticks.map((t) => (
            <span key={t} style={{ left: screenX(t) + 'px' }} />
          ))}
          <span className="graph-grid-start" style={{ left: startX + 'px' }} />
          <span className={'graph-grid-end' + (end.active ? ' graph-now' : '')} style={{ left: endX + 'px' }} />
        </div>
        <div
          className="graph-canvas"
          style={{
            width: layout.width + 'px',
            height: layout.height + 'px',
            transform: `translate(${x}px, ${y}px) scale(${zoom})`,
          }}
        >
          <svg width={layout.width} height={layout.height} aria-hidden="true">
            <defs>
              <marker
                id="graph-arrow"
                viewBox="0 0 10 10"
                refX="9"
                refY="5"
                markerWidth="6"
                markerHeight="6"
                orient="auto-start-reverse"
              >
                <path d="M 0 0 L 10 5 L 0 10 z" />
              </marker>
            </defs>
            {graph.relations.map((edge, i) => {
              const from = layout.positions.get(edge.from);
              const to = layout.positions.get(edge.to);
              if (!from || !to) return null;
              const right = to.x > from.x;
              const x1 = from.x + (right || to.x === from.x ? card.width : 0);
              const x2 = to.x + (right ? 0 : card.width);
              const y1 = from.y + card.height / 2;
              const y2 = to.y + card.height / 2;
              const bend = to.x === from.x ? x1 + 36 : (x1 + x2) / 2;
              return (
                <path
                  key={i}
                  className={edge.kind === 'blocks' || edge.kind === 'waits_for' ? 'graph-gate' : undefined}
                  d={`M ${x1} ${y1} C ${bend} ${y1}, ${bend} ${y2}, ${x2} ${y2}`}
                  markerEnd="url(#graph-arrow)"
                >
                  <title>
                    {edge.from} {relationLabel(edge.kind)} {edge.to}
                  </title>
                </path>
              );
            })}
          </svg>
          {graph.nodes.map((node) => {
            const position = layout.positions.get(node.id)!;
            const wait = (node.fields as { wait?: { reason?: string } } | undefined)?.wait;
            const detail = stateDetail(node);
            const step = stepName(node.fields);
            return (
              <article
                key={node.id}
                className={'graph-card' + (selected === node.id ? ' graph-selected' : '')}
                style={{
                  left: position.x + 'px',
                  top: position.y + 'px',
                  width: card.width + 'px',
                  height: card.height + 'px',
                }}
                data-ticket={node.id}
                data-status={node.status}
              >
                <div className="graph-select hosted">
                  <button
                    className="hit"
                    aria-label={'Inspect ' + node.id}
                    aria-describedby={'graph-status-' + node.id}
                    aria-pressed={selected === node.id}
                    onClick={() => onSelect(node.id)}
                  />
                  <strong className="graph-title" title={parseTitle(node.title).text}>
                    {node.title ? <TitleText title={node.title} /> : node.kind}
                  </strong>
                  <span className="graph-meta">
                    <StatusLetter ticket={node} id={'graph-status-' + node.id} />
                    <span className="mono">{node.id}</span>
                    <span>
                      {node.kind}
                      {node.pool ? ' · ' + node.pool : ''}
                      {step ? ' · ' + step : ''}
                    </span>
                  </span>
                  <span className="graph-meta muted">
                    <time dateTime={node.created_at} title={exactLabel(node.created_at)}>
                      {shortLabel(node.created_at)}
                    </time>
                    {node.closed_at && (
                      <span>
                        →{' '}
                        <time dateTime={node.closed_at} title={exactLabel(node.closed_at)}>
                          {shortLabel(node.closed_at)}
                        </time>
                      </span>
                    )}
                  </span>
                </div>
                {(detail || wait?.reason) && (
                  <div className="graph-state" data-status={node.status}>
                    {detail && <Markdown inline text={detail} />}
                    {wait?.reason && (
                      <div>
                        <Markdown inline text={wait.reason} />
                      </div>
                    )}
                  </div>
                )}
                <div className="graph-links">
                  {(node.holder || node.kind === 'session') && (
                    <button title={node.holder || node.id} onClick={() => onWatch(node.holder || node.id)}>
                      view session ↗
                    </button>
                  )}
                  {safeLinks(node).map((href) => (
                    <a key={href} href={href} target="_blank" rel="noopener noreferrer">
                      {linkLabel(href)}
                    </a>
                  ))}
                </div>
              </article>
            );
          })}
        </div>
        <div
          className={'graph-axis' + (stacked ? ' graph-axis-stacked' : '')}
          aria-label="Creation time axis"
          role="group"
        >
          {ticks.map((t) => (
            <span key={t} className="graph-mark" style={{ left: screenX(t) + 'px' }} />
          ))}
          <span
            className={'graph-mark graph-mark-end' + (end.active ? ' graph-now' : '')}
            style={{ left: endX + 'px' }}
          />
          <span className="graph-mark graph-mark-start" style={{ left: startX + 'px' }} />
          {ticks.map((t) => (
            <span key={'l' + t} style={{ display: 'contents' }}>
              {stamp(t, relativeLabel(t - layout.start), 'tick', screenX(t), screenX(t))}
            </span>
          ))}
          <span
            className={'graph-stamp-box' + (startX < startLeft - 1 ? ' graph-off-left' : '')}
            ref={measure(setStartBox)}
            style={{ left: startLeft + 'px' }}
          >
            {stamp(layout.start, 'start ' + fullLabel(layout.start), 'start', 0, startLeft + startBox / 2)}
          </span>
          <span
            className={
              'graph-stamp-box graph-stamp-box-end' +
              (endX > endLeft + endBox + 1 ? ' graph-off-right' : '') +
              (endX < endLeft - 1 ? ' graph-off-left' : '')
            }
            ref={measure(setEndBox)}
            style={{ left: endLeft + 'px' }}
          >
            {stamp(
              end.at,
              (end.active ? 'now ' : 'ended ') + fullLabel(end.at),
              end.active ? 'now' : 'end',
              0,
              endLeft + endBox / 2,
            )}
          </span>
        </div>
        {tip && (
          <div
            role="tooltip"
            className="graph-tip"
            ref={measure(setTipBox)}
            style={{ left: Math.max(pad, Math.min(width - pad - tipBox, tip.left - tipBox / 2)) + 'px' }}
          >
            {tip.text}
          </div>
        )}
      </div>
    </section>
  );
}
