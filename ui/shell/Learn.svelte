<script lang="ts">
  import type { Learning, MemoryEdit } from "../../packages/jobshop-client/rpc.js";
  import { ago } from "./model.js";
  import {
    build,
    matches,
    summarize,
    sections,
    metricRows,
    isFresh,
    type Stage,
  } from "./learn.js";

  let { learnings, edits, now }: {
    learnings: Learning[];
    edits: MemoryEdit[];
    now: number;
  } = $props();

  // Rows shown per section before "more": dense data stays scannable.
  const PAGE = 10;
  let filter = $state("");
  let only: Stage | "" = $state("");
  let toggled: Record<string, boolean> = $state({});
  let shown: Record<string, number> = $state({});

  const units = $derived(build(learnings, edits));
  const summary = $derived(summarize(units, learnings, edits, now));
  const visible = $derived(units.filter((u) => matches(u, filter)));
  const groups = $derived(
    sections
      .map((s) => ({ ...s, items: visible.filter((u) => u.stage === s.stage) }))
      .filter((g) => g.items.length && (!only || g.stage === only)),
  );
  const counts = $derived(
    sections
      .map((s) => ({ ...s, n: visible.filter((u) => u.stage === s.stage).length }))
      .filter((s) => s.n > 0),
  );
  const notHelped = $derived(units.filter((u) => u.stage === "not-helped").length);

  // A search or a chosen chip opens what it found; otherwise recent and
  // unresolved sections are open and history is closed.
  function isOpen(stage: string, dflt: boolean) {
    if (stage in toggled) return toggled[stage];
    return !!filter || !!only || dflt;
  }
  function toggle(stage: string, dflt: boolean) {
    toggled = { ...toggled, [stage]: !isOpen(stage, dflt) };
  }
  function choose(stage: Stage) {
    only = only === stage ? "" : stage;
    toggled = {};
  }
  function when(value?: string) {
    return value && !value.startsWith("0001") ? new Date(value).toLocaleString() : "not recorded";
  }
  // A gate result with no steps in its window still has a window.
  function measured(edit: MemoryEdit) {
    return !!edit.after && !edit.after.from.startsWith("0001");
  }
  function span(m?: { from: string; to: string }) {
    return m && !m.from.startsWith("0001") ? `${when(m.from)} to ${when(m.to)}` : "";
  }
</script>

<div class="tag">
  <span class="mono">/learn</span>
  <input
    aria-label="Filter learn"
    bind:value={filter}
    placeholder="problem, fix, pr, ticket or pool"
  /><button onclick={() => { filter = ""; only = ""; toggled = {}; }}>clear</button>
  <span class="push-right muted">{visible.length}{visible.length !== units.length ? " of " + units.length : ""}</span>
