import { Dialog } from '@base-ui/react/dialog';
import { Popover } from '@base-ui/react/popover';
import { useAtomValue } from 'jotai';
import { useEffect, useEffectEvent, useMemo, useRef, useState, type ClipboardEvent, type KeyboardEvent } from 'react';
import type { SessionInfo, SessionSendResult } from '../../../packages/jobshop-client/rpc.js';
import { client } from '../lib/api.ts';
import { isLargePaste, payload, previewOf, revoke, validate, type Attachment } from '../lib/attachments.ts';
import { createSearcher, type SearchSpec } from '../lib/search.ts';
import { sessionActivity, type PendingMessage } from '../lib/session.ts';
import { useSessionFrames } from '../lib/useSessionFrames.ts';
import { useDeskRoute } from '../lib/route.ts';
import { navigate, openTrace, setChatParam, setFilter } from '../state/actions.ts';
import { chatBridge } from '../state/chatBridge.ts';
import { onDeskCommand } from '../state/live.ts';
import { store } from '../state/store.ts';
import { chatListAtom, collapsedAtom, focusedAtom, type Chat as ChatInfo } from '../state/ui.ts';
import { Transcript } from './Transcript.tsx';
import type { Event } from '../../../packages/jobshop-client/rpc.js';

function chatType(chat: ChatInfo) {
  if (chat.named) return 'chat';
  if (chat.key.startsWith('escalation:')) return 'escalation';
  if (chat.key.startsWith('attention:')) return 'attention';
  return 'chat';
}

const byRecent = (a: ChatInfo, b: ChatInfo) =>
  (b.updated || '').localeCompare(a.updated || '') || a.title.localeCompare(b.title);

// Exact ids first, then weighted title and key matches, then a loose match
// of the title's letters in order. Only the loaded chats are searched.
const chatSearch: SearchSpec<ChatInfo> = {
  fields: [
    { name: 'title', weight: 3, get: (chat) => chat.title },
    { name: 'key', weight: 1, get: (chat) => chat.key },
    { name: 'type', weight: 0.5, get: chatType },
  ],
  ids: (chat) => [chat.id, chat.key],
  label: (chat) => chat.title,
};

function managedChat(info: SessionInfo, previousKey: string): ChatInfo {
  const fields = info.ticket?.fields as { key?: string; suspended?: boolean; named?: boolean } | undefined;
  return {
    id: info.id,
    key: fields?.key || previousKey,
    named: fields?.named,
    title: info.display_name || info.ticket?.title || info.id,
    updated: info.ticket?.updated_at || '',
    alive: info.alive,
    state:
      info.ticket?.status === 'closed' ? 'archived' : info.busy ? 'working' : fields?.suspended ? 'suspended' : 'idle',
  };
}

