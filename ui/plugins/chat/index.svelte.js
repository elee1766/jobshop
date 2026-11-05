// Built-in chat plugin (#535): the session panel body (slot session), the
// Chat view and a drawer section. Each element mounts an ordinary Svelte
// component; CSS loads through a <link> because the CSP has no inline
// styles. Everything goes through the api object (#517).
import { mount, unmount } from 'svelte';
import Pane from './Pane.svelte';
import ChatList from './ChatList.svelte';
import DrawerChat from './DrawerChat.svelte';

const css = new URL(/* @vite-ignore */ './style.css', import.meta.url).href;

/** @param {any} Component */
function element(Component) {
  return class extends HTMLElement {
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
      this.app = mount(Component, { target: this, props: this.props });
    }
    disconnectedCallback() {
      if (this.app) unmount(this.app);
      this.app = null;
    }
  };
}

for (const [name, c] of /** @type {[string, any][]} */ ([['desk-chat-pane', Pane], ['desk-chat-view', ChatList], ['desk-chat-drawer', DrawerChat]])) {
  if (!customElements.get(name)) customElements.define(name, element(c));
}
