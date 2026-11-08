// Compact labels for bare GitHub PR and Linear issue URLs (JOB-125).
import { Marked } from 'marked';

const githubPull = /^\/([A-Za-z0-9][A-Za-z0-9-]*)\/([A-Za-z0-9_.-]+)\/pull\/([1-9][0-9]*)\/?$/;
const linearIssue = /^\/([A-Za-z0-9][A-Za-z0-9_-]*)\/issue\/([A-Za-z][A-Za-z0-9]*-[1-9][0-9]*)(?:\/[^/]+)?\/?$/;

/**
 * Short label for a GitHub pull request or Linear issue URL, or '' when the
 * URL is anything else. Only the real hosts over https count, so lookalike
 * hosts, ports and credentials keep the full URL as their label.
 * @param {string} href
 */
export function compactLabel(href) {
  let url;
  try { url = new URL(href); } catch { return ''; }
  if (url.protocol !== 'https:' || url.port || url.username || url.password) return '';
  if (url.hostname === 'github.com') {
    const m = githubPull.exec(url.pathname);
    return m ? `PR:${m[1]}/${m[2]}#${m[3]}` : '';
  }
  if (url.hostname === 'linear.app') {
    const m = linearIssue.exec(url.pathname);
    return m ? `${m[1]}@${m[2]}` : '';
  }
  return '';
}

/**
 * Relabel autolinks (bare URLs and <url>) only. Authored [label](url) links
 * keep their label, and code spans and blocks never produce link tokens.
 * The href is left as written, query and fragment included.
 * @param {import('marked').Token} token
 */
function compactAutolink(token) {
  if (token.type !== 'link' || !token.autolink) return;
  const label = compactLabel(token.href);
  if (!label) return;
  token.text = label;
  token.tokens = [{ type: 'text', raw: label, text: label }];
}

export const marked = new Marked({ async: false, gfm: true, breaks: true, walkTokens: compactAutolink });
