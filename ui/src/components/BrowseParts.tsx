import { useAtomValue } from 'jotai';
import type { BrowseStatus } from '../../../packages/jobshop-client/rpc.js';
import { codeLabel, statusNotice } from '../lib/browse.ts';
import { browseAtom, nowAtom } from '../state/live.ts';

// Pieces both Linear and GitHub screens use (JOB-163).

// WritesButton is a write control of the design that is not connected: it
// stays in the layout, can be focused, and says why it does nothing. The
// reason is also in the page as text (see WritesNote), because a tooltip is
// not available on touch screens.
export function WritesButton({ label, className }: { label: string; className?: string }) {
  const writes = useAtomValue(browseAtom)?.writes;
  return (
    <button
      type="button"
      className={className ? `writes-off ${className}` : 'writes-off'}
      aria-disabled="true"
      title={writes?.reason}
      onClick={(e) => e.preventDefault()}
    >
      {label}
    </button>
  );
}

export function WritesNote() {
  const writes = useAtomValue(browseAtom)?.writes;
  if (!writes || writes.enabled) return null;
  return <p className="browse-writes muted">read only. {writes.reason}</p>;
}

// StatusLine says what a read could not do. It is nothing for a fresh or
// cached read.
export function StatusLine({
  status,
  error,
  onRetry,
}: {
  status?: BrowseStatus;
  error?: string;
  onRetry?: () => void;
}) {
  const now = useAtomValue(nowAtom);
  const notice = statusNotice(status, now);
  if (!notice && !error) return null;
  return (
    <output className="browse-status block" data-state={status?.state ?? 'error'} data-code={status?.code ?? ''}>
      <b className="mono">{status?.code ? (codeLabel[status.code] ?? status.code) : 'error'}</b> {error || notice}
      {onRetry && status?.code !== 'not_found' && (
        <>
          {' '}
          <button onClick={onRetry}>retry</button>
        </>
      )}
    </output>
  );
}
