import { useCallback, useSyncExternalStore } from 'react';

// Whether a media query matches now, kept current as the window changes.
export function useMedia(query: string): boolean {
  const subscribe = useCallback(
    (notify: () => void) => {
      const list = matchMedia(query);
      list.addEventListener('change', notify);
      return () => list.removeEventListener('change', notify);
    },
    [query],
  );
  return useSyncExternalStore(
    subscribe,
    () => matchMedia(query).matches,
    () => false,
  );
}
