// Markdown and prompt history for the chat plugin, after jcode-web's
// format.ts and prompt-history.ts (#535). md lives in ui/kit, shared
// with the attention plugin (JOB-101).
export { md } from '../../kit/markdown.js';

/** @param {number} n */
export function tokens(n) {
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M';
  if (n >= 1e3) return Math.round(n / 1e3) + 'k';
  return String(n);
}

const storeKey = (/** @type {string} */ id) => 'desk-chat-prompts.' + id;

/**
 * Sent prompts newest first: the transcript's, then ones kept locally
 * that the transcript has not caught up with.
 * @param {string} id @param {string[]} transcript
 */
export function promptHistory(id, transcript) {
  /** @type {string[]} */
  let saved = [];
  try {
    const v = JSON.parse(localStorage.getItem(storeKey(id)) ?? '[]');
    if (Array.isArray(v)) saved = v.filter((x) => typeof x === 'string');
  } catch { /* storage off or corrupt */ }
  return [...new Set([...transcript, ...saved.filter((p) => !transcript.includes(p))].filter(Boolean).reverse())].slice(0, 100);
}

/** @param {string} id @param {string} text */
export function rememberPrompt(id, text) {
  try {
    const h = promptHistory(id, []).filter((p) => p !== text);
    localStorage.setItem(storeKey(id), JSON.stringify([...h.reverse(), text].slice(-100)));
  } catch { /* sending must not depend on storage */ }
}
