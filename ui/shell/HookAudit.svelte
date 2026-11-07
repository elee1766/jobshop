<script lang="ts">
  import type { AutomationAudit, AutomationList, AutomationRecord } from '../../packages/jobshop-client/index.js';
  let { summaries, selected, detail, onselect }: {
    summaries?: AutomationList;
    selected: string;
    detail: AutomationAudit | { id: string; error: string } | null;
    onselect: (id: string) => void;
  } = $props();
  let filter = $state('');
  const items = $derived((summaries?.items || []).filter(item =>
    `${item.key} ${item.state} ${item.trigger || ''} ${item.run || ''} ${item.event?.delivery || ''}`.toLowerCase().includes(filter.toLowerCase())));
  const audit = $derived(detail && !('error' in detail) && detail.id === selected ? detail : null);
  const invocation = $derived(audit?.invocation.value || {});
  const unknown = $derived(audit?.calls.some(call => call.state.toUpperCase() === 'UNKNOWN') || audit?.attempts.some(attempt => attempt.state.toUpperCase() === 'UNKNOWN'));
  const tickets = $derived.by(() => {
    const linked = new Map<string, string>();
    if (invocation.run) linked.set(invocation.run, 'run');
    if (invocation.step) linked.set(invocation.step, 'step');
    if (audit?.invocation.state === 'done') {
      for (const target of Array.isArray(invocation.targets) ? invocation.targets : []) {
        if (typeof target === 'string') linked.set(target, 'ticket');
      }
    }
    return [...linked].map(([id, label]) => ({ id, label }));
  });

  function lookup(event: SubmitEvent) {
    event.preventDefault();
    const id = new FormData(event.currentTarget as HTMLFormElement).get('invocation')?.toString().trim();
    if (id) onselect(id);
  }
  function duration(value: Record<string, any>) {
    const ns = value.elapsed_ns || value.observation?.duration_ns;
    if (typeof ns === 'number') return (ns / 1e6).toFixed(1) + ' ms';
    if (value.started_at && value.finished_at) return Math.max(0, Date.parse(value.finished_at) - Date.parse(value.started_at)).toFixed(1) + ' ms';
    return 'unfinished';
  }
  function payload(value: unknown) {
    const text = JSON.stringify(value, null, 2) || 'null';
    return text.length > 65536 ? text.slice(0, 65536) + '\n[display limited to 64 KiB]' : text;
  }
  function stamp(value: string) {
    return value && !value.startsWith('0001-') ? new Date(value).toLocaleString() : 'not recorded';
  }
  function started(record: AutomationRecord) {
    return record.value.started_at || record.created_at;
  }
</script>

