// Session events to transcript blocks, after jcode-web's
// transcript.svelte.ts (#535). Events come from api.session.stream (live)
// and api.session.older (history, "Load older"); both are the
// controller's decoded journal (rpc.SessionEvent).

/**
 * @typedef {{ kind: 'user', text: string, source: string, context: string, seq: number }} UserBlock
 * @typedef {{ kind: 'assistant', text: string, live: boolean, seq: number }} AssistantBlock
 * @typedef {{ kind: 'tool', id: string, name: string, input: string, output: string, state: string, seq: number }} ToolBlock
 * @typedef {{ kind: 'system', text: string, seq: number }} SystemBlock
 * @typedef {{ kind: 'error', text: string, seq: number }} ErrorBlock
 * @typedef {UserBlock | AssistantBlock | ToolBlock | SystemBlock | ErrorBlock} Block
 */

const messagesHeader = 'Messages for you:';

/** ACP tool status to the row state. @param {string} s */
function toolState(s) {
  switch (s) {
    case 'pending': return 'input';
    case 'in_progress': return 'running';
    case 'completed': return 'done';
    case 'failed': return 'failed';
    default: return s || 'input';
  }
}

/**
 * Who a prompt came from: messages delivered by the controller start
 * with its header and name each sender; anything else is the
 * controller's own prompt (ticket, reminder, pool prompt).
 * @param {string} text
 */
