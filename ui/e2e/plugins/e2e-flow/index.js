// Test plugin for e2e/plugins.spec.js (#542): one view that walks the
// api end to end. Each step writes its result into a data-testid line.
// Needs desk_commands.e2e-echo (command: cat) in jobshop.yaml; see
// docs/verify/desk-plugin-commands.md.

class Flow extends HTMLElement {
  connectedCallback() {
    const steps = ['create', 'list', 'event', 'exec', 'stream'];
    this.innerHTML = '<h1>E2E flow</h1>' + steps.map((s) => `<div data-testid="flow-${s}">…</div>`).join('');
    this.run();
  }

  disconnectedCallback() {
    this.off?.();
    this.es?.close();
  }

  /** @param {string} step @param {string} text */
  set(step, text) {
    const el = /** @type {HTMLElement} */ (this.querySelector(`[data-testid=flow-${step}]`));
    el.textContent = text;
  }

  async run() {
    const api = this.api;
    const tag = 'e2e-' + Date.now().toString(36);
    // Subscribe before the update so the event cannot be missed.
    /** @type {any} */
    let t;
    const gotEvent = new Promise((resolve) => {
      this.off = api.events({ kind: 'ticket', op: 'comment' }, (ev) => { if (t && ev.id === t.id) resolve(ev); });
    });
    try {
      t = await api.call('ticket.create', { title: 'e2e plugin ' + tag, meta: { e2e: { tag } } });
      this.set('create', 'create ok');
    } catch (e) { this.set('create', 'create failed ' + e); return; }
    try {
      const list = await api.call('ticket.list', { newest: true, limit: 50 });
      const hit = list.filter((/** @type {any} */ x) => x.meta?.e2e?.tag === tag);
      this.set('list', hit.length === 1 && hit[0].id === t.id ? 'list ok' : 'list missing ' + hit.length);
    } catch (e) { this.set('list', 'list failed ' + e); }
    try {
      await api.call('comment.add', { id: t.id, body: 'from the e2e plugin' });
      const ev = await Promise.race([gotEvent, new Promise((_, rej) => setTimeout(() => rej(new Error('no event in 10s')), 10000))]);
      this.set('event', 'event ok ' + ev.op);
    } catch (e) { this.set('event', 'event failed ' + e); }
    try {
      const out = await api.exec('e2e-echo', { tag }, { ticket: t.id });
      this.set('exec', out?.tag === tag ? 'exec ok' : 'exec wrong ' + JSON.stringify(out));
    } catch (e) { this.set('exec', 'exec failed ' + e); }
    try {
      await api.call('ticket.close', { id: t.id, resolution: 'done' });
    } catch { /* cleanup only */ }
    this.stream();
  }

  async stream() {
    try {
      const sessions = await this.api.get('api/sessions');
      const s = sessions.find((/** @type {any} */ x) => x.ticket?.status !== 'closed') || sessions[0];
      if (!s) { this.set('stream', 'stream skip no sessions'); return; }
      this.es = new EventSource(this.api.url('api/session/' + encodeURIComponent(s.id) + '/stream?tail=5'));
      let frames = 0;
      const done = (/** @type {string} */ msg) => { this.set('stream', msg); this.es?.close(); };
      this.es.addEventListener('open', () => setTimeout(() => done('stream ok ' + frames), 1500));
      this.es.addEventListener('session', () => { frames++; });
      this.es.addEventListener('end', () => done('stream ok ' + frames));
    } catch (e) { this.set('stream', 'stream failed ' + e); }
  }
}

customElements.define('e2e-flow-view', Flow);
