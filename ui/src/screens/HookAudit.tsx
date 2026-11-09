import { useMemo, useState, type FormEvent } from 'react';
import type { AutomationAudit, AutomationList, AutomationRecord } from '../../../packages/jobshop-client/index.js';

type Detail = AutomationAudit | { id: string; error: string } | null;
type Json = Record<string, any>;

function duration(value: Json) {
  const ns = value.elapsed_ns || value.observation?.duration_ns;
  if (typeof ns === 'number') return (ns / 1e6).toFixed(1) + ' ms';
  if (value.started_at && value.finished_at) {
    return Math.max(0, Date.parse(value.finished_at) - Date.parse(value.started_at)).toFixed(1) + ' ms';
  }
  return 'unfinished';
}
// Evidence is bounded so one large body cannot freeze the page.
function payload(value: unknown) {
  const text = JSON.stringify(value, null, 2) || 'null';
  return text.length > 65536 ? text.slice(0, 65536) + '\n[display limited to 64 KiB]' : text;
}
function stamp(value?: string) {
  return value && !value.startsWith('0001-') ? new Date(value).toLocaleString() : 'not recorded';
}
const started = (record: AutomationRecord) => record.value.started_at || record.created_at;
const isUnknown = (state: string) => state.toUpperCase() === 'UNKNOWN';

