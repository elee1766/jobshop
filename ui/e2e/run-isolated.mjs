// Runs the Desk browser checks that need an isolated controller (JOB-162):
// the ticket context menu, sender headers and configured iframe tabs.
//
//   JOBSHOP_BIN=/path/to/jobshop FAKE_AGENT_BIN=/path/to/control.test \
//     node e2e/run-isolated.mjs
//
// It makes a throwaway root under the temp dir, starts a controller, Desk and
// the embed fixture on loopback, runs Playwright and stops everything. It never
// touches the root, socket or session of the environment it runs in. Build the
// binaries from the branch under test: `go build -o jobshop ./cmd/jobshop` after
// `npm run build`, and `go test -c -o control.test ./pkg/control` for the ACP
// test agent that chat sessions run (docs/verify/desk-after-react.md).
import { spawn, spawnSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { setTimeout as delay } from 'node:timers/promises';
import { fileURLToPath } from 'node:url';

const ui = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const bin = need('JOBSHOP_BIN');
const agent = need('FAKE_AGENT_BIN');
const port = Number(process.env.DESK_PORT || 7461);
const desk = `http://127.0.0.1:${port}`;

function need(name) {
  const value = process.env[name];
  if (!value) {
    console.error(`set ${name} (see the header of e2e/run-isolated.mjs)`);
    process.exit(2);
  }
  return resolve(value);
}

// Short path: the controller's unix sockets live under the root.
const dir = mkdtempSync(join(tmpdir(), 'jse2e-'));
const root = join(dir, 'root');
const work = join(dir, 'work');
mkdirSync(work);

// The session env points at a live controller when this runs inside a jobshop
// session: drop it so nothing here can reach one.
const env = { ...process.env };
for (const key of ['JOBSHOP_SOCKET', 'JOBSHOP_SESSION', 'JOBSHOP_ROOT']) delete env[key];

const jobshop = (...args) => spawnSync(bin, ['--root', root, ...args], { env, encoding: 'utf8' });

const children = [];
function start(name, command, args, opts = {}) {
  const child = spawn(command, args, { env, cwd: ui, stdio: ['ignore', 'pipe', 'pipe'], ...opts });
  const tag = (chunk) => String(chunk).split('\n').filter(Boolean).map((l) => `[${name}] ${l}`).join('\n');
  child.stdout.on('data', (d) => console.log(tag(d)));
  child.stderr.on('data', (d) => console.log(tag(d)));
  children.push(child);
  return child;
}

async function until(what, check) {
  for (let i = 0; i < 150; i++) {
    if (await check()) return;
    await delay(200);
  }
  throw new Error(`timed out waiting for ${what}`);
}

const reachable = (url) => async () => {
  try {
    return (await fetch(url)).ok;
  } catch {
    return false;
  }
};

async function stop() {
  for (const child of children) child.kill('SIGTERM');
  for (const child of children) {
    if (child.exitCode === null && child.signalCode === null) {
      await Promise.race([new Promise((done) => child.once('exit', done)), delay(10_000)]);
      child.kill('SIGKILL');
    }
  }
  rmSync(dir, { recursive: true, force: true });
}

function playwright(args, extraEnv = {}) {
  return spawnSync('npx', ['playwright', 'test', ...args], {
    cwd: ui,
    stdio: 'inherit',
    env: { ...env, DESK_URL: desk, DESK_MUTATE: '1', DESK_LINK_WORKFLOW: 'link-acceptance', ...extraEnv },
  }).status ?? 1;
}

let code = 1;
try {
  const init = jobshop('init', root);
  if (init.status !== 0) throw new Error('jobshop init: ' + init.stderr);
  // Work definitions live in the pack; jobshop.yaml keeps only service
  // settings. The reload spec rewrites everything from `desk:` on, so keep
  // `desk:` last.
  writeFileSync(
    join(root, 'pack.yaml'),
    `default_project: js
projects:
  js: {}
agents:
  fake:
    command: ${JSON.stringify(agent)}
    args: [fake-agent]
pools:
  operator:
    project: js
    agent: fake
    type: chat
    attach: rw
    max_sessions: 200
    prompt: "Operator pool instructions"
    dir: ${JSON.stringify(work)}
workflows:
  # Stays open for a day: a row for the menu spec, with no outside effect.
  link-acceptance:
    vars: [repo, pr, issue]
    steps:
      - id: hold
        type: command
        run: 'exit 75'
        retry_after: 24h
`,
  );
  writeFileSync(
    join(root, 'jobshop.yaml'),
    `pack: pack.yaml
desk:
  embeds:
    - {id: cross, label: Cross app, url: "http://localhost:7497/app"}
    - {id: same, label: Same host, url: "http://127.0.0.1:7498/app", sandbox: [allow-same-origin]}
    - {id: denied, label: Denied XFO, url: "http://localhost:7497/deny"}
    - {id: ancestor, label: Denied CSP, url: "http://localhost:7497/ancestor"}
    - {id: extra-long-label-for-wrap, label: A rather long embed label, url: "http://localhost:7497/app"}
`,
  );

  // A 1s tick lets escalation chats and step retries move within a test.
  start('serve', bin, ['--root', root, 'serve', '--tick=1s']);
  await until('the controller', async () => jobshop('reload').status === 0);
  start('desk', bin, ['--root', root, 'desk', '--listen', `127.0.0.1:${port}`]);
  start('fixture', 'node', ['e2e/embed-fixture.mjs']);
  await until('desk', reachable(`${desk}/api/embeds`));
  await until('the embed fixture', reachable('http://127.0.0.1:7498/seen'));

  code = playwright(['ticket-menu.spec.js', 'message-headers.spec.js', 'embeds.spec.js', '--workers=2']);
  // The reload spec rewrites this root's config and reloads it, so it runs
  // alone, after the others.
  if (code === 0) {
    code = playwright(['embed-reload.spec.js', '--project=desktop', '--workers=1'], {
      EMBED_ROOT: root,
      EMBED_JOBSHOP: bin,
    });
  }
} catch (err) {
  console.error(err);
} finally {
  await stop();
}
process.exit(code);
