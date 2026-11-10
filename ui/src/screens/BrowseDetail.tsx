import { useAtomValue } from 'jotai';
import { useMemo, useRef, useState } from 'react';
import type {
  BrowseCheck,
  BrowseComment,
  BrowseDetail,
  BrowseFile,
  BrowseLinked,
  BrowseThread,
} from '../../../packages/jobshop-client/rpc.js';
import { StatusLine, WritesButton, WritesNote } from '../components/BrowseParts.tsx';
import { Markdown } from '../components/Markdown.tsx';
import { CheckMark } from '../components/StatusLetter.tsx';
import { openTicketMenu } from '../state/ticketMenu.ts';
import { parseTitle } from '../../kit/titles.js';
import { age, duration, newComments, providerName, whereIs, type Place, type Provider } from '../lib/browse.ts';
import { copyText } from '../lib/clipboard.ts';
import { currentURL, setParams, useDeskRoute } from '../lib/route.ts';
import { discussProvider, navigate, openAudit, openTrace } from '../state/actions.ts';
import { useBrowseDetail } from '../state/browse.ts';
import { browseAtom, nowAtom, systemAtom } from '../state/live.ts';
import { store } from '../state/store.ts';
import { noticeAtom } from '../state/ticketMenu.ts';

// One Linear issue or GitHub PR or issue (JOB-163): header, title, properties,
// body, the real links the provider and Jobshop know, checks and files, and
// the discussion. Read-only: the controls of the design that would write are
// present and disabled, and the reason is on the page.

export function Detail({ provider, item }: { provider: Provider; item: string }) {
  const state = useBrowseDetail(provider, item, '');
  const { detail } = state;
  const now = useAtomValue(nowAtom);
  const route = useDeskRoute();

  if (state.loading && !detail) {
    return (
      <>
        <Header provider={provider} item={item} detail={null} />
        <output className="section-body muted block">loading</output>
      </>
    );
  }
  if (!detail || (detail.status.state === 'error' && !detail.item.title)) {
    return (
      <>
        <Header provider={provider} item={item} detail={detail} />
        <StatusLine status={detail?.status} error={state.error} onRetry={state.refresh} />
        <p className="section-body muted">
          {detail?.status.code === 'not_found'
            ? 'This item does not exist, or the service identity cannot see it.'
            : 'The item could not be read.'}
        </p>
      </>
    );
  }
  const github = provider === 'github';
  const tab = github && (route.tab === 'files' || route.tab === 'checks') ? route.tab : 'discussion';
  return (
    <div className="browse-body" data-kind={detail.item.kind}>
      <Header provider={provider} item={item} detail={detail} />
      <StatusLine status={detail.status} error={state.error} onRetry={state.refresh} />
      <h2 className="browse-title-text">{detail.item.title}</h2>
      {github ? <GitHubMeta detail={detail} /> : <LinearProps detail={detail} />}
      {github && <Tabs detail={detail} tab={tab} />}
      {tab === 'discussion' && (
        <>
          <Body detail={detail} />
          <Linked detail={detail} provider={provider} />
          {github && detail.checks.length > 0 && <ChecksSummary detail={detail} />}
          <Discussion
            detail={detail}
            seen={state.seen}
            now={now}
            older={state.older}
            loadOlder={state.loadOlder}
            acknowledge={state.acknowledge}
            what={github ? 'discussion' : 'activity'}
          />
        </>
      )}
      {tab === 'files' && <Files detail={detail} />}
      {tab === 'checks' && <Checks detail={detail} now={now} />}
    </div>
  );
}

