<script lang="ts">
  import { onMount, tick, untrack } from "svelte";
  import { client } from "./lib/api.svelte.js";
  import { md } from "../kit/markdown.js";
  import type {
    SessionEvent,
    Event,
    SessionInfo,
  } from "../../packages/jobshop-client/rpc.js";

  export interface Chat {
    id: string;
    key: string;
    title: string;
    state?: string;
    updated?: string;
  }
  let {
    requested = null,
    chats = [],
    onactive,
    oncommand,
    onhide,
  }: {
    requested?: Chat | null;
    chats?: Chat[];
    onactive: (chat: Chat | null) => void;
    oncommand: (ev: Event) => void;
    onhide: () => void;
  } = $props();
  let active: Chat | null = $state(null);
  let frames: SessionEvent[] = $state([]);
  let commands: Event[] = $state([]);
  let refs: string[] = $state([]);
  let text = $state("");
  let name = $state("");
  let creating = $state(false);
  let draftName = $state("");
  let newButton: HTMLButtonElement;
  let newPopover: HTMLDialogElement;
  let nameInput: HTMLInputElement;
  let archived = $state(false);
  const archivedChats = $derived(
    chats
      .filter((chat) => chat.state === "archived")
      .sort((a, b) => (b.updated || "").localeCompare(a.updated || "")),
  );
  let error = $state("");
  let sending = $state(false);
  let connected = $state("closed");
  let transcript: HTMLDivElement;
  let composer: HTMLTextAreaElement;
  let generation = 0;
  let lastSeq = 0;
  let loadingOlder = $state(false);
  let closeStream = () => {};

  function managedChat(info: SessionInfo): Chat {
    const fields = info.ticket?.fields as
      | { key?: string; suspended?: boolean }
      | undefined;
    return {
      id: info.id,
      key: fields?.key || active?.key || "",
      title: info.display_name || info.ticket?.title || info.id,
      state:
        info.ticket?.status === "closed"
          ? "archived"
          : info.busy
            ? "working"
            : fields?.suspended
              ? "suspended"
              : "idle",
    };
  }
  $effect(() => {
    const list = chats;
    untrack(() => {
      const updated = list.find((entry) => entry.id === active?.id);
      if (updated) {
        active = updated;
        onactive(updated);
      }
    });
  });
  export function receiveCommand(ev: Event) {
    if (ev.session !== active?.id) return;
    commands = [...commands, ev];
    oncommand(ev);
    if (ev.op === "attach")
      attach((ev.args?.refs || "").split(/[ ,]+/).filter(Boolean));
  }
  async function select(chat: Chat) {
    const selectedGeneration = ++generation;
    closeStream();
    active = chat;
    frames = [];
    commands = [];
    refs = [];
    creating = false;
    error = "";
    onactive(chat);
    lastSeq = 0;
    connect(chat.id, selectedGeneration);
  }
  function connect(id: string, selectedGeneration: number) {
    closeStream();
    closeStream = client.session(
      id,
      (event, kind) => {
        if (kind !== "session" || generation !== selectedGeneration) return;
        const ev = event as SessionEvent;
        if (ev.seq) lastSeq = Math.max(lastSeq, ev.seq);
        if (ev.seq && frames.some((f) => f.seq === ev.seq)) return;
        const atBottom =
          !transcript ||
          transcript.scrollHeight -
            transcript.scrollTop -
            transcript.clientHeight <
            80;
        frames = [...frames, ev];
        if (atBottom)
          void tick().then(() => {
            if (transcript && generation === selectedGeneration)
              transcript.scrollTop = transcript.scrollHeight;
          });
      },
      {
        since: lastSeq,
        onState: (state) => {
          if (generation === selectedGeneration) connected = state;
        },
      },
    );
  }
  $effect(() => {
    const target = requested;
    untrack(() => {
      if (target && target.id !== active?.id) void select(target);
    });
  });
  export function start(draft = "", contextRefs: string[] = [], chatName = "") {
    generation++;
    closeStream();
    active = null;
    frames = [];
    commands = [];
    refs = [...new Set(contextRefs)];
    text = draft;
    name = chatName;
    creating = true;
    onactive(null);
    void tick().then(() => composer?.focus());
  }

  function positionNewChat() {
    if (!newButton || !newPopover) return;
    const rect = newButton.getBoundingClientRect();
    newPopover.style.top = rect.bottom + 4 + "px";
    newPopover.style.left = Math.max(8, Math.min(rect.left, window.innerWidth - 348)) + "px";
  }

  export function attach(values: string[]) {
    refs = [...new Set([...refs, ...values])];
  }
  async function older() {
    if (!active || loadingOlder) return;
    const selectedGeneration = generation;
    const height = transcript?.scrollHeight || 0;
    const top = transcript?.scrollTop || 0;
    const before = frames.find((f) => (f.seq || 0) > 0)?.seq;
    if (!before) return;
    loadingOlder = true;
    try {
      const data = await client.get(
        `api/session/${encodeURIComponent(active.id)}/events?before=${before}&limit=200`,
      );
      if (generation !== selectedGeneration) return;
      frames = [
        ...(data.events || []).filter(
          (event: SessionEvent) =>
            !event.seq || !frames.some((frame) => frame.seq === event.seq),
        ),
        ...frames,
      ];
      await tick();
      if (generation === selectedGeneration && transcript)
        transcript.scrollTop = top + transcript.scrollHeight - height;
    } catch (e) {
      if (generation === selectedGeneration) error = String(e);
    } finally {
      loadingOlder = false;
    }
  }
  async function manage(action: string) {
    if (!active) return;
    const newName =
      action === "rename" ? window.prompt("Chat name", active.title) : "";
    if (action === "rename" && !newName) return;
    try {
      const result = await client.post<SessionInfo>("api/chat/manage", {
        action,
        id: active.id,
        name: newName,
      });
      await select(managedChat(result));
    } catch (e) {
      error = String(e);
    }
  }
  async function send() {
    if (!text.trim() || sending) return;
    if (creating && !name.trim()) {
      error = "Give this chat a name.";
      return;
    }
    sending = true;
    error = "";
    const message = text;
    const sentGeneration = generation;
    try {
      if (creating || !active) {
        const result = await client.post<SessionInfo>("api/chat/manage", {
          action: "new",
          name,
          text: message,
          refs,
        });
        if (generation !== sentGeneration) return;
        text = "";
        refs = [];
        await select(managedChat(result));
      } else {
        await client.post("api/chat/send", {
          key: active.key,
          text: message,
          refs,
        });
      }
      if (generation === sentGeneration) {
        text = "";
        refs = [];
      }
    } catch (e) {
      if (generation === sentGeneration) error = String(e);
    } finally {
      sending = false;
      composer?.focus();
    }
  }
  async function stop() {
    if (!active) return;
    try {
      await client.post("api/chat/cancel", { session: active.id });
    } catch (e) {
      error = String(e);
    }
  }
  function keydown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey && !e.isComposing) {
      e.preventDefault();
      void send();
    }
    if (e.key === "Escape") {
      e.preventDefault();
      void stop();
    }
  }
  onMount(() => {
    const offline = () => {
      closeStream();
      connected = "reconnecting";
    };
    const online = () => {
      if (active) connect(active.id, generation);
    };
    window.addEventListener("offline", offline);
    window.addEventListener("online", online);
    return () => {
      generation++;
      closeStream();
      window.removeEventListener("offline", offline);
      window.removeEventListener("online", online);
    };
  });
  const messages = $derived.by(() => {
    const out: {
      key: string;
      kind: string;
      text: string;
      tool?: SessionEvent["tool"];
    }[] = [];
    for (const frame of frames) {
      if (frame.type === "tool" && frame.tool) {
        const previous = out.find((m) => m.tool?.id === frame.tool?.id);
        if (previous) {
          previous.tool = { ...previous.tool, ...frame.tool };
          continue;
        }
      }
      if (frame.type === "text" && out.at(-1)?.kind === "text")
        out[out.length - 1].text += frame.text || "";
      else if (
        ["turn_start", "text", "tool"].includes(frame.type) ||
        frame.error
      )
        out.push({
          key: String(frame.seq),
          kind: frame.type,
          text: frame.text || frame.error || "",
          tool: frame.tool,
        });
    }
    return out;
  });
