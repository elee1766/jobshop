<script lang="ts">
  import { onMount } from "svelte";
  import Chat, { type Chat as ChatInfo } from "./Chat.svelte";
  import Trace from "./Trace.svelte";
  import { client } from "./lib/api.svelte.js";
  import {
    initPush,
    enablePush,
    disablePush,
    push,
  } from "./lib/push.svelte.js";
  import { ago, flights } from "./model.js";
  import { md } from "../kit/markdown.js";
  import type {
    AttentionItem,
    RunResult,
    Event,
  } from "../../packages/jobshop-client/rpc.js";

  import type { LiveEvent } from "../../packages/jobshop-client/index.js";
  let chats: ChatInfo[] = $state([]);
  let view = $state("home");
  let phoneTab = $state("waiting");
  let attention: AttentionItem[] = $state([]);
  let runs: RunResult[] = $state([]);
  let system: any = $state(null);
  let improvements: any = $state(null);
  let trace: any = $state(null);
  let traceRef = $state("");
  let selected: any = $state(null);
  let journal: any[] = $state([]);
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
  const rows = $derived(flights(runs));
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
              `${r.step} ${r.repo} ${r.ref} ${r.waiting}`
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
    history.replaceState(null, "", u);
    updateStream();
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
        break;
      case "scheduler":
        if (system && event.data.system)
          system = { ...system, status: event.data.system };
        break;
      case "chats":
        chats = event.data.sessions || [];
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
        if (!selected || selected.id === event.data.ticket.id)
          selected = event.data.ticket;
        break;
      case "desk":
        chat?.receiveCommand(event.data);
        break;
    }
  }
  function openTrace(ref: string) {
    traceRef = ref;
    selected = null;
    trace = null;
    journal = [];
    navigate("trace");
  }
  async function inspect(id: string) {
    journal = [];
    try {
      const result = await client.call("ticket.show", { id });
      selected = result.ticket;
    } catch (e) {
      error = String(e);
    }
  }
  async function loadJournal(id: string) {
    try {
      const result = await client.call("session.journal", { id, limit: 100 });
      journal = result.frames || [];
    } catch (e) {
      error = String(e);
    }
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
    busy = true;
    try {
      const result = await client.post("api/attention/answer", {
        kind: current.kind,
        id: current.id,
        option,
        expected_title: current.options?.find((choice) => choice.id === option)
          ?.title,
        note,
      });
      note = "";
      if (result.session)
        requested = {
          id: result.session,
          key: result.key,
          title: current.ref || current.id,
          state: "idle",
        };
    } catch (e) {
      error = String(e);
    } finally {
      busy = false;
    }
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
  onMount(() => {
    try {
      hidden = localStorage.getItem("desk.chat.hidden") === "true";
    } catch {
      /* Storage may be disabled. */
    }
    const params = new URLSearchParams(location.search);
    const item = params.get("item");
    if (item) focused = item.slice(item.indexOf(":") + 1);
    if (
      ["home", "trace", "improvements", "system"].includes(
        params.get("view") || "",
      )
    )
      view = params.get("view")!;
    void client
      .get("api/focus")
      .then((d) => (lastFocus = d.last_focus))
      .catch(() => {});
    void initPush();
    let close = () => {};
    const connect = () => {
      close();
      if (!navigator.onLine) return;
      close = client.live(receive, {
        trace: view === "trace" ? traceRef : undefined,
        onState: (state, detail) => {
          connected = state === "connected";
          if (state === "error")
            error = detail?.error || "Live updates unavailable";
          if (state === "connected") error = "";
        },
      });
    };
    const offline = () => {
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

<svelte:window onkeydown={keydown} />
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
          >{page[0].toUpperCase() + page.slice(1)}</button
        >{/each}<span class="divider">|</span><button
        class:chosen={sinceAway}
        onclick={() => {
          sinceAway = !sinceAway;
          navigate("home");
        }}>Since-away</button
      >
    </nav>
    <span class="push-right status mono"
      >{connected ? "sys ok" : "reconnecting"}{#if system}
        · tick {system.status?.tick_seconds?.toFixed(1) || "0"}s{/if}</span
    >
  </header>
  {#if error}<div class="error banner" role="alert">
      {error}<button onclick={() => (error = "")}>Dismiss</button>
    </div>{/if}
  <div class="columns">
    <main class="work">
      {#if view === "home"}
        {#if sinceAway}<div class="tag">
            Since {lastFocus && !lastFocus.startsWith("0001")
              ? new Date(lastFocus).toLocaleString()
              : "first visit"}<button onclick={() => (sinceAway = false)}
              >Clear</button
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
              >Answer</button
            ><button onclick={() => discuss()}>Discuss</button><button
              class="wide"
              onclick={() => attach(attention.map((a) => a.ref || a.id))}
              >Triage-in-chat</button
            ><span class="push-right muted">{attention.length}</span>
          </div>
          {#if !collapsed.waiting}<div class="section-body">
              {#if !attention.length}<p class="muted">
                  Nothing waiting on you.
                </p>{/if}
              {#each attention as item (item.kind + ":" + item.id)}
                <article class:focused={current?.id === item.id}>
                  <button class="item-title" onclick={() => (focused = item.id)}
                    ><b>{item.ref || item.id}</b>
                    {item.title}
                    <span class="muted">{ago(item.since)}</span></button
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
                      <summary>Add answer note</summary><label for="answer-note"
                        >Note sent with your choice</label
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
              onclick={() => attach(stalled.map((r) => r.ref))}>Diagnose</button
            ><button onclick={() => (hideHolds = !hideHolds)}
              >{hideHolds ? "Show-holds" : "Hide-holds"}</button
            ><span class="push-right muted">{stalled.length}</span>
          </div>
          {#if !collapsed["not-moving"]}<div class="section-body">
              {#each stalled as row}<p class:muted={row.hold}>
                  <button class="mono" onclick={() => openTrace(row.id)}
                    >{row.ref}</button
                  >
                  {ago(row.since)}
                  {row.diagnosis}{row.hold ? " (external hold?)" : ""}
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
              onclick={() => (sortOldest = !sortOldest)}>Sort</button
            ><button
              onclick={() => {
                if (filtered[0]) void openTrace(filtered[0].id);
              }}>Trace</button
            ><input
              aria-label="Filter inflight"
              bind:value={filter}
              placeholder="step, repo or ref"
            /><button onclick={() => (filter = "")}>Clear</button><span
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
                      ><th>ticket</th><th class="wide">pr</th><th>step</th><th
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
                        ><td
                          ><button
                            onclick={(e) =>
                              e.shiftKey
                                ? attach([row.ref])
                                : openTrace(row.id)}>{row.ref}</button
                          ></td
                        ><td class="wide"
                          >{row.pr
                            ? row.repo.split("/").at(-1) + "#" + row.pr
                            : "–"}</td
                        ><td>{row.step}</td><td>{row.waiting}</td><td
                          title={row.source}>{ago(row.since)}</td
                        ><td class="wide"
                          ><button
                            disabled={!row.holder}
                            onclick={() => openTrace(row.holder)}
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
        <div class="tag">
          <span class="mono">/trace</span>
          <form
            onsubmit={(e) => {
              e.preventDefault();
              void openTrace(traceRef);
            }}
          >
            <input
              aria-label="Trace reference"
              bind:value={traceRef}
              placeholder="ticket, session, repo#PR"
            /><button type="submit">Open</button>
          </form>
        </div>
        {#if trace}<div class="trace-layout">
            <div class="trace-tree">
              {#if trace.trace}<Trace
                  node={trace.trace}
                  onselect={inspect}
                />{/if}
            </div>
            <div class="inspector">
              {#if selected}<h2 class="mono">{selected.id} {selected.title}</h2>
                <p>{selected.status} {selected.resolution || ""}</p>
                <p class="muted">
                  Created {new Date(selected.created_at).toLocaleString()}
                </p>
                <div class="markdown">{@html md(selected.body || "")}</div>
                {#if selected.kind === "session"}<button
                    onclick={() => loadJournal(selected.id)}
                    >Load journal</button
                  >{/if}
                <details open>
                  <summary>Fields and observed conditions</summary>
                  <pre>{JSON.stringify(selected.fields, null, 2)}</pre>
                </details>
                {#each journal as frame}<pre>{JSON.stringify(
                      frame,
                      null,
                      2,
                    )}</pre>{/each}{/if}
            </div>
          </div>{/if}
      {:else if view === "improvements"}
        <div class="tag">
          <span class="mono">/improvements</span>
        </div>
        <div class="section-body">
          <h2>Memory changes</h2>
          {#each improvements?.edits || [] as edit}<article>
              <b>{edit.pr || edit.path || "Edit " + edit.id}</b> · {edit.state}
              <span class="muted">{ago(edit.created_at)}</span>
              <p>{edit.reason || "Gate has not evaluated this change yet."}</p>
              <details>
                <summary>Before / after measurements</summary>
                <pre>{JSON.stringify(
                    { before: edit.baseline, after: edit.after },
                    null,
                    2,
                  )}</pre>
              </details>
            </article>{:else}<p class="muted">
              No memory changes recorded.
            </p>{/each}
          <h2>Proposals</h2>
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
            >Push: {push.state}</button
          ><button
            onclick={() =>
              (document.documentElement.dataset.theme =
                document.documentElement.dataset.theme === "dark"
                  ? "light"
                  : "dark")}>Theme</button
          >
        </div>
        <div class="section-body">
          <h2>Pools</h2>
          <div class="table-scroll">
            <table class="mono">
              <thead
                ><tr><th>pool</th><th>status</th><th>details</th></tr></thead
              ><tbody
                >{#each system?.pools?.pools || [] as pool}<tr
                    ><td>{pool.name || pool.pool}</td><td
                      >{system.pools.quarantined?.[pool.name]
                        ? "quarantined"
                        : "ok"}</td
                    ><td><pre>{JSON.stringify(pool)}</pre></td></tr
                  >{/each}</tbody
              >
            </table>
          </div>
          <h2>Resources</h2>
          <pre>{JSON.stringify(system?.resources, null, 2)}</pre>
          <h2>Scheduler</h2>
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
          <h2>Sessions</h2>
          {#each system?.sessions || [] as session}<p>
              <button class="mono" onclick={() => openTrace(session.id)}
                >{session.display_name || session.id}</button
              >
              {session.busy ? "working" : "idle"}
              <button onclick={() => closeSession(session.id)}>Close</button>
            </p>{/each}
          <p class="muted">
            Costs are not recorded by this controller. Context usage is
            available in session journals.
          </p>
        </div>
      {/if}
    </main>
    <aside class="chat-column" aria-label="Chats">
      <button class="show-chats" onclick={showChat}>Chats ▸</button>
      <div class="chat-content">
        <Chat
          bind:this={chat}
          {requested}
          {chats}
          onactive={(c) => (activeChat = c)}
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
          ? "In flight"
          : tab[0].toUpperCase() + tab.slice(1)}{tab === "waiting"
          ? " " + attention.length
          : ""}</button
      >{/each}
  </nav>
</div>
