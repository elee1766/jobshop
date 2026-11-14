// Which Desk e2e specs does a change touch? (JOB-213)
//
//   npm run e2e:affected                    run them on an isolated controller
//   node e2e/affected.mjs                   list them (one per line)
//   node e2e/affected.mjs --base origin/master --files ui/src/screens/Chat.tsx
//
// The files are the branch's diff against the merge base with `--base`
// (default origin/master), plus uncommitted and untracked files, or the ones
// after `--files` (repo-relative paths). Running needs JOBSHOP_BIN and
// FAKE_AGENT_BIN, see the header of run-isolated.mjs.
//
// Desk PRs broke specs their authors did not run (message headers, turn
// phase, graph cards). The table below says which specs cover which source.
// A ui/ source file with no rule runs every spec: add a rule rather than
// rely on that. affected.test.mjs fails for a source file nothing maps and for
// a spec the table names that run-isolated.mjs does not run.
import { spawnSync } from 'node:child_process';
import { dirname, resolve } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { isolated } from './groups.mjs';

const all = 'all';

// Every spec: shared code, styling and the build, so anything can change.
const everything = [
  /^ui\/src\/(app|index|tokens)\.css$/,
  /^ui\/src\/(App|main)\.tsx$/,
  /^ui\/src\/state\/(live|store|entities|ui|actions)\.ts$/,
  /^ui\/src\/lib\/(api|model|route|reconcile|useMedia|useSize|clipboard)\.ts$/,
  /^ui\/src\/components\/(LiveOwner|StatusLetter)\.tsx$/,
  /^ui\/(package\.json|package-lock\.json|vite\.config\.ts|tsconfig\.json|index\.html|playwright\.config\.js)$/,
  /^ui\/public\//,
  /^ui\/e2e\/(groups|run-isolated|embed-fixture|prefix-proxy|trace-view)\.m?js$/,
  /^ui\/e2e\/fixtures\//,
  // The API Desk talks to and the client for it.
  /^packages\/jobshop-client\//,
  /^pkg\/transport\/desk\/(?!dist\/)/,
];

const trace = [
  'timeline.spec.js',
  'trace-axis.spec.js',
  'trace-detail.spec.js',
  'trace-whole.spec.js',
  'title-links.spec.js',
  'search.spec.js',
  'live.spec.js',
];
const chat = ['message-headers.spec.js', 'turn-phase.spec.js', 'desk.spec.js', 'live.spec.js'];
const browse = ['browse.spec.js', 'browse-live.spec.js'];
const embed = ['embeds.spec.js', 'embed-reload.spec.js'];
const titles = ['title-links.spec.js', 'ticket-menu.spec.js', 'browse.spec.js', 'timeline.spec.js'];