export function promptSource(text) {
  if (!text.startsWith(messagesHeader)) return { source: 'controller', text };
  const from = /\n\nFrom ([^\s(]+) \(/.exec(text);
  const who = from ? from[1] : '';
  const body = text.slice(messagesHeader.length).replace(/^\s*\n\nFrom [^\n]*\n/, '').replace(/\n\nFrom [^\n]*\n/g, '\n\n');
  return { source: who.startsWith('desk:') ? 'desk' : who ? 'operator' : 'message', text: body.trim(), who };
}

/** Builds blocks from events. Used for live and older pages alike. Blocks
 * are deep state, so a tool row or streaming text updates in place. */
export class Builder {
  /** @type {Block[]} */
  blocks = $state([]);
  busy = $state(false);
  /** @type {{ used: number, size: number } | null} */
  usage = $state(null);

  /** @template {Block['kind']} K @param {K} kind @returns {Extract<Block, { kind: K }> | undefined} */
  last(kind) {
    const b = this.blocks[this.blocks.length - 1];
    return b && b.kind === kind ? /** @type {any} */ (b) : undefined;
  }

  /** @param {string} id */
  tool(id) {
    for (let i = this.blocks.length - 1; i >= 0; i--) {
      const b = this.blocks[i];
      if (b.kind === 'tool' && b.id === id) return b;
    }
    return undefined;
  }

  settle() {
    for (const b of this.blocks) if (b.kind === 'assistant' && b.live) b.live = false;
  }

  /** @param {any} ev */
  apply(ev) {
    const seq = ev.seq || 0;
    switch (ev.type) {
      case 'turn_start': {
        this.settle();
        this.busy = true;
        const p = promptSource(ev.text || '');
        if (ev.context) this.blocks.push({ kind: 'system', text: 'Context given to the session:\n' + ev.context, seq });
        if (p.text) this.blocks.push({ kind: 'user', text: p.text, source: p.source, context: ev.context || '', seq });
        break;
      }
      case 'text': {
        this.busy = true;
        const b = this.last('assistant');
        if (b?.live) b.text += ev.text;
        else this.blocks.push({ kind: 'assistant', text: ev.text || '', live: true, seq });
        break;
      }
      case 'reasoning':
        // Thinking is not shown; it still means the agent is working.
        this.busy = true;
        break;
      case 'tool': {
        const tl = ev.tool;
        if (!tl) break;
        this.busy = true;
        const id = tl.id || tl.title;
        const t = this.tool(id);
        if (t) {
          if (tl.title) t.name = tl.title;
          t.state = toolState(tl.status);
          if (tl.input) t.input = tl.input;
          if (tl.output) t.output = tl.output;
        } else {
          this.settle();
          this.blocks.push({ kind: 'tool', id, name: tl.title || id, input: tl.input || '', output: tl.output || '', state: toolState(tl.status), seq });
        }
        break;
      }
      case 'turn_end':
        this.settle();
        this.busy = false;
        if (ev.error) this.blocks.push({ kind: 'error', text: 'Turn failed: ' + ev.error, seq });
        else if (ev.stop_reason === 'cancelled') this.blocks.push({ kind: 'system', text: 'Stopped', seq });
        else if (ev.stop_reason && ev.stop_reason !== 'end_turn') this.blocks.push({ kind: 'system', text: 'Turn ended: ' + ev.stop_reason, seq });
        break;
      case 'usage':
        if (ev.usage) this.usage = ev.usage;
        break;
      case 'state':
        if (ev.status === 'closed') {
          this.busy = false;
          this.blocks.push({ kind: 'system', text: 'Session closed' + (ev.resolution ? ' · ' + ev.resolution : ''), seq });
        }
        break;
      default:
        return false;
    }
    return true;
  }
}

/** One session's transcript: live stream plus older pages. */
export class Transcript {
  status = $state('connecting');
  /** The session's resolution once a state event said it closed, else ''. */
  ended = $state('');
  /** Lowest journal seq loaded; older entries are 1..first-1. */
  first = $state(0);
  loadingOlder = $state(false);
  error = $state('');
  /** Bumped on every change, for scroll following. */
  version = $state(0);

  #b = new Builder();
  /** @type {(() => void) | null} */
  #close = null;

  get blocks() { return this.#b.blocks; }
  get busy() { return this.#b.busy; }
  set busy(v) { this.#b.busy = v; }
  get usage() { return this.#b.usage; }

  /** @param {any} api @param {string} id */
  constructor(api, id) {
    this.api = api;
    this.id = id;
  }

  open(tail = 200) {
    this.close();
    this.#close = this.api.session.stream(this.id, { tail }, (/** @type {{ type: string, event?: any }} */ m) => {
      if (m.type === 'open') this.status = 'live';
      else if (m.type === 'end') { this.status = 'closed'; this.busy = false; }
      else if (m.type === 'error') { if (this.status !== 'closed') this.status = 'reconnecting'; }
      else if (m.type === 'event') {
        this.#apply(m.event);
        // The session's state said closed: it shows its end, never working.
        if (m.event?.type === 'state' && m.event.status === 'closed') this.ended = m.event.resolution || 'closed';
      }
    });
  }

  close() {
    this.#close?.();
    this.#close = null;
  }

  /** @param {any} ev */
  #apply(ev) {
    if (ev.seq && (!this.first || ev.seq < this.first)) this.first = ev.seq;
    if (!this.#b.apply(ev)) return;
    this.version++;
  }

  /** Earlier journal entries than the oldest loaded. */
  get earlier() {
    return this.first > 1 ? this.first - 1 : 0;
  }

  /** Prepends the page before the oldest loaded event. */
  async loadOlder() {
    if (this.loadingOlder || !this.earlier) return;
    this.loadingOlder = true;
    this.error = '';
    try {
      const page = await this.api.session.older(this.id, this.first, 400);
      const old = new Builder();
      for (const ev of page.events || []) old.apply(ev);
      old.settle();
      this.#b.blocks = [...$state.snapshot(old.blocks), ...this.#b.blocks];
      this.first = Math.max(1, page.first || 1);
      this.version++;
    } catch (e) {
      this.error = String(/** @type {any} */ (e).message || e);
    } finally {
      this.loadingOlder = false;
    }
  }

  /** Text of the user prompts, oldest first, for prompt history. */
  prompts() {
    return this.blocks.filter((b) => b.kind === 'user' && b.source !== 'controller').map((b) => /** @type {UserBlock} */ (b).text);
  }
}
