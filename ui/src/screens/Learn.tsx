import { useAtomValue } from 'jotai';
import { useMemo, useState } from 'react';
import type { MemoryEdit } from '../../../packages/jobshop-client/rpc.js';
import { ago } from '../lib/model.ts';
import { build, isFresh, matches, metricRows, sections, summarize, type Stage } from '../lib/learn.ts';
import { improvementsAtom, nowAtom } from '../state/live.ts';

// Rows shown per section before "more": dense data stays scannable.
const PAGE = 10;

function when(value?: string) {
  return value && !value.startsWith('0001') ? new Date(value).toLocaleString() : 'not recorded';
}
// A gate result with no steps in its window still has a window.
function measured(edit: MemoryEdit) {
  return !!edit.after && !edit.after.from.startsWith('0001');
}
function span(m?: { from: string; to: string }) {
  return m && !m.from.startsWith('0001') ? `${when(m.from)} to ${when(m.to)}` : '';
}

export function Learn() {
  const data = useAtomValue(improvementsAtom);
  const now = useAtomValue(nowAtom);
  const learnings = useMemo(() => data?.learnings || [], [data]);
  const edits = useMemo(() => data?.edits || [], [data]);
  const [filter, setFilter] = useState('');
  const [only, setOnly] = useState<Stage | ''>('');
  const [toggled, setToggled] = useState<Record<string, boolean>>({});
  const [shown, setShown] = useState<Record<string, number>>({});

  const units = useMemo(() => build(learnings, edits), [learnings, edits]);
  const summary = useMemo(() => summarize(units, learnings, edits, now), [units, learnings, edits, now]);
  const visible = useMemo(() => units.filter((u) => matches(u, filter)), [units, filter]);
  const groups = useMemo(
    () =>
      sections
        .map((s) => ({ ...s, items: visible.filter((u) => u.stage === s.stage) }))
        .filter((g) => g.items.length && (!only || g.stage === only)),
    [visible, only],
  );
  const counts = useMemo(
    () => sections.map((s) => ({ ...s, n: visible.filter((u) => u.stage === s.stage).length })).filter((s) => s.n > 0),
    [visible],
  );
  const notHelped = useMemo(() => units.filter((u) => u.stage === 'not-helped').length, [units]);

  // A search or a chosen chip opens what it found; otherwise recent and
  // unresolved sections are open and history is closed.
  const isOpen = (stage: string, dflt: boolean) => (stage in toggled ? toggled[stage] : !!filter || !!only || dflt);

  return (
    <>
      <div className="tag">
        <span className="mono">/learn</span>
        <input
          aria-label="Filter learn"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder="problem, fix, pr, ticket or pool"
        />
        <button
          onClick={() => {
            setFilter('');
            setOnly('');
            setToggled({});
          }}
        >
          clear
        </button>
        <span className="push-right muted">
          {visible.length}
          {visible.length !== units.length ? ' of ' + units.length : ''}
        </span>
      </div>
      <div className="section-body learn">
        {!units.length ? (
          <p className="muted">
            Nothing recorded yet. Learnings appear after a failed or corrected step, and memory edits after the
            curator's change merges.
          </p>
        ) : (
          <>
            <p className="learn-summary" aria-label="Summary">
              {summary.items} items
              <span className="muted">
                {' '}
                · {summary.learnings} learnings, {summary.edits} memory edits · {summary.fresh} new in 7 days
              </span>
              {notHelped > 0 && <b> · {notHelped} not helped</b>}
            </p>
            <div className="step-strip mono" role="group" aria-label="Stage">
              {counts.map((s) => (
                <button
                  key={s.stage}
                  className={only === s.stage ? 'chosen' : undefined}
                  aria-pressed={only === s.stage}
                  title={s.hint}
                  onClick={() => {
                    setOnly(only === s.stage ? '' : s.stage);
                    setToggled({});
                  }}
                >
                  {s.title} {s.n}
                </button>
              ))}
            </div>
            {!groups.length && <p className="muted">Nothing matches the filter.</p>}
            {groups.map((g) => {
              const open = isOpen(g.stage, g.open);
              const limit = shown[g.stage] || PAGE;
              return (
                <section className="learn-group" aria-label={g.title} key={g.stage}>
                  <h2>
                    <button
                      className="learn-toggle"
                      aria-expanded={open}
                      onClick={() => setToggled({ ...toggled, [g.stage]: !open })}
                    >
                      <span aria-hidden="true">{open ? '▾' : '▸'}</span> {g.title}{' '}
                      <span className="muted mono">{g.items.length}</span>
                    </button>{' '}
                    <span className="muted learn-hint">{g.hint}</span>
                  </h2>
                  {open && (
                    <>
                      <div className="rows">
                        {g.items.slice(0, limit).map((u) => (
                          <LearnRow key={u.key} u={u} now={now} />
                        ))}
                      </div>
                      {g.items.length > limit && (
                        <button className="learn-more" onClick={() => setShown({ ...shown, [g.stage]: limit + PAGE })}>
                          more ({g.items.length - limit})
                        </button>
                      )}
                    </>
                  )}
                </section>
              );
            })}
          </>
        )}
      </div>
    </>
  );
}

