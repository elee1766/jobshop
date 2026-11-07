import type { Ticket, RunResult } from "../../packages/jobshop-client/rpc.js";

export function ago(value?: string) {
  if (!value || value.startsWith("0001")) return "unknown";
  const minutes = Math.max(
    0,
    Math.floor((Date.now() - Date.parse(value)) / 60000),
  );
  return minutes < 60
    ? minutes + "m"
    : minutes < 1440
      ? Math.floor(minutes / 60) + "h"
      : Math.floor(minutes / 1440) + "d";
}

export interface Flight {
  id: string;
  ref: string;
  repo: string;
  pr: string;
  step: string;
  waiting: string;
  since: string;
  source: string;
  holder: string;
  state: string;
  stalled: boolean;
  diagnosis: string;
  hold: boolean;
  run: RunResult;
}

export function flights(runs: RunResult[], threshold = 3 * 3600000): Flight[] {
  const rows: Flight[] = [];
  for (const run of runs) {
    const f = (run.run.fields as Record<string, any>) || {};
    if (run.run.status === "closed" || ["done", "cancelled"].includes(f.state))
      continue;
    const vars = f.vars || {};
    const step =
      run.steps?.find((s) => s.status === "in_progress") ||
      run.steps?.find((s) => s.status !== "closed");
    const sf = (step?.fields as Record<string, any>) || {};
    const wait = sf.wait;
    const meta = (run.run.meta as Record<string, any>) || {};
    const since =
      meta.github?.head_at || step?.updated_at || run.run.updated_at;
    const source = meta.github?.head_at ? "last push" : "last run activity";
    const rw = run.waits?.find((w) => w.ticket === step?.id);
    let waiting =
      f.state === "needs_human"
        ? "you"
        : rw
          ? rw.key + " slot"
          : wait?.reason || (step?.holder ? "agent" : "queued");
    const observed = wait?.conds
      ?.filter((c: any) => !c.met)
      .map((c: any) => c.note || c.type)
      .join(", ");
    const repeated = (sf.failures || 0) > 1 || (sf.review_rounds || 0) > 1;
    const old = Date.now() - Date.parse(since) > threshold;
    const hold = /external|other.team/i.test(wait?.owner || wait?.reason || "");
    rows.push({
      id: run.run.id,
      ref: vars.issue || meta.linear?.issue || run.run.id,
      repo: vars.repo || meta.github?.repo || "",
      pr: vars.pr || meta.github?.pr || "",
      step: sf.step || "queued",
      waiting,
      since,
      source,
      holder: step?.holder || "",
      state: step?.holder ? "active" : f.state || "queued",
      stalled: old || repeated,
      diagnosis: rw
        ? `${rw.key} ${rw.held}/${rw.capacity}, queued behind ${rw.ahead}`
        : observed ||
          (repeated
            ? `${sf.failures || sf.review_rounds} repeated attempts`
            : waiting),
      hold,
      run,
    });
  }
  return rows;
}

export function refOf(t: Ticket) {
  const f = (t.fields as Record<string, any>) || {};
  const m = (t.meta as Record<string, any>) || {};
  return f.vars?.issue || m.linear?.issue || t.id;
}
