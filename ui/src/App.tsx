import { Provider, useAtom, useAtomValue } from 'jotai';
import { lazy, Suspense, useEffect, type ComponentType } from 'react';
import { Chat } from './screens/Chat.tsx';
import { Embed } from './screens/Embed.tsx';
import { Home } from './screens/Home.tsx';
import { LiveOwner, useTabLifecycle } from './components/LiveOwner.tsx';
import { useDeskRoute, viewLabel, views } from './lib/route.ts';
import { goSection, hideChat, navigate, openTrace, showChat, showSection, toggleSince } from './state/actions.ts';
import { connectedAtom, embedsAtom, errorAtom, systemAtom, attention } from './state/live.ts';
import { store } from './state/store.ts';
import { chatHiddenAtom, chatOpenAtom, keyboardAtom, phoneTabAtom } from './state/ui.ts';

// Home and chat are the first screens and ship in the entry chunk. The graph,
// learn and system screens are separate chunks, which keeps the first load
// near the framework floor (JOB-150 budgets). They are fetched when the
// browser is idle after the first render, so a first switch finds them cached.
//
// A lazy() component that already loaded still suspends once on its first
// render, and React holds a revealed Suspense boundary for 300 ms. So a loaded
// screen is kept here and rendered directly.
function screen(load: () => Promise<ComponentType>) {
  let loaded: ComponentType | undefined;
  let promise: Promise<ComponentType> | undefined;
  const start = () =>
    (promise ??= load().then((c) => {
      loaded = c;
      return c;
    }));
  const Lazy = lazy(() => start().then((c) => ({ default: c })));
  return {
    start,
    Screen: () => {
      const Loaded = loaded;
      return Loaded ? <Loaded /> : <Lazy />;
    },
  };
}

const traceScreen = screen(() => import('./screens/TraceScreen.tsx').then((m) => m.Trace));
const learnScreen = screen(() => import('./screens/Learn.tsx').then((m) => m.Learn));
const systemScreen = screen(() => import('./screens/System.tsx').then((m) => m.System));
const { Screen: Trace } = traceScreen;
const { Screen: Learn } = learnScreen;
const { Screen: System } = systemScreen;

function usePreloadScreens() {
  useEffect(() => {
    const load = () => [traceScreen, learnScreen, systemScreen].forEach((s) => void s.start());
    if ('requestIdleCallback' in window) {
      const id = requestIdleCallback(load, { timeout: 1500 });
      return () => cancelIdleCallback(id);
    }
    const id = setTimeout(load, 300);
    return () => clearTimeout(id);
  }, []);
}

export function App() {
  return (
    <Provider store={store}>
      <LiveOwner />
      <Desk />
    </Provider>
  );
}