function LearnRow({ u, now }: { u: ReturnType<typeof build>[number]; now: number }) {
  return (
    <article className="learn-row">
      <p className="learn-problem">
        {u.problem ? <b>{u.problem}</b> : <b className="muted">problem not recorded</b>}
        {u.more > 0 && <span className="muted"> +{u.more} more</span>}
      </p>
      <p className="learn-fix">
        <span className="muted">{u.fixLabel}</span> {u.fix ? u.fix : <span className="muted">not recorded</span>}
      </p>
      <p className="learn-meta muted">
        <span className="mono">{u.outcome}</span>
        {u.where && (
          <>
            {' · '}
            {u.where.href ? (
              <a href={u.where.href} target="_blank" rel="noopener noreferrer">
                {u.where.text}
              </a>
            ) : (
              <span className="mono">{u.where.text}</span>
            )}
          </>
        )}
        {' · '}
        <span title={when(u.at)}>{ago(u.at, now)}</span>
        {isFresh(u.at, now) && <span className="mono"> · new</span>}
        {u.kind && <> · {u.kind}</>}
      </p>
      <details>
        <summary>details</summary>
        <dl className="learn-facts mono">
          {u.edit && (
            <>
              <dt>memory edit</dt>
              <dd>
                #{u.edit.id}, merged {when(u.edit.merged_at)}
              </dd>
              {!!u.edit.pools?.length && (
                <>
                  <dt>pools</dt>
                  <dd>{u.edit.pools.join(', ')}</dd>
                </>
              )}
              {u.edit.evaluated_at && !u.edit.evaluated_at.startsWith('0001') && (
                <>
                  <dt>evaluated</dt>
                  <dd>{when(u.edit.evaluated_at)}</dd>
                </>
              )}
            </>
          )}
          {u.scope && (
            <>
              <dt>scope</dt>
              <dd>{u.scope}</dd>
            </>
          )}
        </dl>
        {u.learnings.map((l) => (
          <div className="learn-cited" key={l.id}>
            <p>
              <span className="muted">problem #{l.id}</span> {l.summary}
            </p>
            <p>
              <span className="muted">fix</span> {l.fix ? l.fix : <span className="muted">not recorded</span>}
            </p>
            <p className="muted mono">
              {l.kind} · {l.source} · {l.status}
              {l.gate && <> · gate {l.gate}</>}
              {l.ticket && (
                <>
                  {' · '}
                  <a href={'?view=trace&ref=' + encodeURIComponent(l.ticket)}>{l.ticket}</a>
                </>
              )}
              {l.session && <> · {l.session}</>} · {when(l.created_at)}
            </p>
          </div>
        ))}
        {u.missing.length > 0 && (
          <p className="muted">Also cites {u.missing.map((id) => '#' + id).join(', ')}, not in the loaded list.</p>
        )}
        {u.edit && <EditDetails edit={u.edit} />}
      </details>
    </article>
  );
}

function EditDetails({ edit }: { edit: MemoryEdit }) {
  return (
    <>
      <div className="table-scroll">
        <table className="mono rows">
          <thead>
            <tr>
              <th>measure</th>
              <th>before</th>
              <th>after</th>
            </tr>
          </thead>
          <tbody>
            {metricRows.map((row) => (
              <tr key={row.key}>
                <td>{row.label}</td>
                <td>{edit.baseline[row.key]}</td>
                <td>{measured(edit) ? edit.after?.[row.key] : 'not measured'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <p className="muted mono">
        before {span(edit.baseline)}
        {measured(edit) && <>; after {span(edit.after)}</>}
      </p>
      {edit.before && (
        <details>
          <summary>text the edit replaced</summary>
          <pre>{edit.before.length > 8000 ? edit.before.slice(0, 8000) + '\n[display limited]' : edit.before}</pre>
        </details>
      )}
    </>
  );
}
