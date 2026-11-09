import { Menu } from '@base-ui/react/menu';
import { useAtomValue } from 'jotai';
import { useEffect, useRef, useState } from 'react';
import { copyText } from '../lib/clipboard.ts';
import { useDeskRoute } from '../lib/route.ts';
import { attach, discuss, inspect, openTrace } from '../state/actions.ts';
import { errorAtom, selectedAtom, traceAtom } from '../state/live.ts';
import { store } from '../state/store.ts';
import {
  closeTicketMenu,
  noticeAtom,
  openTicketMenu,
  ticketMenuAtom,
  type TicketMenuState,
  type TicketMenuTarget,
} from '../state/ticketMenu.ts';
import { watchedSessionAtom } from '../state/ui.ts';

// The ticket context menu (JOB-136). Rows open it with onTicketContextMenu
// (right click, the context-menu key, Shift+F10, and long press where the
// browser sends contextmenu) or with a TicketMenuButton, the visible way in
// on touch screens and for keyboard users. One menu is mounted for the page.

// Cases where the browser's own menu is the right one: Shift+right click, a
// real link (open, copy address), text entry, and text the reader selected.
function nativeMenuWanted(event: MouseEvent) {
  if (event.shiftKey) return true;
  const el = event.target instanceof Element ? event.target : null;
  if (!el) return true;
  if (el.closest('a[href], input, textarea, select, [contenteditable]:not([contenteditable="false"])')) return true;
  const selection = window.getSelection();
  return (
    !!selection && !selection.isCollapsed && selection.toString().trim() !== '' && selection.containsNode(el, true)
  );
}

// A ref for a row: ref={ticketMenuRef(target)}. The listener is native, not
// onContextMenu: rows are plain article, tr and li elements, and the menu is
// an extra route to actions their own controls already offer. The innermost
// row handles the event first and marks it, so a nested row opens one menu.
export function ticketMenuRef(target: TicketMenuTarget) {
  return (row: HTMLElement | null) => {
    if (!row) return;
    const listener = (event: MouseEvent) => showMenu(event, row, target);
    row.addEventListener('contextmenu', listener);
    return () => row.removeEventListener('contextmenu', listener);
  };
}

function showMenu(event: MouseEvent, row: HTMLElement, target: TicketMenuTarget) {
  if (event.defaultPrevented || nativeMenuWanted(event)) return;
  event.preventDefault();
  // Some browsers report no position for a keyboard invocation: sit under
  // the row instead of the corner of the page.
  const placed = event.clientX !== 0 || event.clientY !== 0;
  const box = row.getBoundingClientRect();
  const x = placed ? event.clientX : box.left;
  const y = placed ? event.clientY : box.bottom;
  const active = document.activeElement;
  openTicketMenu({
    target,
    anchor: { getBoundingClientRect: () => new DOMRect(x, y, 0, 0) },
    returnTo: active instanceof HTMLElement && row.contains(active) ? active : null,
  });
}

// The visible, focusable way to open the same menu for a row.
export function TicketMenuButton({ target }: { target: TicketMenuTarget }) {
  return (
    <button
      type="button"
      className="ticket-menu-button"
      aria-haspopup="menu"
      aria-label={'Actions for ' + target.id}
      title="actions"
      onClick={(event) => {
        event.stopPropagation();
        openTicketMenu({ target, anchor: event.currentTarget, returnTo: event.currentTarget });
      }}
    >
      ⋯
    </button>
  );
}

// "copied <id>" for a few seconds after a copy.
export function Notice() {
  const notice = useAtomValue(noticeAtom);
  useEffect(() => {
    if (!notice) return;
    const timer = setTimeout(() => store.set(noticeAtom, ''), 3000);
    return () => clearTimeout(timer);
  }, [notice]);
  return <output className="notice">{notice}</output>;
}

export function TicketMenu() {
  const state = useAtomValue(ticketMenuAtom);
  return state ? <Popup state={state} /> : null;
}

function Popup({ state }: { state: TicketMenuState }) {
  const { target } = state;
  const route = useDeskRoute();
  const selected = useAtomValue(selectedAtom);
  const root = useAtomValue(traceAtom)?.graph?.root;
  const [failed, setFailed] = useState(false);
  // On the trace page the root is already open and the inspected ticket is
  // already inspected: offering either again would do nothing.
  const onTrace = route.view === 'trace';
  const isRoot = onTrace && root === target.id;
  const isSelected = onTrace && (selected?.id ?? root) === target.id;

  // A navigation while the menu is open (Back, a live redirect) dismisses it
  // rather than leave it pointing at a row that is gone. Base UI's modal
  // menu already blocks scrolling and clicks on the page behind it.
  const place = route.view + '|' + route.ref + '|' + route.inspect;
  const opened = useRef(place);
  useEffect(() => {
    if (opened.current !== place) closeTicketMenu();
    opened.current = place;
  }, [place]);

  async function copy() {
    if (await copyText(target.id)) {
      store.set(noticeAtom, 'copied ' + target.id);
      closeTicketMenu();
    } else {
      setFailed(true);
      store.set(errorAtom, `Could not copy ${target.id}. Select the id at the bottom of the menu and copy it by hand.`);
    }
  }
  function viewSession() {
    if (!target.inspectable) openTrace(target.id);
    store.set(watchedSessionAtom, target.session || '');
  }

  return (
    <Menu.Root
      open={state.open}
      onOpenChange={(open) => {
        if (!open) closeTicketMenu();
      }}
    >
      <Menu.Portal>
        <Menu.Positioner
          anchor={state.anchor}
          side="bottom"
          align="start"
          sideOffset={2}
          collisionPadding={8}
          className="popup-positioner"
        >
          <Menu.Popup
            className="ticket-menu"
            aria-label={'Actions for ' + target.id}
            // Escape and keyboard selection return to the row's control; a
            // click elsewhere keeps focus where the click put it.
            finalFocus={(closeType) =>
              closeType === 'keyboard' ? (state.returnTo?.isConnected ? state.returnTo : true) : false
            }
          >
            {!isRoot && (
              <Menu.Item className="ticket-menu-item" onClick={() => openTrace(target.id)}>
                {target.inspectable ? 'open as graph root' : 'open in trace'}
              </Menu.Item>
            )}
            {target.inspectable && !isSelected && (
              <Menu.Item className="ticket-menu-item" onClick={() => inspect(target.id)}>
                inspect
              </Menu.Item>
            )}
            {target.session && (
              <Menu.Item className="ticket-menu-item" onClick={viewSession}>
                view session
              </Menu.Item>
            )}
            {target.item && (
              <Menu.Item className="ticket-menu-item" onClick={() => void discuss(target.item)}>
                discuss
              </Menu.Item>
            )}
            <Menu.Item className="ticket-menu-item" onClick={() => attach([target.ref || target.id])}>
              attach to chat
            </Menu.Item>
            <Menu.Item className="ticket-menu-item" closeOnClick={false} onClick={() => void copy()}>
              copy ticket id
            </Menu.Item>
            <div className="ticket-menu-id">
              <span className="muted">ticket id</span>
              <code>{target.id}</code>
              {failed && (
                <span role="alert" className="warn">
                  could not copy
                </span>
              )}
            </div>
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  );
}