</script>

<svelte:window onresize={positionNewChat} />
<div class="tag chat-tag">
  <span class="mono">/chats</span><button
    bind:this={newButton}
    popovertarget="new-chat"
    onclick={() => {
      if (!newPopover.matches(":popover-open")) draftName = "";
      positionNewChat();
    }}>New</button
  ><button disabled={!active} onclick={() => manage("rename")}>Rename</button
  ><button
    disabled={!active || active.state === "archived"}
    onclick={() => manage("archive")}>Archive</button
  ><button
    class:chosen={archived}
    aria-expanded={archived}
    onclick={() => (archived = !archived)}
    >Archived({archivedChats.length})</button
  ><button class="push-right" onclick={onhide}>Hide</button>
</div>
<dialog
  id="new-chat"
  class="new-chat-popover"
  bind:this={newPopover}
  popover="auto"
  aria-label="New chat"
  ontoggle={() => {
    if (newPopover.matches(":popover-open")) nameInput?.focus();
  }}
>
<form onsubmit={(event) => {
    event.preventDefault();
    if (!draftName.trim()) return;
    newPopover.hidePopover();
    start("", [], draftName.trim());
  }}
>
  <label for="new-chat-name">Name</label>
  <input id="new-chat-name" bind:this={nameInput} bind:value={draftName} maxlength="120" placeholder="ops" required />
  <button type="submit">Start</button>
  <button type="button" popovertarget="new-chat" popovertargetaction="hide">Cancel</button>
