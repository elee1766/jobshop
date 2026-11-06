// Test plugin for e2e/plugins.spec.js (JOB-87): adds to the built-in
// Work and Chat views through their plugin-defined slots instead of
// replacing them. Plain JS, no build step.

class Button extends HTMLElement {
  connectedCallback() {
    const b = document.createElement('button');
    b.dataset.testid = 'e2e-slot-button';
    b.textContent = 'E2E action';
    b.addEventListener('click', () => this.api.toast('e2e toolbar ' + (this.context?.state || 'all')));
    this.replaceChildren(b);
  }
}

class Aside extends HTMLElement {
  set context(c) {
    this.dataset.testid = 'e2e-slot-aside';
    this.className = 'card';
    this.textContent = 'aside runs ' + (c?.work ? c.work.runs.length : '…');
  }
}

class Panel extends HTMLElement {
  connectedCallback() {
    this.dataset.testid = 'e2e-slot-panel';
    this.className = 'card';
    this.style.display = 'block';
    this.style.padding = '12px';
    this.textContent = 'E2E chat panel';
  }
}

class Action extends HTMLElement {
  set context(c) {
    this.dataset.testid = 'e2e-slot-action';
    this.textContent = 'E2E ' + (c?.session || '?');
  }
}

customElements.define('e2e-slots-button', Button);
customElements.define('e2e-slots-aside', Aside);
customElements.define('e2e-slots-panel', Panel);
customElements.define('e2e-slots-action', Action);