</div>
<div class="section-body learn">
  {#if !units.length}
    <p class="muted">Nothing recorded yet. Learnings appear after a failed or corrected step, and memory edits after the curator's change merges.</p>
  {:else}
    <p class="learn-summary" aria-label="Summary">
      {summary.items} items
      <span class="muted">
        · {summary.learnings} learnings, {summary.edits} memory edits
        · {summary.fresh} new in 7 days
      </span>
      {#if notHelped}<b>· {notHelped} not helped</b>{/if}
    </p>
    <div class="step-strip mono" role="group" aria-label="Stage">
      {#each counts as s}<button
          class:chosen={only === s.stage}
          aria-pressed={only === s.stage}
          title={s.hint}
          onclick={() => choose(s.stage)}>{s.title} {s.n}</button
        >{/each}
    </div>
    {#if !groups.length}<p class="muted">Nothing matches the filter.</p>{/if}
    {#each groups as g (g.stage)}
      {@const open = isOpen(g.stage, g.open)}
      {@const limit = shown[g.stage] || PAGE}
      <section class="learn-group" aria-label={g.title}>
        <h2>
          <button
            class="learn-toggle"
            aria-expanded={open}
            onclick={() => toggle(g.stage, g.open)}
            ><span aria-hidden="true">{open ? "▾" : "▸"}</span> {g.title}
            <span class="muted mono">{g.items.length}</span></button
          >
          <span class="muted learn-hint">{g.hint}</span>
        </h2>
        {#if open}
          <div class="rows">
          {#each g.items.slice(0, limit) as u (u.key)}
            <article class="learn-row">
              <p class="learn-problem">
                {#if u.problem}<b>{u.problem}</b>{:else}<b class="muted">problem not recorded</b>{/if}
                {#if u.more}<span class="muted"> +{u.more} more</span>{/if}
              </p>
              <p class="learn-fix">
                <span class="muted">{u.fixLabel}</span>
                {#if u.fix}{u.fix}{:else}<span class="muted">not recorded</span>{/if}
              </p>
              <p class="learn-meta muted">
                <span class="mono">{u.outcome}</span>
                {#if u.where}·
                  {#if u.where.href}<a href={u.where.href} target="_blank" rel="noopener noreferrer">{u.where.text}</a>{:else}<span class="mono">{u.where.text}</span>{/if}
                {/if}
                · <span title={when(u.at)}>{ago(u.at, now)}</span>
                {#if isFresh(u.at, now)}<span class="mono">· new</span>{/if}
                {#if u.kind}· {u.kind}{/if}
              </p>
              <details>
                <summary>details</summary>
                <dl class="learn-facts mono">
                  {#if u.edit}
                    <dt>memory edit</dt><dd>#{u.edit.id}, merged {when(u.edit.merged_at)}</dd>
                    {#if u.edit.pools?.length}<dt>pools</dt><dd>{u.edit.pools.join(", ")}</dd>{/if}
                    {#if u.edit.evaluated_at && !u.edit.evaluated_at.startsWith("0001")}<dt>evaluated</dt><dd>{when(u.edit.evaluated_at)}</dd>{/if}
                  {/if}
                  {#if u.scope}<dt>scope</dt><dd>{u.scope}</dd>{/if}
                </dl>
                {#each u.learnings as l (l.id)}
                  <div class="learn-cited">
                    <p><span class="muted">problem #{l.id}</span> {l.summary}</p>
                    <p><span class="muted">fix</span> {#if l.fix}{l.fix}{:else}<span class="muted">not recorded</span>{/if}</p>
                    <p class="muted mono">
                      {l.kind} · {l.source} · {l.status}{#if l.gate} · gate {l.gate}{/if}
                      {#if l.ticket}· <a href={"?view=trace&ref=" + encodeURIComponent(l.ticket)}>{l.ticket}</a>{/if}
                      {#if l.session}· {l.session}{/if}
                      · {when(l.created_at)}
                    </p>
                  </div>
                {/each}
                {#if u.missing.length}<p class="muted">Also cites {u.missing.map((id) => "#" + id).join(", ")}, not in the loaded list.</p>{/if}
                {#if u.edit}
                  <div class="table-scroll">
                    <table class="mono rows">
                      <thead><tr><th>measure</th><th>before</th><th>after</th></tr></thead>
                      <tbody>
                        {#each metricRows as row}<tr>
                            <td>{row.label}</td>
                            <td>{u.edit.baseline[row.key]}</td>
                            <td>{measured(u.edit) ? u.edit.after?.[row.key] : "not measured"}</td>
                          </tr>{/each}
                      </tbody>
                    </table>
                  </div>
                  <p class="muted mono">before {span(u.edit.baseline)}{#if measured(u.edit)}; after {span(u.edit.after)}{/if}</p>
                  {#if u.edit.before}<details>
                      <summary>text the edit replaced</summary>
                      <pre>{u.edit.before.length > 8000 ? u.edit.before.slice(0, 8000) + "\n[display limited]" : u.edit.before}</pre>
                    </details>{/if}
                {/if}
              </details>
            </article>
          {/each}
          </div>
          {#if g.items.length > limit}<button class="learn-more" onclick={() => (shown = { ...shown, [g.stage]: limit + PAGE })}
              >more ({g.items.length - limit})</button
            >{/if}
        {/if}
      </section>
    {/each}
  {/if}
</div>
