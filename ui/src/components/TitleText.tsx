import { parseTitle } from '../../kit/titles.js';

// A title's text with its Markdown links as real anchors (JOB-135). No
// wrapper element and no row behavior: headings and graph cards use this.
export function TitleText({ title, fallback = '' }: { title?: string; fallback?: string }) {
  const parsed = parseTitle(title || fallback);
  return (
    <>
      {parsed.parts.map((part, i) =>
        part.link ? (
          <a
            key={i}
            href={part.link.href}
            target={part.link.external ? '_blank' : undefined}
            rel={part.link.external ? 'noopener noreferrer' : undefined}
          >
            {part.text}
          </a>
        ) : (
          part.text
        ),
      )}
    </>
  );
}
