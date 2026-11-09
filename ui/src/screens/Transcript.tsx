import { memo, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import type { SessionEvent } from '../../../packages/jobshop-client/rpc.js';
import { md } from '../../kit/markdown.js';
import { client } from '../lib/api.ts';
import { reconcile } from '../lib/reconcile.ts';
import type { PendingMessage } from '../lib/session.ts';

type Message = SessionEvent & { key: string; pending?: PendingMessage };

const shown = ['turn_start', 'text', 'tool', 'image'];

// Rows come from frames: streamed text frames join into one row, a tool's
// updates replace its row, and a row keeps its key when an older page
// completes it, so a streamed row never remounts.
function build(frames: SessionEvent[], pending: PendingMessage[], rowKeys: Map<number, string>): Message[] {
  const out: Message[] = [];
  const tools = new Map<string, Message>();
  const delivered = new Set(frames.flatMap((frame) => (frame.messages || []).map((message) => message.id)));
  let text: Message | undefined;
  for (const frame of frames) {
    if (frame.type === 'text' && text) {
      text.text = (text.text || '') + (frame.text || '');
      text.key = rowKeys.get(frame.seq!) || text.key;
      rowKeys.set(frame.seq!, text.key);
      continue;
    }
    if (frame.type === 'turn_start' || frame.type === 'turn_end') tools.clear();
    if (['turn_start', 'turn_end', 'tool', 'image'].includes(frame.type) || frame.error) text = undefined;
    if (frame.type === 'tool' && frame.tool) {
      const previous = tools.get(frame.tool.id);
      if (previous) {
        previous.tool = { ...previous.tool!, ...frame.tool };
        previous.key = rowKeys.get(frame.seq!) || previous.key;
        rowKeys.set(frame.seq!, previous.key);
        if (frame.images?.length) previous.images = frame.images;
        continue;
      }
    }
    if (shown.includes(frame.type) || frame.error) {
      const message: Message = { ...frame, key: rowKeys.get(frame.seq!) || 'frame-' + frame.seq };
      rowKeys.set(frame.seq!, message.key);
      out.push(message);
      if (frame.tool) tools.set(frame.tool.id, message);
      if (frame.type === 'text') text = message;
    }
  }
  for (const message of pending) {
    if (!delivered.has(message.id)) out.push({ session: '', type: 'pending', key: message.id, pending: message });
  }
  return out;
}

const dataURL = (image: { mime_type: string; data: string }) => 'data:' + image.mime_type + ';base64,' + image.data;

// A virtualized transcript with variable-height rows, anchored to the end:
// it follows new rows while the reader is at the bottom and keeps the place
// when older history loads above (TanStack Virtual's anchorTo and
// followOnAppend). Frames are owned by the caller; older pages go through
// onOlder, which prepends them.
export function Transcript({
  frames,
  viewport,
  pending = [],
  onOlder,
}: {
  frames: SessionEvent[];
  viewport: HTMLDivElement | null;
  pending?: PendingMessage[];
  onOlder: (older: SessionEvent[]) => void;
}) {
  const rowKeys = useRef(new Map<number, string>());
  const previous = useRef<Message[]>([]);
  const messages = useMemo(() => {
    const next = reconcile(previous.current, build(frames, pending, rowKeys.current));
    previous.current = next;
    return next;
  }, [frames, pending]);

  const [expanded, setExpanded] = useState<ReadonlySet<string>>(new Set());
  const remember = (key: string, open: boolean) =>
    setExpanded((set) => {
      if (set.has(key) === open) return set;
      const next = new Set(set);
      if (open) next.add(key);
      else next.delete(key);
      return next;
    });

  const virtualizer = useVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: messages.length,
    getScrollElement: () => viewport,
    getItemKey: (index) => messages[index].key,
    estimateSize: () => 90,
    paddingStart: 28,
    anchorTo: 'end',
    followOnAppend: true,
    scrollEndThreshold: 80,
    overscan: 6,
    useFlushSync: false,
  });

  // The first rows start at the end of the conversation.
  const [started, setStarted] = useState(false);
  useLayoutEffect(() => {
    if (!viewport || !messages.length || started) return;
    virtualizer.scrollToEnd();
    setStarted(true);
  }, [viewport, messages.length, started, virtualizer]);

  // History: load older pages while the reader is near the top.
  const [before, setBefore] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);
  const [historyError, setHistoryError] = useState('');
  const alive = useRef(true);
  const abort = useRef(new AbortController());
  useEffect(() => {
    alive.current = true;
    abort.current = new AbortController();
    return () => {
      alive.current = false;
      abort.current.abort();
    };
  }, []);
  useEffect(() => {
    if (before !== null) return;
    const first = frames.find((frame) => frame.first)?.first ?? frames.find((frame) => frame.seq)?.seq;
    if (first) setBefore(first);
  }, [frames, before]);

  const offset = virtualizer.scrollOffset || 0;
  const total = virtualizer.getTotalSize();
  const nearStart = offset < 160 || (!!viewport && total <= viewport.clientHeight);

  async function older() {
    if (before === null || before <= 1 || loading) return;
    setLoading(true);
    setHistoryError('');
    let cursor = before;
    try {
      let entries: SessionEvent[] = [];
      // A page of protocol bookkeeping is not the beginning of the journal.
      do {
        const data = await client.get(
          `api/session/${encodeURIComponent(frames[0].session)}/events?before=${cursor}&limit=200`,
          {
            signal: abort.current.signal,
          },
        );
        if (!alive.current) return;
        if (!Number.isInteger(data.first) || data.first < 1 || data.first >= cursor)
          throw new Error('History cursor did not advance.');
        cursor = data.first;
        entries = [...(data.events || []), ...entries];
      } while (cursor > 1 && !entries.some((frame) => shown.includes(frame.type) || frame.error));
      onOlder(entries);
      setBefore(cursor);
    } catch (cause) {
      if (alive.current) setHistoryError(String(cause));
    } finally {
      if (alive.current) setLoading(false);
    }
  }
  const olderRef = useRef(older);
  olderRef.current = older;
  useEffect(() => {
    if (
      (started || !messages.length) &&
      nearStart &&
      before !== null &&
      before > 1 &&
      !loading &&
      !historyError &&
      frames.length
    ) {
      void olderRef.current();
    }
  }, [started, messages.length, nearStart, before, loading, historyError, frames.length]);

  return (
    <div className="virtual-transcript" style={{ height: total + 'px' }}>
      <div className="history-loader">
        {historyError ? (
          <>
            <button onClick={() => void older()}>retry history</button>
            <span className="error" role="alert" title={historyError}>
              {historyError}
            </span>
          </>
        ) : loading ? (
          <output>loading history…</output>
        ) : before !== null && before > 1 ? (
          <button onClick={() => void older()}>load older</button>
        ) : before === 1 ? (
          <span className="muted">start of conversation</span>
        ) : null}
      </div>
      {virtualizer.getVirtualItems().map((row) => {
        const message = messages[row.index];
        if (!message) return null;
        const afterTool = !!messages[row.index - 1]?.tool;
        return (
          <div
            key={row.key}
            ref={virtualizer.measureElement}
            className={
              'transcript-row' +
              (afterTool && message.tool ? ' tool-next' : '') +
              (afterTool && !message.tool ? ' after-tools' : '')
            }
            data-index={row.index}
            data-row-key={message.key}
            style={{ transform: 'translateY(' + row.start + 'px)' }}
          >
            <Row message={message} expanded={expanded} remember={remember} />
          </div>
        );
      })}
    </div>
  );
}