</form>
</dialog>
{#if archived}
  <div class="chat-archive">
    <table aria-label="Archived chats">
      <thead>
        <tr><th scope="col">Chat</th><th scope="col">Updated</th></tr>
      </thead>
      <tbody>
        {#each archivedChats as chat (chat.id)}
          <tr class:chosen={active?.id === chat.id}>
            <td>
              <button
                aria-pressed={active?.id === chat.id}
                onclick={() => select(chat)}>{chat.title}</button
              >
            </td>
            <td>
              {#if chat.updated}
                <time datetime={chat.updated}
                  >{new Date(chat.updated).toLocaleDateString()}</time
                >
              {:else}–{/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
    {#if !archivedChats.length}<p class="muted">No archived chats.</p>{/if}
  </div>
{:else}
  <div class="chat-tabs" role="tablist" aria-label="Named chats">
    {#each chats.filter((c) => c.state !== "archived") as chat (chat.id)}<button
        role="tab"
        aria-selected={active?.id === chat.id}
        class:chosen={active?.id === chat.id}
        onclick={() => select(chat)}
        >{chat.title}{chat.state === "working" ? " ●" : ""}</button
      >{/each}
    {#if !chats.some((chat) => chat.state !== "archived")}
      <span class="muted">No chats yet. Choose New.</span>
    {/if}
  </div>
{/if}
{#if active && !creating}
  <div class="tag">
    <span class="mono">/chats/{active.title}</span><button onclick={stop}
      >Stop</button
    ><button
      onclick={() => {
        const ref = window.prompt("Attach ticket, PR or session reference");
        if (ref) attach([ref]);
      }}>Attach</button
    ><button
      onclick={() =>
        window.open(
          client.url(
            "api/session/" + encodeURIComponent(active!.id) + "/journal",
          ),
          "_blank",
          "noopener",
        )}>Journal</button
    ><span class="push-right muted"
      >{active.state === "archived"
        ? "archived"
        : connected === "reconnecting"
          ? "reconnecting"
          : active.state || "idle"}</span
    >
  </div>
{/if}
<div
  class="transcript"
  bind:this={transcript}
  aria-live="polite"
  aria-label="Chat transcript"
>
  {#if frames.some((f) => (f.seq || 0) > 1)}<button
      disabled={loadingOlder}
      onclick={older}>Load older</button
    >{/if}
  {#if active?.state === "archived"}<p>
      This chat is archived. <button onclick={() => manage("restore")}
        >Restore</button
      > to continue the same conversation.
    </p>{/if}
  {#each messages as message (message.key)}
    {#if message.tool}<details class="tool mono">
        <summary
          >{message.tool.intent || message.tool.title || "Tool"} · {message.tool
            .status}</summary
        >
        <pre>{message.tool.target || ""}\n{message.tool.input || ""}\n{message
            .tool.output || ""}</pre>
      </details>
    {:else if message.kind === "turn_start"}<div class="operator">
        {message.text}
      </div>
    {:else}<div class="markdown">{@html md(message.text)}</div>{/if}
  {/each}
  {#each commands as command}<p class="tool mono">
      desk ▸ {command.op}
      {Object.entries(command.args || {})
        .map(([k, v]) => k + "=" + v)
        .join(" ")}
    </p>{/each}
</div>
{#if error}<p class="error" role="alert">{error}</p>{/if}
<form
  class="composer"
  onsubmit={(e) => {
    e.preventDefault();
    void send();
  }}
>
  {#if refs.length}<div class="context">
      ctx: {#each refs as ref}<button
          title="Remove context"
          onclick={() => (refs = refs.filter((r) => r !== ref))}>{ref} ×</button
        >{/each}
    </div>{/if}
  <label for="chat-message" class="muted"
    >to {creating ? name || "new chat" : active?.title || "new chat"}</label
  >
  <div class="compose-line">
    <textarea
      id="chat-message"
      bind:this={composer}
      bind:value={text}
      onkeydown={keydown}
      placeholder={active || creating
        ? "Message"
        : "Choose New to start a named chat"}
      disabled={(!active && !creating) ||
        (active?.state === "archived" && !creating)}
      rows="3"
    ></textarea><button
      class="send"
      type="submit"
      disabled={sending || !text.trim()}>Send</button
    >
  </div>
</form>
