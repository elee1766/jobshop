// Session chat panel layout: width on desktop (persisted) and whether the
// viewport is the phone layout, where the panel is full screen.

const FLOOR = 300; // narrow windows squeeze the page rather than the chat
const MAX = 760;
const SIDEBAR = 212;
const MAIN_MIN = 360;
const DEFAULT = 440;

/** @param {number} w */
function clamp(w) {
  const room = window.innerWidth - SIDEBAR - MAIN_MIN;
  return Math.round(Math.max(FLOOR, Math.min(MAX, room, w)));
}

const saved = Number(localStorage.getItem('desk-panel-w'));
const mq = window.matchMedia('(max-width: 768px)');

export const panel = $state({ w: clamp(saved > 0 ? saved : DEFAULT), mobile: mq.matches });

mq.addEventListener('change', (e) => { panel.mobile = e.matches; });
window.addEventListener('resize', () => { panel.w = clamp(panel.w); });

/** @param {number} w */
export function setWidth(w) {
  panel.w = clamp(w);
  localStorage.setItem('desk-panel-w', String(panel.w));
}

export function resetWidth() {
  setWidth(DEFAULT);
}
