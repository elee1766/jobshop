import type { Ticket, RunResult } from "../../packages/jobshop-client/rpc.js";

export function ago(value?: string, now = Date.now()) {
  if (!value || value.startsWith("0001")) return "unknown";
  const minutes = Math.max(0, Math.floor((now - Date.parse(value)) / 60000));
  return minutes < 60
    ? minutes + "m"
    : minutes < 1440
      ? Math.floor(minutes / 60) + "h"
      : Math.floor(minutes / 1440) + "d";
}

export interface PRLink {
  url: string;
  label: string;
}

const pullURL = /^https:\/\/github\.com\/([^/]+)\/([^/]+)\/pull\/(\d+)$/;

// prLinks are the run's GitHub PR links, labelled repo#n from the link
// itself: a PR may live in another repo than the run's (JOB-131).
export function prLinks(links?: string[]): PRLink[] {
  const out: PRLink[] = [];
  for (const url of links || []) {
    const m = pullURL.exec(url);
    if (m) out.push({ url, label: `${m[2]}#${m[3]}` });
  }
  return out;
}

export interface Flight {
  id: string;
  title: string;
  ref: string;
  repo: string;
  prs: PRLink[];
  issueURL: string;
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

export function flights(
  runs: RunResult[],
  threshold = 3 * 3600000,
  now = Date.now(),
): Flight[] {
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
      meta.github?.head_at ||
      wait?.since ||
      step?.updated_at ||
      run.run.updated_at;
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
    const old = now - Date.parse(since) > threshold;
    const hold =
      /external|other.team/i.test(wait?.owner || wait?.reason || "") ||
      run.run.labels?.some((label) =>
        /^(external-hold|hold:external)$/.test(label),
      ) ||
      false;
    const mismatch = wait?.conds?.some(
      (c: any) =>
        !c.met &&
        c.type === "github.review" &&
        /observed (CHANGES_REQUESTED|DISMISSED)/.test(c.note || ""),
    );
    const repo = vars.repo || meta.github?.repo || "";
    const issue = vars.issue || meta.linear?.issue;
    rows.push({
      id: run.run.id,
      title: run.run.title || "",
      ref: issue
        ? /^\d+$/.test(String(issue)) && repo
          ? `${repo.split("/").at(-1)}#${issue}`
          : String(issue)
        : run.run.id,
      repo,
      prs: prLinks(run.links),
      issueURL: (run.links || []).find(link => /^https:\/\/linear\.app\//.test(link)) || '',
      step: sf.step || "queued",
      waiting,
      since,
      source,
      holder: step?.holder || "",
      state: step?.holder ? "active" : f.state || "queued",
      stalled: f.state !== "needs_human" && (old || repeated || mismatch),
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
