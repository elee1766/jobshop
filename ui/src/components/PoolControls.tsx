import { useAtomValue } from 'jotai';
import { useState } from 'react';
import type { PoolScale } from '../../../packages/jobshop-client/rpc.js';
import { cordonSession, scalePool } from '../state/actions.ts';
import { pendingAtom, rpcMethodsAtom } from '../state/ui.ts';

// Whether this Desk lets the page call a method (desk.rpc). The controller
// still checks its own grant when the call is made, and its refusal is shown.
export function useAllowed(method: string) {
  return !!useAtomValue(rpcMethodsAtom)?.includes(method);
}

// Desired sessions for one ticket pool: set a number (capped by the pool's
// max_sessions) or return to auto. Renders nothing without the grant.
export function ScaleControl({ pool }: { pool: PoolScale }) {
  const allowed = useAllowed('pool.scale');
  const busy = !!useAtomValue(pendingAtom)['pool:' + pool.name];
  const [draft, setDraft] = useState('');
  if (!allowed) return null;
  const n = Number(draft);
  const valid = draft !== '' && Number.isInteger(n) && n >= 0 && (!pool.max || n <= pool.max);
  return (
    <form
      className="pool-scale"
      onSubmit={(event) => {
        event.preventDefault();
        if (valid) void scalePool(pool.name, n).then((ok) => ok && setDraft(''));
      }}
    >
      <input
        type="number"
        inputMode="numeric"
        min={0}
        max={pool.max || undefined}
        step={1}
        placeholder={String(pool.desired)}
        aria-label={`Desired sessions for ${pool.name}`}
        value={draft}
        disabled={busy}
        onChange={(event) => setDraft(event.target.value)}
      />
      <button type="submit" disabled={busy || !valid}>
        {busy ? 'setting…' : 'set'}
      </button>
      <button
        type="button"
        disabled={busy || pool.target === undefined}
        title={pool.target === undefined ? 'already following ready work' : 'follow ready work again'}
        onClick={() => void scalePool(pool.name, 'auto')}
      >
        auto
      </button>
    </form>
  );
}

// Takes one session out of rotation or puts it back. A cordoned session
// finishes the ticket it holds and claims nothing new. Renders nothing
// without the grant.
export function CordonControl({ id, name, cordon }: { id: string; name: string; cordon?: string }) {
  const allowed = useAllowed('session.cordon');
  const busy = !!useAtomValue(pendingAtom)['session:' + id];
  if (!allowed) return null;
  const label = cordon ? 'uncordon' : 'cordon';
  return (
    <button
      type="button"
      aria-label={`${label} ${name}`}
      disabled={busy}
      onClick={() => void cordonSession(id, name, !!cordon)}
    >
      {busy ? 'working…' : label}
    </button>
  );
}