export function Chat({ onHide }: { onHide: () => void }) {
  const list = useAtomValue(chatListAtom);
  const route = useDeskRoute();
  // The chat the reader opened. The open chat is derived from the live list,
  // so its state and title follow the controller with no sync effect. `opened`
  // is kept for a chat the list does not have yet (one just created), and
  // `wanted` is the chat the URL named on load, used until the reader picks one.
  const [opened, setOpened] = useState<ChatInfo | null>(null);
  const [wanted, setWanted] = useState(route.chat);
  const active =
    (opened && (list.find((entry) => entry.id === opened.id) ?? opened)) ||
    (wanted ? (list.find((entry) => entry.id === wanted) ?? null) : null);
  const { frames, connection, prepend } = useSessionFrames(active?.id ?? null);

  const [commands, setCommands] = useState<Event[]>([]);
  const [refs, setRefs] = useState<string[]>([]);
  const [text, setText] = useState('');
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [adding, setAdding] = useState(0);
  const [name, setName] = useState('');
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState('');
  const [sending, setSending] = useState(false);
  const [pending, setPending] = useState<PendingMessage[]>([]);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [newOpen, setNewOpen] = useState(false);
  const [pasteOpen, setPasteOpen] = useState(false);
  const [transcript, setTranscript] = useState<HTMLDivElement | null>(null);
  const composer = useRef<HTMLTextAreaElement>(null);
  const fileInput = useRef<HTMLInputElement>(null);
  const previousDraft = useRef<{ text: string; refs: string[]; attachments: Attachment[] } | null>(null);
  // The attachments, mirrored for async code that outlives a render (a file
  // read in progress). Written only where the list changes.
  const attachmentsNow = useRef<Attachment[]>([]);
  const generation = useRef(0);
  const attachmentEpoch = useRef(0);
  const pasteNumber = useRef(0);
  const submittedNew = useRef(false);

  const activity = sessionActivity(frames, active?.state);
  const focusComposer = () => requestAnimationFrame(() => composer.current?.focus());

  function putAttachments(next: Attachment[]) {
    attachmentsNow.current = next;
    setAttachments(next);
  }

  function select(chat: ChatInfo | null) {
    attachmentEpoch.current++;
    generation.current++;
    revoke(attachmentsNow.current);
    revoke(previousDraft.current?.attachments || []);
    previousDraft.current = null;
    setWanted('');
    setOpened(chat);
    setPending([]);
    setCommands([]);
    setRefs([]);
    putAttachments([]);
    setCreating(false);
    setText('');
    setError('');
    setPickerOpen(false);
    setChatParam(chat?.id || '');
  }

  function start(draft = '', contextRefs: string[] = [], chatName = '') {
    attachmentEpoch.current++;
    if (!creating) previousDraft.current = { text, refs: [...refs], attachments: [...attachments] };
    else revoke(attachments);
    putAttachments([]);
    setRefs([...new Set(contextRefs)]);
    setText(draft);
    setName(chatName);
    setCreating(true);
    setError('');
    focusComposer();
  }
  function attach(values: string[]) {
    if (!active?.key.startsWith('desk:')) {
      setError('Open a named chat before attaching context.');
      return;
    }
    setRefs((all) => [...new Set([...all, ...values])]);
  }

  // Other screens reach the panel through the bridge, and the controller's
  // commands drive Desk. Effect events always run the newest handlers, so the
  // registrations are made once.
  const bridgeOpen = useEffectEvent((chat: ChatInfo) => select(chat));
  const bridgeStart = useEffectEvent((draft: string, contextRefs: string[], chatName: string) =>
    start(draft, contextRefs, chatName),
  );
  const bridgeAttach = useEffectEvent((values: string[]) => attach(values));
  const command = useEffectEvent((ev: Event) => {
    if (ev.session !== active?.id) return;
    setCommands((all) => [...all, ev]);
    const args = ev.args || {};
    if (ev.op === 'navigate') {
      if (args.ref) openTrace(args.ref);
      else navigate(args.view as 'home');
    }
    if (ev.op === 'filter') setFilter([args.repo, args.step, args.waiting_on].filter(Boolean).join(' '));
    if (ev.op === 'select' && args.refs) store.set(focusedAtom, args.refs.split(/[ ,]+/)[0]);
    if (ev.op === 'expand' || ev.op === 'collapse') {
      store.set(collapsedAtom, (all) => ({ ...all, [args.window.replace(/^\//, '')]: ev.op === 'collapse' }));
    }
    if (ev.op === 'attach') bridgeAttach((args.refs || '').split(/[ ,]+/).filter(Boolean));
  });
  const release = useEffectEvent(() => {
    attachmentEpoch.current++;
    generation.current++;
    revoke(attachmentsNow.current);
    revoke(previousDraft.current?.attachments || []);
  });
  useEffect(() => {
    chatBridge.open = (chat) => bridgeOpen(chat);
    chatBridge.start = (draft, contextRefs, chatName) => bridgeStart(draft, contextRefs, chatName);
    chatBridge.attach = (values) => bridgeAttach(values);
    const unsubscribe = onDeskCommand((ev) => command(ev));
    return () => {
      unsubscribe();
      delete chatBridge.open;
      delete chatBridge.start;
      delete chatBridge.attach;
      // Release object URLs when the panel goes away.
      release();
    };
  }, []);

  function dismissDraft() {
    if (sending) return;
    attachmentEpoch.current++;
    revoke(attachments);
    const back = previousDraft.current;
    setCreating(false);
    setName('');
    putAttachments(back?.attachments || []);
    setText(back?.text || '');
    setRefs(back?.refs || []);
    previousDraft.current = null;
    setError('');
    focusComposer();
  }

  async function addFiles(files: File[]) {
    if (sending) return false;
    const epoch = attachmentEpoch.current;
    setAdding((n) => n + 1);
    try {
      await validate(attachmentsNow.current, files);
      if (epoch !== attachmentEpoch.current) return false;
      // Another add may have landed while the files were read.
      await validate(attachmentsNow.current, files);
      putAttachments([...attachmentsNow.current, ...files.map(previewOf)]);
      setError('');
      return true;
    } catch (e) {
      if (epoch === attachmentEpoch.current) setError(e instanceof Error ? e.message : String(e));
      return false;
    } finally {
      setAdding((n) => n - 1);
    }
  }
  function removeAttachment(item: Attachment) {
    revoke([item]);
    putAttachments(attachmentsNow.current.filter((file) => file !== item));
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
    const pasted = clipboard.getData('text/plain');
    if (isLargePaste(pasted)) {
      e.preventDefault();
      void addFiles([new File([pasted], `pasted-text-${++pasteNumber.current}.txt`, { type: 'text/plain' })]);
    }
  }

  async function manage(action: string) {
    if (!active || sending) return;
    if (
      action === 'resolve' &&
      !window.confirm(
        'Resolve escalation ' +
          active.key.slice('escalation:'.length) +
          ' and close this chat? This closes the linked ticket. It does not resume runs or unquarantine pools.',
      )
    )
      return;
    const newName = action === 'rename' || action === 'convert' ? window.prompt('Chat name', active.title) : '';
    if ((action === 'rename' || action === 'convert') && !newName) return;
    const selected = generation.current;
    setSending(true);
    setError('');
    try {
      const result = await client.post<SessionInfo>('api/chat/manage', { action, id: active.id, name: newName });
      if (generation.current === selected) {
        select(action === 'archive' || action === 'resolve' ? null : managedChat(result, active.key));
      }
    } catch (e) {
      if (generation.current === selected) setError(String(e));
    } finally {
      setSending(false);
    }
  }

  async function send() {
    if ((!text.trim() && !attachments.length) || sending || adding) return;
    if (creating && !name.trim()) {
      setError('Give this chat a name.');
      return;
    }
    setSending(true);
    setError('');
    const message = text;
    const sent = generation.current;
    const draft: PendingMessage | null =
      active && !creating
        ? {
            id: crypto.randomUUID(),
            text: message,
            attachments: attachments.map((item) => item.file.name),
            state: 'sending',
          }
        : null;
    if (draft) setPending((all) => [...all, draft]);
    try {
      const files = await payload(attachments);
      if (creating || !active) {
        const result = await client.post<SessionInfo>('api/chat/manage', {
          action: 'new',
          name,
          text: message,
          refs,
          attachments: files,
        });
        if (generation.current !== sent) return;
        setText('');
        setRefs([]);
        select(managedChat(result, ''));
      } else {
        const accepted = await client.post<SessionSendResult>('api/chat/send', {
          session: active.id,
          key: active.key,
          text: message,
          refs,
          attachments: files,
        });
        if (generation.current !== sent) return;
        setPending((all) =>
          all.map((item) =>
            item.id === draft?.id ? { ...item, id: accepted.message.id, state: 'queued' as const } : item,
          ),
        );
      }
      if (generation.current === sent) {
        setText('');
        setRefs([]);
        revoke(attachments);
        putAttachments([]);
      }
    } catch (e) {
      if (generation.current === sent) {
        setPending((all) => all.filter((item) => item.id !== draft?.id));
        setError(String(e));
      }
    } finally {
      setSending(false);
      composer.current?.focus();
    }
  }

  async function stop() {
    if (!active) return;
    try {
      await client.post('api/chat/cancel', { session: active.id });
    } catch (e) {
      setError(String(e));
    }
  }

  function composerKey(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey && !e.nativeEvent.isComposing) {
      e.preventDefault();
      void send();
    }
    if (e.key === 'Escape' && !e.nativeEvent.isComposing) {
      e.preventDefault();
      void stop();
    }
  }

  const disabledAttach = sending || (!active && !creating) || (active?.state === 'archived' && !creating);
  const label = creating ? 'new chat: ' + name : active?.title || 'choose chat';

  return (
    <>
      <div className="tag chat-tag">
        <span className="muted">to:</span>
        <ChatPicker
          open={pickerOpen}
          onOpenChange={setPickerOpen}
          disabled={sending}
          label={label}
          active={active}
          creating={creating}
          chats={list}
          onSelect={select}
        />
        <Popover.Root open={newOpen} onOpenChange={setNewOpen}>
          <Popover.Trigger disabled={sending}>new</Popover.Trigger>
          <NewChat
            onStart={(chatName) => {
              submittedNew.current = true;
              setNewOpen(false);
              start('', [], chatName);
            }}
            finalFocus={() => {
              const submitted = submittedNew.current;
              submittedNew.current = false;
              return submitted ? composer.current : true;
            }}
          />
        </Popover.Root>
        <button className="push-right" onClick={onHide}>
          hide
        </button>
      </div>
      {active && !creating && (
        <div className="tag chat-toolbar" role="toolbar" aria-label="Chat actions">
          <div className="chat-actions">
            <button disabled={sending} onClick={() => void manage('rename')}>
              rename
            </button>
            <button
              title="Close this chat without resolving its ticket"
              disabled={sending || active.state === 'archived'}
              onClick={() => void manage('archive')}
            >
              {chatType(active) === 'chat' ? 'archive' : 'close'}
            </button>
            {chatType(active) === 'escalation' && (
              <>
                <button
                  title="Resolve the escalation and close its chat"
                  disabled={sending || active.state === 'archived'}
                  onClick={() => void manage('resolve')}
                >
                  resolve
                </button>
                <button
                  title="Keep this conversation as a named chat without resolving the escalation"
                  disabled={sending || active.state === 'archived'}
                  onClick={() => void manage('convert')}
                >
                  convert
                </button>
              </>
            )}
            <button onClick={() => void stop()}>stop</button>
            <button
              onClick={() => {
                const ref = window.prompt('Attach ticket, PR or session reference');
                if (ref) attach([ref]);
              }}
            >
              attach
            </button>
            <button
              onClick={() =>
                window.open(
                  client.url('api/session/' + encodeURIComponent(active.id) + '/journal'),
                  '_blank',
                  'noopener',
                )
              }
            >
              journal
            </button>
          </div>
          <output className={'chat-state' + (activity === 'working' ? ' working' : '')}>
            {active.state === 'archived' ? 'archived' : connection === 'reconnecting' ? 'reconnecting' : activity}
          </output>
        </div>
      )}
      <div className="transcript" ref={setTranscript} aria-live="polite" aria-label="Chat transcript">
        {creating ? (
          <section className="new-chat-draft" aria-label="New chat draft">
            <div>
              <strong>new chat: {name}</strong>
              <button disabled={sending} onClick={dismissDraft} aria-label="Dismiss new chat">
                dismiss ×
              </button>
            </div>
            <p>Your first message will create this chat. Nothing has been sent yet.</p>
          </section>
        ) : (
          <>
            {active?.state === 'archived' && (
              <p>
                This chat is archived. <button onClick={() => void manage('restore')}>restore</button> to continue the
                same conversation.
              </p>
            )}
            {active && (
              <Transcript key={active.id} frames={frames} viewport={transcript} pending={pending} onOlder={prepend} />
            )}
            {commands.map((entry, i) => (
              <p className="tool mono" key={i}>
                desk ▸ {entry.op}{' '}
                {Object.entries(entry.args || {})
                  .map(([k, v]) => k + '=' + v)
                  .join(' ')}
              </p>
            ))}
          </>
        )}
      </div>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <form
        className="composer"
        onSubmit={(e) => {
          e.preventDefault();
          void send();
        }}
      >
        <input
          type="file"
          hidden
          multiple
          ref={fileInput}
          aria-label="Attach files"
          accept="image/png,image/jpeg,image/gif,text/*,.txt,.md,.json,.yaml,.yml,.log,.csv,.xml"
          onChange={() => {
            const files = Array.from(fileInput.current?.files || []);
            if (fileInput.current) fileInput.current.value = '';
            void addFiles(files);
          }}
        />
        <div className="attachment-actions">
          <button
            type="button"
            title="Attach images or text files"
            disabled={disabledAttach}
            onClick={() => fileInput.current?.click()}
          >
            attach
          </button>
          <button
            type="button"
            title="Paste text as a document"
            disabled={disabledAttach}
            onClick={() => setPasteOpen(true)}
          >
            paste
          </button>
        </div>
        {attachments.length > 0 && (
          <ul className="chat-attachments" aria-label="Message attachments">
            {attachments.map((item, i) => (
              <li key={item.preview || item.file.name + i}>
                {item.preview && <img src={item.preview} alt={item.file.name} />}
                <span>
                  {item.file.name}
                  <small>
                    {Math.max(1, Math.ceil(item.file.size / 1024))} KiB · {item.preview ? 'image' : 'read on demand'}
                  </small>
                </span>
                <button
                  type="button"
                  disabled={sending}
                  aria-label={'Remove attachment ' + item.file.name}
                  onClick={() => removeAttachment(item)}
                >
                  ×
                </button>
              </li>
            ))}
          </ul>
        )}
        {adding > 0 && <output>Preparing attachments…</output>}
        {refs.length > 0 && (
          <div className="context">
            Attached work:{' '}
            {refs.map((ref) => (
              <button
                key={ref}
                title="Remove attached reference"
                aria-label={'Remove attached reference ' + ref}
                onClick={() => setRefs((all) => all.filter((r) => r !== ref))}
              >
                {ref} ×
              </button>
            ))}
          </div>
        )}
        <div className="compose-line">
          <textarea
            id="chat-message"
            aria-label={creating ? 'First message to new chat: ' + name : 'to ' + (active?.title || 'new chat')}
            ref={composer}
            value={text}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={composerKey}
            onPaste={paste}
            placeholder={active || creating ? 'message' : 'choose new to start a named chat'}
            disabled={sending || (!active && !creating) || (active?.state === 'archived' && !creating)}
            rows={3}
          />
          <button
            className="send"
            type="submit"
            disabled={sending || adding > 0 || (!text.trim() && !attachments.length)}
          >
            {sending ? 'sending…' : creating ? 'create & send' : 'send'}
          </button>
        </div>
      </form>
      <PasteDocument
        open={pasteOpen}
        onOpenChange={setPasteOpen}
        disabled={sending || adding > 0}
        onAttach={async (content) => {
          const ok = await addFiles([
            new File([content], `pasted-text-${++pasteNumber.current}.txt`, { type: 'text/plain' }),
          ]);
          if (ok) {
            setPasteOpen(false);
            composer.current?.focus();
          }
          return ok;
        }}
      />
    </>
  );
}

function NewChat({
  onStart,
  finalFocus,
}: {
  onStart: (name: string) => void;
  finalFocus: () => HTMLElement | boolean | null;
}) {
  const [draftName, setDraftName] = useState('');
  const input = useRef<HTMLInputElement>(null);
  return (
    <Popover.Portal>
      <Popover.Positioner side="bottom" align="start" sideOffset={4} collisionPadding={8} className="popup-positioner">
        <Popover.Popup className="new-chat-popover" aria-label="New chat" initialFocus={input} finalFocus={finalFocus}>
          <form
            onSubmit={(event) => {
              event.preventDefault();
              const chosen = draftName.trim();
              if (!chosen) return;
              setDraftName('');
              onStart(chosen);
            }}
          >
            <label htmlFor="new-chat-name">name</label>
            <input
              id="new-chat-name"
              ref={input}
              value={draftName}
              onChange={(e) => setDraftName(e.target.value)}
              maxLength={120}
              placeholder="ops"
              required
            />
            <button type="submit">start</button>
            <Popover.Close type="button">cancel</Popover.Close>
          </form>
        </Popover.Popup>
      </Popover.Positioner>
    </Popover.Portal>
  );
}

function ChatPicker({
  open,
  onOpenChange,
  disabled,
  label,
  active,
  creating,
  chats,
  onSelect,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  disabled: boolean;
  label: string;
  active: ChatInfo | null;
  creating: boolean;
  chats: ChatInfo[];
  onSelect: (chat: ChatInfo) => void;
}) {
  const [search, setSearch] = useState('');
  const [archived, setArchived] = useState(false);
  const body = useRef<HTMLDivElement>(null);
  const searchInput = useRef<HTMLInputElement>(null);
  const archivedCount = useMemo(() => chats.filter((chat) => chat.state === 'archived').length, [chats]);
  const pool = useMemo(() => {
    const all = active && !chats.some((chat) => chat.id === active.id) ? [...chats, active] : chats;
    return all.filter((chat) => (chat.state === 'archived') === archived);
  }, [chats, active, archived]);
  const searcher = useMemo(() => createSearcher(pool, chatSearch), [pool]);
  const shown = useMemo(
    () => (search.trim() ? searcher.search(search) : pool.toSorted(byRecent)),
    [pool, searcher, search],
  );

  return (
    <Popover.Root
      open={open}
      onOpenChange={(next) => {
        if (next) {
          setSearch('');
          setArchived(active?.state === 'archived');
        }
        onOpenChange(next);
      }}
    >
      <Popover.Trigger className="chat-picker-button" aria-label="Choose chat" disabled={disabled} title={label}>
        {label} ▾
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Positioner
          side="bottom"
          align="start"
          sideOffset={4}
          collisionPadding={8}
          className="popup-positioner"
        >
          <Popover.Popup className="chat-picker" aria-label="Choose chat" initialFocus={searchInput}>
            <div className="chat-picker-search">
              <input
                ref={searchInput}
                type="search"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                aria-label="Search chats"
                placeholder="search chats"
                onKeyDown={(event) => {
                  if (event.key === 'Enter' && shown[0]) {
                    event.preventDefault();
                    onSelect(shown[0]);
                  }
                  if (event.key === 'ArrowDown') {
                    event.preventDefault();
                    body.current?.querySelector<HTMLButtonElement>('tbody button')?.focus();
                  }
                }}
              />
            </div>
            <div className="chat-picker-filter">
              <button
                className={!archived ? 'chosen' : undefined}
                aria-pressed={!archived}
                onClick={() => setArchived(false)}
              >
                recent
              </button>
              <button
                className={archived ? 'chosen' : undefined}
                aria-pressed={archived}
                onClick={() => setArchived(true)}
              >
                archived ({archivedCount})
              </button>
              <span className="muted">newest first</span>
            </div>
            <div className="chat-archive" ref={body}>
              <table className="rows" aria-label={archived ? 'Archived chats' : 'Recent chats'}>
                <thead>
                  <tr>
                    <th scope="col">chat</th>
                    <th scope="col">type</th>
                    <th scope="col">updated</th>
                  </tr>
                </thead>
                <tbody>
                  {shown.map((chat) => {
                    const chosen = !creating && active?.id === chat.id;
                    return (
                      <tr key={chat.id} className={chosen ? 'chosen' : undefined}>
                        <td>
                          <button aria-pressed={chosen} onClick={() => onSelect(chat)}>
                            {chat.title}
                            {chat.state === 'working' ? ' ●' : ''}
                          </button>
                        </td>
                        <td className="muted">{chatType(chat)}</td>
                        <td>
                          {chat.updated ? (
                            <time dateTime={chat.updated}>{new Date(chat.updated).toLocaleDateString()}</time>
                          ) : (
                            '–'
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
              {!shown.length && (
                <p className="muted">
                  {search ? 'No matching chats.' : archived ? 'No archived chats.' : 'No chats yet. Choose New.'}
                </p>
              )}
            </div>
          </Popover.Popup>
        </Popover.Positioner>
      </Popover.Portal>
    </Popover.Root>
  );
}

function PasteDocument({
  open,
  onOpenChange,
  disabled,
  onAttach,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  disabled: boolean;
  onAttach: (content: string) => Promise<boolean>;
}) {
  const [content, setContent] = useState('');
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Backdrop className="dialog-backdrop" />
        <Dialog.Popup className="paste-document" aria-label="Paste document">
          <label htmlFor="document-text">paste text to attach as a document</label>
          <textarea id="document-text" value={content} onChange={(e) => setContent(e.target.value)} rows={10} />
          <div>
            <button
              disabled={!content || disabled}
              onClick={async () => {
                if (await onAttach(content)) setContent('');
              }}
            >
              attach document
            </button>
            <Dialog.Close>cancel</Dialog.Close>
          </div>
        </Dialog.Popup>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