function Desk() {
  useTabLifecycle();
  usePreloadScreens();
  const route = useDeskRoute();
  const [phoneTab, setPhoneTab] = useAtom(phoneTabAtom);
  const hidden = useAtomValue(chatHiddenAtom);
  const chatOpen = useAtomValue(chatOpenAtom);
  const keyboard = useAtomValue(keyboardAtom);
  const [error, setError] = useAtom(errorAtom);

  // The URL decides the phone tab on a first load, back and forward.
  useEffect(() => {
    if (route.view === 'home') {
      setPhoneTab((tab) => (tab === 'chats' ? tab : route.section || 'waiting'));
    } else setPhoneTab('more');
  }, [route.view, route.section, setPhoneTab]);

  // A hash on a fresh load or Back/Forward scrolls to its section.
  useEffect(() => {
    if (route.view === 'home') showSection(route.section);
  }, [route.view, route.section]);

  // A link to a ticket trace opens it in place, with no page load.
  useEffect(() => {
    const click = (event: MouseEvent) => {
      if (
        event.defaultPrevented ||
        event.button !== 0 ||
        event.metaKey ||
        event.ctrlKey ||
        event.shiftKey ||
        event.altKey
      )
        return;
      const anchor = event.target instanceof Element ? event.target.closest('a') : null;
      if (!anchor || anchor.target || anchor.hasAttribute('download')) return;
      const url = new URL(anchor.href);
      if (
        url.origin !== location.origin ||
        url.pathname !== location.pathname ||
        url.searchParams.get('view') !== 'trace'
      )
        return;
      const ref = url.searchParams.get('ref');
      if (!ref) return;
      event.preventDefault();
      openTrace(ref);
    };
    window.addEventListener('click', click);
    return () => window.removeEventListener('click', click);
  }, []);

  return (
    <div
      className={['desk', keyboard && 'keyboard', hidden && 'chat-hidden', chatOpen && 'chat-open']
        .filter(Boolean)
        .join(' ')}
      data-phone={phoneTab}
    >
      <Header />
      {error && (
        <div className="error banner" role="alert">
          {error}
          <button onClick={() => setError('')}>dismiss</button>
        </div>
      )}
      <div className="columns">
        <main className="work">
          {route.view === 'home' && <Home />}
          <Suspense fallback={null}>
            {route.view === 'trace' && <Trace />}
            {route.view === 'improvements' && <Learn />}
            {route.view === 'system' && <System />}
            {route.view === 'embed' && <Embed />}
          </Suspense>
        </main>
        <aside className="chat-column" aria-label="Chats">
          <button className="show-chats" onClick={showChat}>
            chats ▸
          </button>
          <div className="chat-content">
            <Chat onHide={hideChat} />
          </div>
        </aside>
      </div>
      <PhoneNav />
    </div>
  );
}

function Header() {
  const route = useDeskRoute();
  const connected = useAtomValue(connectedAtom);
  const system = useAtomValue(systemAtom);
  const embeds = useAtomValue(embedsAtom);
  return (
    <header className="tag top">
      <b className="mono">desk</b>
      <nav aria-label="Main" className={embeds && embeds.length > 0 ? 'has-embeds' : undefined}>
        {views.map((page) => (
          <button
            key={page}
            className={route.view === page ? 'chosen' : undefined}
            onClick={() => (page === 'home' ? goSection('') : navigate(page))}
          >
            {viewLabel[page]}
          </button>
        ))}
        {embeds && embeds.length > 0 && (
          // The configured tabs (JOB-151) follow the built-in views in config order.
          <span className="embeds">
            {embeds.map((embed) => {
              const chosen = route.view === 'embed' && route.embed === embed.id;
              return (
                <button
                  key={embed.id}
                  className={chosen ? 'chosen' : undefined}
                  aria-current={chosen ? 'page' : undefined}
                  onClick={() => navigate('embed', '', { embed: embed.id })}
                >
                  {embed.label}
                </button>
              );
            })}
          </span>
        )}
        <span className="divider">|</span>
        <button className={route.since ? 'chosen' : undefined} onClick={toggleSince}>
          catch-up
        </button>
      </nav>
      <span className="push-right status mono">
        {connected ? 'sys ok' : 'reconnecting'}
        {system && <> · tick {system.status?.tick_seconds?.toFixed(1) || '0'}s</>}
      </span>
    </header>
  );
}

function PhoneNav() {
  const phoneTab = useAtomValue(phoneTabAtom);
  const [, setPhoneTab] = useAtom(phoneTabAtom);
  const waiting = useAtomValue(attention.ids).length;
  return (
    <nav className="phone-nav" aria-label="Mobile">
      {['waiting', 'chats', 'inflight', 'more'].map((tab) => (
        <button
          key={tab}
          className={phoneTab === tab ? 'chosen' : undefined}
          onClick={() => {
            if (tab === 'chats') {
              setPhoneTab(tab);
              showChat();
            } else if (tab === 'more') setPhoneTab(tab);
            else goSection(tab as 'waiting' | 'inflight');
          }}
        >
          {tab === 'inflight' ? 'in flight' : tab}
          {tab === 'waiting' ? ' ' + waiting : ''}
        </button>
      ))}
    </nav>
  );
}
