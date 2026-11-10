import type { HistoryEntry } from '../../../packages/jobshop-client/rpc.js';

// Turn rows (JOB-184) are written twice per agent turn on every ticket a
// session holds. They would bury the history list, so they are folded into
// the claim they ran in: the claim's line says how many turns and how long.
export interface ClaimTurns {
  count: number;
  active: number; // ms in finished turns
}

export function foldTurns(entries: HistoryEntry[]) {
  const shown: HistoryEntry[] = [];
  const claims = new Map<number, ClaimTurns>(); // by the claim's seq
  const lastClaim = new Map<string, number>(); // actor -> seq of its newest claim
  const pending = new Map<string, { at: number; claim: number }>(); // actor -> unfinished turn
  const finish = (actor: string, at: number) => {
    const turn = pending.get(actor);
    if (!turn) return;
    pending.delete(actor);
    const claim = claims.get(turn.claim);
    if (claim) claim.active += Math.max(0, at - turn.at);
  };
  for (const entry of entries) {
    const at = Date.parse(entry.at);
    if (entry.op === 'turn_start') {
      finish(entry.actor, at); // the controller stopped mid-turn: cut here
      const claim = lastClaim.get(entry.actor);
      if (claim === undefined) continue; // outside any claim
      const totals = claims.get(claim) || { count: 0, active: 0 };
      totals.count++;
      claims.set(claim, totals);
      pending.set(entry.actor, { at, claim });
      continue;
    }
    if (entry.op === 'turn_end') {
      finish(entry.actor, at);
      continue;
    }
    if (entry.op === 'claim') lastClaim.set(entry.actor, entry.seq);
    shown.push(entry);
  }
  return { entries: shown, claims };
}
