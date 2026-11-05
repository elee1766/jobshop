// Entry of the built-in deploys plugin (#517). Defines the element by
// hand with mount(), so the component stays an ordinary Svelte
// component and no customElement compile mode is needed. The build
// emits its CSS as style.css next to this file; a <link> loads it
// because the CSP has no 'unsafe-inline' for styles.
import { mount, unmount } from 'svelte';
import Deploys from './Deploys.svelte';

const css = new URL(/* @vite-ignore */ './style.css', import.meta.url).href;

class DeskDeploysView extends HTMLElement {
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
    this.app = mount(Deploys, { target: this, props: this.props });
  }
  disconnectedCallback() {
    if (this.app) unmount(this.app);
    this.app = null;
  }
}

if (!customElements.get('desk-deploys-view')) customElements.define('desk-deploys-view', DeskDeploysView);
