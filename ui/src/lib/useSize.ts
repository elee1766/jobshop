import { useCallback, useLayoutEffect, useState } from 'react';

// The border-box-free client size of an element, kept current by a
// ResizeObserver. Returns a callback ref and the size. A hidden element
// (display: none reads 0 by 0) keeps its last size, so a chart that is
// swapped out and back does not reflow while it is away.
export function useSize<T extends HTMLElement>(): [(node: T | null) => void, { width: number; height: number }] {
  const [node, setNode] = useState<T | null>(null);
  const [size, setSize] = useState({ width: 0, height: 0 });
  const ref = useCallback((el: T | null) => setNode(el), []);
  useLayoutEffect(() => {
    if (!node) return;
    const read = () => {
      const next = { width: node.clientWidth, height: node.clientHeight };
      if (next.width === 0 && next.height === 0) return;
      setSize((prev) => (prev.width === next.width && prev.height === next.height ? prev : next));
    };
    read();
    const observer = new ResizeObserver(read);
    observer.observe(node);
    return () => observer.disconnect();
  }, [node]);
  return [ref, size];
}
