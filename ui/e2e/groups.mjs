// The specs `run-isolated.mjs` runs, and how it groups them (JOB-213).
// `affected.mjs` selects from `isolated`, so a spec that is not listed here
// is never picked by the preflight.
//
// Not listed, because they fail on this controller today and need their own
// setup: chat-dismiss, chat-responsiveness, home-sections, run-links and
// trace-session. attachments and chat need a real agent. Add one here when its
// setup is part of `run-isolated.mjs`.

// Run together on one controller.
export const shared = [
  'ticket-menu.spec.js',
  'message-headers.spec.js',
  'embeds.spec.js',
  'timeline.spec.js',
  'browse.spec.js',
  'browse-live.spec.js',
  'turn-phase.spec.js',
];

// The trace, search and layout specs. `run-isolated.mjs trace` runs them on a
// controller of their own, as a second CI step: on one controller with the
// others, pool-controls and turn-phase's mobile cases timed out on RPCs.
export const trace = [
  'desk.spec.js',
  'live.spec.js',
  'search.spec.js',
  'shell.spec.js',
  'title-links.spec.js',
  'trace-axis.spec.js',
  'trace-detail.spec.js',
  'trace-whole.spec.js',
];

// Scales the shared worker pool: one worker, after the shared group.
export const poolControls = ['pool-controls.spec.js'];

// Rewrites the controller's config and reloads it: alone, last.
export const reload = ['embed-reload.spec.js'];

// What `run-isolated.mjs` runs with no arguments.
export const defaults = [...shared, ...poolControls, ...reload];

export const isolated = [...shared, ...trace, ...poolControls, ...reload];
