import { createStore } from 'jotai';

// One store for the tab. It is created once, above the router, so navigation
// never recreates it or drops live data (JOB-150).
export const store = createStore();
