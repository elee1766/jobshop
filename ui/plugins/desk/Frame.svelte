<script>
  // The built-in Desk app (JOB-85): the page column (main view and session
  // console), the ticket drawer, and the docked nav. The nav is a plugin
  // (JOB-86): this frame mounts the last `nav` contribution, docks it by
  // its placement (left, right, top, bottom, or none to let it position
  // itself) and switches placement at the phone breakpoint. With no nav
  // enabled the page column fills the screen. Everything goes through the
  // api, so a replacement app can do the same.
  import { params, slot } from '../../kit/plugin.svelte.js';
  import { panel } from './panel.svelte.js';
  import Console from './Console.svelte';
  import Drawer from './Drawer.svelte';

  /** @type {{ api: any, context?: any, config?: any }} */
  let { api } = $props();
  const url = params(() => api);

  // Contributions are fixed once the shell has loaded plugins, which it
  // does before mounting the app.
  /** @type {{ plugin: string, element: string, id?: string, title?: string }[]} */
  const pviews = $derived(api.contributions('view'));
  const view = $derived(url.view || 'attention');
  const pview = $derived(pviews.find((v) => v.id === view));

  // One nav, the last by plugin name, as for session.
  /** @type {{ plugin: string, element: string, placement?: { wide?: string, narrow?: string } } | undefined} */
  const nav = $derived(api.contributions('nav').at(-1));
  const narrow = $derived(panel.mobile);
  const dock = $derived(nav ? (narrow ? nav.placement?.narrow || 'bottom' : nav.placement?.wide || 'left') : 'none');
  // The brand sits above a side nav, and is the top bar over a phone's bottom nav.
  const brand = $derived(dock === 'left' || dock === 'right' || (narrow && dock === 'bottom'));

  /** @param {KeyboardEvent} e */
  function onkey(e) {
    // The ticket drawer sits on top and closes first; it marks its Escape handled.
    if (e.key === 'Escape' && url.session && !url.ticket && !e.defaultPrevented) api.closeSession();
  }
</script>

<svelte:window onkeydown={onkey} />

<div class="shell" class:narrow class:hasbrand={brand} data-testid="desk-app" data-dock={dock}>
  {#if brand}<div class="brand"><span class="mark"></span>Jobshop <span class="faint">Desk</span></div>{/if}
  {#if nav}
    {#key nav.plugin + nav.element}
      <div class="navhost plugin-host" data-plugin={nav.plugin} data-testid="nav-host" use:slot={{ api, slot: 'nav', target: { plugin: nav.plugin, element: nav.element }, context: { dock, narrow } }}></div>
    {/key}
  {/if}
  <div class="col">
    <main>
      {#if pview}
        {#key pview.plugin + pview.element}
          <div class="plugin-host" data-plugin={pview.plugin} use:slot={{ api, slot: 'view', target: { plugin: pview.plugin, element: pview.element } }}></div>
        {/key}
      {:else}<p class="faint pad" data-testid="view-missing">No plugin shows <span class="mono">{view}</span>.</p>
      {/if}
    </main>
    <Console {api} {url} />
  </div>
  {#if url.ticket}<Drawer {api} id={url.ticket} />{/if}
</div>

<style>
  .shell {
    position: fixed; left: 0; right: 0; top: var(--vvtop, 0px); height: var(--vvh, 100dvh);
    display: grid; align-content: start;
    grid-template-columns: minmax(0, 1fr); grid-template-rows: minmax(0, 1fr);
    grid-template-areas: 'col';
  }
  .shell > * { min-width: 0; }
  .brand { grid-area: brand; display: flex; align-items: center; gap: 8px; padding: 14px 16px 10px; font-weight: 600; font-size: 13px; background: var(--side); }
  .mark { width: 14px; height: 14px; border-radius: 4px; background: var(--accent); }
  .navhost { grid-area: nav; min-width: 0; min-height: 0; display: flex; flex-direction: column; }
  /* A bar takes its width from the page, so the nav can adapt to it (@container in the nav).
     Not on a phone's bottom dock: a container is a containing block for position: fixed, and
     the built-in tab bar draws its status controls into the top bar that way. */
  .shell[data-dock='top'] .navhost, .shell:not(.narrow)[data-dock='bottom'] .navhost { container: nav / inline-size; }
  /* "none": the nav positions itself (a floating menu, a palette). */
  .shell[data-dock='none'] .navhost { display: contents; }
  .col { grid-area: col; min-height: 0; display: flex; flex-direction: column; }
  main { flex: 1; min-height: 0; overflow-y: auto; overscroll-behavior: contain; display: flex; flex-direction: column; background: var(--bg); }

  .shell[data-dock='left'] { grid-template-columns: auto minmax(0, 1fr); grid-template-rows: minmax(0, 1fr); grid-template-areas: 'nav col'; }
  .shell[data-dock='left'].hasbrand { grid-template-rows: auto minmax(0, 1fr); grid-template-areas: 'brand col' 'nav col'; }
  .shell[data-dock='left'] .brand { border-right: 1px solid var(--border); }
  .shell[data-dock='right'] { grid-template-columns: minmax(0, 1fr) auto; grid-template-rows: minmax(0, 1fr); grid-template-areas: 'col nav'; }
  .shell[data-dock='right'].hasbrand { grid-template-rows: auto minmax(0, 1fr); grid-template-areas: 'col brand' 'col nav'; }
  .shell[data-dock='right'] .brand { border-left: 1px solid var(--border); }
  .shell[data-dock='top'] { grid-template-rows: auto minmax(0, 1fr); grid-template-areas: 'nav' 'col'; }
  .shell[data-dock='bottom'] { grid-template-rows: minmax(0, 1fr) auto; grid-template-areas: 'col' 'nav'; }

  @media (max-width: 768px) {
    .shell[data-dock='bottom'].hasbrand {
      grid-template-rows: auto minmax(0, 1fr) auto; grid-template-areas: 'brand' 'col' 'nav';
    }
    /* The top bar: the brand, with the nav's status controls drawn over its right end. */
    .shell[data-dock='bottom'] .brand {
      padding: env(safe-area-inset-top) 12px 0; height: calc(40px + env(safe-area-inset-top)); box-sizing: border-box;
      border-bottom: 1px solid var(--border);
    }
  }
</style>
