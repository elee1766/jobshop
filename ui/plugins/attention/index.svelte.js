// Built-in Attention plugin (#535). The element mounts an ordinary Svelte
// component; CSS loads through a <link> (the CSP has no inline styles).
import { mount, unmount } from 'svelte';
import Attention from './Attention.svelte';

const css = new URL(/* @vite-ignore */ './style.css', import.meta.url).href;

class DeskAttentionView extends HTMLElement {
  props = $state({ api: /** @type {any} */ (null), context: /** @type {any} */ ({}), config: /** @type {any} */ ({}) });
  /** @type {Record<string, any> | null} */
  app = null;
  set api(v) { this.props.api = v; }
  get api() { return this.props.api; }
  set context(v) { this.props.context = v; }
  get context() { return this.props.context; }
  set config(v) { this.props.config = v; }
  get config() { return this.props.config; }
  connectedCallback() {
    if (!document.querySelector(`link[href="${css}"]`)) {
      const l = document.createElement('link');
      l.rel = 'stylesheet';
      l.href = css;
      document.head.append(l);
    }
    this.style.display = 'contents';
    this.app = mount(Attention, { target: this, props: this.props });
  }
  disconnectedCallback() {
    if (this.app) unmount(this.app);
    this.app = null;
  }
}

if (!customElements.get('desk-attention-view')) customElements.define('desk-attention-view', DeskAttentionView);