// First match wins. Paths are repo-relative.
export const rules = [
  [/^ui\/src\/screens\/(Timeline|TimelineMarker|TraceScreen|TraceSearch|TicketDetails|TicketEvents|TicketHistory)\.tsx$/, [...trace, 'message-headers.spec.js', 'turn-phase.spec.js']],
  [/^ui\/src\/lib\/(trace|timeline|history|detail|search)\.ts$/, [...trace, 'message-headers.spec.js', 'turn-phase.spec.js']],
  [/^ui\/src\/screens\/(Chat|Transcript|SessionViewer)\.tsx$/, [...chat, 'timeline.spec.js']],
  [/^ui\/src\/lib\/(session|useSessionFrames|activity|attachments|push)\.ts$/, [...chat, 'timeline.spec.js']],
  [/^ui\/src\/state\/chatBridge\.ts$/, chat],
  [/^ui\/src\/screens\/Home\.tsx$/, ['desk.spec.js', 'shell.spec.js', 'ticket-menu.spec.js', 'title-links.spec.js', 'live.spec.js']],
  [/^ui\/src\/lib\/shortcuts\.ts$/, ['shell.spec.js', 'desk.spec.js']],
  [/^ui\/src\/(screens\/(BrowseDetail|Browse)|components\/BrowseParts|state\/browse|lib\/browse)\.tsx?$/, browse],
  [/^ui\/src\/screens\/Embed\.tsx$/, embed],
  [/^ui\/src\/components\/PoolControls\.tsx$/, ['pool-controls.spec.js']],
  [/^ui\/src\/screens\/System\.tsx$/, ['pool-controls.spec.js', 'embed-reload.spec.js', 'shell.spec.js', 'desk.spec.js']],
  [/^ui\/src\/(components\/TicketMenu\.tsx|state\/ticketMenu\.ts)$/, ['ticket-menu.spec.js', 'browse.spec.js', 'shell.spec.js']],
  [/^ui\/(src\/components\/(TitleAction|TitleText|Markdown)\.tsx|kit\/(titles|markdown|links)\.js)$/, [...titles, ...chat]],
  [/^ui\/src\/screens\/(Learn|HookAudit)\.tsx$/, ['desk.spec.js', 'shell.spec.js']],
  [/^ui\/src\/lib\/learn\.ts$/, ['desk.spec.js', 'shell.spec.js']],
  // Unit tests and type-only files run no browser.
  [/^ui\/(src|kit)\/.*\.test\.(ts|js)$/, []],
  [/^ui\/e2e\/(?<spec>[^/]+\.spec\.js)$/, 'self'],
  // Scripts that only produce screenshots.
  [/^ui\/e2e\/[^/]*(shots|affected)[^/]*\.mjs$/, []],
  [/^ui\/perf\//, []],
];

// Specs `file` touches, in groups.isolated order. Not a Desk file: none.
export function specsFor(file) {
  if (everything.some((re) => re.test(file))) return [...isolated];
  for (const [re, specs] of rules) {
    const m = re.exec(file);
    if (!m) continue;
    if (specs === 'self') return isolated.filter((s) => s === m.groups.spec);
    return isolated.filter((s) => specs.includes(s));
  }
  // An unmapped ui source file: be safe, not quiet.
  if (/^ui\/(src|kit|e2e)\//.test(file)) return [...isolated];
  return [];
}

export function affected(files) {
  const set = new Set(files.flatMap(specsFor));
  return isolated.filter((s) => set.has(s));
}

// Whether the table says anything for `file`: false means it fell through to
// the run-everything default.
export function mapped(file) {
  return everything.some((re) => re.test(file)) || rules.some(([re]) => re.test(file));
}

function git(root, ...args) {
  const r = spawnSync('git', args, { cwd: root, encoding: 'utf8' });
  if (r.status !== 0) throw new Error(`git ${args.join(' ')}: ${r.stderr.trim()}`);
  return r.stdout.split('\n').filter(Boolean);
}

export function changedFiles(base) {
  const here = dirname(fileURLToPath(import.meta.url));
  const root = git(here, 'rev-parse', '--show-toplevel')[0];
  const [mergeBase] = git(root, 'merge-base', base, 'HEAD');
  return [...new Set([...git(root, 'diff', '--name-only', mergeBase), ...git(root, 'ls-files', '--others', '--exclude-standard')])];
}

function main(argv) {
  let base = 'origin/master';
  let files = null;
  let run = false;
  for (let i = 0; i < argv.length; i++) {
    if (argv[i] === '--base') base = argv[++i];
    else if (argv[i] === '--files') files = [];
    else if (argv[i] === '--run') run = true;
    else if (files) files.push(argv[i]);
    else {
      console.error('usage: affected.mjs [--run] [--base REF] [--files path...]');
      return 2;
    }
  }
  const specs = affected(files ?? changedFiles(base));
  if (!run) {
    for (const spec of specs) console.log(spec);
    return 0;
  }
  if (specs.length === 0) {
    console.log('no Desk e2e spec is affected by this change');
    return 0;
  }
  console.log(`running ${specs.length} spec(s): ${specs.join(' ')}`);
  const here = dirname(fileURLToPath(import.meta.url));
  return spawnSync('node', [resolve(here, 'run-isolated.mjs'), ...specs], { stdio: 'inherit' }).status ?? 1;
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  process.exit(main(process.argv.slice(2)));
}