export function HookAudit({
  summaries,
  selected,
  detail,
  onSelect,
}: {
  summaries?: AutomationList;
  selected: string;
  detail: Detail;
  onSelect: (id: string) => void;
}) {
  const [filter, setFilter] = useState('');
  const items = useMemo(
    () =>
      (summaries?.items || []).filter((item) =>
        `${item.key} ${item.state} ${item.trigger || ''} ${item.run || ''} ${item.event?.delivery || ''}`
          .toLowerCase()
          .includes(filter.toLowerCase()),
      ),
    [summaries, filter],
  );
  const audit = detail && !('error' in detail) && detail.id === selected ? detail : null;

  function lookup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const id = new FormData(event.currentTarget).get('invocation')?.toString().trim();
    if (id) onSelect(id);
  }

  return (
    <section className="hook-audit" aria-label="Automation audit">
      <h2>automation audit</h2>
      <p className="muted">
        Events are recorded before tickets exist. Select an invocation to inspect its attempts and capability calls.
      </p>
      <div className="audit-controls">
        <label>
          filter{' '}
          <input
            aria-label="Filter automation audit"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder="state, trigger, run or delivery"
          />
        </label>
        <form onSubmit={lookup}>
          <label>
            invocation{' '}
            <input
              name="invocation"
              aria-label="Automation invocation ID"
              defaultValue={selected}
              key={selected}
              required
            />
          </label>
          <button type="submit">inspect</button>
        </form>
      </div>
      <div className="table-scroll">
        <table className="rows" aria-label="Automation invocations">
          <thead>
            <tr>
              <th>invocation</th>
              <th>event / run</th>
              <th>state</th>
              <th>attempts</th>
              <th>updated</th>
            </tr>
          </thead>
          <tbody>
            {items.map((item) => (
              <tr key={item.key} className={selected === item.key ? 'chosen' : undefined}>
                <td>
                  <button
                    className="audit-id"
                    aria-label={'Inspect invocation ' + item.key}
                    aria-pressed={selected === item.key}
                    onClick={() => onSelect(item.key)}
                  >
                    {item.key.slice(0, 12)}
                  </button>
                  <small>{item.trigger || item.entry.split('/').at(-1)}</small>
                </td>
                <td>
                  {item.event ? (
                    <>
                      <span>
                        {item.event.source} / {item.event.event}
                      </span>
                      <small>{item.event.delivery}</small>
                    </>
                  ) : item.run ? (
                    <a href={'?view=trace&ref=' + encodeURIComponent(item.run)}>{item.run}</a>
                  ) : (
                    <span className="muted">no ticket</span>
                  )}
                </td>
                <td className={item.state === 'failed' || isUnknown(item.state) ? 'warn' : undefined}>{item.state}</td>
                <td>{item.attempts}</td>
                <td>
                  <time dateTime={item.updated_at}>{stamp(item.updated_at)}</time>
                </td>
              </tr>
            ))}
            {!items.length && (
              <tr>
                <td colSpan={5} className="muted">
                  {filter ? 'No matching invocations.' : 'No automation invocations.'}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      {summaries?.truncated && (
        <p className="muted">Showing the latest 50 invocations. Inspect an older invocation by its ID.</p>
      )}
      {selected && (
        <section className="audit-detail" aria-label="Automation invocation detail">
          <div className="audit-heading">
            <h3>
              invocation <span className="mono">{selected}</span>
            </h3>
            <button onClick={() => onSelect('')}>dismiss</button>
          </div>
          {detail && 'error' in detail ? (
            <p role="alert" className="error">
              {detail.error}
            </p>
          ) : !audit ? (
            <output className="block">Loading invocation…</output>
          ) : (
            <AuditDetail audit={audit} onSelect={onSelect} />
          )}
        </section>
      )}
    </section>
  );
}

// A timestamp that keeps its machine-readable value.
function Time({ value }: { value?: string }) {
  return <time dateTime={value}>{stamp(value)}</time>;
}
const time = (value?: string) => <Time value={value} />;

function AuditDetail({ audit, onSelect }: { audit: AutomationAudit; onSelect: (id: string) => void }) {
  const invocation: Json = useMemo(() => audit.invocation.value || {}, [audit]);
  const unknown = audit.calls.some((c) => isUnknown(c.state)) || audit.attempts.some((a) => isUnknown(a.state));
  const tickets = useMemo(() => {
    const linked = new Map<string, string>();
    if (invocation.run) linked.set(invocation.run, 'run');
    if (invocation.step) linked.set(invocation.step, 'step');
    if (audit.invocation.state === 'done') {
      for (const target of Array.isArray(invocation.targets) ? invocation.targets : []) {
        if (typeof target === 'string') linked.set(target, 'ticket');
      }
    }
    return [...linked].map(([id, label]) => ({ id, label }));
  }, [audit, invocation]);

  return (
    <>
      <p>
        <b>{audit.invocation.state}</b> · revision {audit.invocation.revision} · {audit.invocation.attempts} attempts
      </p>
      {invocation.parent_invocation && (
        <p>
          parent invocation{' '}
          <button className="mono" onClick={() => onSelect(invocation.parent_invocation)}>
            {invocation.parent_invocation}
          </button>
        </p>
      )}
      <p className="mono">{invocation.entry}</p>
      <p className="muted">
        created {time(audit.invocation.created_at)} · updated {time(audit.invocation.updated_at)}
      </p>
      {invocation.event && (
        <>
          <h3>event</h3>
          <dl>
            <dt>source / event</dt>
            <dd>
              {invocation.event.source} / {invocation.event.event}
            </dd>
            <dt>delivery</dt>
            <dd className="mono">{invocation.event.delivery}</dd>
            <dt>trigger</dt>
            <dd>{invocation.trigger}</dd>
          </dl>
        </>
      )}
      {unknown && (
        <output className="block warn">
          UNKNOWN: an attempt or external call has no confirmed outcome. Reconcile it before retrying.
        </output>
      )}
      {invocation.error && <p className="error">{invocation.error}</p>}
      {invocation.artifact && (
        <p>
          snapshot artifact reference: <code>{invocation.artifact}</code>
        </p>
      )}
      <details>
        <summary>event identity</summary>
        <pre>{payload(invocation.event)}</pre>
      </details>
      {invocation.decisions && (
        <>
          <h3>decisions</h3>
          {Array.isArray(invocation.decisions) ? (
            <ol>
              {invocation.decisions.map((decision: Json, i: number) => (
                <li key={i}>
                  <b>{decision.choice}</b>
                  {decision.defaulted && ' · defaulted'}
                  {decision.confidence != null && ` · confidence ${decision.confidence}`}
                  <p>{decision.reason}</p>
                </li>
              ))}
            </ol>
          ) : (
            <pre>{payload(invocation.decisions)}</pre>
          )}
        </>
      )}
      <h3>attempts</h3>
      <ol className="audit-attempts rows">
        {audit.attempts.map((attempt) => (
          <li key={attempt.key}>
            <b>attempt {attempt.value.attempt}</b> ·{' '}
            <span className={isUnknown(attempt.state) ? 'warn' : undefined}>{attempt.state}</span> ·{' '}
            {duration(attempt.value)}
            <p className="muted">
              {time(started(attempt))}
              {attempt.value.finished_at && <> → {time(attempt.value.finished_at)}</>}
            </p>
            {attempt.value.error && <p className="error">{attempt.value.error}</p>}
            <details>
              <summary>attempt evidence</summary>
              <pre>
                {payload({
                  result_metadata: attempt.value.result_metadata,
                  result_output: attempt.value.result_output,
                  output: attempt.value.output,
                  decisions: attempt.value.decisions,
                  actions: attempt.value.actions,
                  targets: attempt.value.targets,
                  reused_calls: attempt.value.reused_calls,
                })}
              </pre>
            </details>
          </li>
        ))}
        {!audit.attempts.length && <li className="muted">No attempts recorded.</li>}
      </ol>
      <h3>ordered capability calls</h3>
      <p className="muted">
        Capability inputs and results use the runtime's credential redaction. Large bodies remain artifact references.
      </p>
      <ol className="audit-calls rows">
        {audit.calls.map((call) => {
          const observation: Json = call.value.observation || {};
          return (
            <li key={call.key} data-call-index={observation.index}>
              <p>
                <b>
                  #{observation.index} {observation.capability}
                </b>{' '}
                · attempt {call.value.attempt} ·{' '}
                <span className={isUnknown(call.state) ? 'warn' : undefined}>{call.state}</span> ·{' '}
                {duration(call.value)}
              </p>
              <p className="muted">
                {time(started(call))}
                {call.value.finished_at && <> → {time(call.value.finished_at)}</>}
              </p>
              {observation.error && <p className="error">{observation.error}</p>}
              {observation.redacted && <p className="muted">Credential values were redacted.</p>}
              {call.value.artifact && (
                <p>
                  body artifact reference: <code>{call.value.artifact}</code>
                </p>
              )}
              <details>
                <summary>redacted input / result</summary>
                <pre>
                  {payload({
                    input_digest: observation.input_digest,
                    inputs: observation.inputs,
                    result: observation.result,
                  })}
                </pre>
              </details>
            </li>
          );
        })}
        {!audit.calls.length && <li className="muted">No capability calls recorded.</li>}
      </ol>
      {audit.truncated && <p className="warn">Trace limited to 50 attempts and 200 capability calls.</p>}
      <h3>resulting tickets</h3>
      <ul className="rows">
        {tickets.map((ticket) => (
          <li key={ticket.id}>
            <a href={'?view=trace&ref=' + encodeURIComponent(ticket.id)}>{ticket.id}</a> · {ticket.label}
          </li>
        ))}
        {!tickets.length && <li className="muted">No resulting tickets recorded.</li>}
      </ul>
      {invocation.targets && !Array.isArray(invocation.targets) && <pre>{payload(invocation.targets)}</pre>}
      {invocation.actions && (
        <details>
          <summary>planned actions</summary>
          <pre>{payload(invocation.actions)}</pre>
        </details>
      )}
      {(invocation.result_metadata || invocation.result_output) && (
        <details>
          <summary>callback result</summary>
          <pre>{payload({ metadata: invocation.result_metadata, output: invocation.result_output })}</pre>
        </details>
      )}
      {invocation.output && (
        <details>
          <summary>hook output</summary>
          <pre>{payload(invocation.output)}</pre>
        </details>
      )}
    </>
  );
}
