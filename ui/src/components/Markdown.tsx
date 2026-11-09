import { memo, useMemo } from 'react';
import { md } from '../../kit/markdown.js';

// Sanitized Markdown (the shared kit renderer). The HTML is DOMPurify output
// with GitHub reference linking, the same as before the port. With a
// className it renders a div of that class, otherwise a span for inline text
// or a div.markdown for a block.
export const Markdown = memo(function Markdown({
  text,
  inline = false,
  className,
}: {
  text: string;
  inline?: boolean;
  className?: string;
}) {
  const html = useMemo(() => md(text, inline), [text, inline]);
  if (className) return <div className={className} dangerouslySetInnerHTML={{ __html: html }} />;
  return inline ? (
    <span dangerouslySetInnerHTML={{ __html: html }} />
  ) : (
    <div className="markdown" dangerouslySetInnerHTML={{ __html: html }} />
  );
});
