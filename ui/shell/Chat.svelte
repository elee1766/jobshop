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
  type Attachment = { file: File; preview?: string };
  let attachments: Attachment[] = $state([]);
  let fileInput: HTMLInputElement;
  let documentPopover: HTMLDialogElement;
  let documentText = $state("");
  let pasteNumber = 0;
  let attachmentEpoch = 0;
  let adding = $state(0);
  let name = $state("");
  let creating = $state(false);
  let draftName = $state("");
  let newButton: HTMLButtonElement;
  let newPopover: HTMLDialogElement;
  let nameInput: HTMLInputElement;
  let archived = $state(false);
  let search = $state("");
  let pickerButton: HTMLButtonElement;
  let picker: HTMLDialogElement;
  let searchInput: HTMLInputElement;
  let previousDraft: { text: string; refs: string[]; attachments: Attachment[] } | null = null;
  const archivedChats = $derived(chats.filter((chat) => chat.state === "archived"));
  const recentChats = $derived.by(() => {
    const all = active && !chats.some(chat => chat.id === active?.id) ? [...chats, active] : chats;
    const query = search.toLocaleLowerCase().replace(/\s+/g, "");
    return all.filter(chat => {
      if ((chat.state === "archived") !== archived) return false;
      let index = 0;
      for (const char of chat.title.toLocaleLowerCase()) {
        if (char === query[index]) index++;
      }
      return index === query.length;
    }).sort((a, b) => (b.updated || "").localeCompare(a.updated || "") || a.title.localeCompare(b.title));
  });
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
      updated: info.ticket?.updated_at,
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
    attachmentEpoch++;
    const selectedGeneration = ++generation;
    closeStream();
    active = chat;
    frames = [];
    commands = [];
    refs = [];
    clearAttachments(attachments);
    clearAttachments(previousDraft?.attachments || []);
    attachments = [];
    creating = false;
    previousDraft = null;
    text = "";
    error = "";
    picker?.hidePopover();
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
    attachmentEpoch++;
    if (!creating) previousDraft = { text, refs: [...refs], attachments: [...attachments] };
    else clearAttachments(attachments);
    attachments = [];
    refs = [...new Set(contextRefs)];
    text = draft;
    name = chatName;
    creating = true;
    error = "";
    void tick().then(() => composer?.focus());
  }

  function dismissDraft() {
    if (sending) return;
    attachmentEpoch++;
    creating = false;
    name = "";
    clearAttachments(attachments);
    attachments = previousDraft?.attachments || [];
    text = previousDraft?.text || "";
    refs = previousDraft?.refs || [];
    previousDraft = null;
    error = "";
    void tick().then(() => composer?.focus());
  }

  function clearAttachments(files: Attachment[]) {
    for (const item of files) if (item.preview) URL.revokeObjectURL(item.preview);
  }
  function removeAttachment(item: Attachment) {
    clearAttachments([item]);
    attachments = attachments.filter(file => file !== item);
  }
  async function addFiles(files: File[]) {
    if (sending) return false;
    const epoch = attachmentEpoch;
    adding++;
    try {
      if (attachments.length + files.length > 8 || [...attachments.map(item => item.file), ...files].reduce((n, file) => n + file.size, 0) > 8 * 1024 * 1024)
        throw new Error("Use at most 8 attachments totaling 8 MiB.");
      for (const file of files) {
        if (!file.size) throw new Error("Empty files cannot be attached.");
        if (file.type.startsWith("image/")) {
          if (!["image/png", "image/jpeg", "image/gif"].includes(file.type)) throw new Error("Use PNG, JPEG or GIF images.");
        } else {
          if (file.size > 2 * 1024 * 1024) throw new Error("Text documents must be at most 2 MiB.");
          const content = new TextDecoder("utf-8", { fatal: true }).decode(await file.arrayBuffer());
          if (content.includes("\0")) throw new Error("Use UTF-8 text documents, not binary files.");
        }
      }
      if (epoch !== attachmentEpoch) return false;
      if (attachments.length + files.length > 8 || [...attachments.map(item => item.file), ...files].reduce((n, file) => n + file.size, 0) > 8 * 1024 * 1024)
        throw new Error("Use at most 8 attachments totaling 8 MiB.");
      attachments = [...attachments, ...files.map(file => ({file, preview: file.type.startsWith("image/") ? URL.createObjectURL(file) : undefined}))];
      error = "";
      return true;
    } catch (e) { if (epoch === attachmentEpoch) error = String(e); return false; }
    finally { adding--; }
  }
  function paste(e: ClipboardEvent) {
    if (sending) return;
    const clipboard = e.clipboardData;
    if (!clipboard) return;
    const files = Array.from(clipboard.files);
    if (files.length) {
      e.preventDefault();
      void addFiles(files);
      return;
    }
    const pasted = clipboard.getData("text/plain");
    if (pasted.length > 4000 || pasted.split("\n").length >= 40) {
      e.preventDefault();
      void addFiles([new File([pasted], `pasted-text-${++pasteNumber}.txt`, {type:"text/plain"})]);
    }
  }
  async function attachmentPayload() {
    return Promise.all(attachments.map(async ({file}) => {
      const bytes = new Uint8Array(await file.arrayBuffer());
      let binary = "";
      for (let i = 0; i < bytes.length; i += 8192) binary += String.fromCharCode(...bytes.subarray(i, i + 8192));
      return {name: file.name, mime_type: file.type.startsWith("image/") ? file.type : "text/plain", data: btoa(binary)};
    }));
  }

  function positionPicker() {
    if (!pickerButton || !picker) return;
    const rect = pickerButton.getBoundingClientRect();
    const width = Math.min(380, window.innerWidth - 16);
    picker.style.width = width + "px";
    picker.style.top = rect.bottom + 4 + "px";
    picker.style.left = Math.max(8, Math.min(rect.left, window.innerWidth - width - 8)) + "px";
    picker.style.maxHeight = Math.max(120, window.innerHeight - rect.bottom - 16) + "px";
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
    if ((!text.trim() && !attachments.length) || sending || adding) return;
    if (creating && !name.trim()) {
      error = "Give this chat a name.";
      return;
    }
    sending = true;
    error = "";
    const message = text;
    const sentGeneration = generation;
    try {
      const files = await attachmentPayload();
      if (creating || !active) {
        const result = await client.post<SessionInfo>("api/chat/manage", {
          action: "new",
          name,
          text: message,
          refs,
          attachments: files,
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
          attachments: files,
        });
      }
      if (generation === sentGeneration) {
        text = "";
        refs = [];
        clearAttachments(attachments);
        attachments = [];
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
      attachmentEpoch++;
      clearAttachments(attachments);
      clearAttachments(previousDraft?.attachments || []);
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
      images?: SessionEvent["images"];
    }[] = [];
    for (const frame of frames) {
      if (frame.type === "tool" && frame.tool) {
        const previous = out.find((m) => m.tool?.id === frame.tool?.id);
        if (previous) {
          previous.tool = { ...previous.tool, ...frame.tool };
          if (frame.images?.length) previous.images = frame.images;
          continue;
        }
      }
      if (frame.type === "text" && out.at(-1)?.kind === "text")
        out[out.length - 1].text += frame.text || "";
      else if (
        ["turn_start", "text", "tool", "image"].includes(frame.type) ||
        frame.error
      )
        out.push({
          key: String(frame.seq),
          kind: frame.type,
          text: frame.text || frame.error || "",
          tool: frame.tool,
          images: frame.images,
        });
    }
    return out;
  });