<section class="hook-audit" aria-label="Automation audit">
  <h2>automation audit</h2>
  <p class="muted">Events are recorded before tickets exist. Select an invocation to inspect its attempts and capability calls.</p>
  <div class="audit-controls">
    <label>filter <input aria-label="Filter automation audit" bind:value={filter} placeholder="state, trigger, run or delivery" /></label>
    <form onsubmit={lookup}>
      <label>invocation <input name="invocation" aria-label="Automation invocation ID" value={selected} required /></label>
      <button type="submit">inspect</button>
    </form>
  </div>
  <div class="table-scroll">
    <table aria-label="Automation invocations">
      <thead><tr><th>invocation</th><th>event / run</th><th>state</th><th>attempts</th><th>updated</th></tr></thead>
      <tbody>
        {#each items as item (item.key)}
          <tr class:chosen={selected === item.key}>
            <td><button class="audit-id" aria-label={'Inspect invocation ' + item.key} aria-pressed={selected === item.key} onclick={() => onselect(item.key)}>{item.key.slice(0, 12)}</button><small>{item.trigger || item.entry.split('/').at(-1)}</small></td>
            <td>{#if item.event}<span>{item.event.source} / {item.event.event}</span><small>{item.event.delivery}</small>{:else if item.run}<a href={'?view=trace&ref=' + encodeURIComponent(item.run)}>{item.run}</a>{:else}<span class="muted">no ticket</span>{/if}</td>
            <td class:warn={item.state === 'failed' || item.state.toUpperCase() === 'UNKNOWN'}>{item.state}</td>
            <td>{item.attempts}</td>
            <td><time datetime={item.updated_at}>{stamp(item.updated_at)}</time></td>
          </tr>
        {:else}
          <tr><td colspan="5" class="muted">{filter ? 'No matching invocations.' : 'No automation invocations.'}</td></tr>
        {/each}
      </tbody>
    </table>
  </div>
  {#if summaries?.truncated}<p class="muted">Showing the latest 50 invocations. Inspect an older invocation by its ID.</p>{/if}
  {#if selected}
    <section class="audit-detail" aria-label="Automation invocation detail">
      <div class="audit-heading"><h3>invocation <span class="mono">{selected}</span></h3><button onclick={() => onselect('')}>dismiss</button></div>
      {#if detail && 'error' in detail}
        <p role="alert" class="error">{detail.error}</p>
      {:else if !audit}
        <p role="status">Loading invocation…</p>
      {:else}
        <p><b>{audit.invocation.state}</b> · revision {audit.invocation.revision} · {audit.invocation.attempts} attempts</p>
        {#if invocation.parent_invocation}<p>parent invocation <button class="mono" onclick={() => onselect(invocation.parent_invocation)}>{invocation.parent_invocation}</button></p>{/if}
        <p class="mono">{invocation.entry}</p>
        <p class="muted">created <time datetime={audit.invocation.created_at}>{stamp(audit.invocation.created_at)}</time> · updated <time datetime={audit.invocation.updated_at}>{stamp(audit.invocation.updated_at)}</time></p>
        {#if invocation.event}
          <h3>event</h3>
          <dl><dt>source / event</dt><dd>{invocation.event.source} / {invocation.event.event}</dd><dt>delivery</dt><dd class="mono">{invocation.event.delivery}</dd><dt>trigger</dt><dd>{invocation.trigger}</dd></dl>
        {/if}
        {#if unknown}<p class="warn" role="status">UNKNOWN: an attempt or external call has no confirmed outcome. Reconcile it before retrying.</p>{/if}
        {#if invocation.error}<p class="error">{invocation.error}</p>{/if}
        {#if invocation.artifact}<p>snapshot artifact reference: <code>{invocation.artifact}</code></p>{/if}
        <details><summary>event identity</summary><pre>{payload(invocation.event)}</pre></details>
        {#if invocation.decisions}
          <h3>decisions</h3>
          {#if Array.isArray(invocation.decisions)}
            <ol>{#each invocation.decisions as decision}<li><b>{decision.choice}</b>{#if decision.defaulted} · defaulted{/if}{#if decision.confidence != null} · confidence {decision.confidence}{/if}<p>{decision.reason}</p></li>{/each}</ol>
          {:else}<pre>{payload(invocation.decisions)}</pre>{/if}
        {/if}
        <h3>attempts</h3>
        <ol class="audit-attempts">
          {#each audit.attempts as attempt (attempt.key)}
            <li><b>attempt {attempt.value.attempt}</b> · <span class:warn={attempt.state.toUpperCase() === 'UNKNOWN'}>{attempt.state}</span> · {duration(attempt.value)}<p class="muted"><time datetime={started(attempt)}>{stamp(started(attempt))}</time>{#if attempt.value.finished_at} → <time datetime={attempt.value.finished_at}>{stamp(attempt.value.finished_at)}</time>{/if}</p>{#if attempt.value.error}<p class="error">{attempt.value.error}</p>{/if}<details><summary>attempt evidence</summary><pre>{payload({ output: attempt.value.output, decisions: attempt.value.decisions, actions: attempt.value.actions, targets: attempt.value.targets, reused_calls: attempt.value.reused_calls })}</pre></details></li>
          {:else}<li class="muted">No attempts recorded.</li>{/each}
        </ol>
        <h3>ordered capability calls</h3>
        <p class="muted">Capability inputs and results use the runtime's credential redaction. Large bodies remain artifact references.</p>
        <ol class="audit-calls">
          {#each audit.calls as call (call.key)}
            {@const observation = call.value.observation || {}}
            <li data-call-index={observation.index}>
              <p><b>#{observation.index} {observation.capability}</b> · attempt {call.value.attempt} · <span class:warn={call.state.toUpperCase() === 'UNKNOWN'}>{call.state}</span> · {duration(call.value)}</p>
              <p class="muted"><time datetime={started(call)}>{stamp(started(call))}</time>{#if call.value.finished_at} → <time datetime={call.value.finished_at}>{stamp(call.value.finished_at)}</time>{/if}</p>
              {#if observation.error}<p class="error">{observation.error}</p>{/if}
              {#if observation.redacted}<p class="muted">Credential values were redacted.</p>{/if}
              {#if call.value.artifact}<p>body artifact reference: <code>{call.value.artifact}</code></p>{/if}
              <details><summary>redacted input / result</summary><pre>{payload({ input_digest: observation.input_digest, inputs: observation.inputs, result: observation.result })}</pre></details>
            </li>
          {:else}<li class="muted">No capability calls recorded.</li>{/each}
        </ol>
        {#if audit.truncated}<p class="warn">Trace limited to 50 attempts and 200 capability calls.</p>{/if}
        <h3>resulting tickets</h3>
        <ul>{#each tickets as ticket}<li><a href={'?view=trace&ref=' + encodeURIComponent(ticket.id)}>{ticket.id}</a> · {ticket.label}</li>{:else}<li class="muted">No resulting tickets recorded.</li>{/each}</ul>
        {#if invocation.targets && !Array.isArray(invocation.targets)}<pre>{payload(invocation.targets)}</pre>{/if}
        {#if invocation.actions}<details><summary>planned actions</summary><pre>{payload(invocation.actions)}</pre></details>{/if}
        {#if invocation.output}<details><summary>hook output</summary><pre>{payload(invocation.output)}</pre></details>{/if}
      {/if}
    </section>
  {/if}
</section>

<style>
  .hook-audit { min-width: 0; border-bottom: 1px solid var(--rule); padding-bottom: var(--space-3); }
  .audit-controls, .audit-controls form, .audit-heading { display: flex; flex-wrap: wrap; align-items: center; gap: var(--space-2); }
  .audit-controls { margin-bottom: var(--space-2); }
  .audit-controls label { display: flex; align-items: center; gap: var(--space-2); min-width: 0; }
  .audit-controls input { width: 22ch; }
  .audit-detail { margin-top: var(--space-3); border-top: 1px solid var(--rule); overflow-wrap: anywhere; }
  .audit-heading { justify-content: space-between; }
  h3 { font-size: 1rem; margin: 12px 0 6px; }
  .audit-heading h3 { min-width: 0; flex: 1; }
  small { display: block; color: var(--muted); overflow-wrap: anywhere; }
  .audit-id { padding-left: 0; }
  dl { display: grid; grid-template-columns: max-content minmax(0, 1fr); gap: 4px 12px; }
  dd { margin: 0; }
  .audit-calls > li, .audit-attempts > li { margin-block: 8px; }
  .audit-detail pre { max-height: 24rem; overflow: auto; }
  code { overflow-wrap: anywhere; }
</style>