function Header({ provider, item, detail }: { provider: Provider; item: string; detail: BrowseDetail | null }) {
  const url = detail?.item.url;
  const ref = detail?.item.ref || item;
  const kind = detail?.item.kind;
  const label = provider === 'github' ? `/${kind === 'issue' ? 'issue' : 'pr'}/${ref.split('#')[1] ?? ref}` : `/${ref}`;
  const name = providerName[provider];
  return (
    <div className="tag browse-head">
      <button className="browse-back" onClick={() => setParams({ item: '', tab: '' })} aria-label="Back to the list">
        ‹ {provider === 'github' ? 'prs' : 'issues'}
      </button>
      <span className="mono" title={ref}>
        {provider === 'github' ? label : `/${ref}`}
      </span>
      {url && (
        <a href={url} target="_blank" rel="noopener noreferrer">
          open in {name.toLowerCase()} ↗
        </a>
      )}
      <button onClick={() => discussProvider(provider, ref)} title="Start a chat with this attached">
        chat
      </button>
      <button
        onClick={async () => {
          if (await copyText(ref)) store.set(noticeAtom, 'copied ' + ref);
        }}
      >
        copy
      </button>
      {provider === 'linear' && <WritesButton label="edit" />}
      <button
        aria-haspopup="menu"
        onClick={(e) =>
          openTicketMenu({
            target: { id: ref, ref: provider + ':' + ref, provider: { name, url: url || '' } },
            anchor: e.currentTarget,
            returnTo: e.currentTarget,
          })
        }
      >
        actions ▾
      </button>
      <button className="push-right browse-dismiss" onClick={() => setParams({ item: '', tab: '' })}>
        dismiss
      </button>
    </div>
  );
}

// The properties of a Linear issue. Each is a disabled control, the design's
// dropdowns, with the value the provider has.
function LinearProps({ detail }: { detail: BrowseDetail }) {
  const p = detail.props;
  return (
    <fieldset className="browse-props">
      <legend className="sr-only">Properties (read only)</legend>
      <WritesButton className="prop" label={`${detail.item.state} ▾`} />
      <WritesButton className="prop" label={`${p.assignee || 'unassigned'} ▾`} />
      <WritesButton className="prop" label={`${p.priority || 'no priority'} ▾`} />
      <WritesButton className="prop" label={`project: ${p.project || 'none'} ▾`} />
      <WritesButton className="prop" label={`labels: ${p.labels.length ? p.labels.join(', ') : 'none'} ▾`} />
    </fieldset>
  );
}

function GitHubMeta({ detail }: { detail: BrowseDetail }) {
  const { props: p, item } = detail;
  const parts = [
    item.state,
    p.author && (item.owner_bot ? p.author + ' (app)' : p.author),
    item.kind === 'pr' && p.head && p.base ? `${p.head} → ${p.base}` : p.base && `→ ${p.base}`,
  ].filter(Boolean);
  return (
    <p className="browse-meta">
      <span>{parts.join(' · ')}</span>
      {item.kind === 'pr' && (
        <>
          {' · '}
          <span className="add">+{p.additions ?? 0}</span> <span className="del">−{p.deletions ?? 0}</span>
          {' · '}
          {p.changed_files ?? 0} files
        </>
      )}
      {item.review && <span> · review {item.review === 'changes' ? 'changes requested' : item.review}</span>}
      {p.mergeable && p.mergeable !== 'unknown' && item.state === 'open' && <span> · {p.mergeable.toLowerCase()}</span>}
      {item.labels.length > 0 && <span className="muted"> · {item.labels.join(', ')}</span>}
    </p>
  );
}

