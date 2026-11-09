import type { MouseEvent, ReactNode } from 'react';
import { parseTitle } from '../../kit/titles.js';
import { TitleText } from './TitleText.tsx';

// A ticket or run title that is also the row's action (JOB-135).
//
// Plain title: the same button or link as before. Title with Markdown
// links: a link cannot sit inside a button or another link, so the action
// becomes an empty overlay (one tab stop, named by the plain text) behind
// the title text, and each title link is its own control above it.
export function TitleAction({
  title,
  fallback = '',
  className = '',
  tip,
  href,
  onActivate,
  children,
}: {
  title?: string;
  fallback?: string;
  className?: string;
  tip?: string;
  // Set for a link action, omitted for a button action.
  href?: string;
  onActivate?: (event: MouseEvent) => void;
  // Trailing content inside the action: ids, ages.
  children?: ReactNode;
}) {
  const parsed = parseTitle(title || fallback);
  if (!parsed.links.length) {
    return href !== undefined ? (
      <a className={className} title={tip} href={href}>
        <span>{parsed.text}</span>
        {children}
      </a>
    ) : (
      <button className={className} title={tip} onClick={onActivate}>
        <span>{parsed.text}</span>
        {children}
      </button>
    );
  }
  return (
    <span className={`${className} hosted`} title={tip}>
      {href !== undefined ? (
        <a className="hit" href={href} aria-label={parsed.text}>
          {' '}
        </a>
      ) : (
        <button className="hit" aria-label={parsed.text} onClick={onActivate} />
      )}
      <span>
        <TitleText title={title} fallback={fallback} />
      </span>
      {children}
    </span>
  );
}