</script>

<svelte:window onresize={() => { positionNewChat(); positionPicker(); }} />
<div class="tag chat-tag">
  <span class="muted">To:</span>
  <button
    class="chat-picker-button"
    bind:this={pickerButton}
    popovertarget="chat-picker"
    aria-label="Choose chat"
    disabled={sending}
    onclick={() => {
      if (!picker.matches(":popover-open")) { search = ""; archived = active?.state === "archived"; }
      positionPicker();
    }}
    title={creating ? "New chat: " + name : active?.title || "Choose chat"}
    >{creating ? "New chat: " + name : active?.title || "Choose chat"} ▾</button
  ><button
    bind:this={newButton}
    popovertarget="new-chat"
    disabled={sending}
    onclick={() => {
      if (!newPopover.matches(":popover-open")) draftName = "";
      positionNewChat();
    }}>New</button
  ><button class="push-right" onclick={onhide}>Hide</button>
</div>
<dialog id="chat-picker" class="chat-picker" bind:this={picker} popover="auto" aria-label="Choose chat" ontoggle={() => { if (picker.matches(":popover-open")) searchInput?.focus(); }}>
  <div class="chat-picker-search"><input bind:this={searchInput} type="search" bind:value={search} aria-label="Search chats" placeholder="Search chats" onkeydown={(event) => {
    if (event.key === "Enter" && recentChats[0]) { event.preventDefault(); void select(recentChats[0]); }
    if (event.key === "ArrowDown") { event.preventDefault(); picker.querySelector<HTMLButtonElement>("tbody button")?.focus(); }
  }} /></div>
  <div class="chat-picker-filter">
    <button class:chosen={!archived} aria-pressed={!archived} onclick={() => archived = false}>Recent</button>
    <button class:chosen={archived} aria-pressed={archived} onclick={() => archived = true}>Archived ({archivedChats.length})</button>
    <span class="muted">Newest first</span>
  </div>
  <div class="chat-archive">
    <table aria-label={archived ? "Archived chats" : "Recent chats"}>
      <thead><tr><th scope="col">Chat</th><th scope="col">Updated</th></tr></thead>
      <tbody>{#each recentChats as chat (chat.id)}
        <tr class:chosen={!creating && active?.id === chat.id}>
          <td><button aria-pressed={!creating && active?.id === chat.id} onclick={() => select(chat)}>{chat.title}{chat.state === "working" ? " ●" : ""}</button></td>
          <td>{#if chat.updated}<time datetime={chat.updated}>{new Date(chat.updated).toLocaleDateString()}</time>{:else}–{/if}</td>
        </tr>
      {/each}</tbody>
    </table>
    {#if !recentChats.length}<p class="muted">{search ? "No matching chats." : archived ? "No archived chats." : "No chats yet. Choose New."}</p>{/if}
  </div>
</dialog>
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
{#if active && !creating}
  <div class="tag chat-toolbar" role="toolbar" aria-label="Chat actions">
    <div class="chat-actions"><button disabled={sending} onclick={() => manage("rename")}>Rename</button><button disabled={sending || active.state === "archived"} onclick={() => manage("archive")}>Archive</button><button onclick={stop}
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
        )}>Journal</button></div
    ><span class="chat-state muted"
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
  {#if creating}
    <section class="new-chat-draft" aria-label="New chat draft">
      <div><strong>New chat: {name}</strong><button disabled={sending} onclick={dismissDraft} aria-label="Dismiss new chat">Dismiss ×</button></div>
      <p>Your first message will create this chat. Nothing has been sent yet.</p>
    </section>
  {:else}
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
        {#each message.images || [] as image}<img class="transcript-image" src={"data:" + image.mime_type + ";base64," + image.data} alt="Tool output" loading="lazy" />{/each}
      </details>
    {:else if message.kind === "turn_start"}<div class="operator">
        {message.text}
        {#each message.images || [] as image}<img class="transcript-image" src={"data:" + image.mime_type + ";base64," + image.data} alt="User attachment" loading="lazy" />{/each}
      </div>
    {:else}<div class="markdown">{@html md(message.text)}{#each message.images || [] as image}<img class="transcript-image" src={"data:" + image.mime_type + ";base64," + image.data} alt="Agent output" loading="lazy" />{/each}</div>{/if}
  {/each}
  {#each commands as command}<p class="tool mono">
      desk ▸ {command.op}
      {Object.entries(command.args || {})
        .map(([k, v]) => k + "=" + v)
        .join(" ")}
    </p>{/each}
  {/if}
</div>
{#if error}<p class="error" role="alert">{error}</p>{/if}
<form
  class="composer"
  onsubmit={(e) => {
    e.preventDefault();
    void send();
  }}
>
  <input type="file" hidden multiple bind:this={fileInput} aria-label="Attach files" accept="image/png,image/jpeg,image/gif,text/*,.txt,.md,.json,.yaml,.yml,.log,.csv,.xml" onchange={() => {
    const files = Array.from(fileInput.files || []);
    fileInput.value = "";
    void addFiles(files);
  }} />
  <div class="attachment-actions">
    <button type="button" title="Attach images or text files" disabled={sending || (!active && !creating) || (active?.state === "archived" && !creating)} onclick={() => fileInput.click()}>Attach</button>
    <button type="button" title="Paste text as a document" disabled={sending || (!active && !creating) || (active?.state === "archived" && !creating)} popovertarget="paste-document">Paste</button>
  </div>
  {#if attachments.length}<ul class="chat-attachments" aria-label="Message attachments">
    {#each attachments as item}<li>
      {#if item.preview}<img src={item.preview} alt={item.file.name} />{/if}
      <span>{item.file.name}<small>{Math.max(1, Math.ceil(item.file.size / 1024))} KiB · {item.preview ? "image" : "read on demand"}</small></span>
      <button type="button" disabled={sending} aria-label={"Remove attachment " + item.file.name} onclick={() => removeAttachment(item)}>×</button>
    </li>{/each}
  </ul>{/if}
  {#if adding}<span role="status">Preparing attachments…</span>{/if}
  {#if refs.length}<div class="context">
      Attached work: {#each refs as ref}<button
          title="Remove attached reference"
          aria-label={"Remove attached reference " + ref}
          onclick={() => (refs = refs.filter((r) => r !== ref))}>{ref} ×</button
        >{/each}
    </div>{/if}
  <div class="compose-line">
    <textarea
      id="chat-message"
      aria-label={creating ? "First message to new chat: " + name : "to " + (active?.title || "new chat")}
      bind:this={composer}
      bind:value={text}
      onkeydown={keydown}
      onpaste={paste}
      placeholder={active || creating
        ? "Message"
        : "Choose New to start a named chat"}
      disabled={sending || (!active && !creating) ||
        (active?.state === "archived" && !creating)}
      rows="3"
    ></textarea><button
      class="send"
      type="submit"
      disabled={sending || adding > 0 || (!text.trim() && !attachments.length)}>{creating ? "Create & send" : "Send"}</button
    >
  </div>
</form>
<dialog id="paste-document" popover="auto" class="paste-document" bind:this={documentPopover} aria-label="Paste document">
  <label for="document-text">Paste text to attach as a document</label>
  <textarea id="document-text" bind:value={documentText} rows="10"></textarea>
  <div><button disabled={!documentText || sending || adding > 0} onclick={async () => {
    if (await addFiles([new File([documentText], `pasted-text-${++pasteNumber}.txt`, {type:"text/plain"})])) {
      documentText = "";
      documentPopover.hidePopover();
      composer?.focus();
    }
  }}>Attach document</button><button onclick={() => documentPopover.hidePopover()}>Cancel</button></div>
</dialog>