function Tabs({ detail, tab }: { detail: BrowseDetail; tab: string }) {
  const isPR = detail.item.kind === 'pr';
  const files = detail.files_total || detail.files.length;
  const checks = detail.checks_total || detail.checks.length;
  const tabs = [
    ['discussion', 'discussion'],
    ...(isPR ? [['files', `files ${files}`]] : []),
    ...(isPR ? [['checks', `checks ${checks}`]] : []),
  ];
  return (
    <div className="tag browse-tabs" role="tablist" aria-label="Pull request">
      {tabs.map(([id, label]) => (
        <button
          key={id}
          role="tab"
          aria-selected={tab === id}
          className={tab === id ? 'chosen' : undefined}
          onClick={() => setParams({ tab: id === 'discussion' ? '' : id })}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

function Body({ detail }: { detail: BrowseDetail }) {
  if (!detail.body.trim()) return <p className="section-body muted">No description.</p>;
  return (
    <div className="section-body browse-text">
      <Markdown text={detail.body} />
      {detail.body_truncated && (
        <p className="muted">
          The description is cut here.{' '}
          {detail.item.url && (
            <a href={detail.item.url} target="_blank" rel="noopener noreferrer">
              read the rest in {providerName[detail.item.provider as Provider]} ↗
            </a>
          )}
        </p>
      )}
    </div>
  );
}

function placeHref(place: Place) {
  if (!place.view) return place.url;
  const url = currentURL();
  url.search = '';
  url.searchParams.set('view', place.view);
  for (const [key, value] of Object.entries(place.params ?? {})) if (value) url.searchParams.set(key, value);
  url.hash = '';
  return url.pathname + url.search;
}

function Linked({ detail, provider }: { detail: BrowseDetail; provider: Provider }) {
  const scope = useAtomValue(browseAtom);
  const automation = useAtomValue(systemAtom)?.automation?.items;
  const [open, setOpen] = useState(true);
  if (!detail.linked.length) return null;
  return (
    <section aria-label="Linked work">
      <div className="tag">
        <button aria-expanded={open} aria-label="Toggle linked work" onClick={() => setOpen(!open)}>
          {open ? '▾' : '▸'}
        </button>
        <span>linked work</span> <span className="muted">{detail.linked.length}</span>
      </div>
      {open && (
        <div className="section-body table-scroll">
          <table className="mono rows browse-linked">
            <thead>
              <tr>
                <th>kind</th>
                <th>title</th>
                <th>relation</th>
                <th>state</th>
              </tr>
            </thead>
            <tbody>
              {detail.linked.map((link) => (
                <tr key={link.kind + link.ref + link.relation}>
                  <td>{link.kind}</td>
                  <td className="browse-linked-title">
                    <LinkedTitle link={link} scope={scope} provider={provider} />
                    {link.kind === 'run' &&
                      (automation ?? [])
                        .filter((a) => a.run === link.ref)
                        .slice(0, 3)
                        .map((a) => (
                          <button
                            key={a.key}
                            className="browse-audit"
                            onClick={() => openAudit(a.key)}
                            title={'Hook audit ' + a.key}
                          >
                            audit
                          </button>
                        ))}
                  </td>
                  <td className="muted">{link.relation}</td>
                  <td>{link.state || <span className="muted">-</span>}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function LinkedTitle({
  link,
  scope,
  provider,
}: {
  link: BrowseLinked;
  scope: ReturnType<typeof useAtomValue<typeof browseAtom>>;
  provider: Provider;
}) {
  const text = link.title ? `${link.ref} · ${link.title}` : link.ref;
  if (link.kind === 'run') {
    // A Jobshop run opens its trace, which is where its graph, sessions and
    // history are. Its title may hold Markdown links: show their labels.
    return (
      <a
        href={`?view=trace&ref=${encodeURIComponent(link.ref)}`}
        title={'Open the trace of ' + link.ref}
        onClick={(e) => {
          if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
          e.preventDefault();
          openTrace(link.ref);
        }}
      >
        {parseTitle(link.title).text || link.ref}
      </a>
    );
  }
  const place = whereIs(link, scope);
  if (place.view) {
    return (
      <>
        <a
          href={placeHref(place)}
          onClick={(e) => {
            if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
            e.preventDefault();
            navigate(place.view!, '', place.params ?? {});
          }}
        >
          {text}
        </a>
        {place.url && (
          <>
            {' '}
            <a
              href={place.url}
              target="_blank"
              rel="noopener noreferrer"
              aria-label={'Open ' + link.ref + ' at the provider'}
            >
              ↗
            </a>
          </>
        )}
      </>
    );
  }
  if (place.url) {
    return (
      <a
        href={place.url}
        target="_blank"
        rel="noopener noreferrer"
        title={provider === 'linear' ? 'Outside the teams Desk reads' : 'Outside the repositories Desk reads'}
      >
        {text} ↗
      </a>
    );
  }
  return <span>{text}</span>;
}

const checkMark: Record<string, [string, string]> = {
  pass: ['✓', 'passed'],
  fail: ['✗', 'failed'],
  run: ['…', 'running'],
  skip: ['-', 'skipped'],
};

function ChecksSummary({ detail }: { detail: BrowseDetail }) {
  const counts = { pass: 0, fail: 0, run: 0, skip: 0 } as Record<string, number>;
  for (const c of detail.checks) counts[c.state] = (counts[c.state] ?? 0) + 1;
  const total = detail.checks_total || detail.checks.length;
  const words = [
    counts.fail && `${counts.fail} failed`,
    counts.run && `${counts.run} running`,
    counts.pass && `${counts.pass} passed`,
    counts.skip && `${counts.skip} skipped`,
  ].filter(Boolean);
  return (
    <div className="tag browse-checks-line">
      <span>checks</span>
      <span>{words.join(', ') || `${total}`}</span>
      {detail.item.url && (
        <a href={detail.item.url + '/checks'} target="_blank" rel="noopener noreferrer">
          details on github ↗
        </a>
      )}
    </div>
  );
}

function Checks({ detail, now }: { detail: BrowseDetail; now: number }) {
  const rows: BrowseCheck[] = detail.checks;
  if (!rows.length) return <p className="section-body muted">No checks reported for the head commit.</p>;
  return (
    <div className="section-body table-scroll">
      <table className="mono rows browse-checks">
        <thead>
          <tr>
            <th>s</th>
            <th>check</th>
            <th>time</th>
            <th>details</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((c, i) => {
            const [mark, word] = checkMark[c.state] ?? ['?', c.state];
            return (
              <tr key={c.name + i} data-state={c.state}>
                <td>
                  <CheckMark state={c.state} mark={mark} label={word} />
                </td>
                <td>{c.name}</td>
                <td>
                  {c.state === 'run' && c.started_at
                    ? age(c.started_at, now).replace(' ago', '')
                    : duration(c.seconds) || '-'}
                </td>
                <td>
                  {c.url?.startsWith('https://') ? (
                    <a href={c.url} target="_blank" rel="noopener noreferrer">
                      logs ↗
                    </a>
                  ) : (
                    <span className="muted">-</span>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
      {(detail.checks_total ?? 0) > rows.length && (
        <p className="muted">
          {rows.length} of {detail.checks_total} checks shown.
        </p>
      )}
    </div>
  );
}

function Files({ detail }: { detail: BrowseDetail }) {
  const rows: BrowseFile[] = detail.files;
  if (!rows.length) return <p className="section-body muted">No changed files.</p>;
  return (
    <div className="section-body table-scroll">
      <table className="mono rows browse-files">
        <thead>
          <tr>
            <th>file</th>
            <th>change</th>
            <th>+</th>
            <th>−</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((f) => (
            <tr key={f.path}>
              <td className="browse-path">{f.path}</td>
              <td className="muted">{f.change}</td>
              <td className="add">{f.additions}</td>
              <td className="del">{f.deletions}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {(detail.files_total ?? 0) > rows.length && (
        <p className="muted">
          {rows.length} of {detail.files_total} files shown.{' '}
          {detail.item.url && (
            <a href={detail.item.url + '/files'} target="_blank" rel="noopener noreferrer">
              all files on github ↗
            </a>
          )}
        </p>
      )}
    </div>
  );
}

// --- discussion -----------------------------------------------------------

function Discussion({
  detail,
  seen,
  now,
  older,
  loadOlder,
  acknowledge,
  what,
}: {
  detail: BrowseDetail;
  seen: Set<string>;
  now: number;
  older: boolean;
  loadOlder: () => void;
  acknowledge: () => void;
  what: string;
}) {
  const [open, setOpen] = useState(true);
  const [oldestFirst, setOldestFirst] = useState(false);
  const ids = useMemo(() => detail.threads.flatMap((t) => t.comments.map((c) => c.id)), [detail.threads]);
  const fresh = useMemo(() => new Set(newComments(seen, ids)), [seen, ids]);
  const threads = useMemo(
    () => (oldestFirst ? detail.threads.toReversed() : detail.threads),
    [detail.threads, oldestFirst],
  );
  const total = ids.length;
  const list = useRef<HTMLDivElement>(null);
  return (
    <section aria-label={what} className="browse-discussion">
      <div className="tag">
        <button aria-expanded={open} aria-label={'Toggle ' + what} onClick={() => setOpen(!open)}>
          {open ? '▾' : '▸'}
        </button>
        <span>{what}</span>
        <button onClick={() => setOldestFirst(!oldestFirst)} title="Order of the threads">
          {oldestFirst ? 'oldest activity' : 'newest activity'} ▾
        </button>
        {fresh.size > 0 && (
          <output>
            {fresh.size} new ·{' '}
            <button
              onClick={() => {
                list.current?.querySelector('[data-new="true"]')?.scrollIntoView({ block: 'center' });
                acknowledge();
              }}
            >
              show
            </button>
          </output>
        )}
        <WritesButton className="push-right" label="new comment" />
        <WritesButton label="preview" />
        <WritesButton label="send" />
      </div>
      {open && (
        <div className="section-body browse-thread-list" ref={list}>
          <textarea
            className="browse-composer"
            aria-label="New comment (writes are not connected)"
            placeholder="write a new comment... (writes not connected)"
            disabled
            rows={3}
          />
          <WritesNote />
          {!threads.length && <p className="muted">No {what} yet.</p>}
          {threads.map((t) => (
            <Thread key={t.id} thread={t} fresh={fresh} now={now} />
          ))}
          {detail.older && (
            <p>
              <button onClick={loadOlder} disabled={older}>
                {older ? 'loading' : '▸ show older ' + what}
              </button>
            </p>
          )}
          {total > 0 && <p className="muted browse-count">{total} loaded</p>}
        </div>
      )}
    </section>
  );
}

function Thread({ thread, fresh, now }: { thread: BrowseThread; fresh: Set<string>; now: number }) {
  const [shut, setShut] = useState(false);
  const first = thread.comments[0];
  if (!first) return null;
  const rest = thread.comments.slice(1);
  const unread = thread.comments.filter((c) => fresh.has(c.id)).length;
  return (
    <article className="browse-thread" data-kind={thread.kind} data-resolved={thread.resolved ? 'true' : undefined}>
      {thread.path && (
        <p className="mono browse-thread-path">
          {thread.path}
          {thread.line ? ':' + thread.line : ''}
          {thread.resolved && <span className="muted"> · resolved</span>}
          {thread.outdated && <span className="muted"> · outdated</span>}
        </p>
      )}
      <Message comment={first} fresh={fresh.has(first.id)} now={now} head>
        {rest.length > 0 && (
          <button className="browse-fold" aria-expanded={!shut} onClick={() => setShut(!shut)}>
            {shut ? `▸ show ${rest.length} replies${unread ? `, ${unread} new` : ''}` : '▾'}
          </button>
        )}
      </Message>
      {!shut && (
        <div className="browse-replies">
          {rest.map((c) => (
            <Message key={c.id} comment={c} fresh={fresh.has(c.id)} now={now} />
          ))}
        </div>
      )}
      {thread.hidden ? (
        <p className="muted browse-hidden">
          {thread.hidden} more {thread.hidden === 1 ? 'reply' : 'replies'} not loaded.{' '}
          {thread.url && (
            <a href={thread.url} target="_blank" rel="noopener noreferrer">
              read on the provider ↗
            </a>
          )}
        </p>
      ) : null}
    </article>
  );
}

function Message({
  comment,
  fresh,
  now,
  head,
  children,
}: {
  comment: BrowseComment;
  fresh: boolean;
  now: number;
  head?: boolean;
  children?: React.ReactNode;
}) {
  return (
    <div
      className={'browse-message' + (head ? ' head' : '')}
      data-new={fresh ? 'true' : undefined}
      data-bot={comment.bot ? 'true' : undefined}
    >
      <p className="browse-by muted">
        {children}
        <span title={comment.at}>
          <b>{comment.author}</b>
          {comment.bot ? ' (bot)' : ''}
          {comment.reply_to_author && ` · replying to ${comment.reply_to_author}`}
          {comment.state &&
            comment.state !== 'commented' &&
            ` · ${comment.state === 'changes' ? 'requested changes' : comment.state}`}
          {' · '}
          {age(comment.at, now)}
          {comment.edited_at && ' · edited'}
          {fresh && <strong className="browse-new"> · new</strong>}
        </span>
      </p>
      <Markdown text={comment.body} />
      {comment.truncated && <p className="muted">This message is cut here.</p>}
      <p className="browse-actions muted">
        <button className="writes-off" aria-disabled="true" onClick={(e) => e.preventDefault()}>
          reply (disabled)
        </button>
        {comment.url?.startsWith('https://') && (
          <>
            {' · '}
            <a href={comment.url} target="_blank" rel="noopener noreferrer">
              link
            </a>
          </>
        )}
      </p>
    </div>
  );
}
