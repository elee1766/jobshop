import { useAtomValue } from 'jotai';
import { ago } from '../lib/model.ts';
import { useDeskRoute } from '../lib/route.ts';
import { disablePush, enablePush, pushStateAtom } from '../lib/push.ts';
import { closeSession, openAudit, openTrace } from '../state/actions.ts';
import { auditAtom, nowAtom, systemAtom } from '../state/live.ts';
import { HookAudit } from './HookAudit.tsx';

function toggleTheme() {
  const root = document.documentElement;
  root.dataset.theme = root.dataset.theme === 'dark' ? 'light' : 'dark';
}

export function System() {
  const system = useAtomValue(systemAtom);
  const audit = useAtomValue(auditAtom);
  const now = useAtomValue(nowAtom);
  const push = useAtomValue(pushStateAtom);
  const route = useDeskRoute();
  return (
    <>
      <div className="tag">
        <span className="mono">/system</span>
        <button onClick={() => void (push === 'on' ? disablePush() : enablePush())}>push: {push}</button>
        <button onClick={toggleTheme}>theme</button>
      </div>
      <div className="section-body">
        <HookAudit summaries={system?.automation} selected={route.audit} detail={audit} onSelect={openAudit} />
        <h2>pools</h2>
        <div className="table-scroll">
          <table className="mono rows">
            <thead>
              <tr>
                <th>pool</th>
                <th>status</th>
                <th>open / max</th>
                <th>working</th>
                <th>idle</th>
                <th>desired</th>
              </tr>
            </thead>
            <tbody>
              {(system?.pools?.pools || []).map((pool) => (
                <tr key={pool.name}>
                  <td>{pool.name}</td>
                  <td>{system?.pools.quarantined?.[pool.name] ? 'quarantined' : 'ok'}</td>
                  <td>
                    {pool.open} / {pool.max || 'uncapped'}
                  </td>
                  <td>{pool.working}</td>
                  <td>{pool.idle}</td>
                  <td>{pool.desired}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <h2>resources</h2>
        <div className="table-scroll">
          <table className="mono rows">
            <thead>
              <tr>
                <th>resource</th>
                <th>held / capacity</th>
                <th>waiting</th>
              </tr>
            </thead>
            <tbody>
              {(system?.resources?.use || []).map((resource) => (
                <tr key={resource.key}>
                  <td>{resource.key}</td>
                  <td>
                    {resource.held} / {resource.capacity || 'uncapped'}
                  </td>
                  <td>{resource.waiting}</td>
                </tr>
              ))}
              {!system?.resources?.use?.length && (
                <tr>
                  <td colSpan={3} className="muted">
                    No resource use.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
          <table className="mono rows">
            <thead>
              <tr>
                <th>slot</th>
                <th>resource</th>
                <th>state</th>
                <th>ticket</th>
              </tr>
            </thead>
            <tbody>
              {(system?.resources?.slots || []).map((slot) => (
                <tr key={slot.id}>
                  <td>{slot.id}</td>
                  <td>{slot.pool}</td>
                  <td>{slot.state}</td>
                  <td>
                    <button disabled={!slot.ticket} onClick={() => slot.ticket && openTrace(slot.ticket)}>
                      {slot.ticket || '–'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <h2>ingest</h2>
        <p>
          {system?.ingest?.mode || 'controller'} ·{' '}
          {Object.entries(system?.ingest?.counts || {})
            .map(([state, count]) => `${state}: ${count}`)
            .join(' · ') || 'Inbox empty'}
        </p>
        <div className="table-scroll">
          <table className="mono rows">
            <thead>
              <tr>
                <th>worker</th>
                <th>host</th>
                <th>last seen</th>
                <th>admitted</th>
              </tr>
            </thead>
            <tbody>
              {(system?.ingest?.workers || []).map((worker) => (
                <tr key={worker.id}>
                  <td>{worker.id}</td>
                  <td>{worker.host}</td>
                  <td>{ago(worker.seen, now)}</td>
                  <td>{worker.admitted}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <h2>scheduler</h2>
        <p>
          Last {system?.status?.tick_seconds?.toFixed(3) || '0'}s · p95 {system?.status?.p95_seconds?.toFixed(3) || '0'}
          s ({system?.status?.samples || 0} passes)
        </p>
        <table className="mono rows">
          <tbody>
            {Object.entries(system?.status?.steps || {}).map(([name, seconds]) => (
              <tr key={name}>
                <td>{name}</td>
                <td>{Number(seconds).toFixed(3)}s</td>
              </tr>
            ))}
          </tbody>
        </table>
        <h2>sessions</h2>
        <div className="rows">
          {(system?.sessions || []).map((session) => (
            <p key={session.id}>
              <button className="mono" onClick={() => openTrace(session.id)}>
                {session.display_name || session.id}
              </button>{' '}
              {session.busy ? 'working' : 'idle'} <button onClick={() => void closeSession(session.id)}>close</button>
            </p>
          ))}
        </div>
        <p className="muted">
          Costs are not recorded by this controller. Context usage is available in session journals.
        </p>
      </div>
    </>
  );
}
