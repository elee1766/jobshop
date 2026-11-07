<script lang="ts">
  import { onMount } from "svelte";
  import Chat, { type Chat as ChatInfo } from "./Chat.svelte";
  import Trace from "./Trace.svelte";
  import HookAudit from "./HookAudit.svelte";
  import TraceSearch from "./TraceSearch.svelte";
  import TicketDetails from "./TicketDetails.svelte";
  import SessionViewer from "./SessionViewer.svelte";
  import { client } from "./lib/api.svelte.js";
  import {
    initPush,
    enablePush,
    disablePush,
    push,
  } from "./lib/push.svelte.js";
  import { ago, flights } from "./model.js";
  import { state as ticketState, statusLetter } from "./trace.js";
  import { md } from "../kit/markdown.js";
  import type {
    AttentionItem,
    RunResult,
    Event,
  } from "../../packages/jobshop-client/rpc.js";

  import type { LiveEvent, AutomationAudit, SystemView } from "../../packages/jobshop-client/index.js";
  let chats: ChatInfo[] = $state([]);
  let view = $state("home");
  let phoneTab = $state("waiting");
  let attention: AttentionItem[] = $state([]);
  let runs: RunResult[] = $state([]);
  let now = $state(Date.now());
  let system: SystemView | null = $state(null);
  let auditRef = $state("");
  let audit: AutomationAudit | { id: string; error: string } | null = $state(null);
  let improvements: any = $state(null);
  let trace: any = $state(null);
  let traceRef = $state("");
  let inspectRef = $state("");
  let selected: any = $state(null);
  let watchedSession = $state("");
  let focused = $state("");
  let note = $state("");
  let busy = $state(false);
  let error = $state("");
  let connected = $state(false);
  let truncated = $state(false);
  let hidden = $state(false);
  let hideHolds = $state(false);
  let chatOpen = $state(false);
  let activeChat: ChatInfo | null = $state(null);
  let requested: ChatInfo | null = $state(null);
  let chat: Chat | undefined = $state();
  let collapsed: Record<string, boolean> = $state({});
  let filter = $state("");
  let sortOldest = $state(true);
  let sinceAway = $state(false);
  let lastFocus = $state("");
  let keyboard = $state(false);
  let updateStream = () => {};
  let restoreChat = "";
  const rows = $derived(flights(runs, undefined, now));
  const stalled = $derived(
    rows.filter((r) => r.stalled && (!hideHolds || !r.hold)),
  );
  const filtered = $derived(
    rows
      .filter(
        (r) =>
          filter
            .toLowerCase()
            .split(/\s+/)
            .filter(Boolean)
            .every((term) =>
              `${r.title} ${r.id} ${r.step} ${r.repo} ${r.ref} ${r.waiting}`
                .toLowerCase()
                .includes(term),
            ) &&
          (!sinceAway ||
            !lastFocus ||
            Date.parse(r.run.run.updated_at) > Date.parse(lastFocus)),
      )
      .sort(
        (a, b) =>
          (sortOldest ? 1 : -1) * (Date.parse(a.since) - Date.parse(b.since)),
      ),
  );
  const current = $derived(
    attention.find((a) => a.id === focused) || attention[0],
  );
  const steps = $derived([...new Set(rows.map((r) => r.step))]);

  function showChat() {
    hidden = false;
    chatOpen = true;
    phoneTab = "chats";
    saveHidden();
  }
  function saveHidden() {
    try {
      localStorage.setItem("desk.chat.hidden", String(hidden));
    } catch {
      /* Storage may be disabled. */
    }
  }
  function navigate(next: string) {
    view = next;
    if (next !== "home") phoneTab = "more";
    const u = new URL(location.href);
    u.searchParams.set("view", next);
    for (const [key, value] of Object.entries({
      ref: traceRef,
      inspect: inspectRef,
      audit: auditRef,
      filter,
      since: sinceAway ? "1" : "",
    })) {
      if (value) u.searchParams.set(key, value);
      else u.searchParams.delete(key);
    }
    if (u.href !== location.href) history.pushState(null, "", u);
    updateStream();
  }
  function traceClick(event: MouseEvent) {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    const anchor = event.target instanceof Element ? event.target.closest("a") : null;
    if (!anchor || anchor.target || anchor.hasAttribute("download")) return;
    const url = new URL(anchor.href);
    if (url.origin !== location.origin || url.pathname !== location.pathname || url.searchParams.get("view") !== "trace") return;
    const ref = url.searchParams.get("ref");
    if (!ref) return;
    event.preventDefault();
    openTrace(ref);
  }
  function receive(event: LiveEvent) {
    switch (event.kind) {
      case "attention":
        attention = (event.data.items || []).filter((item) =>
          [
            "question",
            "decision",
            "ticket",
            "run",
            "wait",
            "guard",
            "pool",
          ].includes(item.kind),
        );
        break;
      case "overview":
        runs = event.data.runs || [];
        truncated = event.data.truncated;
        break;
      case "system":
        system = event.data;
        now = Date.now();
        break;
      case "automation":
        if (event.data.id === auditRef) audit = event.data;
        break;
      case "scheduler":
        now = Date.now();
        if (system && event.data.system)
          system = { ...system, status: event.data.system };
        break;
      case "chats":
        chats = event.data.sessions || [];
        if (restoreChat) {
          requested = chats.find((entry) => entry.id === restoreChat) || null;
          if (requested) restoreChat = "";
        }
        break;
      case "improvements":
        improvements = event.data;
        break;
      case "trace":
        if (view !== "trace") break;
        if ("error" in event.data) {
          error = event.data.error;
          break;
        }
        trace = event.data;
        selected = event.data.inspected || event.data.ticket;
        break;
      case "desk":
        chat?.receiveCommand(event.data);
        break;
    }
  }
  function openAudit(id: string) {
    if (auditRef !== id) audit = null;
    auditRef = id;
    navigate("system");
  }
  function openTrace(ref: string) {
    if (ref !== traceRef || inspectRef) {
      selected = null;
      trace = null;
      watchedSession = "";
    }
    traceRef = ref;
    inspectRef = "";
    chatOpen = false;
    navigate("trace");
  }
  function inspect(id: string) {
    const node = trace?.graph?.nodes.find((node: any) => node.id === id);
    watchedSession = node?.holder || (node?.kind === 'session' ? id : '');
    if (inspectRef === id) return;
    inspectRef = id;
    selected = null;
    navigate("trace");
  }
  async function discuss(item = current) {
    if (!item) return;
    try {
      const result = await client.post("api/attention/chat", {
        kind: item.kind,
        id: item.id,
      });
      requested = {
        id: result.session,
        key: result.key,
        title: item.ref || item.id,
        state: "idle",
      };
      showChat();
    } catch (e) {
      error = String(e);
    }
  }
  async function answer(option: number) {
    if (!current || busy) return;
    const item = current;
    busy = true;
    try {
      const result = await client.post("api/attention/answer", {
        kind: item.kind,
        id: item.id,
        option,
        expected_title: item.options?.find((choice) => choice.id === option)
          ?.title,
        note,
      });
      note = "";
      if (result.session)
        requested = {
          id: result.session,
          key: result.key,
          title: item.ref || item.id,
          state: "idle",
        };
      if (result.session) showChat();
    } catch (e) {
      error = String(e);
    } finally {
      busy = false;
    }
  }
  function triage() {
    showChat();
    chat?.start(
      "Summarize these waiting items, recommend an order, and identify decisions I need to make. Do not change work until I choose an action.",
      attention.map((item) => item.id),
      "triage " + new Date().toISOString().slice(0, 19),
    );
  }
  function activeChanged(value: ChatInfo | null) {
    activeChat = value;
    const url = new URL(location.href);
    if (value) url.searchParams.set("chat", value.id);
    else url.searchParams.delete("chat");
    history.replaceState(null, "", url);
  }
  function attach(refs: string[]) {
    if (!activeChat?.key.startsWith("desk:")) {
      error = "Open a named chat before attaching context.";
      return;
    }
    showChat();
    chat?.attach(refs);
  }
  function command(ev: Event) {
    if (ev.session !== activeChat?.id) return;
    const args = ev.args || {};
    if (ev.op === "navigate") {
      if (args.ref) void openTrace(args.ref);
      else navigate(args.view);
    }
    if (ev.op === "filter")
      filter = [args.repo, args.step, args.waiting_on]
        .filter(Boolean)
        .join(" ");
    if (ev.op === "select" && args.refs) focused = args.refs.split(/[ ,]+/)[0];
    if (ev.op === "expand" || ev.op === "collapse")
      collapsed = {
        ...collapsed,
        [args.window.replace(/^\//, "")]: ev.op === "collapse",
      };
  }
  async function closeSession(id: string) {
    try {
      const dry = await client.call("session.close", { id, dry_run: true });
      if (
        !window.confirm(
          `Close ${id}? Its turn stops and its slot is freed. ${dry.held?.length || 0} held tickets will be released. The transcript remains.`,
        )
      )
        return;
      await client.call("session.close", { id, reason: "closed in Desk" });
    } catch (e) {
      error = String(e);
    }
  }
  function keydown(e: KeyboardEvent) {
    if (
      (e.target as HTMLElement)?.closest(
        "input, textarea, select, [contenteditable]",
      ) ||
      e.ctrlKey ||
      e.metaKey ||
      e.altKey
    )
      return;
    if (view !== "home") return;
    if (
      /^[1-5]$/.test(e.key) &&
      current?.options?.some((o) => o.id === Number(e.key))
    ) {
      e.preventDefault();
      void answer(Number(e.key));
    }
    if (e.key === "d") void discuss();
    if (e.key === "j" || e.key === "k") {
      e.preventDefault();
      const index = attention.findIndex((a) => a.id === current?.id);
      focused =
        attention[
          Math.max(
            0,
            Math.min(attention.length - 1, index + (e.key === "j" ? 1 : -1)),
          )
        ]?.id || "";
    }
  }
  function restoreRoute() {
    const previousTrace = traceRef + ":" + inspectRef;
    const params = new URLSearchParams(location.search);
    const item = params.get("item");
    if (item) focused = item.slice(item.indexOf(":") + 1);
    view = ["home", "trace", "improvements", "system"].includes(params.get("view") || "")
      ? params.get("view")! : "home";
    traceRef = params.get("ref") || "";
    inspectRef = params.get("inspect") || "";
    const nextAudit = params.get("audit") || "";
    if (auditRef !== nextAudit) audit = null;
    auditRef = nextAudit;
    filter = params.get("filter") || "";
    sinceAway = params.get("since") === "1";
    restoreChat = params.get("chat") || "";
    phoneTab = view === "home" ? "waiting" : "more";
    chatOpen = false;
    if (previousTrace !== traceRef + ":" + inspectRef) {
      trace = null;
      selected = null;
      watchedSession = "";
    }
    updateStream();
  }
  onMount(() => {
    try {
      hidden = localStorage.getItem("desk.chat.hidden") === "true";
    } catch {
      /* Storage may be disabled. */
    }
    restoreRoute();
    void client
      .get("api/focus")
      .then((d) => (lastFocus = d.last_focus))
      .catch(() => {});
    void initPush();
    let close = () => {};
    let streamKey: string | null = null;
    const connect = () => {
      const key = view === "trace" ? "trace:" + traceRef + ":" + inspectRef : view === "system" && auditRef ? "audit:" + auditRef : "";
      if (streamKey === key && navigator.onLine) return;
      close();
      streamKey = null;
      if (!navigator.onLine) return;
      streamKey = key;
      close = client.live(receive, {
        trace: view === "trace" ? traceRef : undefined,
        inspect: view === "trace" ? inspectRef : undefined,
        audit: view === "system" ? auditRef : undefined,
        onState: (state, detail) => {
          connected = state === "connected";
          if (state === "error")
            error = detail?.error || "Live updates unavailable";
          if (state === "connected") error = "";
        },
      });
    };
    const offline = () => {
      streamKey = null;
      close();
      connected = false;
    };
    updateStream = connect;
    window.addEventListener("online", connect);
    window.addEventListener("offline", offline);
    if (navigator.onLine) connect();
    const vv = window.visualViewport;
    let fullHeight = vv?.height || window.innerHeight;
    const fit = () => {
      const height = vv?.height || window.innerHeight;
      document.documentElement.style.setProperty("--vvh", height + "px");
      document.documentElement.style.setProperty(
        "--vvtop",
        (vv?.offsetTop || 0) + "px",
      );
      if (!document.activeElement?.matches("textarea,input"))
        fullHeight = Math.max(fullHeight, height);
      keyboard =
        height < fullHeight * 0.78 &&
        !!document.activeElement?.matches("textarea,input");
    };
    const visibility = () => {
      if (document.visibilityState === "hidden")
        void client.post("api/focus", {}).catch(() => {});
    };
    fit();
    vv?.addEventListener("resize", fit);
    vv?.addEventListener("scroll", fit);
    document.addEventListener("visibilitychange", visibility);
    return () => {
      close();
      updateStream = () => {};
      window.removeEventListener("online", connect);
      window.removeEventListener("offline", offline);
      vv?.removeEventListener("resize", fit);
      vv?.removeEventListener("scroll", fit);
      document.removeEventListener("visibilitychange", visibility);
    };
  });
</script>

<svelte:window onkeydown={keydown} onclick={traceClick} onpopstate={restoreRoute} />
<div
  class="desk"
  class:keyboard
  class:chat-hidden={hidden}
  class:chat-open={chatOpen}
  data-phone={phoneTab}
>
  <header class="tag top">
    <b class="mono">desk</b>
    <nav aria-label="Main">
      {#each ["home", "trace", "improvements", "system"] as page}<button
          class:chosen={view === page}
          onclick={() => navigate(page)}
          >{{ home: 'home', trace: 'trace', improvements: 'learn', system: 'sys' }[page]}</button
        >{/each}<span class="divider">|</span><button
        class:chosen={sinceAway}
        onclick={() => {
          sinceAway = !sinceAway;
          navigate("home");
        }}>catch-up</button
      >
    </nav>
    <span class="push-right status mono"
      >{connected ? "sys ok" : "reconnecting"}{#if system}
        · tick {system.status?.tick_seconds?.toFixed(1) || "0"}s{/if}</span
    >
  </header>
  {#if error}<div class="error banner" role="alert">
      {error}<button onclick={() => (error = "")}>dismiss</button>
    </div>{/if}
  <div class="columns">
    <main class="work">
      {#if view === "home"}
        {#if sinceAway}<div class="tag">
            Since {lastFocus && !lastFocus.startsWith("0001")
              ? new Date(lastFocus).toLocaleString()
              : "first visit"}<button onclick={() => (sinceAway = false)}
              >clear</button
            >
          </div>{/if}
        <section class="waiting" aria-label="Waiting">
          <div class="tag">
            <button
              aria-label="Toggle waiting"
              onclick={() =>
                (collapsed = { ...collapsed, waiting: !collapsed.waiting })}
              >{collapsed.waiting ? "▸" : "▾"}</button
            ><span class="mono">/waiting</span><button
              onclick={() => document.getElementById("answer-note")?.focus()}
              >answer</button
            ><button onclick={() => discuss()}>discuss</button><button
              class="wide"
              onclick={triage}>triage-in-chat</button
            ><span class="push-right muted">{attention.length}</span>
          </div>
          {#if !collapsed.waiting}<div class="section-body">
              {#if !attention.length}<p class="muted">
                  Nothing waiting on you.
                </p>{/if}
              {#each attention as item (item.kind + ":" + item.id)}
                <article class:focused={current?.id === item.id}>
                  {#if item.url && /^https?:\/\//i.test(item.url)}<a href={item.url} target="_blank" rel="noopener noreferrer"><b>{item.ref || item.id}</b></a>{:else}<b>{item.ref || item.id}</b>{/if}
                  {#if item.kind !== "pool"}<a
                      href={"?view=trace&ref=" + encodeURIComponent(item.id)}
                      aria-label={"Trace " + (item.ref || item.id)}
                      >trace</a
                    >{/if}
                  <button class="item-title" onclick={() => (focused = item.id)}
                    >
                    {item.title}
                    <span class="muted">{ago(item.since, now)}</span></button
                  >
                  {#if current?.id === item.id}<div
                      class="attention-context markdown"
                    >
                      {@html md(item.context || item.summary || "")}
                    </div>
                    <div class="options">
                      {#each item.options || [] as option}<button
                          class:recommended={option.recommended}
                          disabled={busy}
                          onclick={() => answer(option.id)}
                          >{option.id}
                          {option.title}{option.recommended
                            ? " (recommended)"
                            : ""}{option.description
                            ? ": " + option.description
                            : ""}</button
                        >{/each}
                    </div>
                    <details class="answer-note">
                      <summary>add answer note</summary><label for="answer-note"
                        >note sent with your choice</label
                      ><textarea id="answer-note" bind:value={note} rows="2"
                      ></textarea>
                    </details>
                  {:else}<span class="compact-options muted"
                      >{(item.options || [])
                        .map((o) => o.title)
                        .join(" · ")}</span
                    >{/if}
                </article>
              {/each}
            </div>{/if}
        </section>
        <section class="not-moving" aria-label="Not moving">
          <div class="tag">
            <button
              aria-label="Toggle not moving"
              onclick={() =>
                (collapsed = {
                  ...collapsed,
                  "not-moving": !collapsed["not-moving"],
                })}>{collapsed["not-moving"] ? "▸" : "▾"}</button
            ><span class="mono">/not-moving</span><button
              onclick={() => attach(stalled.map((r) => r.ref))}>diagnose</button
            ><button onclick={() => (hideHolds = !hideHolds)}
              >{hideHolds ? "show-holds" : "hide-holds"}</button
            ><span class="push-right muted">{stalled.length}</span>
          </div>
          {#if !collapsed["not-moving"]}<div class="section-body">
              {#each stalled as row}<p class:muted={row.hold}>
                  <button class="mono" onclick={() => openTrace(row.id)}
                    >{row.ref}</button
                  >
                  {ago(row.since, now)}
                  {@html md(row.diagnosis, true)}{row.hold ? " (external hold?)" : ""}
                </p>{:else}<p class="muted">
                  No stalled work. Threshold: 3h since last push or run
                  activity.
                </p>{/each}
            </div>{/if}
        </section>
        <section class="inflight" aria-label="In flight">
          <div class="tag">
            <button
              aria-label="Toggle inflight"
              onclick={() =>
                (collapsed = { ...collapsed, inflight: !collapsed.inflight })}
              >{collapsed.inflight ? "▸" : "▾"}</button
            ><span class="mono">/inflight</span><button
              onclick={() => (sortOldest = !sortOldest)}>sort</button
            ><button
              onclick={() => {
                if (filtered[0]) void openTrace(filtered[0].id);
              }}>trace</button
            ><input
              aria-label="Filter inflight"
              bind:value={filter}
              placeholder="step, repo or ref"
            /><button onclick={() => (filter = "")}>clear</button><span
              class="push-right muted">{filtered.length}</span
            >
          </div>
          {#if !collapsed.inflight}<div class="section-body">
              <div class="step-strip mono">
                {#each steps as step}<button
                    class:chosen={filter === step}
                    onclick={() => (filter = filter === step ? "" : step)}
                    >{step} {rows.filter((r) => r.step === step).length}</button
                  >{/each}
              </div>
              {#if truncated}<p class="warn">
                  Showing the first 100 runs. Narrow the controller query for
                  more.
                </p>{/if}
              <div class="table-scroll">
                <table class="mono">
                  <thead
                    ><tr
                      ><th title="Status: O open, R running, C closed">s</th><th>ticket</th><th class="wide">pr</th><th>step</th><th
                        >waiting</th
                      ><th
                        ><button onclick={() => (sortOldest = !sortOldest)}
                          >age</button
                        ></th
                      ><th class="wide">session</th></tr
                    ></thead
                  ><tbody
                    >{#each filtered as row (row.id)}<tr
                        class:stalled={row.stalled}
                        ><td class="ticket-status" title={ticketState(row.run.run)} aria-label={ticketState(row.run.run)}>{statusLetter(row.run.run)}</td><td
                          ><button class="ticket-label" title={row.ref === row.id ? row.id : `${row.id} · ${row.ref}`}
                            onclick={(e) =>
                              e.shiftKey
                                ? attach([row.ref])
                                : openTrace(row.id)}><span>{row.title || row.ref}</span></button
                          >{#if row.issueURL}<a class="ticket-issue" href={row.issueURL} target="_blank" rel="noopener noreferrer">{row.ref}</a>{/if}{#if row.prURL}<a class="ticket-issue compact-pr" href={row.prURL} target="_blank" rel="noopener noreferrer">{row.repo.split("/").at(-1)}#{row.pr || row.prURL.split('/').at(-1)}</a>{/if}</td
                        ><td class="wide"
                          >{#if row.prURL}<a href={row.prURL} target="_blank" rel="noopener noreferrer">{row.repo.split("/").at(-1)}#{row.pr || row.prURL.split('/').at(-1)}</a>{:else}{row.pr ? row.repo.split("/").at(-1) + "#" + row.pr : "–"}{/if}</td
                        ><td>{row.step}</td><td>{@html md(row.waiting, true)}</td><td
                          title={row.source}>{ago(row.since, now)}</td
                        ><td class="wide"
                          ><button
                            disabled={!row.holder}
                            onclick={() => { openTrace(row.id); watchedSession = row.holder; }}
                            >{row.state}</button
                          ></td
                        ></tr
                      >{/each}</tbody
                  >
                </table>
              </div>
            </div>{/if}
        </section>
      {:else if view === "trace"}
        <TraceSearch reference={traceRef} onopen={openTrace} />
        {#if trace?.graph}<div class="trace-layout" class:watching={!!watchedSession}>
            <Trace graph={trace.graph} selected={selected?.id || trace.graph.root} onselect={inspect} onwatch={(id) => watchedSession = id} />
            <div class="trace-bottom" class:watching={!!watchedSession}>
            {#if selected}<TicketDetails
              detail={trace.inspection || trace}
              graph={trace.graph}
              onselect={inspect}
              ontrace={openTrace}
              onwatch={(id) => watchedSession = id}
            />{/if}
            {#if watchedSession}{#key watchedSession}<SessionViewer id={watchedSession} onclose={() => watchedSession = ''} />{/key}{/if}
            </div>
          </div>{/if}
        {#if selected?.meta?.automation?.invocation}
          <p class="section-body">automation invocation <button class="mono" onclick={() => openAudit(selected.meta.automation.invocation)}>{selected.meta.automation.invocation}</button></p>
        {/if}
      {:else if view === "improvements"}
        <div class="tag">
          <span class="mono">/improvements</span>
        </div>
        <div class="section-body">
          <h2>memory changes</h2>
          {#each improvements?.edits || [] as edit}<article>
              <b>{edit.pr || edit.path || "Edit " + edit.id}</b> · {edit.state}
              <span class="muted">{ago(edit.created_at, now)}</span>
              <p>{@html md(edit.reason || "Gate has not evaluated this change yet.", true)}</p>
              <details>
                <summary>before / after measurements</summary>
                <pre>{JSON.stringify(
                    { before: edit.baseline, after: edit.after },
                    null,
                    2,
                  )}</pre>
              </details>
            </article>{:else}<p class="muted">
              No memory changes recorded.
            </p>{/each}
          <h2>proposals</h2>
          {#each improvements?.learnings || [] as learning}<article>
              <b>{learning.summary}</b> · {learning.status}
              <p>{learning.fix || ""}</p>
              <p class="muted">
                {learning.pr || learning.ticket}
                {learning.gate || ""}
              </p>
            </article>{/each}
        </div>
      {:else if view === "system"}
        <div class="tag">
          <span class="mono">/system</span><button
            onclick={() => (push.state === "on" ? disablePush() : enablePush())}
            >push: {push.state}</button
          ><button
            onclick={() =>
              (document.documentElement.dataset.theme =
                document.documentElement.dataset.theme === "dark"
                  ? "light"
                  : "dark")}>theme</button
          >
        </div>
        <div class="section-body">
          <HookAudit summaries={system?.automation} selected={auditRef} detail={audit} onselect={openAudit} />
          <h2>pools</h2>
          <div class="table-scroll">
            <table class="mono">
              <thead
                ><tr
                  ><th>pool</th><th>status</th><th>open / max</th><th
                    >working</th
                  ><th>idle</th><th>desired</th></tr
                ></thead
              ><tbody
                >{#each system?.pools?.pools || [] as pool}<tr
                    ><td>{pool.name}</td><td
                      >{system?.pools.quarantined?.[pool.name]
                        ? "quarantined"
                        : "ok"}</td
                    ><td>{pool.open} / {pool.max || "uncapped"}</td><td
                      >{pool.working}</td
                    ><td>{pool.idle}</td><td>{pool.desired}</td></tr
                  >{/each}</tbody
              >
            </table>
          </div>
          <h2>resources</h2>
          <div class="table-scroll">
            <table class="mono">
              <thead
                ><tr
                  ><th>resource</th><th>held / capacity</th><th>waiting</th></tr
                ></thead
              >
              <tbody
                >{#each system?.resources?.use || [] as resource}<tr
                    ><td>{resource.key}</td><td
                      >{resource.held} / {resource.capacity || "uncapped"}</td
                    ><td>{resource.waiting}</td></tr
                  >{:else}<tr
                    ><td colspan="3" class="muted">No resource use.</td></tr
                  >{/each}</tbody
              >
            </table>
            <table class="mono">
              <thead
                ><tr
                  ><th>slot</th><th>resource</th><th>state</th><th>ticket</th
                  ></tr
                ></thead
              >
              <tbody
                >{#each system?.resources?.slots || [] as slot}<tr
                    ><td>{slot.id}</td><td>{slot.pool}</td><td>{slot.state}</td
                    ><td
                      ><button
                        disabled={!slot.ticket}
                        onclick={() => slot.ticket && openTrace(slot.ticket)}
                        >{slot.ticket || "–"}</button
                      ></td
                    ></tr
                  >{/each}</tbody
              >
            </table>
          </div>
          <h2>ingest</h2>
          <p>
            {system?.ingest?.mode || "controller"} · {Object.entries(
              system?.ingest?.counts || {},
            )
              .map(([state, count]) => `${state}: ${count}`)
              .join(" · ") || "Inbox empty"}
          </p>
          <div class="table-scroll">
            <table class="mono">
              <thead
                ><tr
                  ><th>worker</th><th>host</th><th>last seen</th><th
                    >admitted</th
                  ></tr
                ></thead
              ><tbody
                >{#each system?.ingest?.workers || [] as worker}<tr
                    ><td>{worker.id}</td><td>{worker.host}</td><td
                      >{ago(worker.seen, now)}</td
                    ><td>{worker.admitted}</td></tr
                  >{/each}</tbody
              >
            </table>
          </div>
          <h2>scheduler</h2>
          <p>
            Last {system?.status?.tick_seconds?.toFixed(3) || "0"}s · p95 {system?.status?.p95_seconds?.toFixed(
              3,
            ) || "0"}s ({system?.status?.samples || 0} passes)
          </p>
          <table class="mono">
            <tbody
              >{#each Object.entries(system?.status?.steps || {}) as [name, seconds]}<tr
                  ><td>{name}</td><td>{Number(seconds).toFixed(3)}s</td></tr
                >{/each}</tbody
            >
          </table>
          <h2>sessions</h2>
          {#each system?.sessions || [] as session}<p>
              <button class="mono" onclick={() => openTrace(session.id)}
                >{session.display_name || session.id}</button
              >
              {session.busy ? "working" : "idle"}
              <button onclick={() => closeSession(session.id)}>close</button>
            </p>{/each}
          <p class="muted">
            Costs are not recorded by this controller. Context usage is
            available in session journals.
          </p>
        </div>
      {/if}
    </main>
    <aside class="chat-column" aria-label="Chats">
      <button class="show-chats" onclick={showChat}>chats ▸</button>
      <div class="chat-content">
        <Chat
          bind:this={chat}
          {requested}
          {chats}
          onactive={activeChanged}
          oncommand={command}
          onhide={() => {
            hidden = true;
            chatOpen = false;
            saveHidden();
          }}
        />
      </div>
    </aside>
  </div>
  <nav class="phone-nav" aria-label="Mobile">
    {#each ["waiting", "chats", "inflight", "more"] as tab}<button
        class:chosen={phoneTab === tab}
        onclick={() => {
          phoneTab = tab;
          if (tab === "chats") showChat();
          else if (tab !== "more") view = "home";
        }}
        >{tab === "inflight"
          ? "in flight"
          : tab}{tab === "waiting"
          ? " " + attention.length
          : ""}</button
      >{/each}
  </nav>
</div>
