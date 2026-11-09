import { Provider, useAtom, useAtomValue } from 'jotai';
import { useEffect } from 'react';
import { Chat } from './screens/Chat.tsx';
import { Home } from './screens/Home.tsx';
import { Learn } from './screens/Learn.tsx';
import { System } from './screens/System.tsx';
import { Trace } from './screens/TraceScreen.tsx';
import { LiveOwner, useTabLifecycle } from './components/LiveOwner.tsx';
import { useDeskRoute, viewLabel, views } from './lib/route.ts';
import { goSection, hideChat, navigate, openTrace, showChat, showSection, toggleSince } from './state/actions.ts';
import { connectedAtom, errorAtom, systemAtom, attention } from './state/live.ts';
import { store } from './state/store.ts';
import { chatHiddenAtom, chatOpenAtom, keyboardAtom, phoneTabAtom } from './state/ui.ts';

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
          {route.view === 'trace' && <Trace />}
          {route.view === 'improvements' && <Learn />}
          {route.view === 'system' && <System />}
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
  return (
    <header className="tag top">
      <b className="mono">desk</b>
      <nav aria-label="Main">
        {views.map((page) => (
          <button
            key={page}
            className={route.view === page ? 'chosen' : undefined}
            onClick={() => (page === 'home' ? goSection('') : navigate(page))}
          >
            {viewLabel[page]}
          </button>
        ))}
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