const Row = memo(function Row({
  message,
  expanded,
  remember,
}: {
  message: Message;
  expanded: ReadonlySet<string>;
  remember: (key: string, open: boolean) => void;
}) {
  const details = (suffix: string) => ({
    open: expanded.has(message.key + suffix),
    onToggle: (event: React.SyntheticEvent<HTMLDetailsElement>) =>
      remember(message.key + suffix, event.currentTarget.open),
  });
  const images = (alt: string) =>
    (message.images || []).map((image, i) => (
      <img key={i} className="transcript-image" src={dataURL(image)} alt={alt} loading="lazy" />
    ));
  if (message.pending) {
    const p = message.pending;
    return (
      <div className="operator" data-message={p.id}>
        {p.text}
        {p.attachments.length > 0 && <small>{p.attachments.join(', ')}</small>}
        <output className="message-delivery muted">{p.state === 'sending' ? 'sending…' : 'queued'}</output>
      </div>
    );
  }
  if (message.type === 'turn_start') {
    return (
      <>
        {message.context && (
          <details className="controller-context" {...details(':context')}>
            <summary>controller context</summary>
            <div className="markdown" dangerouslySetInnerHTML={{ __html: md(message.context) }} />
          </details>
        )}
        {message.raw_prompt ? (
          <>
            {(message.messages || []).map((original) => (
              <div className="operator" key={original.id} data-message={original.id} title={'from ' + original.from}>
                {original.text}
              </div>
            ))}
            <details className="raw-prompt" {...details(':raw')}>
              <summary>raw model prompt</summary>
              <pre>{message.raw_prompt}</pre>
            </details>
          </>
        ) : message.text ? (
          <details className="controller-context legacy-prompt" {...details(':legacy')}>
            <summary>legacy prompt</summary>
            <pre>{message.text}</pre>
          </details>
        ) : null}
        {images('User attachment')}
      </>
    );
  }
  if (message.tool) {
    return (
      <details className="tool mono" {...details(':tool')}>
        <summary>
          {message.tool.intent || message.tool.title || 'tool'} · {message.tool.status}
        </summary>
        <pre>{[message.tool.target, message.tool.input, message.tool.output].filter(Boolean).join('\n')}</pre>
        {images('Tool output')}
      </details>
    );
  }
  return (
    <div className={'markdown' + (message.error ? ' error' : '')}>
      <span
        style={{ display: 'contents' }}
        dangerouslySetInnerHTML={{ __html: md(message.text || message.error || '') }}
      />
      {images('Agent output')}
    </div>
  );
});
