// Stands in for packs/deploy's Deploys view in e2e/plugins.spec.js (#577).
customElements.define('e2e-deploys-view', class extends HTMLElement {
  connectedCallback() { this.innerHTML = '<h1 data-testid="user-deploys">User deploys</h1>'; }
});
