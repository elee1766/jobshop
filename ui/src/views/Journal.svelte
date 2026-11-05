<script>
  import { get } from '../lib/api.svelte.js';
  import Chat from './Chat.svelte';

  /** Attach to a live session: read-only, or read/write when its pool has attach: rw. */
  /** @type {{ id: string }} */
  let { id } = $props();

  let access = $state({ write: false, reason: '' });
  $effect(() => {
    access = { write: false, reason: '' };
    get('api/session/' + encodeURIComponent(id) + '/access').then(
      (a) => { access = { write: !!a.write, reason: a.reason || '' }; },
      () => {},
    );
  });
</script>

<Chat {id} write={access.write} reason={access.reason} />
