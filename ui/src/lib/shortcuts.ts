// Shared keyboard shortcuts (JOB-150), on @tanstack/react-hotkeys.
//
// Scope: a shortcut is registered by the component that owns the screen or
// overlay and is `enabled` only while it is active, so a route change or a
// closing overlay unregisters it by unmounting and no registration repeats.
// Text entry: the library skips single-key and Shift/Alt shortcuts while a
// text input, textarea, select or contenteditable has focus, and skips IME
// composition. Ctrl and Meta chords are left to the browser and OS unless a
// caller registers one on purpose. Base UI owns Escape and focus inside its
// dialogs; nothing here registers Escape globally.
import { useFormatForDisplay, useHotkey, type UseHotkeyDefinition, useHotkeys } from '@tanstack/react-hotkeys';
import type { RegisterableHotkey } from '@tanstack/hotkeys';

export interface ShortcutOptions {
  enabled?: boolean;
}

export function useShortcut(hotkey: RegisterableHotkey, run: () => void, options: ShortcutOptions = {}) {
  useHotkey(hotkey, () => run(), {
    enabled: options.enabled ?? true,
    ignoreInputs: 'auto',
    preventDefault: true,
    stopPropagation: false,
    conflictBehavior: 'warn',
  });
}

export function useShortcuts(defs: UseHotkeyDefinition[], enabled = true) {
  useHotkeys(defs, { enabled, ignoreInputs: 'auto', preventDefault: true, stopPropagation: false });
}

// shortcutLabel formats a hotkey for the platform ("Ctrl+K" or "⌘K").
export function useShortcutLabel() {
  return useFormatForDisplay();
}
